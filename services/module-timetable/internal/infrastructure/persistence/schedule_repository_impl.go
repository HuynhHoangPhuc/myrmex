package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/domain/valueobject"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/infrastructure/persistence/sqlc"
)

// ScheduleRepositoryImpl implements domain/repository.ScheduleRepository.
type ScheduleRepositoryImpl struct {
	q *sqlc.Queries
}

func NewScheduleRepository(q *sqlc.Queries) *ScheduleRepositoryImpl {
	return &ScheduleRepositoryImpl{q: q}
}

func (r *ScheduleRepositoryImpl) Create(ctx context.Context, s *entity.Schedule) (*entity.Schedule, error) {
	row, err := r.q.CreateSchedule(ctx, uuidToPg(s.SemesterID), s.Name, s.Status.String())
	if err != nil {
		return nil, fmt.Errorf("create schedule: %w", err)
	}
	return scheduleToEntity(row), nil
}

func (r *ScheduleRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.Schedule, error) {
	row, err := r.q.GetScheduleByID(ctx, uuidToPg(id))
	if err != nil {
		return nil, fmt.Errorf("get schedule: %w", err)
	}
	return scheduleToEntity(row), nil
}

func (r *ScheduleRepositoryImpl) ListBySemester(ctx context.Context, semesterID uuid.UUID) ([]*entity.Schedule, error) {
	rows, err := r.q.ListSchedulesBySemester(ctx, uuidToPg(semesterID))
	if err != nil {
		return nil, fmt.Errorf("list schedules: %w", err)
	}
	result := make([]*entity.Schedule, len(rows))
	for i, row := range rows {
		result[i] = scheduleToEntity(row)
	}
	return result, nil
}

func (r *ScheduleRepositoryImpl) ListPaged(ctx context.Context, semesterID uuid.UUID, limit, offset int32) ([]*entity.Schedule, error) {
	rows, err := r.q.ListSchedulesPaged(ctx, uuidToPg(semesterID), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list schedules paged: %w", err)
	}
	result := make([]*entity.Schedule, len(rows))
	for i, row := range rows {
		result[i] = scheduleToEntity(row)
	}
	return result, nil
}

func (r *ScheduleRepositoryImpl) CountSchedules(ctx context.Context, semesterID uuid.UUID) (int64, error) {
	n, err := r.q.CountSchedulesPaged(ctx, uuidToPg(semesterID))
	if err != nil {
		return 0, fmt.Errorf("count schedules: %w", err)
	}
	return n, nil
}

func (r *ScheduleRepositoryImpl) UpdateResult(ctx context.Context, id uuid.UUID, score float64, hardViolations int, softPenalty float64) (*entity.Schedule, error) {
	row, err := r.q.UpdateScheduleResult(ctx, uuidToPg(id), score, int32(hardViolations), softPenalty)
	if err != nil {
		return nil, fmt.Errorf("update schedule result: %w", err)
	}
	return scheduleToEntity(row), nil
}

func (r *ScheduleRepositoryImpl) UpdateStatus(ctx context.Context, id uuid.UUID, status valueobject.ScheduleStatus) (*entity.Schedule, error) {
	row, err := r.q.UpdateScheduleStatus(ctx, uuidToPg(id), status.String())
	if err != nil {
		return nil, fmt.Errorf("update schedule status: %w", err)
	}
	return scheduleToEntity(row), nil
}

func (r *ScheduleRepositoryImpl) CreateEntry(ctx context.Context, e *entity.ScheduleEntry) (*entity.ScheduleEntry, error) {
	row, err := r.q.CreateScheduleEntry(ctx, sqlc.CreateScheduleEntryParams{
		ScheduleID:       uuidToPg(e.ScheduleID),
		SubjectID:        uuidToPg(e.SubjectID),
		TeacherID:        uuidToPg(e.TeacherID),
		RoomID:           uuidToPg(e.RoomID),
		TimeSlotID:       uuidToPg(e.TimeSlotID),
		IsManualOverride: e.IsManualOverride,
		SubjectName:      e.SubjectName,
		SubjectCode:      e.SubjectCode,
		TeacherName:      e.TeacherName,
		DepartmentID:     uuidToPg(e.DepartmentID),
	})
	if err != nil {
		return nil, fmt.Errorf("create schedule entry: %w", err)
	}
	return entryToEntity(row), nil
}

