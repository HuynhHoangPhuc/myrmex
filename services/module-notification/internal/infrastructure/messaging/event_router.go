package messaging

import "fmt"

// EventSpec describes how to build a notification from a domain event payload.
type EventSpec struct {
	NotifType  string
	Channels   []string // subset of {"in_app", "email"}
	BuildTitle func(p map[string]any) string
	BuildBody  func(p map[string]any) string
}

// eventSpecs maps NATS subjects to their notification specification.
// Keys match the subjects published by each domain module.
var eventSpecs = map[string]EventSpec{
	"student.enrollment_approved": {
		NotifType: "enrollment.approved",
		Channels:  []string{"in_app", "email"},
		BuildTitle: func(p map[string]any) string { return "Enrollment Approved" },
		BuildBody: func(p map[string]any) string {
			return "Your enrollment request has been approved. You are now officially enrolled. Good luck with your studies!"
		},
	},
	"student.enrollment_rejected": {
		NotifType: "enrollment.rejected",
		Channels:  []string{"in_app", "email"},
		BuildTitle: func(p map[string]any) string { return "Enrollment Not Approved" },
		BuildBody: func(p map[string]any) string {
			note := strFromPayload(p, "AdminNote", "admin_note")
			if note != "" {
				return fmt.Sprintf("Your enrollment request was not approved. Reason: %s", note)
			}
			return "Your enrollment request was not approved. Please contact the academic department for more information."
		},
	},
	"student.enrollment_requested": {
		NotifType: "enrollment.requested",
		Channels:  []string{"in_app"},
		BuildTitle: func(p map[string]any) string { return "New Enrollment Request" },
		BuildBody: func(p map[string]any) string {
			return "A student has submitted an enrollment request that requires your review."
		},
	},
	"student.grade_assigned": {
		NotifType: "grade.posted",
		Channels:  []string{"in_app", "email"},
		BuildTitle: func(p map[string]any) string { return "Grade Posted" },
		BuildBody: func(p map[string]any) string {
			letter := strFromPayload(p, "GradeLetter", "grade_letter")
			if letter != "" {
				return fmt.Sprintf("Your grade has been posted: %s. Log in to view your full transcript.", letter)
			}
			return "Your grade has been posted. Log in to the student portal to view your transcript."
		},
	},
	"hr.teacher.created": {
		NotifType: "teacher.added",
		Channels:  []string{"in_app"},
		BuildTitle: func(p map[string]any) string { return "New Teacher Added" },
		BuildBody: func(p map[string]any) string {
			name := strFromPayload(p, "FullName", "full_name")
			if name != "" {
				return fmt.Sprintf("A new teacher, %s, has been added to your department.", name)
			}
			return "A new teacher has been added to your department."
		},
	},
	"timetable.semester.created": {
		NotifType: "semester.created",
		Channels:  []string{"in_app"},
		BuildTitle: func(p map[string]any) string { return "New Semester Created" },
		BuildBody: func(p map[string]any) string {
			name := strFromPayload(p, "name", "Name")
			if name != "" {
				return fmt.Sprintf("Semester %s has been created. Enrollment and scheduling will begin soon.", name)
			}
			return "A new semester has been created. Enrollment and scheduling will begin soon."
		},
	},
	"timetable.entry.assigned": {
		NotifType: "assignment.changed",
		Channels:  []string{"in_app", "email"},
		BuildTitle: func(p map[string]any) string { return "Teaching Assignment Updated" },
		BuildBody: func(p map[string]any) string {
			return "Your teaching assignment has been updated. Please review the new schedule."
		},
	},
	"timetable.schedule.generated": {
		NotifType: "schedule.published",
		Channels:  []string{"in_app", "email"},
		BuildTitle: func(p map[string]any) string { return "Schedule Published" },
		BuildBody: func(p map[string]any) string {
			return "The teaching schedule has been generated and published. Please log in to review your assigned time slots."
		},
	},
	"core.user.role_updated": {
		NotifType: "role.changed",
		Channels:  []string{"in_app", "email"},
		BuildTitle: func(p map[string]any) string { return "Your Role Has Been Updated" },
		BuildBody: func(p map[string]any) string {
			role := strFromPayload(p, "role")
			if role != "" {
				return fmt.Sprintf("Your account role has been updated to: %s. Please log in again for the changes to take effect.", role)
			}
			return "Your account role has been updated. Please log in again for the changes to take effect."
		},
	},
	"notification.system.announcement": {
		NotifType: "system.announcement",
		Channels:  []string{"in_app", "email"},
		BuildTitle: func(p map[string]any) string {
			return strFromPayload(p, "title")
		},
		BuildBody: func(p map[string]any) string {
			return strFromPayload(p, "body")
		},
	},
}

// GetEventSpec returns the spec for a NATS subject, or false if unknown.
func GetEventSpec(subject string) (EventSpec, bool) {
	spec, ok := eventSpecs[subject]
	return spec, ok
}

// strFromPayload extracts a string value trying each key in order.
func strFromPayload(p map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := p[k]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}
