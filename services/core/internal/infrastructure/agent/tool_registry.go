package agent

import (
	"encoding/json"
	"sync"

	"github.com/myrmex-erp/myrmex/services/core/internal/infrastructure/llm"
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

// HRTools defines the tools exposed by the HR module.
var HRTools = []RegisteredTool{
	{
		Definition: llm.Tool{
			Name:        "hr.list_teachers",
			Description: "List all teachers with optional filters by search keyword, department, or specialization.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"search":         {"type": "string",  "description": "Full-text search by name or email"},
					"department":     {"type": "string",  "description": "Filter by department name"},
					"specialization": {"type": "string",  "description": "Filter by subject specialization"}
				}
			}`),
		},
		ModuleName: "hr",
		MethodName: "list_teachers",
	},
	{
		Definition: llm.Tool{
			Name:        "hr.get_teacher",
			Description: "Get detailed information about a specific teacher by their UUID.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"teacher_id": {"type": "string", "description": "UUID of the teacher"}
				},
				"required": ["teacher_id"]
			}`),
		},
		ModuleName: "hr",
		MethodName: "get_teacher",
	},
}

// SubjectTools defines the tools exposed by the Subject module.
var SubjectTools = []RegisteredTool{
	{
		Definition: llm.Tool{
			Name:        "subject.list_subjects",
			Description: "List all subjects with optional filters by name, code, or credits.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"search":  {"type": "string",  "description": "Full-text search by subject name or code"},
					"credits": {"type": "integer", "description": "Filter by number of credits"}
				}
			}`),
		},
		ModuleName: "subjects",
		MethodName: "list_subjects",
	},
	{
		Definition: llm.Tool{
			Name:        "subject.get_prerequisites",
			Description: "Get the prerequisite subjects for a given subject (returns a DAG).",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"subject_id": {"type": "string", "description": "UUID of the subject"}
				},
				"required": ["subject_id"]
			}`),
		},
		ModuleName: "subjects",
		MethodName: "get_prerequisites",
	},
}

// TimetableTools defines the tools exposed by the Timetable module.
var TimetableTools = []RegisteredTool{
	{
		Definition: llm.Tool{
			Name:        "timetable.generate",
			Description: "Generate a timetable for a semester using CSP solver. Returns a schedule or best-partial on timeout.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"semester_id": {"type": "string", "description": "UUID of the semester to generate schedule for"}
				},
				"required": ["semester_id"]
			}`),
		},
		ModuleName: "timetable",
		MethodName: "generate",
	},
	{
		Definition: llm.Tool{
			Name:        "timetable.suggest_teachers",
			Description: "Suggest available teachers for a given subject and time slot.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"subject_id":  {"type": "string", "description": "UUID of the subject"},
					"semester_id": {"type": "string", "description": "UUID of the semester"}
				},
				"required": ["subject_id", "semester_id"]
			}`),
		},
		ModuleName: "timetable",
		MethodName: "suggest_teachers",
	},
}

// DefaultTools is the full set of tools registered on startup.
var DefaultTools = append(append(HRTools, SubjectTools...), TimetableTools...)
