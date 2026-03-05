package nats

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInferStream tests the inferStream helper function.
func TestInferStream(t *testing.T) {
	tests := []struct {
		subject       string
		expectedName  string
		shouldSucceed bool
	}{
		// HR events
		{"hr.teacher.created", "HR_EVENTS", true},
		{"hr.teacher.updated", "HR_EVENTS", true},
		{"hr.department.added", "HR_EVENTS", true},

		// Subject events
		{"subject.course.created", "SUBJECT_EVENTS", true},
		{"subject.prerequisite.updated", "SUBJECT_EVENTS", true},

		// Student events
		{"student.enrollment_approved", "STUDENT_EVENTS", true},
		{"student.enrollment_rejected", "STUDENT_EVENTS", true},

		// Timetable events
		{"timetable.schedule.generated", "TIMETABLE", true},
		{"timetable.slot.updated", "TIMETABLE", true},

		// Audit events
		{"AUDIT.logs.created", "AUDIT_EVENTS", true},
		{"AUDIT.user.modified", "AUDIT_EVENTS", true},

		// Core events
		{"core.system.started", "CORE_EVENTS", true},
		{"core.migration.completed", "CORE_EVENTS", true},

		// Notification events
		{"notification.email.sent", "NOTIFICATION_EVENTS", true},
		{"notification.push.delivered", "NOTIFICATION_EVENTS", true},

		// Unknown subjects
		{"unknown.subject", "", false},
		{"invalid", "", false},
		{"", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.subject, func(t *testing.T) {
			stream, ok := inferStream(tt.subject)
			if tt.shouldSucceed {
				assert.True(t, ok, "expected inferStream to succeed for subject %q", tt.subject)
				assert.Equal(t, tt.expectedName, stream, "expected stream name %q, got %q", tt.expectedName, stream)
			} else {
				assert.False(t, ok, "expected inferStream to fail for subject %q", tt.subject)
				assert.Empty(t, stream)
			}
		})
	}
}

// TestInferStream_Prefix tests that inferStream matches by prefix.
func TestInferStream_Prefix(t *testing.T) {
	// These should all match "hr." prefix
	subjects := []string{
		"hr.",
		"hr.a",
		"hr.teacher.created",
		"hr.very.long.subject.chain",
	}

	for _, subject := range subjects {
		stream, ok := inferStream(subject)
		assert.True(t, ok, "expected match for %q", subject)
		assert.Equal(t, "HR_EVENTS", stream, "expected HR_EVENTS for %q", subject)
	}
}

// TestStreamWildcard tests the streamWildcard helper function.
func TestStreamWildcard(t *testing.T) {
	tests := []struct {
		stream         string
		expectedWildcard string
	}{
		{"HR_EVENTS", "hr.>"},
		{"SUBJECT_EVENTS", "subject.>"},
		{"STUDENT_EVENTS", "student.>"},
		{"TIMETABLE", "timetable.>"},
		{"AUDIT_EVENTS", "AUDIT.>"},
		{"CORE_EVENTS", "core.>"},
		{"NOTIFICATION_EVENTS", "notification.>"},
		// Unknown streams get default treatment
		{"UNKNOWN_EVENTS", "UNKNOWN_EVENTS.>"},
		{"CUSTOM_STREAM", "CUSTOM_STREAM.>"},
	}

	for _, tt := range tests {
		t.Run(tt.stream, func(t *testing.T) {
			wildcard := streamWildcard(tt.stream)
			assert.Equal(t, tt.expectedWildcard, wildcard, "expected wildcard %q, got %q", tt.expectedWildcard, wildcard)
		})
	}
}

// TestStreamWildcard_Unknown tests default wildcard format for unknown streams.
func TestStreamWildcard_Unknown(t *testing.T) {
	unknownStreams := []string{
		"UNKNOWN",
		"MY_CUSTOM_STREAM",
		"anything-else",
	}

	for _, stream := range unknownStreams {
		wildcard := streamWildcard(stream)
		assert.Equal(t, stream+".>", wildcard, "expected default format for %q", stream)
	}
}

