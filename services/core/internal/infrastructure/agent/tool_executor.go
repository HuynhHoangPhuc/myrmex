package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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
// For MVP: uses HTTP REST calls to core's own HTTP API so the correct router
// and middleware handle the request (gRPC ports do not serve REST).
type ToolExecutor struct {
	registry    *ToolRegistry
	connector   ModuleConnector
	selfURL     string // core's own HTTP base URL, e.g. "http://localhost:8080"
	internalJWT string // service-level JWT for internal requests
	httpClient  *http.Client
	log         *zap.Logger
}

// NewToolExecutor creates a ToolExecutor.
// selfURL is core's own HTTP base URL (e.g. "http://localhost:8080").
// internalJWT is a pre-generated service JWT attached to every internal request.
func NewToolExecutor(registry *ToolRegistry, connector ModuleConnector, selfURL, internalJWT string, log *zap.Logger) *ToolExecutor {
	return &ToolExecutor{
		registry:    registry,
		connector:   connector,
		selfURL:     selfURL,
		internalJWT: internalJWT,
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
// All calls are routed through core's own HTTP API (selfURL) regardless of
// whether a gRPC connection is available, because gRPC ports do not serve REST.
func (e *ToolExecutor) dispatchToModule(ctx context.Context, tool RegisteredTool, args map[string]interface{}) (string, error) {
	return e.invokeViaHTTP(ctx, e.selfURL, tool, args)
}

// invokeViaGRPC is kept for interface compatibility but routes through selfURL.
// gRPC ports do not serve REST, so we always use core's HTTP API instead.
func (e *ToolExecutor) invokeViaGRPC(ctx context.Context, conn *grpc.ClientConn, tool RegisteredTool, args map[string]interface{}) (string, error) {
	return e.invokeViaHTTP(ctx, e.selfURL, tool, args)
}

// invokeViaHTTP calls a module's REST endpoint derived from the tool method name.
func (e *ToolExecutor) invokeViaHTTP(ctx context.Context, baseURL string, tool RegisteredTool, args map[string]interface{}) (string, error) {
	endpoint, httpMethod := buildEndpoint(baseURL, tool.ModuleName, tool.MethodName, args)

	e.log.Debug("executing tool via HTTP",
		zap.String("tool", tool.Definition.Name),
		zap.String("method", httpMethod),
		zap.String("endpoint", endpoint),
	)

	req, err := http.NewRequestWithContext(ctx, httpMethod, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("build HTTP request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if e.internalJWT != "" {
		req.Header.Set("Authorization", "Bearer "+e.internalJWT)
	}

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

// buildEndpoint constructs the REST URL and HTTP method for a module method call.
// Returns (url, httpMethod) — httpMethod is GET for reads, POST for mutations.
func buildEndpoint(baseURL, module, method string, args map[string]interface{}) (string, string) {
	base := strings.TrimRight(baseURL, "/")

	switch module + "." + method {
	case "hr.list_teachers":
		url := base + "/api/hr/teachers"
		if params := buildQueryParams(args, "search", "department", "specialization"); params != "" {
			url += "?" + params
		}
		return url, http.MethodGet

	case "hr.get_teacher":
		id, _ := args["teacher_id"].(string)
		return fmt.Sprintf("%s/api/hr/teachers/%s", base, id), http.MethodGet

	case "subjects.list_subjects":
		url := base + "/api/subjects"
		if params := buildQueryParams(args, "search", "credits"); params != "" {
			url += "?" + params
		}
		return url, http.MethodGet

	case "subjects.get_prerequisites":
		id, _ := args["subject_id"].(string)
		return fmt.Sprintf("%s/api/subjects/%s/prerequisites", base, id), http.MethodGet

	case "timetable.list_semesters":
		return base + "/api/timetable/semesters?page=1&page_size=50", http.MethodGet

	case "timetable.generate":
		// POST: triggers schedule generation for a semester
		id, _ := args["semester_id"].(string)
		return fmt.Sprintf("%s/api/timetable/semesters/%s/generate", base, id), http.MethodPost

	case "timetable.suggest_teachers":
		url := base + "/api/timetable/suggest-teachers"
		if params := buildQueryParams(args, "subject_id"); params != "" {
			url += "?" + params
		}
		return url, http.MethodGet

	default:
		return fmt.Sprintf("%s/api/%s/%s", base, module, strings.ReplaceAll(method, "_", "/")), http.MethodGet
	}
}

// buildQueryParams constructs a properly URL-encoded query string from args for specified keys.
func buildQueryParams(args map[string]interface{}, keys ...string) string {
	q := url.Values{}
	for _, k := range keys {
		if v, ok := args[k]; ok && v != nil {
			q.Set(k, fmt.Sprintf("%v", v))
		}
	}
	return q.Encode()
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
