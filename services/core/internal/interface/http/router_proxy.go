package http

import (
	"fmt"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// healthHandler checks database and Redis connectivity and returns aggregate health status.
func healthHandler(db *pgxpool.Pool, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		checks := gin.H{}
		allOk := true

		if db != nil {
			if err := db.Ping(c.Request.Context()); err != nil {
				checks["database"] = "error: " + err.Error()
				allOk = false
			} else {
				checks["database"] = "ok"
			}
		} else {
			checks["database"] = "unconfigured"
		}

		if rdb != nil {
			if err := rdb.Ping(c.Request.Context()).Err(); err != nil {
				checks["redis"] = "error: " + err.Error()
				allOk = false
			} else {
				checks["redis"] = "ok"
			}
		} else {
			checks["redis"] = "unconfigured"
		}

		status := "ok"
		code := 200
		if !allOk {
			status = "degraded"
			code = 503
		}
		c.JSON(code, gin.H{"status": status, "checks": checks})
	}
}

// newNotificationProxy creates a gin handler that reverse-proxies to module-notification.
// It injects X-User-ID and X-User-Role headers from the validated JWT claims in the gin context.
func newNotificationProxy(targetAddr string) gin.HandlerFunc {
	target, err := url.Parse(targetAddr)
	if err != nil {
		panic(fmt.Sprintf("invalid notification target URL %q: %v", targetAddr, err))
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	return func(c *gin.Context) {
		c.Request.Header.Set("X-User-ID", c.GetString("user_id"))
		c.Request.Header.Set("X-User-Role", c.GetString("user_role"))
		c.Request.Header.Set("X-Department-ID", c.GetString("department_id"))
		// Rewrite path: /api/notifications/... → /notifications/...
		c.Request.URL.Path = "/notifications" + c.Param("path")
		c.Request.Host = target.Host
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

// newAnalyticsProxy creates a gin handler that reverse-proxies to the analytics HTTP service.
func newAnalyticsProxy(targetAddr string) gin.HandlerFunc {
	target, err := url.Parse(targetAddr)
	if err != nil {
		panic(fmt.Sprintf("invalid analytics target URL %q: %v", targetAddr, err))
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	return func(c *gin.Context) {
		// Rewrite path: /api/analytics/workload → /workload
		c.Request.URL.Path = c.Request.URL.Path[len("/api/analytics"):]
		if c.Request.URL.Path == "" {
			c.Request.URL.Path = "/"
		}
		c.Request.Host = target.Host
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
