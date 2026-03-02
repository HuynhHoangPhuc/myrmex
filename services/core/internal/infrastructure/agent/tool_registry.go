package agent

import (
	"encoding/json"
	"sync"

	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/llm"
)

// RegisteredTool pairs an LLM tool definition with its execution metadata.
type RegisteredTool struct {
	Definition llm.Tool
	ModuleName string // e.g. "hr", "subjects", "timetable"
	MethodName string // identifies which handler to invoke
}

// ToolRegistry is a thread-safe registry of tools available to the agent.
type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]RegisteredTool
}

// NewToolRegistry creates an empty registry.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{tools: make(map[string]RegisteredTool)}
}

// Register adds a set of tools for a given module, overwriting any existing entry with the same name.
func (r *ToolRegistry) Register(tools []RegisteredTool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range tools {
		r.tools[t.Definition.Name] = t
	}
}

// Get returns a tool by name.
func (r *ToolRegistry) Get(name string) (RegisteredTool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tools[name]
	return t, ok
}

// GetAllTools returns all tool definitions for sending to the LLM.
func (r *ToolRegistry) GetAllTools() []llm.Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tools := make([]llm.Tool, 0, len(r.tools))
	for _, rt := range r.tools {
		tools = append(tools, rt.Definition)
	}
	return tools
}

// rawSchema is a helper that returns a json.RawMessage from a literal JSON string.
func rawSchema(s string) json.RawMessage {
	return json.RawMessage(s)
}

// DefaultTools is the full set of tools registered on startup.
// Tool definitions live in per-module files: tools_hr.go, tools_subject.go,
// tools_timetable.go, tools_student.go, tools_analytics.go.
var DefaultTools = func() []RegisteredTool {
	var all []RegisteredTool
	all = append(all, HRTools...)
	all = append(all, SubjectTools...)
	all = append(all, TimetableTools...)
	all = append(all, StudentTools...)
	all = append(all, AnalyticsTools...)
	return all
}()
