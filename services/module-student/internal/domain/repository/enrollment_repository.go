package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
)

// EnrollmentRepository defines persistence operations for enrollment requests.
type EnrollmentRepository interface {
	Create(ctx context.Context, enrollment *entity.EnrollmentRequest) (*entity.EnrollmentRequest, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.EnrollmentRequest, error)
	Review(ctx context.Context, id uuid.UUID, status, adminNote string, reviewedBy uuid.UUID) (*entity.EnrollmentRequest, error)
	List(ctx context.Context, studentID, semesterID *uuid.UUID, status *string, limit, offset int32) ([]*entity.EnrollmentRequest, error)
	Count(ctx context.Context, studentID, semesterID *uuid.UUID, status *string) (int64, error)
	ListPassedSubjectIDs(ctx context.Context, studentID uuid.UUID) ([]uuid.UUID, error)
	ListByStudent(ctx context.Context, studentID uuid.UUID, semesterID *uuid.UUID) ([]*entity.EnrollmentRequest, error)
	AppendEvent(ctx context.Context, aggregateID uuid.UUID, eventType string, payload []byte) error
}
