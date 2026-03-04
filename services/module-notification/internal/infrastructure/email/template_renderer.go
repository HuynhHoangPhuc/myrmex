package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strings"
	"time"
)

//go:embed compiled/*.html
var compiledTemplatesFS embed.FS

// subjectByType returns a default email subject for a notification type.
var subjectByType = map[string]string{
	"schedule_published":  "New Schedule Published",
	"enrollment_approved": "Enrollment Approved",
	"enrollment_rejected": "Enrollment Not Approved",
	"grade_posted":        "New Grade Posted",
	"grade_updated":       "Grade Updated",
}

// templateByType maps notification types to compiled HTML template files.
// Types not listed fall back to "generic".
var templateByType = map[string]string{
	"schedule_published":  "schedule-published",
	"enrollment_approved": "enrollment-approved",
	"enrollment_rejected": "enrollment-rejected",
	"grade_posted":        "grade-posted",
}

// TemplateData holds variables substituted into email HTML templates.
type TemplateData struct {
	Title     string
	Body      string
	ActionURL string
	Year      int
}

// TemplateRenderer parses compiled MJML HTML and executes them with data.
type TemplateRenderer struct {
	templates map[string]*template.Template // key = template name (without .html)
}

// NewTemplateRenderer loads and parses all compiled HTML templates at startup.
func NewTemplateRenderer() (*TemplateRenderer, error) {
	entries, err := compiledTemplatesFS.ReadDir("compiled")
	if err != nil {
		return nil, fmt.Errorf("read compiled templates dir: %w", err)
	}

	r := &TemplateRenderer{templates: make(map[string]*template.Template)}
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".html") {
			continue
		}
		data, err := compiledTemplatesFS.ReadFile("compiled/" + e.Name())
		if err != nil {
			return nil, fmt.Errorf("read template %s: %w", e.Name(), err)
		}
		name := strings.TrimSuffix(e.Name(), ".html")
		tmpl, err := template.New(name).Parse(string(data))
		if err != nil {
			return nil, fmt.Errorf("parse template %s: %w", e.Name(), err)
		}
		r.templates[name] = tmpl
	}
	return r, nil
}

// Render executes the template for the given notifType. Falls back to "generic".
// Returns (subject, htmlBody, error).
func (r *TemplateRenderer) Render(notifType, title, body, actionURL string) (subject, html string, err error) {
	// Resolve subject
	subject = subjectByType[notifType]
	if subject == "" {
		subject = title
	}

	// Resolve template
	tmplName := templateByType[notifType]
	if tmplName == "" {
		tmplName = "generic"
	}
	tmpl, ok := r.templates[tmplName]
	if !ok {
		tmpl = r.templates["generic"]
	}
	if tmpl == nil {
		return "", "", fmt.Errorf("template %q not found and no generic fallback", tmplName)
	}

	data := TemplateData{
		Title:     title,
		Body:      body,
		ActionURL: actionURL,
		Year:      time.Now().Year(),
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", "", fmt.Errorf("execute template %s: %w", tmplName, err)
	}
	return subject, buf.String(), nil
}
