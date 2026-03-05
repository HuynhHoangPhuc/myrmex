package http

import (
	"context"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/HuynhHoangPhuc/myrmex/pkg/messaging"
)

// AnnouncementHandler publishes system-wide announcement events (admin only).
type AnnouncementHandler struct {
	pub messaging.Publisher
	log *zap.Logger
}

func NewAnnouncementHandler(pub messaging.Publisher, log *zap.Logger) *AnnouncementHandler {
	return &AnnouncementHandler{pub: pub, log: log}
}

type announceRequest struct {
	Title     string `json:"title"`
	Body      string `json:"body"`
	ActionURL string `json:"action_url"`
}

// HandleAnnounce serves POST /notifications/announce (admin/super_admin only).
// Publishes notification.system.announcement → picked up by EventConsumer.
func (h *AnnouncementHandler) HandleAnnounce(w http.ResponseWriter, r *http.Request) {
	role := r.Header.Get("X-User-Role")
	if role != "admin" && role != "super_admin" {
		writeError(w, http.StatusForbidden, "admin access required")
		return
	}

	var req announceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Title == "" || req.Body == "" {
		writeError(w, http.StatusBadRequest, "title and body are required")
		return
	}

	payload, err := json.Marshal(map[string]string{
		"title":      req.Title,
		"body":       req.Body,
		"action_url": req.ActionURL,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to encode payload")
		return
	}

	if err := h.pub.Publish(context.Background(), "notification.system.announcement", payload); err != nil {
		h.log.Error("announce: publish failed", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to publish announcement")
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{"status": "queued"})
}
