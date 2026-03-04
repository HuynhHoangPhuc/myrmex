package notification

import (
	"bytes"
	"fmt"
	"html/template"
)

// emailTemplate wraps a Go html/template for a single notification type.
type emailTemplate struct {
	subjectTpl *template.Template
	bodyTpl    *template.Template
}

// render executes subject + body templates with the given data.
func (t *emailTemplate) render(data any) (subject, htmlBody string, err error) {
	var subBuf, bodyBuf bytes.Buffer
	if err = t.subjectTpl.Execute(&subBuf, data); err != nil {
		return
	}
	if err = t.bodyTpl.Execute(&bodyBuf, data); err != nil {
		return
	}
	return subBuf.String(), bodyBuf.String(), nil
}

// emailTemplates maps notification type → parsed template pair.
var emailTemplates = map[string]*emailTemplate{
	"schedule_published": mustTemplate(
		`Your schedule for {{.SemesterName}} has been published`,
		`<p>Hello,</p><p>Your teaching schedule for <strong>{{.SemesterName}}</strong> has been published.</p>
<p><a href="{{.Link}}">View your schedule</a></p>`,
	),
	"enrollment_approved": mustTemplate(
		`Enrollment approved: {{.SubjectName}}`,
		`<p>Hello,</p><p>Your enrollment request for <strong>{{.SubjectName}}</strong> has been <strong>approved</strong>.</p>
<p><a href="{{.Link}}">View your enrollments</a></p>`,
	),
	"enrollment_rejected": mustTemplate(
		`Enrollment rejected: {{.SubjectName}}`,
		`<p>Hello,</p><p>Your enrollment request for <strong>{{.SubjectName}}</strong> has been <strong>rejected</strong>.</p>
<p><a href="{{.Link}}">View your enrollments</a></p>`,
	),
	"grade_assigned": mustTemplate(
		`Grade posted for {{.SubjectName}}`,
		`<p>Hello,</p><p>A grade of <strong>{{.GradeValue}}</strong> has been posted for <strong>{{.SubjectName}}</strong>.</p>
<p><a href="{{.Link}}">View your transcript</a></p>`,
	),
}

// mustTemplate panics on template parse error (caught at startup, not runtime).
func mustTemplate(subjectText, bodyText string) *emailTemplate {
	return &emailTemplate{
		subjectTpl: template.Must(template.New("s").Parse(subjectText)),
		bodyTpl:    template.Must(template.New("b").Parse(bodyText)),
	}
}

// RenderEmail returns the rendered subject + HTML body for a given notification type and data.
// Returns an error if the type is unknown or rendering fails.
func RenderEmail(notifType string, data any) (subject, htmlBody string, err error) {
	tpl, ok := emailTemplates[notifType]
	if !ok {
		return "", "", fmt.Errorf("no email template for type %q", notifType)
	}
	return tpl.render(data)
}

// SchedulePublishedData holds template vars for schedule_published emails.
type SchedulePublishedData struct {
	SemesterName string
	Link         string
}

// EnrollmentData holds template vars for enrollment_approved/rejected emails.
type EnrollmentData struct {
	SubjectName string
	Link        string
}

// GradeAssignedData holds template vars for grade_assigned emails.
type GradeAssignedData struct {
	SubjectName string
	GradeValue  string
	Link        string
}
