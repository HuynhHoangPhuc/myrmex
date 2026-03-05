package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"github.com/HuynhHoangPhuc/myrmex/services/module-notification/internal/infrastructure/persistence"
)

// NotificationHandler handles REST endpoints for the notification service.
// User identity is extracted from the X-User-ID header injected by the core reverse proxy.
type NotificationHandler struct {
	repo *persistence.NotificationRepository
	log  *zap.Logger
}

func NewNotificationHandler(repo *persistence.NotificationRepository, log *zap.Logger) *NotificationHandler {
	return &NotificationHandler{repo: repo, log: log}
}

// userIDFromRequest reads the X-User-ID header set by the core reverse proxy.
func userIDFromRequest(r *http.Request) string {
	return r.Header.Get("X-User-ID")
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func parseInt32Query(r *http.Request, key string, fallback int32) int32 {
	if s := r.URL.Query().Get(key); s != "" {
		if v, err := strconv.ParseInt(s, 10, 32); err == nil && v > 0 {
			return int32(v) // safe: ParseInt with bitSize=32 guarantees no overflow
		}
	}
	return fallback
}

// HandleList serves GET /notifications
func (h *NotificationHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "missing user identity")
		return
	}

	page := parseInt32Query(r, "page", 1)
	pageSize := parseInt32Query(r, "page_size", 20)
	if pageSize > 50 {
		pageSize = 50
	}

	rows, total, err := h.repo.ListByUser(r.Context(), userID, page, pageSize)
	if err != nil {
		h.log.Error("list notifications", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to list notifications")
		return
	}
	if rows == nil {
		rows = []persistence.NotificationRow{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data":      rows,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// HandleUnreadCount serves GET /notifications/unread-count
func (h *NotificationHandler) HandleUnreadCount(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "missing user identity")
		return
	}

	count, err := h.repo.CountUnread(r.Context(), userID)
	if err != nil {
		h.log.Error("count unread", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to count unread")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"unread_count": count})
}

// HandleMarkRead serves PATCH /notifications/{id}/read
func (h *NotificationHandler) HandleMarkRead(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "missing user identity")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing notification id")
		return
	}

	if err := h.repo.MarkRead(r.Context(), id, userID); err != nil {
		h.log.Error("mark read", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to mark read")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// HandleMarkAllRead serves POST /notifications/read-all
func (h *NotificationHandler) HandleMarkAllRead(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "missing user identity")
		return
	}

	if err := h.repo.MarkAllRead(r.Context(), userID); err != nil {
		h.log.Error("mark all read", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to mark all read")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
