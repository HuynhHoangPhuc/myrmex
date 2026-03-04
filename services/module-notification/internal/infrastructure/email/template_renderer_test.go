package email

import (
	"strings"
	"testing"
)

func TestTemplateRenderer_KnownTypes(t *testing.T) {
	r, err := NewTemplateRenderer()
	if err != nil {
		t.Fatalf("NewTemplateRenderer: %v", err)
	}

	cases := []struct {
		notifType   string
		wantSubject string
	}{
		{"schedule_published", "New Schedule Published"},
		{"enrollment_approved", "Enrollment Approved"},
		{"enrollment_rejected", "Enrollment Not Approved"},
		{"grade_posted", "New Grade Posted"},
	}

	for _, tc := range cases {
		t.Run(tc.notifType, func(t *testing.T) {
			subject, html, err := r.Render(tc.notifType, "Test Title", "Test body.", "")
			if err != nil {
				t.Fatalf("Render(%s): %v", tc.notifType, err)
			}
			if subject != tc.wantSubject {
				t.Errorf("subject: want %q, got %q", tc.wantSubject, subject)
			}
			if !strings.Contains(html, "Test Title") {
				t.Error("rendered HTML should contain title")
			}
			if !strings.Contains(html, "Test body.") {
				t.Error("rendered HTML should contain body")
			}
		})
	}
}

func TestTemplateRenderer_GenericFallback(t *testing.T) {
	r, err := NewTemplateRenderer()
	if err != nil {
		t.Fatalf("NewTemplateRenderer: %v", err)
	}

	subject, html, err := r.Render("unknown_event_type", "Fallback Title", "Fallback body.", "https://example.com")
	if err != nil {
		t.Fatalf("Render fallback: %v", err)
	}
	if subject != "Fallback Title" {
		t.Errorf("expected title as subject fallback, got %q", subject)
	}
	if !strings.Contains(html, "Fallback Title") {
		t.Error("rendered HTML should contain title")
	}
	if !strings.Contains(html, "https://example.com") {
		t.Error("rendered HTML should contain action URL")
	}
}

func TestTemplateRenderer_NoActionURL(t *testing.T) {
	r, err := NewTemplateRenderer()
	if err != nil {
		t.Fatalf("NewTemplateRenderer: %v", err)
	}

	// Should render without error even when ActionURL is empty
	_, html, err := r.Render("generic", "Title", "Body", "")
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(html, "Title") {
		t.Error("rendered HTML should contain title")
	}
}
