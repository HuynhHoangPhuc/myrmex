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
	endpoint, httpMethod, body := buildEndpoint(baseURL, tool.ModuleName, tool.MethodName, args)

	e.log.Debug("executing tool via HTTP",
		zap.String("tool", tool.Definition.Name),
		zap.String("method", httpMethod),
		zap.String("endpoint", endpoint),
	)

	var bodyReader *bytes.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	var req *http.Request
	var err error
	if bodyReader != nil {
		req, err = http.NewRequestWithContext(ctx, httpMethod, endpoint, bodyReader)
	} else {
		req, err = http.NewRequestWithContext(ctx, httpMethod, endpoint, nil)
	}
	if err != nil {
		return "", fmt.Errorf("build HTTP request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
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

// buildEndpoint constructs the REST URL, HTTP method, and optional request body for a module method call.
// Returns (url, httpMethod, body) — body is nil for GET/read-only, non-nil JSON bytes for mutations.
func buildEndpoint(baseURL, module, method string, args map[string]interface{}) (string, string, []byte) {
	base := strings.TrimRight(baseURL, "/")

	switch module + "." + method {
	case "hr.list_teachers":
		url := base + "/api/hr/teachers"
		if params := buildQueryParams(args, "search", "department_id", "specialization"); params != "" {
			url += "?" + params
		}
		return url, http.MethodGet, nil

	case "hr.get_teacher":
		id, _ := args["teacher_id"].(string)
		return fmt.Sprintf("%s/api/hr/teachers/%s", base, id), http.MethodGet, nil

	case "hr.list_departments":
		return base + "/api/hr/departments", http.MethodGet, nil

	case "hr.get_teacher_availability":
		id, _ := args["teacher_id"].(string)
		return fmt.Sprintf("%s/api/hr/teachers/%s/availability", base, id), http.MethodGet, nil

	case "subjects.list_subjects":
		url := base + "/api/subjects"
		if params := buildQueryParams(args, "search", "credits"); params != "" {
			url += "?" + params
		}
		return url, http.MethodGet, nil

	case "subjects.get_prerequisites":
		id, _ := args["subject_id"].(string)
		return fmt.Sprintf("%s/api/subjects/%s/prerequisites", base, id), http.MethodGet, nil

	case "subjects.get_subject":
		id, _ := args["subject_id"].(string)
		return fmt.Sprintf("%s/api/subjects/%s", base, id), http.MethodGet, nil

	case "subjects.validate_dag":
		return base + "/api/subjects/dag/validate", http.MethodGet, nil

	case "subjects.topological_sort":
		return base + "/api/subjects/dag/topological-sort", http.MethodGet, nil

	case "subjects.full_dag":
		return base + "/api/subjects/dag/full", http.MethodGet, nil

	case "timetable.list_semesters":
		return base + "/api/timetable/semesters?page=1&page_size=50", http.MethodGet, nil

	case "timetable.generate":
		// POST: triggers schedule generation for a semester
		id, _ := args["semester_id"].(string)
		return fmt.Sprintf("%s/api/timetable/semesters/%s/generate", base, id), http.MethodPost, nil

	case "timetable.suggest_teachers":
		url := base + "/api/timetable/suggest-teachers"
		if params := buildQueryParams(args, "subject_id"); params != "" {
			url += "?" + params
		}
		return url, http.MethodGet, nil

	case "timetable.get_semester":
		id, _ := args["semester_id"].(string)
		return fmt.Sprintf("%s/api/timetable/semesters/%s", base, id), http.MethodGet, nil

	case "timetable.list_schedules":
		url := base + "/api/timetable/schedules"
		if params := buildQueryParams(args, "semester_id"); params != "" {
			url += "?" + params
		}
		return url, http.MethodGet, nil

	case "timetable.get_schedule":
		id, _ := args["schedule_id"].(string)
		return fmt.Sprintf("%s/api/timetable/schedules/%s", base, id), http.MethodGet, nil

	case "timetable.list_rooms":
		return base + "/api/timetable/rooms", http.MethodGet, nil

	case "student.list":
		url := base + "/api/students"
		if params := buildQueryParams(args, "department_id", "status"); params != "" {
			url += "?" + params
		}
		return url, http.MethodGet, nil

	case "student.get":
		id, _ := args["id"].(string)
		return fmt.Sprintf("%s/api/students/%s", base, id), http.MethodGet, nil

	case "student.transcript":
		id, _ := args["student_id"].(string)
		return fmt.Sprintf("%s/api/students/%s/transcript", base, id), http.MethodGet, nil

	case "student.list_enrollments":
		url := base + "/api/enrollments"
		if params := buildQueryParams(args, "student_id", "subject_id", "status"); params != "" {
			url += "?" + params
		}
		return url, http.MethodGet, nil

	case "analytics.workload":
		url := base + "/api/analytics/workload"
		if params := buildQueryParams(args, "semester_id"); params != "" {
			url += "?" + params
		}
		return url, http.MethodGet, nil

	case "analytics.utilization":
		url := base + "/api/analytics/utilization"
		if params := buildQueryParams(args, "semester_id"); params != "" {
			url += "?" + params
		}
		return url, http.MethodGet, nil

	case "analytics.dashboard":
		return base + "/api/analytics/dashboard", http.MethodGet, nil

	case "analytics.department_metrics":
		url := base + "/api/analytics/department-metrics"
		if params := buildQueryParams(args, "department_id"); params != "" {
			url += "?" + params
		}
		return url, http.MethodGet, nil

	case "analytics.schedule_metrics":
		url := base + "/api/analytics/schedule-metrics"
		if params := buildQueryParams(args, "schedule_id"); params != "" {
			url += "?" + params
		}
		return url, http.MethodGet, nil

	case "analytics.schedule_heatmap":
		url := base + "/api/analytics/schedule-heatmap"
		if params := buildQueryParams(args, "schedule_id"); params != "" {
			url += "?" + params
		}
		return url, http.MethodGet, nil

	// --- HR mutations ---

	case "hr.create_teacher":
		body, _ := json.Marshal(args)
		return base + "/api/hr/teachers", http.MethodPost, body

	case "hr.update_teacher":
		id := stringArg(args, "teacher_id")
		body, _ := json.Marshal(copyWithout(args, "teacher_id"))
		return fmt.Sprintf("%s/api/hr/teachers/%s", base, id), http.MethodPatch, body

	case "hr.delete_teacher":
		id := stringArg(args, "teacher_id")
		return fmt.Sprintf("%s/api/hr/teachers/%s", base, id), http.MethodDelete, nil

	case "hr.update_teacher_availability":
		id := stringArg(args, "teacher_id")
		body, _ := json.Marshal(copyWithout(args, "teacher_id"))
		return fmt.Sprintf("%s/api/hr/teachers/%s/availability", base, id), http.MethodPut, body

	case "hr.create_department":
		body, _ := json.Marshal(args)
		return base + "/api/hr/departments", http.MethodPost, body

	// --- Subject mutations ---

	case "subjects.create_subject":
		body, _ := json.Marshal(args)
		return base + "/api/subjects", http.MethodPost, body

	case "subjects.update_subject":
		id := stringArg(args, "subject_id")
		body, _ := json.Marshal(copyWithout(args, "subject_id"))
		return fmt.Sprintf("%s/api/subjects/%s", base, id), http.MethodPatch, body

	case "subjects.delete_subject":
		id := stringArg(args, "subject_id")
		return fmt.Sprintf("%s/api/subjects/%s", base, id), http.MethodDelete, nil

	case "subjects.add_prerequisite":
		id := stringArg(args, "subject_id")
		body, _ := json.Marshal(copyWithout(args, "subject_id"))
		return fmt.Sprintf("%s/api/subjects/%s/prerequisites", base, id), http.MethodPost, body

	case "subjects.remove_prerequisite":
		id := stringArg(args, "subject_id")
		prereqID := stringArg(args, "prerequisite_id")
		return fmt.Sprintf("%s/api/subjects/%s/prerequisites/%s", base, id, prereqID), http.MethodDelete, nil

	case "subjects.check_conflicts":
		body, _ := json.Marshal(args)
		return base + "/api/subjects/dag/check-conflicts", http.MethodPost, body

	// --- Timetable mutations ---

	case "timetable.create_semester":
		body, _ := json.Marshal(args)
		return base + "/api/timetable/semesters", http.MethodPost, body

	case "timetable.set_semester_rooms":
		id := stringArg(args, "semester_id")
		body, _ := json.Marshal(copyWithout(args, "semester_id"))
		return fmt.Sprintf("%s/api/timetable/semesters/%s/rooms", base, id), http.MethodPut, body

	case "timetable.create_time_slot":
		id := stringArg(args, "semester_id")
		body, _ := json.Marshal(copyWithout(args, "semester_id"))
		return fmt.Sprintf("%s/api/timetable/semesters/%s/slots", base, id), http.MethodPost, body

	case "timetable.delete_time_slot":
		semID := stringArg(args, "semester_id")
		slotID := stringArg(args, "slot_id")
		return fmt.Sprintf("%s/api/timetable/semesters/%s/slots/%s", base, semID, slotID), http.MethodDelete, nil

	case "timetable.apply_time_slot_preset":
		id := stringArg(args, "semester_id")
		body, _ := json.Marshal(copyWithout(args, "semester_id"))
		return fmt.Sprintf("%s/api/timetable/semesters/%s/slots/preset", base, id), http.MethodPost, body

	case "timetable.add_offered_subject":
		id := stringArg(args, "semester_id")
		body, _ := json.Marshal(copyWithout(args, "semester_id"))
		return fmt.Sprintf("%s/api/timetable/semesters/%s/offered-subjects", base, id), http.MethodPost, body

	case "timetable.remove_offered_subject":
		semID := stringArg(args, "semester_id")
		subjectID := stringArg(args, "subject_id")
		return fmt.Sprintf("%s/api/timetable/semesters/%s/offered-subjects/%s", base, semID, subjectID), http.MethodDelete, nil

	case "timetable.manual_assign":
		schedID := stringArg(args, "schedule_id")
		entryID := stringArg(args, "entry_id")
		body, _ := json.Marshal(copyWithout(args, "schedule_id", "entry_id"))
		return fmt.Sprintf("%s/api/timetable/schedules/%s/entries/%s", base, schedID, entryID), http.MethodPut, body

	// --- Student mutations ---

	case "student.create":
		body, _ := json.Marshal(args)
		return base + "/api/students", http.MethodPost, body

	case "student.update":
		id := stringArg(args, "id")
		body, _ := json.Marshal(copyWithout(args, "id"))
		return fmt.Sprintf("%s/api/students/%s", base, id), http.MethodPatch, body

	case "student.delete":
		id := stringArg(args, "id")
		return fmt.Sprintf("%s/api/students/%s", base, id), http.MethodDelete, nil

	case "student.review_enrollment":
		id := stringArg(args, "enrollment_id")
		body, _ := json.Marshal(copyWithout(args, "enrollment_id"))
		return fmt.Sprintf("%s/api/enrollments/%s/review", base, id), http.MethodPatch, body

	default:
		return fmt.Sprintf("%s/api/%s/%s", base, module, strings.ReplaceAll(method, "_", "/")), http.MethodGet, nil
	}
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
