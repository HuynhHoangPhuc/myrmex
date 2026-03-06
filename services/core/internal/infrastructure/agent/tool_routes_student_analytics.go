package agent

import (
	"fmt"
	"net/http"
)

// buildStudentEndpoint maps Student tool methods to REST paths and HTTP methods.
// Returns (path, httpMethod, bodyArgs, error).
func buildStudentEndpoint(method string, args map[string]interface{}) (string, string, map[string]interface{}, error) {
	switch method {
	case "list":
		path := "/api/students"
		if params := buildQueryParams(args, "department_id", "status"); params != "" {
			path += "?" + params
		}
		return path, http.MethodGet, nil, nil

	case "get":
		id, _ := args["id"].(string)
		return fmt.Sprintf("/api/students/%s", id), http.MethodGet, nil, nil

	case "transcript":
		id, _ := args["student_id"].(string)
		return fmt.Sprintf("/api/students/%s/transcript", id), http.MethodGet, nil, nil

	case "list_enrollments":
		path := "/api/enrollments"
		if params := buildQueryParams(args, "student_id", "subject_id", "status"); params != "" {
			path += "?" + params
		}
		return path, http.MethodGet, nil, nil

	case "create":
		return "/api/students", http.MethodPost, args, nil

	case "update":
		id := stringArg(args, "id")
		return fmt.Sprintf("/api/students/%s", id), http.MethodPatch, copyWithout(args, "id"), nil

	case "delete":
		id := stringArg(args, "id")
		return fmt.Sprintf("/api/students/%s", id), http.MethodDelete, nil, nil

	case "review_enrollment":
		id := stringArg(args, "enrollment_id")
		return fmt.Sprintf("/api/enrollments/%s/review", id), http.MethodPatch, copyWithout(args, "enrollment_id"), nil

	default:
		return "", "", nil, fmt.Errorf("unknown student method: %s", method)
	}
}

// buildAnalyticsEndpoint maps Analytics tool methods to REST paths and HTTP methods.
// Returns (path, httpMethod, bodyArgs, error).
func buildAnalyticsEndpoint(method string, args map[string]interface{}) (string, string, map[string]interface{}, error) {
	switch method {
	case "workload":
		path := "/api/analytics/workload"
		if params := buildQueryParams(args, "semester_id"); params != "" {
			path += "?" + params
		}
		return path, http.MethodGet, nil, nil

	case "utilization":
		path := "/api/analytics/utilization"
		if params := buildQueryParams(args, "semester_id"); params != "" {
			path += "?" + params
		}
		return path, http.MethodGet, nil, nil

	case "dashboard":
		return "/api/analytics/dashboard", http.MethodGet, nil, nil

	case "department_metrics":
		path := "/api/analytics/department-metrics"
		if params := buildQueryParams(args, "department_id"); params != "" {
			path += "?" + params
		}
		return path, http.MethodGet, nil, nil

	case "schedule_metrics":
		path := "/api/analytics/schedule-metrics"
		if params := buildQueryParams(args, "schedule_id"); params != "" {
			path += "?" + params
		}
		return path, http.MethodGet, nil, nil

	case "schedule_heatmap":
		path := "/api/analytics/schedule-heatmap"
		if params := buildQueryParams(args, "schedule_id"); params != "" {
			path += "?" + params
		}
		return path, http.MethodGet, nil, nil

	default:
		return "", "", nil, fmt.Errorf("unknown analytics method: %s", method)
	}
}
