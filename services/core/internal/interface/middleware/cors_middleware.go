package middleware

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware handles Cross-Origin Resource Sharing.
// In production, set CORS_ALLOWED_ORIGINS to a comma-separated list of allowed origins
// (e.g. "https://myrmex-frontend-xxxxx-run.app"). Falls back to wildcard for local dev.
func CORSMiddleware() gin.HandlerFunc {
	allowedOrigins := parseAllowedOrigins(os.Getenv("CORS_ALLOWED_ORIGINS"))

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if len(allowedOrigins) == 0 {
			// Dev mode: allow all origins
			c.Header("Access-Control-Allow-Origin", "*")
		} else if isAllowed(origin, allowedOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func parseAllowedOrigins(env string) []string {
	if env == "" {
		return nil
	}
	var origins []string
	for _, o := range strings.Split(env, ",") {
		if s := strings.TrimSpace(o); s != "" {
			origins = append(origins, s)
		}
	}
	return origins
}

func isAllowed(origin string, allowed []string) bool {
	for _, a := range allowed {
		if a == origin {
			return true
		}
	}
	return false
}
