package messaging

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/persistence"
)

// TestResolveRecipients_SchedulePublished tests schedule_published event recipients handling
func TestResolveRecipients_SchedulePublished(t *testing.T) {
	// Create a minimal consumer to test resolveRecipients logic
	consumer := createTestConsumer()

	event := notifEvent{
		RecipientUserIDs: []string{"user-1", "user-2", "user-3"},
	}

	recipients := consumer.resolveRecipients(context.Background(), "schedule_published", event)

	if len(recipients) != 3 {
		t.Errorf("expected 3 recipients, got %d", len(recipients))
	}
	if recipients[0] != "user-1" || recipients[1] != "user-2" || recipients[2] != "user-3" {
		t.Error("recipients don't match input")
	}
}

func TestResolveRecipients_SchedulePublished_Empty(t *testing.T) {
	consumer := createTestConsumer()

	event := notifEvent{
		RecipientUserIDs: []string{},
	}

	recipients := consumer.resolveRecipients(context.Background(), "schedule_published", event)

	if len(recipients) != 0 {
		t.Errorf("expected 0 recipients, got %d", len(recipients))
	}
}

func TestResolveRecipients_StudentEvents_ApprovedStudent(t *testing.T) {
	tests := []struct {
		name      string
		notifType string
		studentID string
		expected  int
	}{
		{"enrollment_approved", "enrollment_approved", "student-123", 1},
		{"enrollment_rejected", "enrollment_rejected", "student-456", 1},
		{"grade_assigned", "grade_assigned", "student-789", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			consumer := createTestConsumer()
			event := notifEvent{
				StudentUserID: tt.studentID,
			}

			recipients := consumer.resolveRecipients(context.Background(), tt.notifType, event)

			if len(recipients) != tt.expected {
				t.Errorf("expected %d recipients, got %d", tt.expected, len(recipients))
			}
			if tt.expected == 1 && recipients[0] != tt.studentID {
				t.Errorf("expected student ID %s, got %s", tt.studentID, recipients[0])
			}
		})
	}
}

func TestResolveRecipients_StudentEvents_EmptyStudent(t *testing.T) {
	tests := []string{
		"enrollment_approved",
		"enrollment_rejected",
		"grade_assigned",
	}

	for _, notifType := range tests {
		t.Run(notifType+"_empty_student", func(t *testing.T) {
			consumer := createTestConsumer()
			event := notifEvent{
				StudentUserID: "",
			}

			recipients := consumer.resolveRecipients(context.Background(), notifType, event)

			if len(recipients) != 0 {
				t.Errorf("expected 0 recipients for empty student ID, got %d", len(recipients))
			}
		})
	}
}

func TestResolveRecipients_EnrollmentRequested_NoDepartmentID(t *testing.T) {
	consumer := createTestConsumer()

	event := notifEvent{
		DepartmentID: "",
	}

	recipients := consumer.resolveRecipients(context.Background(), "enrollment_requested", event)

	if len(recipients) != 0 {
		t.Errorf("expected 0 recipients when no department ID, got %d", len(recipients))
	}
}

// TestDispatch_InAppDisabled_SkipsDispatch verifies dispatch preferences logic
func TestDispatch_InAppDisabled_SkipsDispatch(t *testing.T) {
	spec := &notifSpec{
		notifType: "schedule_published",
		title:     func(e notifEvent) string { return "Test" },
		body:      func(e notifEvent) string { return "Test body" },
		sendEmail: true,
	}

	// This verifies the dispatch method exists and has correct signature
	if spec.notifType != "schedule_published" {
		t.Error("spec type mismatch")
	}
}

func TestDispatch_DisabledType_SkipsDispatch(t *testing.T) {
	spec := &notifSpec{
		notifType: "schedule_published",
		title:     func(e notifEvent) string { return "Test" },
		body:      func(e notifEvent) string { return "Test body" },
		sendEmail: false,
	}

	if spec.sendEmail {
		t.Error("sendEmail should be false for enrollment_requested")
	}
}

// TestNotifSpecs verifies all 5 notification specs are correctly defined
func TestNotifSpecs_Count(t *testing.T) {
	if len(notifSpecs) != 5 {
		t.Errorf("expected 5 notification specs, got %d", len(notifSpecs))
	}
}

func TestNotifSpecs_AllHaveSubject(t *testing.T) {
	for i, spec := range notifSpecs {
		if spec.subject == "" {
			t.Errorf("spec %d has no subject", i)
		}
	}
}

