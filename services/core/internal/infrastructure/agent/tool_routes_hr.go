package agent

import (
	"fmt"
	"net/http"
)

// buildHREndpoint maps HR tool methods to REST paths and HTTP methods.
// Returns (path, httpMethod, bodyArgs, error).
func buildHREndpoint(method string, args map[string]interface{}) (string, string, map[string]interface{}, error) {
	switch method {
	case "list_teachers":
		path := "/api/hr/teachers"
		if params := buildQueryParams(args, "search", "department_id", "specialization"); params != "" {
			path += "?" + params
		}
		return path, http.MethodGet, nil, nil

	case "get_teacher":
		id, _ := args["teacher_id"].(string)
		return fmt.Sprintf("/api/hr/teachers/%s", id), http.MethodGet, nil, nil

	case "list_departments":
		return "/api/hr/departments", http.MethodGet, nil, nil

	case "get_teacher_availability":
		id, _ := args["teacher_id"].(string)
		return fmt.Sprintf("/api/hr/teachers/%s/availability", id), http.MethodGet, nil, nil

	case "create_teacher":
		return "/api/hr/teachers", http.MethodPost, args, nil

	case "update_teacher":
		id := stringArg(args, "teacher_id")
		return fmt.Sprintf("/api/hr/teachers/%s", id), http.MethodPatch, copyWithout(args, "teacher_id"), nil

	case "delete_teacher":
		id := stringArg(args, "teacher_id")
		return fmt.Sprintf("/api/hr/teachers/%s", id), http.MethodDelete, nil, nil

	case "update_teacher_availability":
		id := stringArg(args, "teacher_id")
		return fmt.Sprintf("/api/hr/teachers/%s/availability", id), http.MethodPut, copyWithout(args, "teacher_id"), nil

	case "create_department":
		return "/api/hr/departments", http.MethodPost, args, nil

	default:
		return "", "", nil, fmt.Errorf("unknown hr method: %s", method)
	}
}
