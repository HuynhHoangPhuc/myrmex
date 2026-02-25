// Package sqlc provides hand-written query helpers over pgx/v5 pool.
// These replace auto-generated sqlc output while keeping the same ergonomics.
package sqlc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Queries wraps a pgxpool.Pool and exposes typed query methods.
type Queries struct {
	pool *pgxpool.Pool
}

// New creates a Queries instance backed by the given pool.
func New(pool *pgxpool.Pool) *Queries {
	return &Queries{pool: pool}
}

// --- Semester queries ---

type CreateSemesterParams struct {
	Name              string
	Year              int32
	Term              int32
	StartDate         pgtype.Date
	EndDate           pgtype.Date
	OfferedSubjectIDs []pgtype.UUID
}

func (q *Queries) CreateSemester(ctx context.Context, p CreateSemesterParams) (TimetableSemester, error) {
	row := q.pool.QueryRow(ctx, `
		INSERT INTO timetable.semesters (name, year, term, start_date, end_date, offered_subject_ids)
		VALUES ($1,$2,$3,$4,$5,$6) RETURNING *`,
		p.Name, p.Year, p.Term, p.StartDate, p.EndDate, p.OfferedSubjectIDs,
	)
	return scanSemester(row)
}

func (q *Queries) GetSemesterByID(ctx context.Context, id pgtype.UUID) (TimetableSemester, error) {
	row := q.pool.QueryRow(ctx, `SELECT * FROM timetable.semesters WHERE id = $1`, id)
	return scanSemester(row)
}

func (q *Queries) ListSemesters(ctx context.Context, limit, offset int32) ([]TimetableSemester, error) {
	rows, err := q.pool.Query(ctx,
		`SELECT * FROM timetable.semesters ORDER BY year DESC, term DESC LIMIT $1 OFFSET $2`,
		limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (TimetableSemester, error) {
		return scanSemesterRow(row)
	})
}

func (q *Queries) CountSemesters(ctx context.Context) (int64, error) {
	var n int64
	err := q.pool.QueryRow(ctx, `SELECT COUNT(*) FROM timetable.semesters`).Scan(&n)
	return n, err
}

func (q *Queries) AddOfferedSubject(ctx context.Context, semesterID, subjectID pgtype.UUID) (TimetableSemester, error) {
	row := q.pool.QueryRow(ctx, `
		UPDATE timetable.semesters SET offered_subject_ids = array_append(offered_subject_ids,$2)
		WHERE id=$1 RETURNING *`, semesterID, subjectID)
	return scanSemester(row)
}

func (q *Queries) RemoveOfferedSubject(ctx context.Context, semesterID, subjectID pgtype.UUID) (TimetableSemester, error) {
	row := q.pool.QueryRow(ctx, `
		UPDATE timetable.semesters SET offered_subject_ids = array_remove(offered_subject_ids,$2)
		WHERE id=$1 RETURNING *`, semesterID, subjectID)
	return scanSemester(row)
}

// --- TimeSlot queries ---

type CreateTimeSlotParams struct {
	SemesterID  pgtype.UUID
	DayOfWeek   int32
	StartPeriod int32
	EndPeriod   int32
}

func (q *Queries) CreateTimeSlot(ctx context.Context, p CreateTimeSlotParams) (TimetableTimeSlot, error) {
	row := q.pool.QueryRow(ctx, `
		INSERT INTO timetable.time_slots (semester_id, day_of_week, start_period, end_period)
		VALUES ($1,$2,$3,$4) RETURNING *`,
		p.SemesterID, p.DayOfWeek, p.StartPeriod, p.EndPeriod)
	var s TimetableTimeSlot
	err := row.Scan(&s.ID, &s.SemesterID, &s.DayOfWeek, &s.StartPeriod, &s.EndPeriod)
	return s, err
}

