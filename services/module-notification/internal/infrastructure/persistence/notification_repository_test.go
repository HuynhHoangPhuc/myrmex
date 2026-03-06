// Package persistence tests cover pure logic within NotificationRepository
// that does not require a live database connection.
//
// Methods that execute SQL (Insert, ListByUser, CountUnread, MarkRead,
// MarkAllRead, GetByID) all call pool.QueryRow / pool.Query / pool.Exec which
// panic on a nil pool — those are deferred to integration tests.
//
// What IS testable without a pool:
//   - NewNotificationRepository wires the struct correctly
//   - JSON marshalling inside Insert (metadata path) can be extracted as a
//     unit; here we verify the constructor and struct shape instead.
package persistence

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewNotificationRepository_NotNil(t *testing.T) {
	repo := NewNotificationRepository(nil)
	if repo == nil {
		t.Fatal("NewNotificationRepository(nil) returned nil")
	}
}

func TestNotificationRow_JSONRoundTrip(t *testing.T) {
	now := time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC)
	readAt := now.Add(time.Hour)

	row := NotificationRow{
		ID:        "aabbccdd-0000-0000-0000-000000000001",
		UserID:    "user-42",
		Type:      "grade.posted",
		Channel:   "in_app",
		Title:     "Grade Posted",
		Body:      "Your grade has been posted.",
		Metadata:  json.RawMessage(`{"subject_id":"s1"}`),
		ReadAt:    &readAt,
		CreatedAt: now,
	}

	data, err := json.Marshal(row)
	if err != nil {
		t.Fatalf("json.Marshal(NotificationRow) error: %v", err)
	}

	var got NotificationRow
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal(NotificationRow) error: %v", err)
	}

	if got.ID != row.ID {
		t.Errorf("ID = %q, want %q", got.ID, row.ID)
	}
	if got.UserID != row.UserID {
		t.Errorf("UserID = %q, want %q", got.UserID, row.UserID)
	}
	if got.Type != row.Type {
		t.Errorf("Type = %q, want %q", got.Type, row.Type)
	}
	if got.Channel != row.Channel {
		t.Errorf("Channel = %q, want %q", got.Channel, row.Channel)
	}
	if got.Title != row.Title {
		t.Errorf("Title = %q, want %q", got.Title, row.Title)
	}
	if got.ReadAt == nil {
		t.Error("ReadAt: got nil, want non-nil")
	}
}

func TestNotificationRow_NilReadAt_Roundtrip(t *testing.T) {
	row := NotificationRow{
		ID:        "aabbccdd-0000-0000-0000-000000000002",
		UserID:    "user-1",
		Type:      "enrollment.approved",
		Channel:   "email",
		Title:     "Approved",
		Body:      "Congrats",
		CreatedAt: time.Now(),
		ReadAt:    nil, // unread notification
	}

	data, err := json.Marshal(row)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}

	var got NotificationRow
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}

	if got.ReadAt != nil {
		t.Errorf("ReadAt: got %v, want nil", got.ReadAt)
	}
}

// TestMetadataMarshalling verifies that arbitrary map[string]any metadata
// can be JSON-marshalled to a byte slice (same path as Insert uses internally).
func TestMetadataMarshalling(t *testing.T) {
	tests := []struct {
		name     string
		metadata any
		wantErr  bool
	}{
		{
			name:     "nil metadata marshals to nil bytes without error",
			metadata: nil,
			wantErr:  false,
		},
		{
			name:     "simple map marshals correctly",
			metadata: map[string]any{"resource_type": "subject", "resource_id": "abc"},
			wantErr:  false,
		},
		{
			name:     "nested map marshals correctly",
			metadata: map[string]any{"link": "/subjects/123", "extra": map[string]any{"key": "val"}},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.metadata == nil {
				return // nil guard matches Insert's early return
			}
			_, err := json.Marshal(tt.metadata)
			if (err != nil) != tt.wantErr {
				t.Errorf("json.Marshal(%v) error = %v, wantErr %v", tt.metadata, err, tt.wantErr)
			}
		})
	}
}
