// Package sqlc holds hand-written model types used in place of sqlc-generated
// code (sqlc is not run in CI for this module; queries are executed via pgx directly).
package sqlc

import (
	"github.com/jackc/pgx/v5/pgtype"
)

// TimetableSemester mirrors the timetable.semesters table row.
type TimetableSemester struct {
	ID                 pgtype.UUID        `db:"id"`
	Name               string             `db:"name"`
	Year               int32              `db:"year"`
	Term               int32              `db:"term"`
	StartDate          pgtype.Date        `db:"start_date"`
	EndDate            pgtype.Date        `db:"end_date"`
	OfferedSubjectIDs  []pgtype.UUID      `db:"offered_subject_ids"`
	CreatedAt          pgtype.Timestamptz `db:"created_at"`
}

// TimetableRoom mirrors the timetable.rooms table row.
type TimetableRoom struct {
	ID       pgtype.UUID `db:"id"`
	Name     string      `db:"name"`
	Capacity int32       `db:"capacity"`
	Type     string      `db:"type"`
	Features []string    `db:"features"`
	IsActive bool        `db:"is_active"`
}

// TimetableTimeSlot mirrors the timetable.time_slots table row.
type TimetableTimeSlot struct {
	ID          pgtype.UUID `db:"id"`
	SemesterID  pgtype.UUID `db:"semester_id"`
	DayOfWeek   int32       `db:"day_of_week"`
	StartPeriod int32       `db:"start_period"`
	EndPeriod   int32       `db:"end_period"`
}

// TimetableSchedule mirrors the timetable.schedules table row.
type TimetableSchedule struct {
	ID             pgtype.UUID        `db:"id"`
	SemesterID     pgtype.UUID        `db:"semester_id"`
	Name           string             `db:"name"`
	Status         string             `db:"status"`
	Score          *float64           `db:"score"`
	HardViolations int32              `db:"hard_violations"`
	SoftPenalty    float64            `db:"soft_penalty"`
	GeneratedAt    pgtype.Timestamptz `db:"generated_at"`
	CreatedAt      pgtype.Timestamptz `db:"created_at"`
}

// TimetableScheduleEntry mirrors the timetable.schedule_entries table row (after migration 007).
type TimetableScheduleEntry struct {
	ID               pgtype.UUID        `db:"id"`
	ScheduleID       pgtype.UUID        `db:"schedule_id"`
	SubjectID        pgtype.UUID        `db:"subject_id"`
	TeacherID        pgtype.UUID        `db:"teacher_id"`
	RoomID           pgtype.UUID        `db:"room_id"`
	TimeSlotID       pgtype.UUID        `db:"time_slot_id"`
	IsManualOverride bool               `db:"is_manual_override"`
	CreatedAt        pgtype.Timestamptz `db:"created_at"`
	// Denormalised columns added by migration 007.
	SubjectName  string      `db:"subject_name"`
	SubjectCode  string      `db:"subject_code"`
	TeacherName  string      `db:"teacher_name"`
	DepartmentID pgtype.UUID `db:"department_id"`
}

// TimetableScheduleEntryRow is returned by ListEntriesBySchedule (JOIN query).
// Extends TimetableScheduleEntry with time-slot and room display fields.
type TimetableScheduleEntryRow struct {
	TimetableScheduleEntry
	DayOfWeek   int32  `db:"day_of_week"`
	StartPeriod int32  `db:"start_period"`
	EndPeriod   int32  `db:"end_period"`
	RoomName    string `db:"room_name"`
}
