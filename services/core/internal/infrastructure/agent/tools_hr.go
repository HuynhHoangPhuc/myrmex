package agent

import "github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/llm"

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
	{
		Definition: llm.Tool{
			Name:        "hr.list_departments",
			Description: "List all departments.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {}
			}`),
		},
		ModuleName: "hr",
		MethodName: "list_departments",
	},
	{
		Definition: llm.Tool{
			Name:        "hr.get_teacher_availability",
			Description: "Get availability schedule for a specific teacher.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"teacher_id": {"type": "string", "description": "UUID of the teacher"}
				},
				"required": ["teacher_id"]
			}`),
		},
		ModuleName: "hr",
		MethodName: "get_teacher_availability",
	},
	{
		Definition: llm.Tool{
			Name:        "hr.create_teacher",
			Description: "Create a new teacher record.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"first_name":       {"type": "string",  "description": "Teacher's first name"},
					"last_name":        {"type": "string",  "description": "Teacher's last name"},
					"email":            {"type": "string",  "description": "Teacher's email address"},
					"title":            {"type": "string",  "description": "Academic title (e.g. Dr., Prof.)"},
					"department_id":    {"type": "string",  "description": "UUID of the department"},
					"specializations":  {"type": "array",   "items": {"type": "string"}, "description": "List of subject specializations"}
				},
				"required": ["first_name", "last_name", "email"]
			}`),
		},
		ModuleName: "hr",
		MethodName: "create_teacher",
	},
	{
		Definition: llm.Tool{
			Name:        "hr.update_teacher",
			Description: "Update an existing teacher's information (partial update).",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"teacher_id":       {"type": "string",  "description": "UUID of the teacher to update"},
					"first_name":       {"type": "string",  "description": "Teacher's first name"},
					"last_name":        {"type": "string",  "description": "Teacher's last name"},
					"email":            {"type": "string",  "description": "Teacher's email address"},
					"title":            {"type": "string",  "description": "Academic title"},
					"department_id":    {"type": "string",  "description": "UUID of the department"},
					"specializations":  {"type": "array",   "items": {"type": "string"}, "description": "List of subject specializations"}
				},
				"required": ["teacher_id"]
			}`),
		},
		ModuleName: "hr",
		MethodName: "update_teacher",
	},
	{
		Definition: llm.Tool{
			Name:        "hr.delete_teacher",
			Description: "Delete a teacher record by UUID.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"teacher_id": {"type": "string", "description": "UUID of the teacher to delete"}
				},
				"required": ["teacher_id"]
			}`),
		},
		ModuleName: "hr",
		MethodName: "delete_teacher",
	},
	{
		Definition: llm.Tool{
			Name:        "hr.update_teacher_availability",
			Description: "Set the weekly availability schedule for a teacher (replaces existing schedule).",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"teacher_id": {"type": "string", "description": "UUID of the teacher"},
					"availability": {
						"type": "array",
						"description": "List of availability time windows",
						"items": {
							"type": "object",
							"properties": {
								"day_of_week": {"type": "integer", "description": "Day of week: 0=Sunday … 6=Saturday"},
								"start_time":  {"type": "string",  "description": "Start time in HH:MM format"},
								"end_time":    {"type": "string",  "description": "End time in HH:MM format"}
							},
							"required": ["day_of_week", "start_time", "end_time"]
						}
					}
				},
				"required": ["teacher_id", "availability"]
			}`),
		},
		ModuleName: "hr",
		MethodName: "update_teacher_availability",
	},
	{
		Definition: llm.Tool{
			Name:        "hr.create_department",
			Description: "Create a new department.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"name": {"type": "string", "description": "Full department name"},
					"code": {"type": "string", "description": "Short department code (e.g. CS, MATH)"}
				},
				"required": ["name", "code"]
			}`),
		},
		ModuleName: "hr",
		MethodName: "create_department",
	},
}
