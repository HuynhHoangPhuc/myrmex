package agent

import (
	"fmt"
	"net/http"
)

// buildSubjectEndpoint maps Subject tool methods to REST paths and HTTP methods.
// Returns (path, httpMethod, bodyArgs, error).
func buildSubjectEndpoint(method string, args map[string]interface{}) (string, string, map[string]interface{}, error) {
	switch method {
	case "list_subjects":
		path := "/api/subjects"
		if params := buildQueryParams(args, "search", "credits"); params != "" {
			path += "?" + params
		}
		return path, http.MethodGet, nil, nil

	case "get_prerequisites":
		id, _ := args["subject_id"].(string)
		return fmt.Sprintf("/api/subjects/%s/prerequisites", id), http.MethodGet, nil, nil

	case "get_subject":
		id, _ := args["subject_id"].(string)
		return fmt.Sprintf("/api/subjects/%s", id), http.MethodGet, nil, nil

	case "validate_dag":
		return "/api/subjects/dag/validate", http.MethodGet, nil, nil

	case "topological_sort":
		return "/api/subjects/dag/topological-sort", http.MethodGet, nil, nil

	case "full_dag":
		return "/api/subjects/dag/full", http.MethodGet, nil, nil

	case "create_subject":
		return "/api/subjects", http.MethodPost, args, nil

	case "update_subject":
		id := stringArg(args, "subject_id")
		return fmt.Sprintf("/api/subjects/%s", id), http.MethodPatch, copyWithout(args, "subject_id"), nil

	case "delete_subject":
		id := stringArg(args, "subject_id")
		return fmt.Sprintf("/api/subjects/%s", id), http.MethodDelete, nil, nil

	case "add_prerequisite":
		id := stringArg(args, "subject_id")
		return fmt.Sprintf("/api/subjects/%s/prerequisites", id), http.MethodPost, copyWithout(args, "subject_id"), nil

	case "remove_prerequisite":
		id := stringArg(args, "subject_id")
		prereqID := stringArg(args, "prerequisite_id")
		return fmt.Sprintf("/api/subjects/%s/prerequisites/%s", id, prereqID), http.MethodDelete, nil, nil

	case "check_conflicts":
		return "/api/subjects/dag/check-conflicts", http.MethodPost, args, nil

	default:
		return "", "", nil, fmt.Errorf("unknown subjects method: %s", method)
	}
}
