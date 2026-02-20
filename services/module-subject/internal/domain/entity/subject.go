package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Subject represents a university course/subject aggregate root.
type Subject struct {
	ID           uuid.UUID
	Code         string
	Name         string
	Credits      int32
	Description  string
	DepartmentID string
	WeeklyHours  int32
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Validate enforces invariants on the Subject entity.
func (s *Subject) Validate() error {
	if s.Code == "" {
		return fmt.Errorf("subject code is required")
	}
	if s.Name == "" {
		return fmt.Errorf("subject name is required")
	}
	if s.Credits < 0 {
		return fmt.Errorf("credits cannot be negative")
	}
	if s.WeeklyHours < 0 {
		return fmt.Errorf("weekly hours cannot be negative")
	}
	return nil
}

// Deactivate marks the subject as inactive.
func (s *Subject) Deactivate() {
	s.IsActive = false
}