func (q *Queries) ListTimeSlotsBySemester(ctx context.Context, semesterID pgtype.UUID) ([]TimetableTimeSlot, error) {
	rows, err := q.pool.Query(ctx,
		`SELECT * FROM timetable.time_slots WHERE semester_id=$1 ORDER BY day_of_week, start_period`,
		semesterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (TimetableTimeSlot, error) {
		var s TimetableTimeSlot
		err := row.Scan(&s.ID, &s.SemesterID, &s.DayOfWeek, &s.StartPeriod, &s.EndPeriod)
		return s, err
	})
}

// --- Room queries ---

type CreateRoomParams struct {
	Name     string
	Capacity int32
	Type     string
	Features []string
}

func (q *Queries) CreateRoom(ctx context.Context, p CreateRoomParams) (TimetableRoom, error) {
	row := q.pool.QueryRow(ctx, `
		INSERT INTO timetable.rooms (name, capacity, type, features)
		VALUES ($1,$2,$3,$4) RETURNING *`,
		p.Name, p.Capacity, p.Type, p.Features)
	return scanRoom(row)
}

func (q *Queries) GetRoomByID(ctx context.Context, id pgtype.UUID) (TimetableRoom, error) {
	row := q.pool.QueryRow(ctx, `SELECT * FROM timetable.rooms WHERE id=$1`, id)
	return scanRoom(row)
}

func (q *Queries) ListRooms(ctx context.Context, limit, offset int32) ([]TimetableRoom, error) {
	rows, err := q.pool.Query(ctx,
		`SELECT * FROM timetable.rooms WHERE is_active=true ORDER BY name LIMIT $1 OFFSET $2`,
		limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (TimetableRoom, error) {
		return scanRoomRow(row)
	})
}

func (q *Queries) CountRooms(ctx context.Context) (int64, error) {
	var n int64
	err := q.pool.QueryRow(ctx, `SELECT COUNT(*) FROM timetable.rooms WHERE is_active=true`).Scan(&n)
	return n, err
}

// --- Schedule queries ---

func (q *Queries) CreateSchedule(ctx context.Context, semesterID pgtype.UUID, name, status string) (TimetableSchedule, error) {
	row := q.pool.QueryRow(ctx, `
		INSERT INTO timetable.schedules (semester_id, name, status)
		VALUES ($1,$2,$3) RETURNING *`,
		semesterID, name, status)
	return scanSchedule(row)
}

func (q *Queries) GetScheduleByID(ctx context.Context, id pgtype.UUID) (TimetableSchedule, error) {
	row := q.pool.QueryRow(ctx, `SELECT * FROM timetable.schedules WHERE id=$1`, id)
	return scanSchedule(row)
}

func (q *Queries) ListSchedulesBySemester(ctx context.Context, semesterID pgtype.UUID) ([]TimetableSchedule, error) {
	rows, err := q.pool.Query(ctx,
		`SELECT * FROM timetable.schedules WHERE semester_id=$1 ORDER BY created_at DESC`, semesterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (TimetableSchedule, error) {
		return scanScheduleRow(row)
	})
}

// ListSchedulesPaged returns schedules with optional semester filter and pagination.
// Pass a zero-value UUID as semesterID to skip the filter.
func (q *Queries) ListSchedulesPaged(ctx context.Context, semesterID pgtype.UUID, limit, offset int32) ([]TimetableSchedule, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT * FROM timetable.schedules
		WHERE ($1::uuid IS NULL OR NOT $1::uuid = '00000000-0000-0000-0000-000000000000'::uuid AND semester_id = $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		semesterID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (TimetableSchedule, error) {
		return scanScheduleRow(row)
	})
}

