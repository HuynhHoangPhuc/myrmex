package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/domain/entity"
)

// SemesterRepository defines persistence operations for Semester aggregates.
type SemesterRepository interface {
	Create(ctx context.Context, s *entity.Semester) (*entity.Semester, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Semester, error)
	List(ctx context.Context, limit, offset int32) ([]*entity.Semester, error)
	Count(ctx context.Context) (int64, error)
	AddOfferedSubject(ctx context.Context, semesterID, subjectID uuid.UUID) (*entity.Semester, error)
	RemoveOfferedSubject(ctx context.Context, semesterID, subjectID uuid.UUID) (*entity.Semester, error)
	CreateTimeSlot(ctx context.Context, ts *entity.TimeSlot) (*entity.TimeSlot, error)
	ListTimeSlots(ctx context.Context, semesterID uuid.UUID) ([]*entity.TimeSlot, error)
	DeleteTimeSlot(ctx context.Context, slotID uuid.UUID) error
	DeleteTimeSlotsBySemester(ctx context.Context, semesterID uuid.UUID) error
}
