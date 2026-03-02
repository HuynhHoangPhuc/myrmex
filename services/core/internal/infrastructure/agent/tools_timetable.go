package agent

import "github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/llm"

// TimetableTools defines the tools exposed by the Timetable module.
var TimetableTools = []RegisteredTool{
	{
		Definition: llm.Tool{
			Name:        "timetable.list_semesters",
			Description: "List all semesters with their UUIDs, names, and dates. Call this first to get a valid semester UUID before calling timetable.generate.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {}
			}`),
		},
		ModuleName: "timetable",
		MethodName: "list_semesters",
	},
	{
		Definition: llm.Tool{
			Name:        "timetable.generate",
			Description: "Generate a timetable for a semester using CSP solver. Requires a valid semester UUID — call timetable.list_semesters first to get valid UUIDs.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"semester_id": {"type": "string", "description": "UUID of the semester (get it from timetable.list_semesters)"}
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
			Description: "Suggest available teachers for a given subject.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"subject_id": {"type": "string", "description": "UUID of the subject"}
				},
				"required": ["subject_id"]
			}`),
		},
		ModuleName: "timetable",
		MethodName: "suggest_teachers",
	},
	{
		Definition: llm.Tool{
			Name:        "timetable.get_semester",
			Description: "Get detailed information about a specific semester by UUID.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"semester_id": {"type": "string", "description": "UUID of the semester"}
				},
				"required": ["semester_id"]
			}`),
		},
		ModuleName: "timetable",
		MethodName: "get_semester",
	},
	{
		Definition: llm.Tool{
			Name:        "timetable.list_schedules",
			Description: "List all schedules, optionally filtered by semester.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"semester_id": {"type": "string", "description": "Optional: filter by semester UUID"}
				}
			}`),
		},
		ModuleName: "timetable",
		MethodName: "list_schedules",
	},
	{
		Definition: llm.Tool{
			Name:        "timetable.get_schedule",
			Description: "Get detailed information about a specific schedule by UUID.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"schedule_id": {"type": "string", "description": "UUID of the schedule"}
				},
				"required": ["schedule_id"]
			}`),
		},
		ModuleName: "timetable",
		MethodName: "get_schedule",
	},
	{
		Definition: llm.Tool{
			Name:        "timetable.list_rooms",
			Description: "List all available rooms.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {}
			}`),
		},
		ModuleName: "timetable",
		MethodName: "list_rooms",
	},
	{
		Definition: llm.Tool{
			Name:        "timetable.create_semester",
			Description: "Create a new semester.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"name":       {"type": "string", "description": "Semester name (e.g. Fall 2025)"},
					"start_date": {"type": "string", "description": "Start date in YYYY-MM-DD format"},
					"end_date":   {"type": "string", "description": "End date in YYYY-MM-DD format"}
				},
				"required": ["name", "start_date", "end_date"]
			}`),
		},
		ModuleName: "timetable",
		MethodName: "create_semester",
	},
	{
		Definition: llm.Tool{
			Name:        "timetable.set_semester_rooms",
			Description: "Set the list of rooms available for a semester (replaces existing room assignments).",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"semester_id": {"type": "string", "description": "UUID of the semester"},
					"room_ids":    {"type": "array", "items": {"type": "string"}, "description": "List of room UUIDs to assign"}
				},
				"required": ["semester_id", "room_ids"]
			}`),
		},
		ModuleName: "timetable",
		MethodName: "set_semester_rooms",
	},
	{
		Definition: llm.Tool{
			Name:        "timetable.create_time_slot",
			Description: "Create a new time slot within a semester.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"semester_id":  {"type": "string",  "description": "UUID of the semester"},
					"day_of_week":  {"type": "integer", "description": "Day of week: 0=Sunday … 6=Saturday"},
					"start_time":   {"type": "string",  "description": "Start time in HH:MM format"},
					"end_time":     {"type": "string",  "description": "End time in HH:MM format"}
				},
				"required": ["semester_id", "day_of_week", "start_time", "end_time"]
			}`),
		},
		ModuleName: "timetable",
		MethodName: "create_time_slot",
	},
	{
		Definition: llm.Tool{
			Name:        "timetable.delete_time_slot",
			Description: "Delete a time slot from a semester.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"semester_id": {"type": "string", "description": "UUID of the semester"},
					"slot_id":     {"type": "string", "description": "UUID of the time slot to delete"}
				},
				"required": ["semester_id", "slot_id"]
			}`),
		},
		ModuleName: "timetable",
		MethodName: "delete_time_slot",
	},
	{
		Definition: llm.Tool{
			Name:        "timetable.apply_time_slot_preset",
			Description: "Apply a named preset to create standard time slots for a semester.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"semester_id": {"type": "string", "description": "UUID of the semester"},
					"preset":      {"type": "string", "description": "Preset name (e.g. standard, mwf, tth)"}
				},
				"required": ["semester_id", "preset"]
			}`),
		},
		ModuleName: "timetable",
		MethodName: "apply_time_slot_preset",
	},
	{
		Definition: llm.Tool{
			Name:        "timetable.add_offered_subject",
			Description: "Add a subject offering to a semester, specifying the number of sections.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"semester_id": {"type": "string",  "description": "UUID of the semester"},
					"subject_id":  {"type": "string",  "description": "UUID of the subject to offer"},
					"sections":    {"type": "integer", "description": "Number of sections/classes to schedule"}
				},
				"required": ["semester_id", "subject_id"]
			}`),
		},
		ModuleName: "timetable",
		MethodName: "add_offered_subject",
	},
	{
		Definition: llm.Tool{
			Name:        "timetable.remove_offered_subject",
			Description: "Remove a subject offering from a semester.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"semester_id": {"type": "string", "description": "UUID of the semester"},
					"subject_id":  {"type": "string", "description": "UUID of the subject to remove"}
				},
				"required": ["semester_id", "subject_id"]
			}`),
		},
		ModuleName: "timetable",
		MethodName: "remove_offered_subject",
	},
	{
		Definition: llm.Tool{
			Name:        "timetable.manual_assign",
			Description: "Manually assign a teacher to a specific schedule entry.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"schedule_id": {"type": "string", "description": "UUID of the schedule"},
					"entry_id":    {"type": "string", "description": "UUID of the schedule entry"},
					"teacher_id":  {"type": "string", "description": "UUID of the teacher to assign"}
				},
				"required": ["schedule_id", "entry_id", "teacher_id"]
			}`),
		},
		ModuleName: "timetable",
		MethodName: "manual_assign",
	},
}
