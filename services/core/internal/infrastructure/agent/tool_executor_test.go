package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/llm"
	"go.uber.org/zap"
)

// --- buildEndpoint tests (pure function, no I/O) ---

func TestBuildEndpoint_HRListTeachers(t *testing.T) {
	url, method := buildEndpoint("http://localhost:8080", "hr", "list_teachers", nil)
	if method != http.MethodGet {
		t.Fatalf("expected GET, got %s", method)
	}
	if url != "http://localhost:8080/api/hr/teachers" {
		t.Fatalf("unexpected url: %s", url)
	}
}

func TestBuildEndpoint_HRListTeachers_WithParams(t *testing.T) {
	args := map[string]interface{}{"search": "alice", "department": "CS"}
	url, method := buildEndpoint("http://localhost:8080", "hr", "list_teachers", args)
	if method != http.MethodGet {
		t.Fatalf("expected GET, got %s", method)
	}
	if url == "http://localhost:8080/api/hr/teachers" {
		t.Fatal("expected query params appended")
	}
}

func TestBuildEndpoint_HRGetTeacher(t *testing.T) {
	args := map[string]interface{}{"teacher_id": "abc-123"}
	url, method := buildEndpoint("http://localhost:8080", "hr", "get_teacher", args)
	if method != http.MethodGet {
		t.Fatalf("expected GET, got %s", method)
	}
	if url != "http://localhost:8080/api/hr/teachers/abc-123" {
		t.Fatalf("unexpected url: %s", url)
	}
}

func TestBuildEndpoint_SubjectsListSubjects(t *testing.T) {
	url, method := buildEndpoint("http://localhost:8080", "subjects", "list_subjects", nil)
	if method != http.MethodGet {
		t.Fatalf("expected GET, got %s", method)
	}
	if url != "http://localhost:8080/api/subjects" {
		t.Fatalf("unexpected url: %s", url)
	}
}

func TestBuildEndpoint_SubjectsGetPrerequisites(t *testing.T) {
	args := map[string]interface{}{"subject_id": "subj-999"}
	url, method := buildEndpoint("http://localhost:8080", "subjects", "get_prerequisites", args)
	if method != http.MethodGet {
		t.Fatalf("expected GET, got %s", method)
	}
	if url != "http://localhost:8080/api/subjects/subj-999/prerequisites" {
		t.Fatalf("unexpected url: %s", url)
	}
}

func TestBuildEndpoint_TimetableGenerate(t *testing.T) {
	args := map[string]interface{}{"semester_id": "sem-42"}
	url, method := buildEndpoint("http://localhost:8080", "timetable", "generate", args)
	if method != http.MethodPost {
		t.Fatalf("expected POST, got %s", method)
	}
	if url != "http://localhost:8080/api/timetable/semesters/sem-42/generate" {
		t.Fatalf("unexpected url: %s", url)
	}
}

func TestBuildEndpoint_TimetableSuggestTeachers(t *testing.T) {
	args := map[string]interface{}{"subject_id": "s1", "semester_id": "sem1"}
	url, method := buildEndpoint("http://localhost:8080", "timetable", "suggest_teachers", args)
	if method != http.MethodGet {
		t.Fatalf("expected GET, got %s", method)
	}
	if url == "http://localhost:8080/api/timetable/suggest-teachers" {
		t.Fatal("expected query params appended")
	}
}

func TestBuildEndpoint_Default(t *testing.T) {
	url, method := buildEndpoint("http://localhost:8080", "unknown", "do_something", nil)
	if method != http.MethodGet {
		t.Fatalf("expected GET for default, got %s", method)
	}
	if url != "http://localhost:8080/api/unknown/do/something" {
		t.Fatalf("unexpected url: %s", url)
	}
}

func TestBuildEndpoint_TrailingSlashStripped(t *testing.T) {
	url, _ := buildEndpoint("http://localhost:8080/", "hr", "list_teachers", nil)
	if url != "http://localhost:8080/api/hr/teachers" {
		t.Fatalf("unexpected url: %s", url)
	}
}

// --- buildQueryParams tests ---

func TestBuildQueryParams_Empty(t *testing.T) {
	result := buildQueryParams(nil, "search")
	if result != "" {
		t.Fatalf("expected empty string, got %s", result)
	}
}

func TestBuildQueryParams_SingleKey(t *testing.T) {
	result := buildQueryParams(map[string]interface{}{"search": "alice"}, "search")
	if result != "search=alice" {
		t.Fatalf("unexpected: %s", result)
	}
}

func TestBuildQueryParams_MultipleKeys_OnePresent(t *testing.T) {
	result := buildQueryParams(map[string]interface{}{"department": "CS"}, "search", "department")
	if result != "department=CS" {
		t.Fatalf("unexpected: %s", result)
	}
}

func TestBuildQueryParams_NilValueSkipped(t *testing.T) {
	result := buildQueryParams(map[string]interface{}{"search": nil, "department": "CS"}, "search", "department")
	if result != "department=CS" {
		t.Fatalf("expected nil values skipped, got: %s", result)
	}
}

// --- Execute tests using httptest server ---

func TestToolExecutor_Execute_ToolNotFound(t *testing.T) {
	r := NewToolRegistry()
	exec := NewToolExecutor(r, nil, "http://localhost", "", zap.NewNop())

	_, err := exec.Execute(context.Background(), llm.ToolCall{ToolName: "nonexistent.tool"})
	if err == nil {
		t.Fatal("expected tool-not-found error")
	}
}

func TestToolExecutor_Execute_HTTPSuccess(t *testing.T) {
	// Spin up a test HTTP server that returns JSON.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"teachers": []interface{}{}})
	}))
	defer srv.Close()

	r := NewToolRegistry()
	r.Register(HRTools) // registers "hr.list_teachers"

	exec := NewToolExecutor(r, nil, srv.URL, "test-jwt", zap.NewNop())

	result, err := exec.Execute(context.Background(), llm.ToolCall{
		ToolName:  "hr.list_teachers",
		Arguments: json.RawMessage(`{}`),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == "" {
		t.Fatal("expected non-empty result")
	}
}

func TestToolExecutor_Execute_HTTPError(t *testing.T) {
	// Point at a server that is not running.
	r := NewToolRegistry()
	r.Register(HRTools)

	exec := NewToolExecutor(r, nil, "http://127.0.0.1:1", "", zap.NewNop())

	result, err := exec.Execute(context.Background(), llm.ToolCall{
		ToolName:  "hr.list_teachers",
		Arguments: json.RawMessage(`{}`),
	})
	// Execute swallows HTTP errors and returns JSON error string — err should be nil.
	if err != nil {
		t.Fatalf("expected err=nil (error swallowed), got: %v", err)
	}
	if result == "" {
		t.Fatal("expected non-empty error JSON result")
	}
}

func TestToolExecutor_Execute_InvalidJSON(t *testing.T) {
	r := NewToolRegistry()
	r.Register(HRTools)

	exec := NewToolExecutor(r, nil, "http://localhost", "", zap.NewNop())

	_, err := exec.Execute(context.Background(), llm.ToolCall{
		ToolName:  "hr.list_teachers",
		Arguments: json.RawMessage(`{invalid json`),
	})
	if err == nil {
		t.Fatal("expected error for invalid JSON arguments")
	}
}

// --- ErrToolNotFound sentinel tests ---

func TestErrToolNotFound_Sentinel(t *testing.T) {
	if ErrToolNotFound == nil {
		t.Fatal("ErrToolNotFound must be non-nil")
	}
	if ErrToolNotFound.Error() == "" {
		t.Fatal("ErrToolNotFound must have a message")
	}
}
