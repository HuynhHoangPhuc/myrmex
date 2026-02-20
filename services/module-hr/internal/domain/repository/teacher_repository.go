package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/myrmex-erp/myrmex/services/module-hr/internal/domain/entity"
)

// TeacherRepository defines persistence operations for the Teacher aggregate.
type TeacherRepository interface {
	Create(ctx context.Context, teacher *entity.Teacher) (*entity.Teacher, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Teacher, error)
	List(ctx context.Context, limit, offset int32) ([]*entity.Teacher, error)
	ListByDepartment(ctx context.Context, deptID uuid.UUID, limit, offset int32) ([]*entity.Teacher, error)
	Count(ctx context.Context) (int64, error)
	Update(ctx context.Context, teacher *entity.Teacher) (*entity.Teacher, error)
	Delete(ctx context.Context, id uuid.UUID) error
	SearchByName(ctx context.Context, name string, limit, offset int32) ([]*entity.Teacher, error)
	SearchBySpecialization(ctx context.Context, spec string) ([]*entity.Teacher, error)

	// Specialization management
	AddSpecialization(ctx context.Context, teacherID uuid.UUID, spec string) error
	RemoveSpecialization(ctx context.Context, teacherID uuid.UUID, spec string) error
	ListSpecializations(ctx context.Context, teacherID uuid.UUID) ([]string, error)

	// Availability management
	UpsertAvailability(ctx context.Context, slot *entity.Availability) (*entity.Availability, error)
	DeleteAvailability(ctx context.Context, teacherID uuid.UUID) error
	ListAvailability(ctx context.Context, teacherID uuid.UUID) ([]*entity.Availability, error)
}
