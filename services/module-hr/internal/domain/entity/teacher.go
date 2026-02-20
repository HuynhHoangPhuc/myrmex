package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Teacher represents the HR teacher aggregate root.
type Teacher struct {
	ID              uuid.UUID
	EmployeeCode    string
	FullName        string
	Email           string
	Phone           string
	Title           string
	DepartmentID    *uuid.UUID
	MaxHoursPerWeek int
	IsActive        bool
	Specializations []string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Validate checks required fields.
func (t *Teacher) Validate() error {
	if t.FullName == "" {
		return fmt.Errorf("full_name is required")
	}
	if t.Email == "" {
		return fmt.Errorf("email is required")
	}
	if t.EmployeeCode == "" {
		return fmt.Errorf("employee_code is required")
	}
	if t.MaxHoursPerWeek <= 0 {
		t.MaxHoursPerWeek = 20
	}
	return nil
}
