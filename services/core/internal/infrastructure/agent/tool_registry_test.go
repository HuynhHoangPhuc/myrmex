package agent

import (
	"sync"
	"testing"

	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/llm"
)

func TestToolRegistry_RegisterAndGet(t *testing.T) {
	r := NewToolRegistry()

	tool := RegisteredTool{
		Definition: llm.Tool{Name: "hr.list_teachers", Description: "List teachers"},
		ModuleName: "hr",
		MethodName: "list_teachers",
	}
	r.Register([]RegisteredTool{tool})

	got, ok := r.Get("hr.list_teachers")
	if !ok {
		t.Fatal("expected tool to be found")
	}
	if got.Definition.Name != "hr.list_teachers" {
		t.Fatalf("name mismatch: got %s", got.Definition.Name)
	}
	if got.ModuleName != "hr" {
		t.Fatalf("module mismatch: got %s", got.ModuleName)
	}
	if got.MethodName != "list_teachers" {
		t.Fatalf("method mismatch: got %s", got.MethodName)
	}
}

func TestToolRegistry_Get_NotFound(t *testing.T) {
	r := NewToolRegistry()
	_, ok := r.Get("nonexistent.tool")
	if ok {
		t.Fatal("expected tool not to be found")
	}
}

func TestToolRegistry_Register_Overwrites(t *testing.T) {
	r := NewToolRegistry()

	r.Register([]RegisteredTool{{
		Definition: llm.Tool{Name: "hr.get_teacher", Description: "v1"},
		ModuleName: "hr",
		MethodName: "get_teacher",
	}})
	r.Register([]RegisteredTool{{
		Definition: llm.Tool{Name: "hr.get_teacher", Description: "v2"},
		ModuleName: "hr",
		MethodName: "get_teacher_v2",
	}})

	got, ok := r.Get("hr.get_teacher")
	if !ok {
		t.Fatal("expected tool to be found")
	}
	if got.Definition.Description != "v2" {
		t.Fatalf("expected overwritten description v2, got %s", got.Definition.Description)
	}
}

func TestToolRegistry_GetAllTools(t *testing.T) {
	r := NewToolRegistry()
	r.Register([]RegisteredTool{
		{Definition: llm.Tool{Name: "a.tool1"}, ModuleName: "a", MethodName: "tool1"},
		{Definition: llm.Tool{Name: "b.tool2"}, ModuleName: "b", MethodName: "tool2"},
	})

	tools := r.GetAllTools()
	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(tools))
	}

	names := map[string]bool{}
	for _, t := range tools {
		names[t.Name] = true
	}
	if !names["a.tool1"] || !names["b.tool2"] {
		t.Fatal("expected both tool names present")
	}
}

func TestToolRegistry_GetAllTools_Empty(t *testing.T) {
	r := NewToolRegistry()
	tools := r.GetAllTools()
	if len(tools) != 0 {
		t.Fatalf("expected 0 tools, got %d", len(tools))
	}
}

func TestToolRegistry_ConcurrentAccess(t *testing.T) {
	r := NewToolRegistry()

	var wg sync.WaitGroup
	// concurrent writers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			r.Register([]RegisteredTool{{
				Definition: llm.Tool{Name: "concurrent.tool"},
				ModuleName: "concurrent",
				MethodName: "tool",
			}})
		}(i)
	}
	// concurrent readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.Get("concurrent.tool")
			r.GetAllTools()
		}()
	}
	wg.Wait()
}

func TestDefaultTools_NotEmpty(t *testing.T) {
	if len(DefaultTools) == 0 {
		t.Fatal("DefaultTools must not be empty")
	}
	for _, tool := range DefaultTools {
		if tool.Definition.Name == "" {
			t.Fatal("every default tool must have a non-empty Name")
		}
		if tool.ModuleName == "" {
			t.Fatalf("tool %s has empty ModuleName", tool.Definition.Name)
		}
		if tool.MethodName == "" {
			t.Fatalf("tool %s has empty MethodName", tool.Definition.Name)
		}
	}
}

func TestToolRegistry_RegisterMultipleAtOnce(t *testing.T) {
	r := NewToolRegistry()
	r.Register(HRTools)
	r.Register(SubjectTools)
	r.Register(TimetableTools)

	tools := r.GetAllTools()
	expected := len(HRTools) + len(SubjectTools) + len(TimetableTools)
	if len(tools) != expected {
		t.Fatalf("expected %d tools, got %d", expected, len(tools))
	}
}
