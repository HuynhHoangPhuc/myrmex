package command

import (
	"context"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/domain/entity"
)

type mockUserRepo struct {
	created   *entity.User
	createErr error
	getUser   *entity.User
	getErr    error

	getByEmailUser *entity.User
	getByEmailErr  error

	listUsers    []*entity.User
	listErr      error
	lastListLimit  int32
	lastListOffset int32

	countTotal int64
	countErr   error

	updated   *entity.User
	updateErr error

	deletedID uuid.UUID
	deleteErr error
}

func mustUUID(n int) uuid.UUID {
	b := [16]byte{}
	b[15] = byte(n)
	return uuid.UUID(b)
}

func (m *mockUserRepo) Create(_ context.Context, user *entity.User) (*entity.User, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	user.ID = mustUUID(1)
	m.created = user
	return user, nil
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

func (m *mockUserRepo) Update(_ context.Context, user *entity.User) (*entity.User, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	m.updated = user
	return user, nil
}

func (m *mockUserRepo) Delete(_ context.Context, id uuid.UUID) error {
	m.deletedID = id
	return m.deleteErr
}