// CountSchedulesPaged returns the total schedule count with optional semester filter.
func (q *Queries) CountSchedulesPaged(ctx context.Context, semesterID pgtype.UUID) (int64, error) {
	var n int64
	err := q.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM timetable.schedules
		WHERE ($1::uuid IS NULL OR NOT $1::uuid = '00000000-0000-0000-0000-000000000000'::uuid AND semester_id = $1)`,
		semesterID).Scan(&n)
	return n, err
}

func (q *Queries) UpdateScheduleResult(ctx context.Context, id pgtype.UUID, score float64, hardViolations int32, softPenalty float64) (TimetableSchedule, error) {
	row := q.pool.QueryRow(ctx, `
		UPDATE timetable.schedules SET score=$2, hard_violations=$3, soft_penalty=$4, generated_at=NOW()
		WHERE id=$1 RETURNING *`,
		id, score, hardViolations, softPenalty)
	return scanSchedule(row)
}

func (q *Queries) UpdateScheduleStatus(ctx context.Context, id pgtype.UUID, status string) (TimetableSchedule, error) {
	row := q.pool.QueryRow(ctx, `UPDATE timetable.schedules SET status=$2 WHERE id=$1 RETURNING *`, id, status)
	return scanSchedule(row)
}

// --- ScheduleEntry queries ---

type CreateScheduleEntryParams struct {
	ScheduleID       pgtype.UUID
	SubjectID        pgtype.UUID
	TeacherID        pgtype.UUID
	RoomID           pgtype.UUID
	TimeSlotID       pgtype.UUID
	IsManualOverride bool
	SubjectName      string
	SubjectCode      string
	TeacherName      string
	DepartmentID     pgtype.UUID
}

func (q *Queries) CreateScheduleEntry(ctx context.Context, p CreateScheduleEntryParams) (TimetableScheduleEntry, error) {
	row := q.pool.QueryRow(ctx, `
		INSERT INTO timetable.schedule_entries
		  (schedule_id, subject_id, teacher_id, room_id, time_slot_id, is_manual_override,
		   subject_name, subject_code, teacher_name, department_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING *`,
		p.ScheduleID, p.SubjectID, p.TeacherID, p.RoomID, p.TimeSlotID, p.IsManualOverride,
		p.SubjectName, p.SubjectCode, p.TeacherName, p.DepartmentID)
	return scanEntry(row)
}

func (q *Queries) GetScheduleEntry(ctx context.Context, id pgtype.UUID) (TimetableScheduleEntry, error) {
	row := q.pool.QueryRow(ctx, `SELECT * FROM timetable.schedule_entries WHERE id=$1`, id)
	return scanEntry(row)
}

// ListEntriesByScheduleEnriched returns entries with time-slot and room fields via JOIN.
func (q *Queries) ListEntriesByScheduleEnriched(ctx context.Context, scheduleID pgtype.UUID) ([]TimetableScheduleEntryRow, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT
		  e.id, e.schedule_id, e.subject_id, e.teacher_id, e.room_id, e.time_slot_id,
		  e.is_manual_override, e.created_at,
		  e.subject_name, e.subject_code, e.teacher_name, e.department_id,
		  ts.day_of_week, ts.start_period, ts.end_period,
		  r.name AS room_name
		FROM timetable.schedule_entries e
		JOIN timetable.time_slots ts ON e.time_slot_id = ts.id
		JOIN timetable.rooms r       ON e.room_id      = r.id
		WHERE e.schedule_id = $1
		ORDER BY ts.day_of_week, ts.start_period`, scheduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (TimetableScheduleEntryRow, error) {
		return scanEntryEnrichedRow(row)
	})
}

func (q *Queries) UpdateScheduleEntry(ctx context.Context, id, teacherID, roomID, timeSlotID pgtype.UUID, isManual bool) (TimetableScheduleEntry, error) {
	row := q.pool.QueryRow(ctx, `
		UPDATE timetable.schedule_entries
		SET teacher_id=$2, room_id=$3, time_slot_id=$4, is_manual_override=$5
		WHERE id=$1 RETURNING *`,
		id, teacherID, roomID, timeSlotID, isManual)
	return scanEntry(row)
}

func (q *Queries) DeleteScheduleEntry(ctx context.Context, id pgtype.UUID) error {
	_, err := q.pool.Exec(ctx, `DELETE FROM timetable.schedule_entries WHERE id=$1`, id)
	return err
}

