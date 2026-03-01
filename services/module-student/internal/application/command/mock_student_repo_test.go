package command

import (
	"context"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
)

type mockStudentRepo struct {
	created   *entity.Student
	createErr error

	updated   *entity.Student
	updateErr error
	deleteErr error
}

func mustStudentUUID(n int) uuid.UUID {
	b := [16]byte{}
	b[15] = byte(n)
	return uuid.UUID(b)
}

func (m *mockStudentRepo) Create(_ context.Context, student *entity.Student) (*entity.Student, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	student.ID = mustStudentUUID(1)
	m.created = student
	return student, nil
}

func (m *mockStudentRepo) GetByID(_ context.Context, _ uuid.UUID) (*entity.Student, error) {
	panic("not implemented")
}

func (m *mockStudentRepo) GetByUserID(_ context.Context, _ uuid.UUID) (*entity.Student, error) {
	panic("not implemented")
}

func (m *mockStudentRepo) List(_ context.Context, _ *uuid.UUID, _ *string, _, _ int32) ([]*entity.Student, error) {
	panic("not implemented")
}

func (m *mockStudentRepo) Count(_ context.Context, _ *uuid.UUID, _ *string) (int64, error) {
	panic("not implemented")
}

func (m *mockStudentRepo) Update(_ context.Context, student *entity.Student) (*entity.Student, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	if m.updated != nil {
		return m.updated, nil
	}
	m.updated = student
	return student, nil
}

func (m *mockStudentRepo) LinkUser(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*entity.Student, error) {
	panic("not implemented")
}

func (m *mockStudentRepo) Delete(_ context.Context, _ uuid.UUID) error {
	return m.deleteErr
}
