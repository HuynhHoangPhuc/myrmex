// Package messaging tests cover message parsing paths in the Consumer.
// Happy-path tests that reach repo calls are omitted here because Consumer
// holds *persistence.AnalyticsRepository (concrete type, not an interface),
// making full mock injection impossible without modifying source.
// The paths covered below verify: bad-JSON early-return semantics,
// unknown-subject no-op semantics, and the mustParseUUID helper.
package messaging

import (
	"testing"

	pkgmsg "github.com/HuynhHoangPhuc/myrmex/pkg/messaging"
	"github.com/google/uuid"
)

// newBadMsg returns a Message with deliberately malformed JSON payload.
func newBadMsg(subject string) *pkgmsg.Message {
	return pkgmsg.NewMessage(subject, []byte("{bad json"), nil, nil)
}

// newUnknownMsg returns a Message with an unrecognised subject.
func newUnknownMsg(domain string) *pkgmsg.Message {
	return pkgmsg.NewMessage(domain+".unknown_event", []byte(`{}`), nil, nil)
}

// consumerWithNilRepo builds a Consumer with a nil repo and nop logger.
// Only test paths that return before touching repo are safe to call.
func consumerWithNilRepo(t *testing.T) *Consumer {
	t.Helper()
	return &Consumer{repo: nil, log: nopLogger()}
}

// --- HR bad-JSON / unknown-subject ---

func TestHandleHRMessage_BadJSON_ReturnsNil(t *testing.T) {
	subjects := []string{
		"hr.teacher.created",
		"hr.teacher.updated",
		"hr.teacher.deleted",
		"hr.department.created",
		"hr.department.updated",
	}
	c := consumerWithNilRepo(t)
	for _, subj := range subjects {
		t.Run(subj, func(t *testing.T) {
			if err := c.handleHRMessage(newBadMsg(subj)); err != nil {
				t.Errorf("handleHRMessage(%q) bad JSON: got error %v, want nil", subj, err)
			}
		})
	}
}

func TestHandleHRMessage_UnknownSubject_ReturnsNil(t *testing.T) {
	c := consumerWithNilRepo(t)
	if err := c.handleHRMessage(newUnknownMsg("hr")); err != nil {
		t.Errorf("handleHRMessage() unknown subject: got error %v, want nil", err)
	}
}

// --- Subject bad-JSON / unknown-subject ---

func TestHandleSubjectMessage_BadJSON_ReturnsNil(t *testing.T) {
	subjects := []string{"subject.created", "subject.updated", "subject.deleted"}
	c := consumerWithNilRepo(t)
	for _, subj := range subjects {
		t.Run(subj, func(t *testing.T) {
			if err := c.handleSubjectMessage(newBadMsg(subj)); err != nil {
				t.Errorf("handleSubjectMessage(%q) bad JSON: got error %v, want nil", subj, err)
			}
		})
	}
}

func TestHandleSubjectMessage_UnknownSubject_ReturnsNil(t *testing.T) {
	c := consumerWithNilRepo(t)
	if err := c.handleSubjectMessage(newUnknownMsg("subject")); err != nil {
		t.Errorf("handleSubjectMessage() unknown subject: got error %v, want nil", err)
	}
}

// --- Timetable bad-JSON / unknown-subject ---

func TestHandleTimetableMessage_BadJSON_ReturnsNil(t *testing.T) {
	subjects := []string{"timetable.semester.created", "timetable.schedule.generated"}
	c := consumerWithNilRepo(t)
	for _, subj := range subjects {
		t.Run(subj, func(t *testing.T) {
			if err := c.handleTimetableMessage(newBadMsg(subj)); err != nil {
				t.Errorf("handleTimetableMessage(%q) bad JSON: got error %v, want nil", subj, err)
			}
		})
	}
}

func TestHandleTimetableMessage_UnknownSubject_ReturnsNil(t *testing.T) {
	c := consumerWithNilRepo(t)
	if err := c.handleTimetableMessage(newUnknownMsg("timetable")); err != nil {
		t.Errorf("handleTimetableMessage() unknown subject: got error %v, want nil", err)
	}
}

// --- Student bad-JSON / unknown-subject ---

func TestHandleStudentMessage_BadJSON_ReturnsNil(t *testing.T) {
	subjects := []string{"student.enrollment_approved", "student.grade_assigned"}
	c := consumerWithNilRepo(t)
	for _, subj := range subjects {
		t.Run(subj, func(t *testing.T) {
			if err := c.handleStudentMessage(newBadMsg(subj)); err != nil {
				t.Errorf("handleStudentMessage(%q) bad JSON: got error %v, want nil", subj, err)
			}
		})
	}
}

func TestHandleStudentMessage_UnknownSubject_ReturnsNil(t *testing.T) {
	c := consumerWithNilRepo(t)
	if err := c.handleStudentMessage(newUnknownMsg("student")); err != nil {
		t.Errorf("handleStudentMessage() unknown subject: got error %v, want nil", err)
	}
}

// --- mustParseUUID helper ---

func TestMustParseUUID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  uuid.UUID
	}{
		{
			name:  "valid uuid",
			input: "11111111-1111-1111-1111-111111111111",
			want:  uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		},
		{
			name:  "empty string returns zero uuid",
			input: "",
			want:  uuid.UUID{},
		},
		{
			name:  "invalid string returns zero uuid",
			input: "not-a-uuid",
			want:  uuid.UUID{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mustParseUUID(tt.input)
			if got != tt.want {
				t.Errorf("mustParseUUID(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
