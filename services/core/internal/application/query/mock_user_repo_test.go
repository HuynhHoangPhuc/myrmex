package query

import (
	"context"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/domain/entity"
)

type mockUserRepo struct {
	getUser *entity.User
	getErr  error

	getByEmailUser *entity.User
	getByEmailErr  error

	listUsers      []*entity.User
	listErr        error
	lastListLimit  int32
	lastListOffset int32

	countTotal int64
	countErr   error
}

func (m *mockUserRepo) Create(_ context.Context, _ *entity.User) (*entity.User, error) {
	panic("not implemented")
}

func (m *mockUserRepo) GetByID(_ context.Context, _ uuid.UUID) (*entity.User, error) {
	return m.getUser, m.getErr
}

func (m *mockUserRepo) GetByEmail(_ context.Context, _ string) (*entity.User, error) {
	return m.getByEmailUser, m.getByEmailErr
}

func (m *mockUserRepo) List(_ context.Context, limit, offset int32) ([]*entity.User, error) {
	m.lastListLimit = limit
	m.lastListOffset = offset
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.listUsers, nil
}

func (m *mockUserRepo) Count(_ context.Context) (int64, error) {
	return m.countTotal, m.countErr
}

func (m *mockUserRepo) Update(_ context.Context, _ *entity.User) (*entity.User, error) {
	panic("not implemented")
}

func (m *mockUserRepo) Delete(_ context.Context, _ uuid.UUID) error {
	panic("not implemented")
}

func (m *mockUserRepo) UpdateRole(_ context.Context, _ uuid.UUID, _ string, _ *uuid.UUID) (*entity.User, error) {
	panic("not implemented")
}

func (m *mockUserRepo) GetTeacherIDByUserID(_ context.Context, _ uuid.UUID) (string, error) {
	return "", nil // no teacher record by default
}

func (m *mockUserRepo) GetByOAuth(_ context.Context, _, _ string) (*entity.User, error) {
	return nil, nil
}

func (m *mockUserRepo) UpsertOAuthUser(_ context.Context, _, _, _, _, _, _ string) (*entity.User, error) {
	panic("not implemented")
}
