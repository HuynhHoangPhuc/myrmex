package query

import (
	"context"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/domain/entity"
)

type mockTeacherRepo struct {
	teacher          *entity.Teacher
	getErr           error
	specializations  []string
	listSpecsErr     error
	listByDepartment []*entity.Teacher
	listByDeptErr    error
	listAll          []*entity.Teacher
	listAllErr       error
	totalCount       int64
	countErr         error
}

func (m *mockTeacherRepo) Create(_ context.Context, _ *entity.Teacher) (*entity.Teacher, error) {
	panic("not implemented")
}

func (m *mockTeacherRepo) GetByID(_ context.Context, _ uuid.UUID) (*entity.Teacher, error) {
	return m.teacher, m.getErr
}

func (m *mockTeacherRepo) List(_ context.Context, _, _ int32) ([]*entity.Teacher, error) {
	return m.listAll, m.listAllErr
}

func (m *mockTeacherRepo) ListByDepartment(_ context.Context, _ uuid.UUID, _, _ int32) ([]*entity.Teacher, error) {
	return m.listByDepartment, m.listByDeptErr
}

func (m *mockTeacherRepo) Count(_ context.Context) (int64, error) {
	return m.totalCount, m.countErr
}

func (m *mockTeacherRepo) Update(_ context.Context, _ *entity.Teacher) (*entity.Teacher, error) {
	panic("not implemented")
}

func (m *mockTeacherRepo) Delete(_ context.Context, _ uuid.UUID) error {
	panic("not implemented")
}

func (m *mockTeacherRepo) SearchByName(_ context.Context, _ string, _, _ int32) ([]*entity.Teacher, error) {
	panic("not implemented")
}

func (m *mockTeacherRepo) SearchBySpecialization(_ context.Context, _ string) ([]*entity.Teacher, error) {
	panic("not implemented")
}

func (m *mockTeacherRepo) AddSpecialization(_ context.Context, _ uuid.UUID, _ string) error {
	panic("not implemented")
}

func (m *mockTeacherRepo) RemoveSpecialization(_ context.Context, _ uuid.UUID, _ string) error {
	panic("not implemented")
}

func (m *mockTeacherRepo) ListSpecializations(_ context.Context, _ uuid.UUID) ([]string, error) {
	return m.specializations, m.listSpecsErr
}

func (m *mockTeacherRepo) UpsertAvailability(_ context.Context, _ *entity.Availability) (*entity.Availability, error) {
	panic("not implemented")
}

func (m *mockTeacherRepo) DeleteAvailability(_ context.Context, _ uuid.UUID) error {
	panic("not implemented")
}

func (m *mockTeacherRepo) ListAvailability(_ context.Context, _ uuid.UUID) ([]*entity.Availability, error) {
	panic("not implemented")
}
