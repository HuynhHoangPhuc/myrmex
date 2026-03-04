package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireDeptScope enforces that dept_head and teacher users have a department_id
// in their JWT claims before accessing scoped resources.
// super_admin, admin, dean, and service roles bypass this check.
// The department_id is already set in context by AuthMiddleware for handler-level filtering.
func RequireDeptScope() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("user_role")
		roleStr, _ := role.(string)

		switch roleStr {
		case "super_admin", "admin", "dean", "service", "manager":
			// Full or read-level access — no dept constraint
			c.Next()
		case "dept_head", "teacher":
			deptID, _ := c.Get("department_id")
			deptIDStr, _ := deptID.(string)
			if deptIDStr == "" {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error": "no department assigned to your account; contact an administrator",
				})
				return
			}
			c.Next()
		default:
			c.Next()
		}
	}
}
