package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/llm"
)

// ErrToolNotFound is returned when an unknown tool name is called.
var ErrToolNotFound = fmt.Errorf("tool not found")

// ModuleConnector provides gRPC connections to module services.
// Satisfied by GatewayProxy from the http interface package via an adapter.
type ModuleConnector interface {
	GetConnection(name string) (*grpc.ClientConn, bool)
}

// ToolExecutor executes tool calls by forwarding to module services.
// For MVP: uses HTTP REST calls to module gRPC-gateway endpoints via the gateway proxy.
// This avoids needing generated gRPC stubs in core for each module.
type ToolExecutor struct {
	registry    *ToolRegistry
	connector   ModuleConnector
	moduleAddrs map[string]string // module name -> base HTTP URL
	httpClient  *http.Client
	log         *zap.Logger
}

// NewToolExecutor creates a ToolExecutor.
// moduleAddrs maps module name to its REST base URL, e.g. {"hr": "http://localhost:8081"}.
func NewToolExecutor(registry *ToolRegistry, connector ModuleConnector, moduleAddrs map[string]string, log *zap.Logger) *ToolExecutor {
	return &ToolExecutor{
		registry:    registry,
		connector:   connector,
		moduleAddrs: moduleAddrs,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		log:         log,
	}
}

// Execute runs a tool call and returns a JSON string result suitable for the LLM context.
func (e *ToolExecutor) Execute(ctx context.Context, call llm.ToolCall) (string, error) {
	tool, ok := e.registry.Get(call.ToolName)
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrToolNotFound, call.ToolName)
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Parse arguments
	var args map[string]interface{}
	if len(call.Arguments) > 0 {
		if err := json.Unmarshal(call.Arguments, &args); err != nil {
			return "", fmt.Errorf("parse tool arguments: %w", err)
		}
	}

	result, err := e.dispatchToModule(ctx, tool, args)
	if err != nil {
		e.log.Warn("tool execution failed",
			zap.String("tool", call.ToolName),
			zap.Error(err),
		)
		return fmt.Sprintf(`{"error": %q}`, err.Error()), nil
	}

	return result, nil
}

// dispatchToModule routes the tool call to the appropriate module handler.
// For MVP: each module handles calls by method name with JSON args.
func (e *ToolExecutor) dispatchToModule(ctx context.Context, tool RegisteredTool, args map[string]interface{}) (string, error) {
	// Try to use direct gRPC connection if available via connector
	conn, hasConn := e.connector.GetConnection(tool.ModuleName)
	if hasConn {
		return e.invokeViaGRPC(ctx, conn, tool, args)
	}

	// Fallback: check static module address map
	baseURL, ok := e.moduleAddrs[tool.ModuleName]
	if !ok {
		return "", fmt.Errorf("module %q has no registered connection or address", tool.ModuleName)
	}
	return e.invokeViaHTTP(ctx, baseURL, tool, args)
}

// invokeViaGRPC invokes a tool by making a generic gRPC call.
// For MVP: uses the connection to make a raw JSON request via grpc.Invoke.
func (e *ToolExecutor) invokeViaGRPC(ctx context.Context, conn *grpc.ClientConn, tool RegisteredTool, args map[string]interface{}) (string, error) {
	// Build a simple HTTP request to the module's REST endpoint through the internal port.
	// The module gRPC connection target gives us the host:port to call.
	target := conn.Target()

	// Strip grpc:// prefix if present; use http for grpc-gateway
	host := strings.TrimPrefix(target, "passthrough:///")
	baseURL := "http://" + host

	return e.invokeViaHTTP(ctx, baseURL, tool, args)
}

// invokeViaHTTP calls a module's REST endpoint derived from the tool method name.
func (e *ToolExecutor) invokeViaHTTP(ctx context.Context, baseURL string, tool RegisteredTool, args map[string]interface{}) (string, error) {
	endpoint := buildEndpoint(baseURL, tool.ModuleName, tool.MethodName, args)

	e.log.Debug("executing tool via HTTP",
		zap.String("tool", tool.Definition.Name),
		zap.String("endpoint", endpoint),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("build HTTP request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	out, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("marshal result: %w", err)
	}

	return string(out), nil
}

// buildEndpoint constructs the REST URL for a module method call.
// Maps method names to REST paths following the project's API conventions.
func buildEndpoint(baseURL, module, method string, args map[string]interface{}) string {
	base := strings.TrimRight(baseURL, "/")

	switch module + "." + method {
	case "hr.list_teachers":
		url := base + "/api/hr/teachers"
		params := buildQueryParams(args, "search", "department", "specialization")
		if params != "" {
			url += "?" + params
		}
		return url

	case "hr.get_teacher":
		id, _ := args["teacher_id"].(string)
		return fmt.Sprintf("%s/api/hr/teachers/%s", base, id)

	case "subjects.list_subjects":
		url := base + "/api/subjects"
		params := buildQueryParams(args, "search", "credits")
		if params != "" {
			url += "?" + params
		}
		return url

	case "subjects.get_prerequisites":
		id, _ := args["subject_id"].(string)
		return fmt.Sprintf("%s/api/subjects/%s/prerequisites", base, id)

	case "timetable.generate":
		id, _ := args["semester_id"].(string)
		return fmt.Sprintf("%s/api/timetable/semesters/%s/generate", base, id)

	case "timetable.suggest_teachers":
		semID, _ := args["semester_id"].(string)
		subID, _ := args["subject_id"].(string)
		return fmt.Sprintf("%s/api/timetable/semesters/%s/suggest?subject_id=%s", base, semID, subID)

	default:
		return fmt.Sprintf("%s/api/%s/%s", base, module, strings.ReplaceAll(method, "_", "/"))
	}
}

// buildQueryParams constructs a query string from args for specified keys.
func buildQueryParams(args map[string]interface{}, keys ...string) string {
	var parts []string
	for _, k := range keys {
		if v, ok := args[k]; ok && v != nil {
			parts = append(parts, fmt.Sprintf("%s=%v", k, v))
		}
	}
	return strings.Join(parts, "&")
}

// NewDirectGRPCConnector creates a ModuleConnector that connects directly to module addresses.
func NewDirectGRPCConnector(addrs map[string]string) (ModuleConnector, error) {
	conns := make(map[string]*grpc.ClientConn, len(addrs))
	for name, addr := range addrs {
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, fmt.Errorf("connect to module %s at %s: %w", name, addr, err)
		}
		conns[name] = conn
	}
	return &staticConnector{conns: conns}, nil
}

// staticConnector is a simple ModuleConnector backed by a map of pre-built connections.
type staticConnector struct {
	conns map[string]*grpc.ClientConn
}

func (s *staticConnector) GetConnection(name string) (*grpc.ClientConn, bool) {
	conn, ok := s.conns[name]
	return conn, ok
}
