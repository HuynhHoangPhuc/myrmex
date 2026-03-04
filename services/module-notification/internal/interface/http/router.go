package http

import (
	"net/http"
)

// NewRouter wires all notification HTTP endpoints onto a new ServeMux.
func NewRouter(notifHandler *NotificationHandler, prefHandler *PreferenceHandler, announceHandler *AnnouncementHandler) *http.ServeMux {
	mux := http.NewServeMux()

	// Health check — no auth required
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Notification REST endpoints — user identity from X-User-ID header (set by core proxy)
	mux.HandleFunc("GET /notifications", notifHandler.HandleList)
	mux.HandleFunc("GET /notifications/unread-count", notifHandler.HandleUnreadCount)
	mux.HandleFunc("POST /notifications/read-all", notifHandler.HandleMarkAllRead)
	mux.HandleFunc("PATCH /notifications/{id}/read", notifHandler.HandleMarkRead)

	// Notification preference endpoints
	mux.HandleFunc("GET /notifications/preferences", prefHandler.HandleGet)
	mux.HandleFunc("PUT /notifications/preferences", prefHandler.HandlePut)

	// System announcement — admin only, checked inside the handler
	mux.HandleFunc("POST /notifications/announce", announceHandler.HandleAnnounce)

	return mux
}