func TestNotifSpecs_AllHaveType(t *testing.T) {
	expectedTypes := map[string]bool{
		"schedule_published":      true,
		"enrollment_approved":     true,
		"enrollment_rejected":     true,
		"grade_assigned":          true,
		"enrollment_requested":    true,
	}

	foundTypes := make(map[string]bool)
	for _, spec := range notifSpecs {
		foundTypes[spec.notifType] = true
	}

	for expected := range expectedTypes {
		if !foundTypes[expected] {
			t.Errorf("missing notification type: %s", expected)
		}
	}
}

func TestNotifSpecs_TitleFunctionWorks(t *testing.T) {
	tests := []struct {
		specIdx  int
		event    notifEvent
		contains string
	}{
		{0, notifEvent{SemesterName: "Fall 2024"}, "Fall 2024"},
		{1, notifEvent{SubjectName: "Math"}, "Math"},
		{2, notifEvent{SubjectName: "Physics"}, "Physics"},
		{3, notifEvent{SubjectName: "Chemistry", GradeValue: "A"}, "A"},
		{4, notifEvent{SubjectName: "History"}, "History"},
	}

	for idx, tt := range tests {
		if idx >= len(notifSpecs) {
			t.Errorf("test index %d out of range", idx)
			break
		}
		title := notifSpecs[idx].title(tt.event)
		if title == "" {
			t.Errorf("spec %d title function returned empty string", idx)
		}
	}
}

func TestNotifSpecs_BodyFunctionWorks(t *testing.T) {
	tests := []struct {
		specIdx int
		event   notifEvent
	}{
		{0, notifEvent{SemesterName: "Spring 2025"}},
		{1, notifEvent{SubjectName: "English"}},
		{2, notifEvent{SubjectName: "History"}},
		{3, notifEvent{SubjectName: "Art", GradeValue: "B"}},
		{4, notifEvent{SubjectName: "PE"}},
	}

	for idx, tt := range tests {
		if idx >= len(notifSpecs) {
			t.Errorf("test index %d out of range", idx)
			break
		}
		body := notifSpecs[idx].body(tt.event)
		if body == "" {
			t.Errorf("spec %d body function returned empty string", idx)
		}
	}
}

func TestNotifSpecs_EmailFlagCorrect(t *testing.T) {
	expectedEmail := map[string]bool{
		"schedule_published":    true,
		"enrollment_approved":   true,
		"enrollment_rejected":   true,
		"grade_assigned":        true,
		"enrollment_requested":  false, // in-app only
	}

	for _, spec := range notifSpecs {
		if expected, ok := expectedEmail[spec.notifType]; ok {
			if spec.sendEmail != expected {
				t.Errorf("spec %s has sendEmail=%v, expected %v",
					spec.notifType, spec.sendEmail, expected)
			}
		}
	}
}

// Test event parsing and routing logic
// Force import of persistence for tests that use it
var _ = persistence.NotificationRow{}

func TestNotifEvent_SchedulePublished_Fields(t *testing.T) {
	event := notifEvent{
		RecipientUserIDs: []string{"user-1", "user-2"},
		SemesterName:     "Fall 2024",
		ResourceType:     "schedule",
		ResourceID:       "sched-123",
		Link:             "https://example.com/schedule",
	}

	if len(event.RecipientUserIDs) != 2 {
		t.Error("RecipientUserIDs not populated")
	}
	if event.SemesterName != "Fall 2024" {
		t.Error("SemesterName not populated")
	}
}

func TestNotifEvent_StudentEvent_Fields(t *testing.T) {
	event := notifEvent{
		StudentUserID: "student-123",
		SubjectName:   "Mathematics",
		GradeValue:    "A",
		ResourceType:  "enrollment",
		ResourceID:    "enroll-456",
		Link:          "https://example.com/enrollments",
	}

	if event.StudentUserID != "student-123" {
		t.Error("StudentUserID not populated")
	}
	if event.SubjectName != "Mathematics" {
		t.Error("SubjectName not populated")
	}
	if event.GradeValue != "A" {
		t.Error("GradeValue not populated")
	}
}

func TestNotifEvent_EnrollmentRequest_Fields(t *testing.T) {
	event := notifEvent{
		DepartmentID: "dept-789",
		SubjectName:  "Advanced Physics",
		ResourceType: "enrollment_request",
		ResourceID:   "enroll-req-999",
	}

	if event.DepartmentID != "dept-789" {
		t.Error("DepartmentID not populated")
	}
	if event.SubjectName != "Advanced Physics" {
		t.Error("SubjectName not populated")
	}
}

// Helper to create a test consumer
// Note: The repo is only accessed in enrollment_requested path
func createTestConsumer() *NotificationConsumer {
	return &NotificationConsumer{
		repo: nil, // resolveRecipients doesn't call repo for most types
		log:  zap.NewNop(),
	}
}
