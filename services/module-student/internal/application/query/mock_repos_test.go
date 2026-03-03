package query

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
)

type mockInviteCodeRepoForValidate struct {
	findErr error
	code    *entity.InviteCode
}

func (m *mockInviteCodeRepoForValidate) FindByCodeHash(_ context.Context, _ string) (*entity.InviteCode, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	if m.code != nil {
		return m.code, nil
	}
	panic("code not set")
}

func (m *mockInviteCodeRepoForValidate) Create(_ context.Context, _ *entity.InviteCode) (*entity.InviteCode, error) {
	panic("not implemented")
}

func (m *mockInviteCodeRepoForValidate) InvalidateStudentCodes(_ context.Context, _ uuid.UUID) error {
	panic("not implemented")
}

func (m *mockInviteCodeRepoForValidate) MarkUsed(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*entity.InviteCode, error) {
	panic("not implemented")
}

type mockStudentRepoForValidate struct {
	getErr  error
	student *entity.Student
}

func (m *mockStudentRepoForValidate) GetByID(_ context.Context, _ uuid.UUID) (*entity.Student, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if m.student != nil {
		return m.student, nil
	}
	return &entity.Student{
		ID:        uuid.New(),
		FullName:  "Test Student",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
	}, nil
}

func (m *mockStudentRepoForValidate) GetByUserID(_ context.Context, _ uuid.UUID) (*entity.Student, error) {
	panic("not implemented")
}

func (m *mockStudentRepoForValidate) Create(_ context.Context, _ *entity.Student) (*entity.Student, error) {
	panic("not implemented")
}

func (m *mockStudentRepoForValidate) List(_ context.Context, _ *uuid.UUID, _ *string, _, _ int32) ([]*entity.Student, error) {
	panic("not implemented")
}

func (m *mockStudentRepoForValidate) Count(_ context.Context, _ *uuid.UUID, _ *string) (int64, error) {
	panic("not implemented")
}

func (m *mockStudentRepoForValidate) Update(_ context.Context, _ *entity.Student) (*entity.Student, error) {
	panic("not implemented")
}

func (m *mockStudentRepoForValidate) LinkUser(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*entity.Student, error) {
	panic("not implemented")
}

func (m *mockStudentRepoForValidate) Delete(_ context.Context, _ uuid.UUID) error {
	panic("not implemented")
}
