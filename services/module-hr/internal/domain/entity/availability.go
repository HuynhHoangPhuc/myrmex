package entity

import (
	"fmt"

	"github.com/google/uuid"
)

// Availability represents a weekly recurring time slot for a teacher.
// DayOfWeek: 0=Monday, 6=Sunday.
// StartPeriod/EndPeriod are period indices (e.g., 1=first period of day).
type Availability struct {
	ID          uuid.UUID
	TeacherID   uuid.UUID
	DayOfWeek   int
	StartPeriod int
	EndPeriod   int
}

// Validate checks the slot is well-formed.
func (a *Availability) Validate() error {
	if a.DayOfWeek < 0 || a.DayOfWeek > 6 {
		return fmt.Errorf("day_of_week must be between 0 and 6")
	}
	if a.StartPeriod >= a.EndPeriod {
		return fmt.Errorf("start_period must be less than end_period")
	}
	return nil
}
