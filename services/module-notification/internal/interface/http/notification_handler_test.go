package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// The handler uses *persistence.NotificationRepository (concrete, requires pgxpool).
// Tests here cover only the middleware logic that runs before any DB call:
//   - missing X-User-ID header → 401 Unauthorized
//   - parseInt32Query with various inputs
//
// Full handler happy-path tests require a real or stubbed pgxpool and are
// intentionally deferred to integration tests.

func TestHandleList_MissingUserID_Returns401(t *testing.T) {
	h := &NotificationHandler{repo: nil, log: nopLogger()}

	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	// Deliberately omit X-User-ID header.
	w := httptest.NewRecorder()

	h.HandleList(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("HandleList() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("response body is not JSON: %v", err)
	}
	if body["error"] == "" {
		t.Error("expected non-empty error field in response body")
	}
}

func TestHandleUnreadCount_MissingUserID_Returns401(t *testing.T) {
	h := &NotificationHandler{repo: nil, log: nopLogger()}

	req := httptest.NewRequest(http.MethodGet, "/notifications/unread-count", nil)
	w := httptest.NewRecorder()

	h.HandleUnreadCount(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("HandleUnreadCount() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandleMarkRead_MissingUserID_Returns401(t *testing.T) {
	h := &NotificationHandler{repo: nil, log: nopLogger()}

	req := httptest.NewRequest(http.MethodPatch, "/notifications/some-id/read", nil)
	w := httptest.NewRecorder()

	h.HandleMarkRead(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("HandleMarkRead() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandleMarkAllRead_MissingUserID_Returns401(t *testing.T) {
	h := &NotificationHandler{repo: nil, log: nopLogger()}

	req := httptest.NewRequest(http.MethodPost, "/notifications/read-all", nil)
	w := httptest.NewRecorder()

	h.HandleMarkAllRead(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("HandleMarkAllRead() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// --- parseInt32Query unit tests ---

func TestParseInt32Query(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		key      string
		fallback int32
		want     int32
	}{
		{
			name:     "valid positive integer",
			query:    "?page=3",
			key:      "page",
			fallback: 1,
			want:     3,
		},
		{
			name:     "missing key uses fallback",
			query:    "?other=5",
			key:      "page",
			fallback: 1,
			want:     1,
		},
		{
			name:     "zero value uses fallback",
			query:    "?page=0",
			key:      "page",
			fallback: 1,
			want:     1,
		},
		{
			name:     "negative value uses fallback",
			query:    "?page=-1",
			key:      "page",
			fallback: 1,
			want:     1,
		},
		{
			name:     "non-numeric value uses fallback",
			query:    "?page=abc",
			key:      "page",
			fallback: 1,
			want:     1,
		},
		{
			name:     "page_size=50 is accepted",
			query:    "?page_size=50",
			key:      "page_size",
			fallback: 20,
			want:     50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/"+tt.query, nil)
			got := parseInt32Query(req, tt.key, tt.fallback)
			if got != tt.want {
				t.Errorf("parseInt32Query(%q, %q, %d) = %d, want %d",
					tt.query, tt.key, tt.fallback, got, tt.want)
			}
		})
	}
}

// --- userIDFromRequest unit tests ---

func TestUserIDFromRequest(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{name: "header present", header: "abc-123", want: "abc-123"},
		{name: "header absent", header: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				req.Header.Set("X-User-ID", tt.header)
			}
			got := userIDFromRequest(req)
			if got != tt.want {
				t.Errorf("userIDFromRequest() = %q, want %q", got, tt.want)
			}
		})
	}
}
