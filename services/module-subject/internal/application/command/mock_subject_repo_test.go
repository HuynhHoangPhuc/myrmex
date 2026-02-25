package command

import (
	"context"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/entity"
)

type mockSubjectRepo struct {
	created    *entity.Subject
	createErr  error
	getByID    *entity.Subject
	getByIDErr error
	updated    *entity.Subject
	updateErr  error
	deleteErr  error
}

func (m *mockSubjectRepo) Create(_ context.Context, s *entity.Subject) (*entity.Subject, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	s.ID = uuid.New()
	m.created = s
	return s, nil
}

func (m *mockSubjectRepo) GetByID(_ context.Context, _ uuid.UUID) (*entity.Subject, error) {
	return m.getByID, m.getByIDErr
}

func (m *mockSubjectRepo) GetByCode(_ context.Context, _ string) (*entity.Subject, error) {
	return nil, nil
}

func (m *mockSubjectRepo) List(_ context.Context, _, _ int32) ([]*entity.Subject, error) {
	return nil, nil
}

func (m *mockSubjectRepo) ListByDepartment(_ context.Context, _ string, _, _ int32) ([]*entity.Subject, error) {
	return nil, nil
}

func (m *mockSubjectRepo) Count(_ context.Context) (int64, error) {
	return 0, nil
}

func (m *mockSubjectRepo) CountByDepartment(_ context.Context, _ string) (int64, error) {
	return 0, nil
}

func (m *mockSubjectRepo) Update(_ context.Context, s *entity.Subject) (*entity.Subject, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	m.updated = s
	return s, nil
}

func (m *mockSubjectRepo) Delete(_ context.Context, _ uuid.UUID) error {
	return m.deleteErr
}

func (m *mockSubjectRepo) ListAllIDs(_ context.Context) ([]uuid.UUID, error) {
	return nil, nil
}
