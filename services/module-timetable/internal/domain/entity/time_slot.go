package entity

import (
	"fmt"

	"github.com/google/uuid"
)

// TimeSlot represents a recurring teaching period within a semester.
// Periods are integer indices (e.g. 1 = 07:00-07:50, 2 = 08:00-08:50 …).
type TimeSlot struct {
	ID          uuid.UUID
	SemesterID  uuid.UUID
	DayOfWeek   int // 0=Monday … 6=Sunday
	StartPeriod int
	EndPeriod   int
}

func (ts *TimeSlot) Validate() error {
	if ts.DayOfWeek < 0 || ts.DayOfWeek > 6 {
		return fmt.Errorf("day_of_week must be 0-6")
	}
	if ts.StartPeriod < 1 {
		return fmt.Errorf("start_period must be >= 1")
	}
	if ts.EndPeriod <= ts.StartPeriod {
		return fmt.Errorf("end_period must be > start_period")
	}
	return nil
}

// OverlapsWith returns true when two slots share at least one period on the same day.
func (ts *TimeSlot) OverlapsWith(other *TimeSlot) bool {
	if ts.DayOfWeek != other.DayOfWeek {
		return false
	}
	return ts.StartPeriod < other.EndPeriod && other.StartPeriod < ts.EndPeriod
}
