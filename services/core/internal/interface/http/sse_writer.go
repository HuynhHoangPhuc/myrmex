package http

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// SSEWriter wraps Gin's response writer for Server-Sent Events.
type SSEWriter struct {
	c *gin.Context
}

// NewSSEWriter sets SSE headers and returns a writer.
func NewSSEWriter(c *gin.Context) *SSEWriter {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	return &SSEWriter{c: c}
}

// SendEvent writes a named SSE event with data payload.
func (w *SSEWriter) SendEvent(eventType string, data []byte) {
	fmt.Fprintf(w.c.Writer, "event: %s\ndata: %s\n\n", eventType, data)
	w.c.Writer.Flush()
}

// SendHeartbeat writes an SSE comment to keep the connection alive.
func (w *SSEWriter) SendHeartbeat() {
	fmt.Fprint(w.c.Writer, ": heartbeat\n\n")
	w.c.Writer.Flush()
}
