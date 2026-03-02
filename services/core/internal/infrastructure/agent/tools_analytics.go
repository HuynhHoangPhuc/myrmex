package agent

import "github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/llm"

// AnalyticsTools defines the tools exposed by the Analytics module.
var AnalyticsTools = []RegisteredTool{
	{
		Definition: llm.Tool{
			Name:        "analytics.workload",
			Description: "Get teacher workload analytics, optionally filtered by semester.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"semester_id": {"type": "string", "description": "Optional: filter by semester UUID"}
				}
			}`),
		},
		ModuleName: "analytics",
		MethodName: "workload",
	},
	{
		Definition: llm.Tool{
			Name:        "analytics.utilization",
			Description: "Get room utilization analytics, optionally filtered by semester.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"semester_id": {"type": "string", "description": "Optional: filter by semester UUID"}
				}
			}`),
		},
		ModuleName: "analytics",
		MethodName: "utilization",
	},
	{
		Definition: llm.Tool{
			Name:        "analytics.dashboard",
			Description: "Get high-level analytics dashboard metrics.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {}
			}`),
		},
		ModuleName: "analytics",
		MethodName: "dashboard",
	},
	{
		Definition: llm.Tool{
			Name:        "analytics.department_metrics",
			Description: "Get metrics for a department, optionally filtered by department.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"department_id": {"type": "string", "description": "Optional: filter by department UUID"}
				}
			}`),
		},
		ModuleName: "analytics",
		MethodName: "department_metrics",
	},
	{
		Definition: llm.Tool{
			Name:        "analytics.schedule_metrics",
			Description: "Get metrics for a schedule, optionally filtered by schedule.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"schedule_id": {"type": "string", "description": "Optional: filter by schedule UUID"}
				}
			}`),
		},
		ModuleName: "analytics",
		MethodName: "schedule_metrics",
	},
	{
		Definition: llm.Tool{
			Name:        "analytics.schedule_heatmap",
			Description: "Get heatmap data for a schedule showing time slot density, optionally filtered by schedule.",
			Parameters: rawSchema(`{
				"type": "object",
				"properties": {
					"schedule_id": {"type": "string", "description": "Optional: filter by schedule UUID"}
				}
			}`),
		},
		ModuleName: "analytics",
		MethodName: "schedule_heatmap",
	},
}
