package entity

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	EnrollmentStatusPending   = "pending"
	EnrollmentStatusApproved  = "approved"
	EnrollmentStatusRejected  = "rejected"
	EnrollmentStatusCompleted = "completed"
)

// EnrollmentRequest tracks a student's request to enroll in an offered subject.
type EnrollmentRequest struct {
	ID               uuid.UUID
	StudentID        uuid.UUID
	SemesterID       uuid.UUID
	OfferedSubjectID uuid.UUID
	SubjectID        uuid.UUID
	Status           string
	RequestNote      string
	AdminNote        string
	RequestedAt      time.Time
	ReviewedAt       *time.Time
	ReviewedBy       *uuid.UUID
}

func (e *EnrollmentRequest) Validate() error {
	if e.StudentID == uuid.Nil {
		return fmt.Errorf("student_id is required")
	}
	if e.SemesterID == uuid.Nil {
		return fmt.Errorf("semester_id is required")
	}
	if e.OfferedSubjectID == uuid.Nil {
		return fmt.Errorf("offered_subject_id is required")
	}
	if e.SubjectID == uuid.Nil {
		return fmt.Errorf("subject_id is required")
	}
	if e.Status == "" {
		e.Status = EnrollmentStatusPending
	}
	switch e.Status {
	case EnrollmentStatusPending, EnrollmentStatusApproved, EnrollmentStatusRejected, EnrollmentStatusCompleted:
		return nil
	default:
		return fmt.Errorf("invalid enrollment status")
	}
}

func NormalizeNote(value string) string {
	return strings.TrimSpace(value)
}
