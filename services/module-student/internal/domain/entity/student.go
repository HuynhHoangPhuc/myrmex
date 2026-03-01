package entity

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	StudentStatusActive    = "active"
	StudentStatusGraduated = "graduated"
	StudentStatusSuspended = "suspended"
)

// Student represents the student aggregate root.
type Student struct {
	ID             uuid.UUID
	StudentCode    string
	UserID         *uuid.UUID
	FullName       string
	Email          string
	DepartmentID   uuid.UUID
	EnrollmentYear int
	Status         string
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Validate checks required fields and allowed values.
func (s *Student) Validate() error {
	if strings.TrimSpace(s.StudentCode) == "" {
		return fmt.Errorf("student_code is required")
	}
	if strings.TrimSpace(s.FullName) == "" {
		return fmt.Errorf("full_name is required")
	}
	if strings.TrimSpace(s.Email) == "" {
		return fmt.Errorf("email is required")
	}
	if s.DepartmentID == uuid.Nil {
		return fmt.Errorf("department_id is required")
	}
	if s.EnrollmentYear < 2000 || s.EnrollmentYear > 2100 {
		return fmt.Errorf("enrollment_year must be between 2000 and 2100")
	}
	if s.Status == "" {
		s.Status = StudentStatusActive
	}
	if s.Status != StudentStatusActive && s.Status != StudentStatusGraduated && s.Status != StudentStatusSuspended {
		return fmt.Errorf("status must be one of active, graduated, suspended")
	}
	return nil
}
