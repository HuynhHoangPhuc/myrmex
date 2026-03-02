package agent

import "github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/llm"

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
	{
		Definition: llm.Tool{
			Name:        "subject.get_subject",
			Description: "Get detailed information about a specific subject by UUID.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"subject_id": {"type": "string", "description": "UUID of the subject"}
				},
				"required": ["subject_id"]
			}`),
		},
		ModuleName: "subjects",
		MethodName: "get_subject",
	},
	{
		Definition: llm.Tool{
			Name:        "subject.validate_dag",
			Description: "Validate the subject prerequisite DAG for cycles or inconsistencies.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {}
			}`),
		},
		ModuleName: "subjects",
		MethodName: "validate_dag",
	},
	{
		Definition: llm.Tool{
			Name:        "subject.topological_sort",
			Description: "Return subjects in topological order based on prerequisites.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {}
			}`),
		},
		ModuleName: "subjects",
		MethodName: "topological_sort",
	},
	{
		Definition: llm.Tool{
			Name:        "subject.full_dag",
			Description: "Return the full prerequisite DAG for all subjects.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {}
			}`),
		},
		ModuleName: "subjects",
		MethodName: "full_dag",
	},
	{
		Definition: llm.Tool{
			Name:        "subject.create_subject",
			Description: "Create a new subject/course.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"name":        {"type": "string",  "description": "Subject name"},
					"code":        {"type": "string",  "description": "Short subject code (e.g. CS101)"},
					"credits":     {"type": "integer", "description": "Number of credit hours"},
					"description": {"type": "string",  "description": "Optional description of the subject"}
				},
				"required": ["name", "code", "credits"]
			}`),
		},
		ModuleName: "subjects",
		MethodName: "create_subject",
	},
	{
		Definition: llm.Tool{
			Name:        "subject.update_subject",
			Description: "Update an existing subject (partial update).",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"subject_id":  {"type": "string",  "description": "UUID of the subject to update"},
					"name":        {"type": "string",  "description": "Subject name"},
					"code":        {"type": "string",  "description": "Short subject code"},
					"credits":     {"type": "integer", "description": "Number of credit hours"},
					"description": {"type": "string",  "description": "Description of the subject"}
				},
				"required": ["subject_id"]
			}`),
		},
		ModuleName: "subjects",
		MethodName: "update_subject",
	},
	{
		Definition: llm.Tool{
			Name:        "subject.delete_subject",
			Description: "Delete a subject by UUID.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"subject_id": {"type": "string", "description": "UUID of the subject to delete"}
				},
				"required": ["subject_id"]
			}`),
		},
		ModuleName: "subjects",
		MethodName: "delete_subject",
	},
	{
		Definition: llm.Tool{
			Name:        "subject.add_prerequisite",
			Description: "Add a prerequisite relationship between two subjects.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"subject_id":       {"type": "string",  "description": "UUID of the subject that requires the prerequisite"},
					"prerequisite_id":  {"type": "string",  "description": "UUID of the prerequisite subject"},
					"type":             {"type": "string",  "description": "Prerequisite type: required | recommended"},
					"priority":         {"type": "integer", "description": "Priority order for prerequisites"}
				},
				"required": ["subject_id", "prerequisite_id"]
			}`),
		},
		ModuleName: "subjects",
		MethodName: "add_prerequisite",
	},
	{
		Definition: llm.Tool{
			Name:        "subject.remove_prerequisite",
			Description: "Remove a prerequisite relationship between two subjects.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"subject_id":      {"type": "string", "description": "UUID of the subject"},
					"prerequisite_id": {"type": "string", "description": "UUID of the prerequisite to remove"}
				},
				"required": ["subject_id", "prerequisite_id"]
			}`),
		},
		ModuleName: "subjects",
		MethodName: "remove_prerequisite",
	},
	{
		Definition: llm.Tool{
			Name:        "subject.check_conflicts",
			Description: "Check whether adding a prerequisite relationship would create a cycle in the DAG.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"subject_id":      {"type": "string", "description": "UUID of the subject that would have the prerequisite added"},
					"prerequisite_id": {"type": "string", "description": "UUID of the candidate prerequisite"}
				},
				"required": ["subject_id", "prerequisite_id"]
			}`),
		},
		ModuleName: "subjects",
		MethodName: "check_conflicts",
	},
}
