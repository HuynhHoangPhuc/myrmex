package agent

import (
	"fmt"
	"net/http"
)

// buildTimetableEndpoint maps Timetable tool methods to REST paths and HTTP methods.
// Returns (path, httpMethod, bodyArgs, error).
func buildTimetableEndpoint(method string, args map[string]interface{}) (string, string, map[string]interface{}, error) {
	switch method {
	case "list_semesters":
		return "/api/timetable/semesters?page=1&page_size=50", http.MethodGet, nil, nil

	case "generate":
		id, _ := args["semester_id"].(string)
		return fmt.Sprintf("/api/timetable/semesters/%s/generate", id), http.MethodPost, nil, nil

	case "suggest_teachers":
		path := "/api/timetable/suggest-teachers"
		if params := buildQueryParams(args, "subject_id"); params != "" {
			path += "?" + params
		}
		return path, http.MethodGet, nil, nil

	case "get_semester":
		id, _ := args["semester_id"].(string)
		return fmt.Sprintf("/api/timetable/semesters/%s", id), http.MethodGet, nil, nil

	case "list_schedules":
		path := "/api/timetable/schedules"
		if params := buildQueryParams(args, "semester_id"); params != "" {
			path += "?" + params
		}
		return path, http.MethodGet, nil, nil

	case "get_schedule":
		id, _ := args["schedule_id"].(string)
		return fmt.Sprintf("/api/timetable/schedules/%s", id), http.MethodGet, nil, nil

	case "list_rooms":
		return "/api/timetable/rooms", http.MethodGet, nil, nil

	case "create_semester":
		return "/api/timetable/semesters", http.MethodPost, args, nil

	case "set_semester_rooms":
		id := stringArg(args, "semester_id")
		return fmt.Sprintf("/api/timetable/semesters/%s/rooms", id), http.MethodPut, copyWithout(args, "semester_id"), nil

	case "create_time_slot":
		id := stringArg(args, "semester_id")
		return fmt.Sprintf("/api/timetable/semesters/%s/slots", id), http.MethodPost, copyWithout(args, "semester_id"), nil

	case "delete_time_slot":
		semID := stringArg(args, "semester_id")
		slotID := stringArg(args, "slot_id")
		return fmt.Sprintf("/api/timetable/semesters/%s/slots/%s", semID, slotID), http.MethodDelete, nil, nil

	case "apply_time_slot_preset":
		id := stringArg(args, "semester_id")
		return fmt.Sprintf("/api/timetable/semesters/%s/slots/preset", id), http.MethodPost, copyWithout(args, "semester_id"), nil

	case "add_offered_subject":
		id := stringArg(args, "semester_id")
		return fmt.Sprintf("/api/timetable/semesters/%s/offered-subjects", id), http.MethodPost, copyWithout(args, "semester_id"), nil

	case "remove_offered_subject":
		semID := stringArg(args, "semester_id")
		subjectID := stringArg(args, "subject_id")
		return fmt.Sprintf("/api/timetable/semesters/%s/offered-subjects/%s", semID, subjectID), http.MethodDelete, nil, nil

	case "manual_assign":
		schedID := stringArg(args, "schedule_id")
		entryID := stringArg(args, "entry_id")
		return fmt.Sprintf("/api/timetable/schedules/%s/entries/%s", schedID, entryID), http.MethodPut, copyWithout(args, "schedule_id", "entry_id"), nil

	default:
		return "", "", nil, fmt.Errorf("unknown timetable method: %s", method)
	}
}