func (r *ScheduleRepositoryImpl) GetEntry(ctx context.Context, id uuid.UUID) (*entity.ScheduleEntry, error) {
	row, err := r.q.GetScheduleEntry(ctx, uuidToPg(id))
	if err != nil {
		return nil, fmt.Errorf("get schedule entry: %w", err)
	}
	return entryToEntity(row), nil
}

func (r *ScheduleRepositoryImpl) ListEntries(ctx context.Context, scheduleID uuid.UUID) ([]*entity.ScheduleEntry, error) {
	rows, err := r.q.ListEntriesByScheduleEnriched(ctx, uuidToPg(scheduleID))
	if err != nil {
		return nil, fmt.Errorf("list entries: %w", err)
	}
	result := make([]*entity.ScheduleEntry, len(rows))
	for i, row := range rows {
		result[i] = enrichedEntryToEntity(row)
	}
	return result, nil
}

func (r *ScheduleRepositoryImpl) UpdateEntry(ctx context.Context, e *entity.ScheduleEntry) (*entity.ScheduleEntry, error) {
	row, err := r.q.UpdateScheduleEntry(ctx,
		uuidToPg(e.ID), uuidToPg(e.TeacherID), uuidToPg(e.RoomID),
		uuidToPg(e.TimeSlotID), e.IsManualOverride)
	if err != nil {
		return nil, fmt.Errorf("update entry: %w", err)
	}
	return entryToEntity(row), nil
}

func (r *ScheduleRepositoryImpl) DeleteEntry(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteScheduleEntry(ctx, uuidToPg(id))
}

func (r *ScheduleRepositoryImpl) DeleteAllEntries(ctx context.Context, scheduleID uuid.UUID) error {
	return r.q.DeleteEntriesBySchedule(ctx, uuidToPg(scheduleID))
}

func (r *ScheduleRepositoryImpl) AppendEvent(ctx context.Context, aggregateID uuid.UUID, aggregateType, eventType string, payload json.RawMessage) error {
	return r.q.AppendEvent(ctx, uuidToPg(aggregateID), aggregateType, eventType, payload)
}

// --- mapping helpers ---

func scheduleToEntity(r sqlc.TimetableSchedule) *entity.Schedule {
	s := &entity.Schedule{
		ID:             pgToUUID(r.ID),
		SemesterID:     pgToUUID(r.SemesterID),
		Name:           r.Name,
		Status:         valueobject.ScheduleStatus(r.Status),
		HardViolations: int(r.HardViolations),
		SoftPenalty:    r.SoftPenalty,
		CreatedAt:      r.CreatedAt.Time,
	}
	if r.Score != nil {
		s.Score = *r.Score
	}
	if r.GeneratedAt.Valid {
		t := r.GeneratedAt.Time
		s.GeneratedAt = &t
	}
	return s
}

func entryToEntity(r sqlc.TimetableScheduleEntry) *entity.ScheduleEntry {
	return &entity.ScheduleEntry{
		ID:               pgToUUID(r.ID),
		ScheduleID:       pgToUUID(r.ScheduleID),
		SubjectID:        pgToUUID(r.SubjectID),
		TeacherID:        pgToUUID(r.TeacherID),
		RoomID:           pgToUUID(r.RoomID),
		TimeSlotID:       pgToUUID(r.TimeSlotID),
		IsManualOverride: r.IsManualOverride,
		CreatedAt:        r.CreatedAt.Time,
		SubjectName:      r.SubjectName,
		SubjectCode:      r.SubjectCode,
		TeacherName:      r.TeacherName,
		DepartmentID:     pgToUUID(r.DepartmentID),
	}
}

func enrichedEntryToEntity(r sqlc.TimetableScheduleEntryRow) *entity.ScheduleEntry {
	e := entryToEntity(r.TimetableScheduleEntry)
	e.DayOfWeek   = int(r.DayOfWeek)
	e.StartPeriod = int(r.StartPeriod)
	e.EndPeriod   = int(r.EndPeriod)
	e.RoomName    = r.RoomName
	return e
}

// pgTimestamptzToTimePtr converts an optional pgtype.Timestamptz to *time.Time.
func pgTimestamptzToTimePtr(ts pgtype.Timestamptz) *time.Time {
	if !ts.Valid {
		return nil
	}
	t := ts.Time
	return &t
}
