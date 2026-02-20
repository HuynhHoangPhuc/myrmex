package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SemesterOffering represents a subject offered in a specific semester.
type SemesterOffering struct {
	ID            uuid.UUID
	SubjectID     uuid.UUID
	SemesterID    uuid.UUID
	MaxEnrollment int32
	CreatedAt     time.Time
}

// Validate enforces invariants on the SemesterOffering entity.
func (o *SemesterOffering) Validate() error {
	if o.SubjectID == uuid.Nil {
		return fmt.Errorf("subject ID is required")
	}
	if o.SemesterID == uuid.Nil {
		return fmt.Errorf("semester ID is required")
	}
	if o.MaxEnrollment < 1 {
		return fmt.Errorf("max enrollment must be at least 1, got %d", o.MaxEnrollment)
	}
	return nil
}
