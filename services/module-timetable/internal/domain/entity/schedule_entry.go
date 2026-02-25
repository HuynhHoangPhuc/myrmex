package entity

import (
	"time"

	"github.com/google/uuid"
)

// ScheduleEntry maps a subject to a teacher, room, and time slot within a schedule.
type ScheduleEntry struct {
	ID               uuid.UUID
	ScheduleID       uuid.UUID
	SubjectID        uuid.UUID
	TeacherID        uuid.UUID
	RoomID           uuid.UUID
	TimeSlotID       uuid.UUID
	IsManualOverride bool
	CreatedAt        time.Time

	// Denormalised write-time fields (persisted in schedule_entries table).
	SubjectName  string
	SubjectCode  string
	TeacherName  string
	DepartmentID uuid.UUID

	// Display-only fields populated via JOIN on read (not persisted directly).
	DayOfWeek   int
	StartPeriod int
	EndPeriod   int
	RoomName    string
}
