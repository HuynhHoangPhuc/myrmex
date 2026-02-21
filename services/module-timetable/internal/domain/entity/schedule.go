package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/domain/valueobject"
)

// Schedule is the aggregate root for a proposed or published timetable.
type Schedule struct {
	ID             uuid.UUID
	SemesterID     uuid.UUID
	Name           string
	Status         valueobject.ScheduleStatus
	Score          float64
	HardViolations int
	SoftPenalty    float64
	GeneratedAt    *time.Time
	CreatedAt      time.Time
	Entries        []*ScheduleEntry
}

func (s *Schedule) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("schedule name is required")
	}
	if !s.Status.IsValid() {
		return fmt.Errorf("invalid schedule status: %s", s.Status)
	}
	return nil
}

// Publish transitions the schedule from Draft to Published.
func (s *Schedule) Publish() error {
	if s.Status != valueobject.ScheduleStatusDraft {
		return fmt.Errorf("only draft schedules can be published, current: %s", s.Status)
	}
	if s.HardViolations > 0 {
		return fmt.Errorf("cannot publish schedule with %d hard violations", s.HardViolations)
	}
	s.Status = valueobject.ScheduleStatusPublished
	return nil
}
