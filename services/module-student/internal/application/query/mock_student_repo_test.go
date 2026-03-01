package query

import (
	"context"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
)

type mockStudentRepo struct {
	student     *entity.Student
	getErr      error
	listResults []*entity.Student
	listErr     error
	totalCount  int64
	countErr    error
}

func (m *mockStudentRepo) Create(_ context.Context, _ *entity.Student) (*entity.Student, error) {
	panic("not implemented")
}

func (m *mockStudentRepo) GetByID(_ context.Context, _ uuid.UUID) (*entity.Student, error) {
	return m.student, m.getErr
}

func (m *mockStudentRepo) GetByUserID(_ context.Context, _ uuid.UUID) (*entity.Student, error) {
	panic("not implemented")
}

func (m *mockStudentRepo) List(_ context.Context, _ *uuid.UUID, _ *string, _, _ int32) ([]*entity.Student, error) {
	return m.listResults, m.listErr
}

func (m *mockStudentRepo) Count(_ context.Context, _ *uuid.UUID, _ *string) (int64, error) {
	return m.totalCount, m.countErr
}

func (m *mockStudentRepo) Update(_ context.Context, _ *entity.Student) (*entity.Student, error) {
	panic("not implemented")
}

func (m *mockStudentRepo) LinkUser(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*entity.Student, error) {
	panic("not implemented")
}

func (m *mockStudentRepo) Delete(_ context.Context, _ uuid.UUID) error {
	panic("not implemented")
}
