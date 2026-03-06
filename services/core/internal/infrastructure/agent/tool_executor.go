package agent

import (
	"bytes"
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
	path, httpMethod, bodyArgs, err := e.buildEndpoint(tool.ModuleName+"."+tool.MethodName, args)
	if err != nil {
		return "", fmt.Errorf("build endpoint: %w", err)
	}

	base := strings.TrimRight(baseURL, "/")
	endpoint := base + path

	e.log.Debug("executing tool via HTTP",
		zap.String("tool", tool.Definition.Name),
		zap.String("method", httpMethod),
		zap.String("endpoint", endpoint),
	)

	var bodyBytes []byte
	if bodyArgs != nil {
		bodyBytes, _ = json.Marshal(bodyArgs)
	}

	var req *http.Request
	if bodyBytes != nil {
		req, err = http.NewRequestWithContext(ctx, httpMethod, endpoint, bytes.NewReader(bodyBytes))
	} else {
		req, err = http.NewRequestWithContext(ctx, httpMethod, endpoint, nil)
	}
	if err != nil {
		return "", fmt.Errorf("build HTTP request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if bodyBytes != nil {
		req.Header.Set("Content-Type", "application/json")
	}
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

// buildEndpoint dispatches to the per-module sub-builder based on the "module.method" name.
// Returns (path, httpMethod, bodyArgs, error) — bodyArgs is nil for GET/read-only operations.
func (e *ToolExecutor) buildEndpoint(name string, args map[string]interface{}) (string, string, map[string]interface{}, error) {
	parts := strings.SplitN(name, ".", 2)
	if len(parts) != 2 {
		return "", "", nil, fmt.Errorf("invalid tool name format %q, expected module.method", name)
	}
	module, method := parts[0], parts[1]

	switch module {
	case "hr":
		return buildHREndpoint(method, args)
	case "subjects":
		return buildSubjectEndpoint(method, args)
	case "timetable":
		return buildTimetableEndpoint(method, args)
	case "student":
		return buildStudentEndpoint(method, args)
	case "analytics":
		return buildAnalyticsEndpoint(method, args)
	default:
		// Fallback: derive path from module/method name
		path := fmt.Sprintf("/api/%s/%s", module, strings.ReplaceAll(method, "_", "/"))
		return path, http.MethodGet, nil, nil
	}
}

// buildEndpoint is a package-level helper used by tests.
// It wraps the method receiver version, prepending baseURL to the returned path.
// Returns (fullURL, httpMethod, bodyBytes) — body is nil for read-only calls.
func buildEndpoint(baseURL, module, method string, args map[string]interface{}) (string, string, []byte) {
	e := &ToolExecutor{}
	path, httpMethod, bodyArgs, err := e.buildEndpoint(module+"."+method, args)
	if err != nil {
		// Fallback: derive URL from module/method like the original
		base := strings.TrimRight(baseURL, "/")
		return fmt.Sprintf("%s/api/%s/%s", base, module, strings.ReplaceAll(method, "_", "/")), http.MethodGet, nil
	}
	base := strings.TrimRight(baseURL, "/")
	var body []byte
	if bodyArgs != nil {
		body, _ = json.Marshal(bodyArgs)
	}
	return base + path, httpMethod, body
}

// stringArg safely extracts a string argument from args.
func stringArg(args map[string]interface{}, key string) string {
	if v, ok := args[key].(string); ok {
		return v
	}
	return ""
}

// copyWithout returns a shallow copy of args with the specified keys removed.
func copyWithout(args map[string]interface{}, keys ...string) map[string]interface{} {
	result := make(map[string]interface{}, len(args))
	skip := make(map[string]bool, len(keys))
	for _, k := range keys {
		skip[k] = true
	}
	for k, v := range args {
		if !skip[k] {
			result[k] = v
		}
	}
	return result
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
