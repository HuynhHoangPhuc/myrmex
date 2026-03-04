package http

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/HuynhHoangPhuc/myrmex/services/module-notification/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-notification/internal/infrastructure/persistence"
)

// PreferenceHandler handles REST endpoints for notification preferences.
type PreferenceHandler struct {
	repo *persistence.PreferenceRepository
	log  *zap.Logger
}

func NewPreferenceHandler(repo *persistence.PreferenceRepository, log *zap.Logger) *PreferenceHandler {
	return &PreferenceHandler{repo: repo, log: log}
}

// HandleGet serves GET /notifications/preferences
// Returns full 24-item matrix, merging DB rows with defaults (missing = enabled).
func (h *PreferenceHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "missing user identity")
		return
	}

	stored, err := h.repo.GetByUser(r.Context(), userID)
	if err != nil {
		h.log.Error("get preferences", zap.String("user_id", userID), zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to get preferences")
		return
	}

	// Build lookup from stored rows
	lookup := make(map[string]bool, len(stored))
	for _, p := range stored {
		lookup[p.EventType+":"+p.Channel] = p.Enabled
	}

	// Merge with full default matrix
	defaults := entity.DefaultPreferences()
	result := make([]map[string]any, 0, len(defaults))
	for _, d := range defaults {
		enabled := true
		if v, ok := lookup[d.EventType+":"+d.Channel]; ok {
			enabled = v
		}
		result = append(result, map[string]any{
			"event_type": d.EventType,
			"channel":    d.Channel,
			"enabled":    enabled,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"preferences": result})
}

type prefItem struct {
	EventType string `json:"event_type"`
	Channel   string `json:"channel"`
	Enabled   bool   `json:"enabled"`
}

// HandlePut serves PUT /notifications/preferences
// Accepts a partial list; only the provided pairs are upserted.
func (h *PreferenceHandler) HandlePut(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "missing user identity")
		return
	}

	var body struct {
		Preferences []prefItem `json:"preferences"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(body.Preferences) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	prefs := make([]entity.Preference, 0, len(body.Preferences))
	for _, p := range body.Preferences {
		prefs = append(prefs, entity.Preference{
			EventType: p.EventType,
			Channel:   p.Channel,
			Enabled:   p.Enabled,
		})
	}

	if err := h.repo.BulkUpsert(r.Context(), userID, prefs); err != nil {
		h.log.Error("upsert preferences", zap.String("user_id", userID), zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to update preferences")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
