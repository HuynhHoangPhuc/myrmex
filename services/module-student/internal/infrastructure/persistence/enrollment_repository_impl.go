package persistence

import (
	"context"
	"fmt"

	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/infrastructure/persistence/sqlc"
	"github.com/google/uuid"
)

// EnrollmentRepositoryImpl implements EnrollmentRepository using sqlc queries.
type EnrollmentRepositoryImpl struct {
	queries *sqlc.Queries
}

func NewEnrollmentRepository(queries *sqlc.Queries) *EnrollmentRepositoryImpl {
	return &EnrollmentRepositoryImpl{queries: queries}
}

func (r *EnrollmentRepositoryImpl) Create(ctx context.Context, enrollment *entity.EnrollmentRequest) (*entity.EnrollmentRequest, error) {
	row, err := r.queries.CreateEnrollmentRequest(ctx, sqlc.CreateEnrollmentRequestParams{
		ID:               uuidToPgtype(enrollment.ID),
		StudentID:        uuidToPgtype(enrollment.StudentID),
		SemesterID:       uuidToPgtype(enrollment.SemesterID),
		OfferedSubjectID: uuidToPgtype(enrollment.OfferedSubjectID),
		SubjectID:        uuidToPgtype(enrollment.SubjectID),
		RequestNote:      enrollment.RequestNote,
	})
	if err != nil {
		return nil, fmt.Errorf("insert enrollment request: %w", err)
	}
	return enrollmentRowToEntity(row), nil
}

func (r *EnrollmentRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.EnrollmentRequest, error) {
	row, err := r.queries.GetEnrollmentRequest(ctx, uuidToPgtype(id))
	if err != nil {
		return nil, fmt.Errorf("get enrollment request: %w", err)
	}
	return enrollmentRowToEntity(row), nil
}

func (r *EnrollmentRepositoryImpl) Review(ctx context.Context, id uuid.UUID, status, adminNote string, reviewedBy uuid.UUID) (*entity.EnrollmentRequest, error) {
	row, err := r.queries.ReviewEnrollmentRequest(ctx, sqlc.ReviewEnrollmentRequestParams{
		Status:     status,
		AdminNote:  adminNote,
		ReviewedBy: uuidToPgtype(reviewedBy),
		ID:         uuidToPgtype(id),
	})
	if err != nil {
		return nil, fmt.Errorf("review enrollment request: %w", err)
	}
	return enrollmentRowToEntity(row), nil
}

func (r *EnrollmentRepositoryImpl) List(ctx context.Context, studentID, semesterID *uuid.UUID, status *string, limit, offset int32) ([]*entity.EnrollmentRequest, error) {
	rows, err := r.queries.ListEnrollmentRequests(ctx, sqlc.ListEnrollmentRequestsParams{
		StudentID:   uuidOptToPgtype(studentID),
		SemesterID:  uuidOptToPgtype(semesterID),
		Status:      textOpt(status),
		OffsetCount: offset,
		LimitCount:  limit,
	})
	if err != nil {
		return nil, fmt.Errorf("list enrollment requests: %w", err)
	}
	return enrollmentRowsToEntities(rows), nil
}

func (r *EnrollmentRepositoryImpl) Count(ctx context.Context, studentID, semesterID *uuid.UUID, status *string) (int64, error) {
	count, err := r.queries.CountEnrollmentRequests(ctx, sqlc.CountEnrollmentRequestsParams{
		StudentID:  uuidOptToPgtype(studentID),
		SemesterID: uuidOptToPgtype(semesterID),
		Status:     textOpt(status),
	})
	if err != nil {
		return 0, fmt.Errorf("count enrollment requests: %w", err)
	}
	return count, nil
}

func (r *EnrollmentRepositoryImpl) ListPassedSubjectIDs(ctx context.Context, studentID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.queries.ListPassedEnrollmentSubjectIDs(ctx, uuidToPgtype(studentID))
	if err != nil {
		return nil, fmt.Errorf("list passed subjects: %w", err)
	}
	ids := make([]uuid.UUID, len(rows))
	for i, id := range rows {
		ids[i] = pgtypeToUUID(id)
	}
	return ids, nil
}

func (r *EnrollmentRepositoryImpl) ListByStudent(ctx context.Context, studentID uuid.UUID, semesterID *uuid.UUID) ([]*entity.EnrollmentRequest, error) {
	rows, err := r.queries.GetStudentEnrollments(ctx, sqlc.GetStudentEnrollmentsParams{
		StudentID:  uuidToPgtype(studentID),
		SemesterID: uuidOptToPgtype(semesterID),
	})
	if err != nil {
		return nil, fmt.Errorf("get student enrollments: %w", err)
	}
	return enrollmentRowsToEntities(rows), nil
}

func (r *EnrollmentRepositoryImpl) AppendEvent(ctx context.Context, aggregateID uuid.UUID, eventType string, payload []byte) error {
	if err := r.queries.AppendEnrollmentEvent(ctx, sqlc.AppendEnrollmentEventParams{
		AggregateID: uuidToPgtype(aggregateID),
		EventType:   eventType,
		Payload:     payload,
	}); err != nil {
		return fmt.Errorf("append enrollment event: %w", err)
	}
	return nil
}

func enrollmentRowsToEntities(rows []sqlc.StudentEnrollmentRequest) []*entity.EnrollmentRequest {
	result := make([]*entity.EnrollmentRequest, len(rows))
	for i, row := range rows {
		result[i] = enrollmentRowToEntity(row)
	}
	return result
}

func enrollmentRowToEntity(row sqlc.StudentEnrollmentRequest) *entity.EnrollmentRequest {
	enrollment := &entity.EnrollmentRequest{
		ID:               pgtypeToUUID(row.ID),
		StudentID:        pgtypeToUUID(row.StudentID),
		SemesterID:       pgtypeToUUID(row.SemesterID),
		OfferedSubjectID: pgtypeToUUID(row.OfferedSubjectID),
		SubjectID:        pgtypeToUUID(row.SubjectID),
		Status:           row.Status,
		RequestNote:      row.RequestNote,
		AdminNote:        row.AdminNote,
		RequestedAt:      row.RequestedAt.Time,
	}
	if row.ReviewedAt.Valid {
		reviewedAt := row.ReviewedAt.Time
		enrollment.ReviewedAt = &reviewedAt
	}
	if row.ReviewedBy.Valid {
		reviewedBy := pgtypeToUUID(row.ReviewedBy)
		enrollment.ReviewedBy = &reviewedBy
	}
	return enrollment
}

