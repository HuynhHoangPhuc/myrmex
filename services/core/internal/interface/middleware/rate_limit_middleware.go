package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimitMiddleware provides a simple in-memory token bucket rate limiter per IP.
func RateLimitMiddleware(maxRequests int, window time.Duration) gin.HandlerFunc {
	type entry struct {
		count   int
		resetAt time.Time
	}
	var mu sync.Mutex
	clients := make(map[string]*entry)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		mu.Lock()
		e, ok := clients[ip]
		if !ok || now.After(e.resetAt) {
			clients[ip] = &entry{count: 1, resetAt: now.Add(window)}
			mu.Unlock()
			c.Next()
			return
		}
		e.count++
		if e.count > maxRequests {
			mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		mu.Unlock()
		c.Next()
	}
}