// TestInferStreamRoundTrip tests that inferStream and streamWildcard are consistent.
func TestInferStreamRoundTrip(t *testing.T) {
	// For known subjects, the wildcard should match the subject
	subjects := []string{
		"hr.teacher.created",
		"subject.course.created",
		"student.enrollment_approved",
		"timetable.schedule.generated",
		"AUDIT.logs.created",
		"core.system.started",
		"notification.email.sent",
	}

	for _, subject := range subjects {
		t.Run(subject, func(t *testing.T) {
			// Get stream from subject
			stream, ok := inferStream(subject)
			require.True(t, ok)

			// Get wildcard from stream
			wildcard := streamWildcard(stream)

			// The subject should match the wildcard pattern
			// For NATS, ">" matches one or more levels
			assert.True(t, matchesWildcard(subject, wildcard),
				"subject %q should match wildcard %q", subject, wildcard)
		})
	}
}

// matchesWildcard is a helper to test if a subject matches a NATS wildcard pattern.
func matchesWildcard(subject, pattern string) bool {
	// Simple wildcard matcher for ">" (matches one or more levels)
	// In NATS: "hr.>" matches "hr.a", "hr.a.b", etc.
	if len(pattern) < 2 {
		return false
	}
	prefix := pattern[:len(pattern)-1] // Remove ">"
	return len(subject) > len(prefix) && subject[:len(prefix)] == prefix
}

// TestInferStream_CaseSensitivity tests that inferStream is case-sensitive where appropriate.
func TestInferStream_CaseSensitivity(t *testing.T) {
	// AUDIT uses uppercase
	auditStream, ok := inferStream("AUDIT.logs.created")
	assert.True(t, ok)
	assert.Equal(t, "AUDIT_EVENTS", auditStream)

	// But "audit.logs" (lowercase) should not match
	auditStream2, ok := inferStream("audit.logs.created")
	assert.False(t, ok, "lowercase 'audit' should not match AUDIT prefix")
	assert.Empty(t, auditStream2)
}

// TestStreamWildcard_Consistency tests that wildcard for a stream is consistent.
func TestStreamWildcard_Consistency(t *testing.T) {
	// Calling streamWildcard multiple times should return same result
	stream := "HR_EVENTS"

	wildcard1 := streamWildcard(stream)
	wildcard2 := streamWildcard(stream)

	assert.Equal(t, wildcard1, wildcard2)
	assert.Equal(t, "hr.>", wildcard1)
}

// BenchmarkInferStream benchmarks the inferStream function.
func BenchmarkInferStream(b *testing.B) {
	subject := "hr.teacher.created"

	for i := 0; i < b.N; i++ {
		inferStream(subject)
	}
}

// BenchmarkStreamWildcard benchmarks the streamWildcard function.
func BenchmarkStreamWildcard(b *testing.B) {
	stream := "HR_EVENTS"

	for i := 0; i < b.N; i++ {
		streamWildcard(stream)
	}
}

// TestNewPublisherFromJS tests NewPublisherFromJS constructor.
func TestNewPublisherFromJS(t *testing.T) {
	// This is a simple nil-safety test since we can't easily mock JetStream
	pub := NewPublisherFromJS(nil, nil)
	assert.NotNil(t, pub)
	assert.Nil(t, pub.conn)
	assert.Nil(t, pub.js)
}

// TestNewConsumerFromJS tests NewConsumerFromJS constructor.
func TestNewConsumerFromJS(t *testing.T) {
	// This is a simple nil-safety test since we can't easily mock JetStream
	con := NewConsumerFromJS(nil, nil)
	assert.NotNil(t, con)
	assert.Nil(t, con.conn)
	assert.Nil(t, con.js)
	// cancels slice is initialized as nil by default (Go zero value)
	assert.Nil(t, con.cancels)
}

// TestPublisher_NilJS tests Publisher methods with nil JetStream (error handling).
func TestPublisher_NilJS(t *testing.T) {
	pub := NewPublisherFromJS(nil, nil)

	// These should panic or return error because js is nil
	// We test the nil-safety of the constructor, not the actual behavior
	assert.NotNil(t, pub)
}

// TestConsumer_EmptyCancels tests Consumer with empty cancels slice.
func TestConsumer_EmptyCancels(t *testing.T) {
	con := NewConsumerFromJS(nil, nil)
	assert.NotNil(t, con)
	assert.Empty(t, con.cancels)

	// Close should be safe to call with empty cancels
	err := con.Close()
	assert.NoError(t, err)
}
