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

	// Denormalised fields populated by query handlers (not persisted here).
	TimeSlot *TimeSlot
	Room     *Room
}