func (q *Queries) DeleteEntriesBySchedule(ctx context.Context, scheduleID pgtype.UUID) error {
	_, err := q.pool.Exec(ctx, `DELETE FROM timetable.schedule_entries WHERE schedule_id=$1`, scheduleID)
	return err
}

// --- Event store ---

func (q *Queries) AppendEvent(ctx context.Context, aggregateID pgtype.UUID, aggregateType, eventType string, payload json.RawMessage) error {
	_, err := q.pool.Exec(ctx, `
		INSERT INTO timetable.event_store (aggregate_id, aggregate_type, event_type, payload)
		VALUES ($1,$2,$3,$4)`,
		aggregateID, aggregateType, eventType, payload)
	return err
}

// --- scan helpers ---

func scanSemester(row pgx.Row) (TimetableSemester, error) {
	var s TimetableSemester
	err := row.Scan(&s.ID, &s.Name, &s.Year, &s.Term, &s.StartDate, &s.EndDate, &s.OfferedSubjectIDs, &s.CreatedAt)
	if err != nil {
		return s, fmt.Errorf("scan semester: %w", err)
	}
	return s, nil
}

func scanSemesterRow(row pgx.CollectableRow) (TimetableSemester, error) {
	var s TimetableSemester
	err := row.Scan(&s.ID, &s.Name, &s.Year, &s.Term, &s.StartDate, &s.EndDate, &s.OfferedSubjectIDs, &s.CreatedAt)
	return s, err
}

func scanRoom(row pgx.Row) (TimetableRoom, error) {
	var r TimetableRoom
	err := row.Scan(&r.ID, &r.Name, &r.Capacity, &r.Type, &r.Features, &r.IsActive)
	if err != nil {
		return r, fmt.Errorf("scan room: %w", err)
	}
	return r, nil
}

func scanRoomRow(row pgx.CollectableRow) (TimetableRoom, error) {
	var r TimetableRoom
	err := row.Scan(&r.ID, &r.Name, &r.Capacity, &r.Type, &r.Features, &r.IsActive)
	return r, err
}

func scanSchedule(row pgx.Row) (TimetableSchedule, error) {
	var s TimetableSchedule
	err := row.Scan(&s.ID, &s.SemesterID, &s.Name, &s.Status,
		&s.Score, &s.HardViolations, &s.SoftPenalty, &s.GeneratedAt, &s.CreatedAt)
	if err != nil {
		return s, fmt.Errorf("scan schedule: %w", err)
	}
	return s, nil
}

func scanScheduleRow(row pgx.CollectableRow) (TimetableSchedule, error) {
	var s TimetableSchedule
	err := row.Scan(&s.ID, &s.SemesterID, &s.Name, &s.Status,
		&s.Score, &s.HardViolations, &s.SoftPenalty, &s.GeneratedAt, &s.CreatedAt)
	return s, err
}

func scanEntry(row pgx.Row) (TimetableScheduleEntry, error) {
	var e TimetableScheduleEntry
	err := row.Scan(
		&e.ID, &e.ScheduleID, &e.SubjectID, &e.TeacherID,
		&e.RoomID, &e.TimeSlotID, &e.IsManualOverride, &e.CreatedAt,
		&e.SubjectName, &e.SubjectCode, &e.TeacherName, &e.DepartmentID,
	)
	if err != nil {
		return e, fmt.Errorf("scan entry: %w", err)
	}
	return e, nil
}

func scanEntryEnrichedRow(row pgx.CollectableRow) (TimetableScheduleEntryRow, error) {
	var e TimetableScheduleEntryRow
	err := row.Scan(
		&e.ID, &e.ScheduleID, &e.SubjectID, &e.TeacherID,
		&e.RoomID, &e.TimeSlotID, &e.IsManualOverride, &e.CreatedAt,
		&e.SubjectName, &e.SubjectCode, &e.TeacherName, &e.DepartmentID,
		&e.DayOfWeek, &e.StartPeriod, &e.EndPeriod, &e.RoomName,
	)
	return e, err
}
