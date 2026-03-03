package command

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
)

type mockStudentRepoForInviteCode struct {
	getErr error
}

func (m *mockStudentRepoForInviteCode) GetByID(_ context.Context, _ uuid.UUID) (*entity.Student, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return &entity.Student{
		ID:        uuid.New(),
		UserID:    nil, // unlinked student
		FullName:  "Test Student",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
	}, nil
}

func (m *mockStudentRepoForInviteCode) GetByUserID(_ context.Context, _ uuid.UUID) (*entity.Student, error) {
	panic("not implemented")
}

func (m *mockStudentRepoForInviteCode) Create(_ context.Context, _ *entity.Student) (*entity.Student, error) {
	panic("not implemented")
}

func (m *mockStudentRepoForInviteCode) List(_ context.Context, _ *uuid.UUID, _ *string, _, _ int32) ([]*entity.Student, error) {
	panic("not implemented")
}

func (m *mockStudentRepoForInviteCode) Count(_ context.Context, _ *uuid.UUID, _ *string) (int64, error) {
	panic("not implemented")
}

func (m *mockStudentRepoForInviteCode) Update(_ context.Context, _ *entity.Student) (*entity.Student, error) {
	panic("not implemented")
}

func (m *mockStudentRepoForInviteCode) LinkUser(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*entity.Student, error) {
	panic("not implemented")
}

func (m *mockStudentRepoForInviteCode) Delete(_ context.Context, _ uuid.UUID) error {
	panic("not implemented")
}

type mockInviteCodeRepoForCreate struct {
	invalidateErr error
	createErr     error
	created       *entity.InviteCode
}

func (m *mockInviteCodeRepoForCreate) Create(_ context.Context, code *entity.InviteCode) (*entity.InviteCode, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	code.ID = uuid.New()
	m.created = code
	return code, nil
}

func (m *mockInviteCodeRepoForCreate) FindByCodeHash(_ context.Context, _ string) (*entity.InviteCode, error) {
	panic("not implemented")
}

func (m *mockInviteCodeRepoForCreate) InvalidateStudentCodes(_ context.Context, _ uuid.UUID) error {
	return m.invalidateErr
}

func (m *mockInviteCodeRepoForCreate) MarkUsed(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*entity.InviteCode, error) {
	panic("not implemented")
}

