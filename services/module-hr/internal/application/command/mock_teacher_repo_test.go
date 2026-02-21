package command

import (
	"context"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/domain/entity"
)

type mockTeacherRepo struct {
	created   *entity.Teacher
	createErr error

	updated   *entity.Teacher
	updateErr error
	deleteErr error
}

func mustUUID(n int) uuid.UUID {
	b := [16]byte{}
	b[15] = byte(n)
	return uuid.UUID(b)
}

func (m *mockTeacherRepo) Create(_ context.Context, teacher *entity.Teacher) (*entity.Teacher, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	teacher.ID = mustUUID(1)
	m.created = teacher
	return teacher, nil
}

func (m *mockTeacherRepo) GetByID(_ context.Context, _ uuid.UUID) (*entity.Teacher, error) {
	panic("not implemented")
}

func (m *mockTeacherRepo) List(_ context.Context, _, _ int32) ([]*entity.Teacher, error) {
	panic("not implemented")
}

func (m *mockTeacherRepo) ListByDepartment(_ context.Context, _ uuid.UUID, _, _ int32) ([]*entity.Teacher, error) {
	panic("not implemented")
}

func (m *mockTeacherRepo) Count(_ context.Context) (int64, error) {
	panic("not implemented")
}

func (m *mockTeacherRepo) Update(_ context.Context, teacher *entity.Teacher) (*entity.Teacher, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	if m.updated != nil {
		return m.updated, nil
	}
	m.updated = teacher
	return teacher, nil
}

func (m *mockTeacherRepo) Delete(_ context.Context, _ uuid.UUID) error {
	return m.deleteErr
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
	panic("not implemented")
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
