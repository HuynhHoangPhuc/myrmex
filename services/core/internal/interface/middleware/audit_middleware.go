package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
)

const (
	// Context keys handlers use to enrich audit entries
	AuditKeyOldValue     = "audit_old"
	AuditKeyNewValue     = "audit_new"
	AuditKeyAction       = "audit_action"
	AuditKeyResourceType = "audit_resource_type"
	AuditKeyResourceID   = "audit_resource_id"
)

// AuditEvent is published to NATS after each mutation.
type AuditEvent struct {
	UserID       string    `json:"user_id"`
	UserRole     string    `json:"user_role"`
	Action       string    `json:"action"`
	ResourceType string    `json:"resource_type"`
	ResourceID   string    `json:"resource_id,omitempty"`
	OldValue     any       `json:"old_value,omitempty"`
	NewValue     any       `json:"new_value,omitempty"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	StatusCode   int       `json:"status_code"`
	CreatedAt    time.Time `json:"created_at"`
}

// AuditMiddleware publishes an AuditEvent to NATS after every state-mutating request.
// Skipped when nc is nil (NATS not configured) or for read-only methods.
func AuditMiddleware(nc *nats.Conn) gin.HandlerFunc {
	if nc == nil {
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		// Only audit mutations
		method := c.Request.Method
		if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
			c.Next()
			return
		}

		c.Next() // execute handler first

		// Extract JWT context set by AuthMiddleware
		userID, _ := c.Get("user_id")
		userRole, _ := c.Get("role")

		userIDStr, _ := userID.(string)
		if userIDStr == "" || userIDStr == "internal-service" {
			return // skip unauthenticated or internal calls
		}

		// Derive action + resource from context (handler can override)
		action, _ := c.Get(AuditKeyAction)
		resourceType, _ := c.Get(AuditKeyResourceType)
		resourceID, _ := c.Get(AuditKeyResourceID)
		oldVal, _ := c.Get(AuditKeyOldValue)
		newVal, _ := c.Get(AuditKeyNewValue)

		actionStr, _ := action.(string)
		if actionStr == "" {
			actionStr = deriveAction(method, c.FullPath())
		}
		resourceTypeStr, _ := resourceType.(string)
		if resourceTypeStr == "" {
			resourceTypeStr = deriveResourceType(c.FullPath())
		}
		resourceIDStr, _ := resourceID.(string)
		userRoleStr, _ := userRole.(string)

		event := AuditEvent{
			UserID:       userIDStr,
			UserRole:     userRoleStr,
			Action:       actionStr,
			ResourceType: resourceTypeStr,
			ResourceID:   resourceIDStr,
			OldValue:     oldVal,
			NewValue:     newVal,
			IPAddress:    c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			StatusCode:   c.Writer.Status(),
			CreatedAt:    time.Now().UTC(),
		}

		data, err := json.Marshal(event)
		if err != nil {
			return
		}

		// Fire-and-forget: never block the response
		_ = nc.Publish("AUDIT.logs", data)
	}
}

// deriveAction builds "resource.verb" from HTTP method + path.
// e.g. POST /api/hr/teachers → "hr.teacher.create"
func deriveAction(method, path string) string {
	resource := deriveResourceType(path)
	var verb string
	switch method {
	case http.MethodPost:
		verb = "create"
	case http.MethodPut, http.MethodPatch:
		verb = "update"
	case http.MethodDelete:
		verb = "delete"
	default:
		verb = strings.ToLower(method)
	}
	// Include module prefix for HR/subject/student paths
	module := deriveModule(path)
	if module != "" {
		return module + "." + resource + "." + verb
	}
	return resource + "." + verb
}

func deriveModule(path string) string {
	switch {
	case strings.Contains(path, "/hr/"):
		return "hr"
	case strings.Contains(path, "/subjects"):
		return "subject"
	case strings.Contains(path, "/students"), strings.Contains(path, "/enrollments"), strings.Contains(path, "/grades"):
		return "student"
	case strings.Contains(path, "/timetable"):
		return "timetable"
	case strings.Contains(path, "/admin"):
		return "admin"
	default:
		return ""
	}
}

func deriveResourceType(path string) string {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	// Skip "api" prefix
	for i, s := range segments {
		if s == "api" && i+1 < len(segments) {
			segments = segments[i+1:]
			break
		}
	}
	// Find first non-module segment that isn't a path param
	for _, s := range segments {
		if s == "" || s == "api" || s == "hr" || s == "timetable" || s == "admin" || s == "auth" || s == "oauth" {
			continue
		}
		if strings.HasPrefix(s, ":") {
			continue
		}
		// Singularize common plural names
		return singularize(s)
	}
	return "resource"
}

func singularize(s string) string {
	switch s {
	case "teachers":
		return "teacher"
	case "departments":
		return "department"
	case "subjects":
		return "subject"
	case "students":
		return "student"
	case "enrollments":
		return "enrollment"
	case "grades":
		return "grade"
	case "users":
		return "user"
	case "roles":
		return "role"
	case "semesters":
		return "semester"
	case "schedules":
		return "schedule"
	default:
		if strings.HasSuffix(s, "s") {
			return s[:len(s)-1]
		}
		return s
	}
}
