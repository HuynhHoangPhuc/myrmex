package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/persistence"
)

// AuditHandler exposes the admin audit log query API.
type AuditHandler struct {
	repo *persistence.AuditLogRepository
}

func NewAuditHandler(repo *persistence.AuditLogRepository) *AuditHandler {
	return &AuditHandler{repo: repo}
}

// ListAuditLogs handles GET /api/audit-logs
// Query params: user_id, resource_type, action, from (RFC3339), to (RFC3339), page, page_size
func (h *AuditHandler) ListAuditLogs(c *gin.Context) {
	filter := persistence.AuditLogFilter{
		UserID:       c.Query("user_id"),
		ResourceType: c.Query("resource_type"),
		Action:       c.Query("action"),
	}

	if v := c.Query("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.From = t
		}
	}
	if v := c.Query("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.To = t
		}
	}

	pageSize := int32(20)
	page := int32(1)
	if v := c.Query("page_size"); v != "" {
		if ps := parseIntQuery(v, 1, 100); ps > 0 {
			pageSize = int32(ps)
		}
	}
	if v := c.Query("page"); v != "" {
		if p := parseIntQuery(v, 1, 10000); p > 0 {
			page = int32(p)
		}
	}
	filter.PageSize = pageSize
	filter.Page = page

	logs, total, err := h.repo.List(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query audit logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func parseIntQuery(s string, min, max int) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	if n < min || n > max {
		return 0
	}
	return n
}
