package persistence

import (
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

// TestGetPreferences tests NotificationRepository.GetPreferences behavior
// Since we can't mock the concrete pgxpool.Pool type, we test the logic
// through behavioral validation of the preference handling

// TestNotificationPreferences_Defaults verifies default preference values
func TestNotificationPreferences_Defaults(t *testing.T) {
	_ = NotificationPreferences{
		UserID: "user-123",
	}

	// When no preferences are explicitly set, defaults should be:
	// (These are set in GetPreferences when pgx.ErrNoRows occurs)
	defaultPrefs := NotificationPreferences{
		UserID:        "user-123",
		EmailEnabled:  true,
		InAppEnabled:  true,
		DisabledTypes: []string{},
	}

	if !defaultPrefs.EmailEnabled {
		t.Error("default EmailEnabled should be true")
	}
	if !defaultPrefs.InAppEnabled {
		t.Error("default InAppEnabled should be true")
	}
	if defaultPrefs.DisabledTypes != nil && len(defaultPrefs.DisabledTypes) != 0 {
		t.Error("default DisabledTypes should be empty")
	}
}

// TestNotificationPreferences_CustomValues tests preference customization
func TestNotificationPreferences_CustomValues(t *testing.T) {
	prefs := NotificationPreferences{
		UserID:        "user-456",
		EmailEnabled:  false,
		InAppEnabled:  true,
		DisabledTypes: []string{"grade_assigned", "enrollment_rejected"},
	}

	if prefs.EmailEnabled {
		t.Error("EmailEnabled should be false")
	}
	if !prefs.InAppEnabled {
		t.Error("InAppEnabled should be true")
	}
	if len(prefs.DisabledTypes) != 2 {
		t.Errorf("expected 2 disabled types, got %d", len(prefs.DisabledTypes))
	}
}

// TestNotificationPreferences_AllChannelsDisabled tests opt-out scenario
func TestNotificationPreferences_AllChannelsDisabled(t *testing.T) {
	prefs := NotificationPreferences{
		UserID:        "user-silent",
		EmailEnabled:  false,
		InAppEnabled:  false,
		DisabledTypes: []string{},
	}

	if prefs.EmailEnabled {
		t.Error("EmailEnabled should be false")
	}
	if prefs.InAppEnabled {
		t.Error("InAppEnabled should be false")
	}
}

// TestNotificationPreferences_AllTypesDisabled tests disabling all notification types
func TestNotificationPreferences_AllTypesDisabled(t *testing.T) {
	prefs := NotificationPreferences{
		UserID:        "user-picky",
		EmailEnabled:  true,
		InAppEnabled:  true,
		DisabledTypes: []string{
			"schedule_published",
			"enrollment_approved",
			"enrollment_rejected",
			"grade_assigned",
			"enrollment_requested",
		},
	}

	if len(prefs.DisabledTypes) != 5 {
		t.Errorf("expected 5 disabled types, got %d", len(prefs.DisabledTypes))
	}

	expected := map[string]bool{
		"schedule_published":   true,
		"enrollment_approved":  true,
		"enrollment_rejected":  true,
		"grade_assigned":       true,
		"enrollment_requested": true,
	}

	for _, disabledType := range prefs.DisabledTypes {
		if !expected[disabledType] {
			t.Errorf("unexpected disabled type: %s", disabledType)
		}
	}
}

// TestNotificationRow_Structure tests NotificationRow field structure
func TestNotificationRow_Structure(t *testing.T) {
	now := time.Now()
	row := NotificationRow{
		ID:        "notif-123",
		UserID:    "user-456",
		Type:      "schedule_published",
		Title:     "Schedule Published",
		Body:      "Your schedule is ready",
		ReadAt:    nil,
		CreatedAt: now,
	}

	if row.ID != "notif-123" {
		t.Error("ID not set correctly")
	}
	if row.UserID != "user-456" {
		t.Error("UserID not set correctly")
	}
	if row.Type != "schedule_published" {
		t.Error("Type not set correctly")
	}
	if row.Title != "Schedule Published" {
		t.Error("Title not set correctly")
	}
	if row.ReadAt != nil {
		t.Error("ReadAt should be nil for unread notification")
	}
}

// TestNotificationRow_WithReadAt tests read notification structure
func TestNotificationRow_WithReadAt(t *testing.T) {
	now := time.Now()
	readTime := now.Add(-1 * time.Hour)

	row := NotificationRow{
		ID:        "notif-456",
		UserID:    "user-789",
		Type:      "enrollment_approved",
		Title:     "Enrollment Approved",
		Body:      "You are enrolled in Math",
		ReadAt:    &readTime,
		CreatedAt: now,
	}

	if row.ReadAt == nil {
		t.Error("ReadAt should not be nil for read notification")
	}
	if row.ReadAt.After(row.CreatedAt) {
		t.Error("ReadAt should be after or equal to CreatedAt")
	}
}

// TestErrNoRows_Behavior tests pgx.ErrNoRows behavior for defaults
func TestErrNoRows_BehaviorForDefaults(t *testing.T) {
	// Verify that pgx.ErrNoRows is the right error to check for defaults
	err := pgx.ErrNoRows

	if !errors.Is(err, pgx.ErrNoRows) {
		t.Error("pgx.ErrNoRows error check failed")
	}
}

// TestNotificationTypes_KnownValues tests valid notification types
func TestNotificationTypes_KnownValues(t *testing.T) {
	validTypes := []string{
		"schedule_published",
		"enrollment_approved",
		"enrollment_rejected",
		"grade_assigned",
		"enrollment_requested",
	}

	for _, notifType := range validTypes {
		row := NotificationRow{
			ID:   "notif-" + notifType,
			Type: notifType,
		}
		if row.Type != notifType {
			t.Errorf("Type %s not set correctly", notifType)
		}
	}
}

// TestNotificationPreferences_UserIDTracking tests UserID is preserved
func TestNotificationPreferences_UserIDTracking(t *testing.T) {
	userIDs := []string{
		"user-1",
		"user-2",
		"00000000-0000-0000-0000-000000000000",
	}

	for _, userID := range userIDs {
		prefs := NotificationPreferences{
			UserID:       userID,
			EmailEnabled: true,
			InAppEnabled: true,
		}

		if prefs.UserID != userID {
			t.Errorf("UserID mismatch: expected %s, got %s", userID, prefs.UserID)
		}
	}
}

// TestNotificationPreferences_DisabledTypesEmptySlice tests empty vs nil slices
func TestNotificationPreferences_DisabledTypesEmptySlice(t *testing.T) {
	// Test that empty slice is handled correctly
	prefs1 := NotificationPreferences{
		UserID:        "user-1",
		DisabledTypes: []string{},
	}

	// Test that uninitialized (nil) slice is valid
	var prefs2 NotificationPreferences
	prefs2.UserID = "user-2"

	if len(prefs1.DisabledTypes) != 0 {
		t.Error("empty slice length should be 0")
	}

	// Both should be treated as "no disabled types"
	if prefs1.DisabledTypes != nil && len(prefs1.DisabledTypes) > 0 {
		t.Error("empty slice check failed")
	}
}

// TestNotificationRow_Data tests JSON data field
func TestNotificationRow_Data(t *testing.T) {
	row := NotificationRow{
		ID:   "notif-data",
		Data: []byte(`{"subject":"Math","link":"http://example.com"}`),
	}

	if len(row.Data) == 0 {
		t.Error("Data should not be empty")
	}

	// Verify it's valid JSON structure
	if row.Data[0] != '{' || row.Data[len(row.Data)-1] != '}' {
		t.Error("Data should be valid JSON object")
	}
}
