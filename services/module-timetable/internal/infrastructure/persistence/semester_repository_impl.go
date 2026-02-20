package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/myrmex-erp/myrmex/services/module-timetable/internal/domain/entity"
	"github.com/myrmex-erp/myrmex/services/module-timetable/internal/infrastructure/persistence/sqlc"
)

// SemesterRepositoryImpl implements domain/repository.SemesterRepository using pgx queries.
type SemesterRepositoryImpl struct {
	q *sqlc.Queries
}

func NewSemesterRepository(q *sqlc.Queries) *SemesterRepositoryImpl {
	return &SemesterRepositoryImpl{q: q}
}

func (r *SemesterRepositoryImpl) Create(ctx context.Context, s *entity.Semester) (*entity.Semester, error) {
	offered := uuidsToPostgres(s.OfferedSubjectIDs)
	row, err := r.q.CreateSemester(ctx, sqlc.CreateSemesterParams{
		Name:              s.Name,
		Year:              int32(s.Year),
		Term:              int32(s.Term),
		StartDate:         timeToPgDate(s.StartDate),
		EndDate:           timeToPgDate(s.EndDate),
		OfferedSubjectIDs: offered,
	})
	if err != nil {
		return nil, fmt.Errorf("create semester: %w", err)
	}
	return semesterToEntity(row), nil
}

func (r *SemesterRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.Semester, error) {
	row, err := r.q.GetSemesterByID(ctx, uuidToPg(id))
	if err != nil {
		return nil, fmt.Errorf("get semester: %w", err)
	}
	return semesterToEntity(row), nil
}

func (r *SemesterRepositoryImpl) List(ctx context.Context, limit, offset int32) ([]*entity.Semester, error) {
	rows, err := r.q.ListSemesters(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list semesters: %w", err)
	}
	result := make([]*entity.Semester, len(rows))
	for i, row := range rows {
		result[i] = semesterToEntity(row)
	}
	return result, nil
}

func (r *SemesterRepositoryImpl) Count(ctx context.Context) (int64, error) {
	return r.q.CountSemesters(ctx)
}

func (r *SemesterRepositoryImpl) AddOfferedSubject(ctx context.Context, semesterID, subjectID uuid.UUID) (*entity.Semester, error) {
	row, err := r.q.AddOfferedSubject(ctx, uuidToPg(semesterID), uuidToPg(subjectID))
	if err != nil {
		return nil, fmt.Errorf("add offered subject: %w", err)
	}
	return semesterToEntity(row), nil
}

func (r *SemesterRepositoryImpl) RemoveOfferedSubject(ctx context.Context, semesterID, subjectID uuid.UUID) (*entity.Semester, error) {
	row, err := r.q.RemoveOfferedSubject(ctx, uuidToPg(semesterID), uuidToPg(subjectID))
	if err != nil {
		return nil, fmt.Errorf("remove offered subject: %w", err)
	}
	return semesterToEntity(row), nil
}

func (r *SemesterRepositoryImpl) CreateTimeSlot(ctx context.Context, ts *entity.TimeSlot) (*entity.TimeSlot, error) {
	row, err := r.q.CreateTimeSlot(ctx, sqlc.CreateTimeSlotParams{
		SemesterID:  uuidToPg(ts.SemesterID),
		DayOfWeek:   int32(ts.DayOfWeek),
		StartPeriod: int32(ts.StartPeriod),
		EndPeriod:   int32(ts.EndPeriod),
	})
	if err != nil {
		return nil, fmt.Errorf("create time slot: %w", err)
	}
	return timeSlotToEntity(row), nil
}

func (r *SemesterRepositoryImpl) ListTimeSlots(ctx context.Context, semesterID uuid.UUID) ([]*entity.TimeSlot, error) {
	rows, err := r.q.ListTimeSlotsBySemester(ctx, uuidToPg(semesterID))
	if err != nil {
		return nil, fmt.Errorf("list time slots: %w", err)
	}
	result := make([]*entity.TimeSlot, len(rows))
	for i, row := range rows {
		result[i] = timeSlotToEntity(row)
	}
	return result, nil
}

// --- mapping helpers ---

func semesterToEntity(r sqlc.TimetableSemester) *entity.Semester {
	s := &entity.Semester{
		ID:        pgToUUID(r.ID),
		Name:      r.Name,
		Year:      int(r.Year),
		Term:      int(r.Term),
		StartDate: r.StartDate.Time,
		EndDate:   r.EndDate.Time,
		CreatedAt: r.CreatedAt.Time,
	}
	for _, pgID := range r.OfferedSubjectIDs {
		s.OfferedSubjectIDs = append(s.OfferedSubjectIDs, pgToUUID(pgID))
	}
	return s
}

func timeSlotToEntity(r sqlc.TimetableTimeSlot) *entity.TimeSlot {
	return &entity.TimeSlot{
		ID:          pgToUUID(r.ID),
		SemesterID:  pgToUUID(r.SemesterID),
		DayOfWeek:   int(r.DayOfWeek),
		StartPeriod: int(r.StartPeriod),
		EndPeriod:   int(r.EndPeriod),
	}
}

func uuidToPg(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func pgToUUID(id pgtype.UUID) uuid.UUID {
	return uuid.UUID(id.Bytes)
}

func uuidsToPostgres(ids []uuid.UUID) []pgtype.UUID {
	result := make([]pgtype.UUID, len(ids))
	for i, id := range ids {
		result[i] = uuidToPg(id)
	}
	return result
}

func timeToPgDate(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, Valid: true}
}
