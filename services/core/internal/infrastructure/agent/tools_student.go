package agent

import "github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/llm"

// StudentTools defines the tools exposed by the Student module.
var StudentTools = []RegisteredTool{
	{
		Definition: llm.Tool{
			Name:        "student.list",
			Description: "List students with optional filters by department or status.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"department_id": {"type": "string", "description": "Filter by department UUID"},
					"status":        {"type": "string", "description": "Filter by status: active | graduated | suspended"}
				}
			}`),
		},
		ModuleName: "student",
		MethodName: "list",
	},
	{
		Definition: llm.Tool{
			Name:        "student.get",
			Description: "Get student details by ID.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"id": {"type": "string", "description": "UUID of the student"}
				},
				"required": ["id"]
			}`),
		},
		ModuleName: "student",
		MethodName: "get",
	},
	{
		Definition: llm.Tool{
			Name:        "student.transcript",
			Description: "Get a student's academic transcript including grades and GPA.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"student_id": {"type": "string", "description": "UUID of the student"}
				},
				"required": ["student_id"]
			}`),
		},
		ModuleName: "student",
		MethodName: "transcript",
	},
	{
		Definition: llm.Tool{
			Name:        "student.list_enrollments",
			Description: "List enrollments with optional filters by student or status.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"student_id": {"type": "string", "description": "Optional: filter by student UUID"},
					"status":     {"type": "string", "description": "Optional: filter by enrollment status"}
				}
			}`),
		},
		ModuleName: "student",
		MethodName: "list_enrollments",
	},
	{
		Definition: llm.Tool{
			Name:        "student.create",
			Description: "Create a new student record.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"student_code":   {"type": "string", "description": "Unique student identification code"},
					"first_name":     {"type": "string", "description": "Student's first name"},
					"last_name":      {"type": "string", "description": "Student's last name"},
					"email":          {"type": "string", "description": "Student's email address"},
					"department_id":  {"type": "string", "description": "UUID of the student's department"}
				},
				"required": ["student_code", "first_name", "last_name", "email"]
			}`),
		},
		ModuleName: "student",
		MethodName: "create",
	},
	{
		Definition: llm.Tool{
			Name:        "student.update",
			Description: "Update an existing student's information (partial update).",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"id":            {"type": "string", "description": "UUID of the student to update"},
					"student_code":  {"type": "string", "description": "Unique student identification code"},
					"first_name":    {"type": "string", "description": "Student's first name"},
					"last_name":     {"type": "string", "description": "Student's last name"},
					"email":         {"type": "string", "description": "Student's email address"},
					"department_id": {"type": "string", "description": "UUID of the student's department"},
					"status":        {"type": "string", "description": "Student status: active | graduated | suspended"}
				},
				"required": ["id"]
			}`),
		},
		ModuleName: "student",
		MethodName: "update",
	},
	{
		Definition: llm.Tool{
			Name:        "student.delete",
			Description: "Delete a student record by UUID.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"id": {"type": "string", "description": "UUID of the student to delete"}
				},
				"required": ["id"]
			}`),
		},
		ModuleName: "student",
		MethodName: "delete",
	},
	{
		Definition: llm.Tool{
			Name:        "student.review_enrollment",
			Description: "Approve or reject a student enrollment request.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"enrollment_id":   {"type": "string", "description": "UUID of the enrollment to review"},
					"status":          {"type": "string", "description": "Decision: approved | rejected"},
					"reviewer_notes":  {"type": "string", "description": "Optional notes explaining the decision"}
				},
				"required": ["enrollment_id", "status"]
			}`),
		},
		ModuleName: "student",
		MethodName: "review_enrollment",
	},
}
