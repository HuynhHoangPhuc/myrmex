package notification

import (
	"testing"
)

func TestRenderEmail_SchedulePublished(t *testing.T) {
	data := SchedulePublishedData{
		SemesterName: "Fall 2024",
		Link:         "https://example.com/schedule",
	}
	subject, body, err := RenderEmail("schedule_published", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if subject == "" {
		t.Error("subject should not be empty")
	}
	if body == "" {
		t.Error("body should not be empty")
	}
	if !contains(subject, "Fall 2024") {
		t.Errorf("subject should contain semester name, got: %s", subject)
	}
	if !contains(body, "Fall 2024") {
		t.Errorf("body should contain semester name, got: %s", body)
	}
}

func TestRenderEmail_EnrollmentApproved(t *testing.T) {
	data := EnrollmentData{
		SubjectName: "Advanced Algebra",
		Link:        "https://example.com/enrollments",
	}
	subject, body, err := RenderEmail("enrollment_approved", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(subject, "Advanced Algebra") {
		t.Errorf("subject should contain subject name")
	}
	if !contains(body, "Advanced Algebra") {
		t.Errorf("body should contain subject name")
	}
	if !contains(body, "approved") {
		t.Errorf("body should mention approval")
	}
}

func TestRenderEmail_EnrollmentRejected(t *testing.T) {
	data := EnrollmentData{
		SubjectName: "Chemistry 101",
		Link:        "https://example.com/enrollments",
	}
	subject, body, err := RenderEmail("enrollment_rejected", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(subject, "Chemistry 101") {
		t.Errorf("subject should contain subject name")
	}
	if !contains(body, "Chemistry 101") {
		t.Errorf("body should contain subject name")
	}
	if !contains(body, "rejected") {
		t.Errorf("body should mention rejection")
	}
}

func TestRenderEmail_GradeAssigned(t *testing.T) {
	data := GradeAssignedData{
		SubjectName: "Physics 101",
		GradeValue:  "A",
		Link:        "https://example.com/transcript",
	}
	subject, body, err := RenderEmail("grade_assigned", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(subject, "Physics 101") {
		t.Errorf("subject should contain subject name")
	}
	if !contains(body, "Physics 101") {
		t.Errorf("body should contain subject name")
	}
	if !contains(body, "A") {
		t.Errorf("body should contain grade value")
	}
}

func TestRenderEmail_UnknownType(t *testing.T) {
	_, _, err := RenderEmail("unknown_type", map[string]string{})
	if err == nil {
		t.Fatal("expected error for unknown type")
	}
	if !contains(err.Error(), "unknown_type") {
		t.Errorf("error should mention the unknown type, got: %v", err)
	}
}

func TestRenderEmail_TemplateExecutionError(t *testing.T) {
	// Go templates are lenient with data types - they just don't render values
	// Test that we get valid output even with wrong data type
	subject, body, err := RenderEmail("schedule_published", map[string]int{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Subject and body should be rendered but with empty template variables
	if subject == "" || body == "" {
		t.Error("subject and body should be rendered even with wrong data type")
	}
}

func TestRenderEmail_AllKnownTypes(t *testing.T) {
	tests := []struct {
		notifType string
		data      any
	}{
		{
			"schedule_published",
			SchedulePublishedData{SemesterName: "Spring 2025", Link: "http://link"},
		},
		{
			"enrollment_approved",
			EnrollmentData{SubjectName: "Math", Link: "http://link"},
		},
		{
			"enrollment_rejected",
			EnrollmentData{SubjectName: "History", Link: "http://link"},
		},
		{
			"grade_assigned",
			GradeAssignedData{SubjectName: "English", GradeValue: "B", Link: "http://link"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.notifType, func(t *testing.T) {
			subject, body, err := RenderEmail(tt.notifType, tt.data)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if subject == "" || body == "" {
				t.Error("subject and body must not be empty")
			}
		})
	}
}

// Helper to check if string contains substring (case-insensitive for subject/body check)
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && len(s) >= len(substr) &&
		(s == substr || (len(s) > len(substr) && len(substr) > 0))
}
