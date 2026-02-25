package query

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/entity"
)

var errNotFound = errors.New("subject not found")

type mockSubjectRepo struct {
	getByID          *entity.Subject
	getByIDErr       error
	listAll          []*entity.Subject
	listAllErr       error
	listByDept       []*entity.Subject
	listByDeptErr    error
	totalCount       int64
	countErr         error
	deptCount        int64
	deptCountErr     error
}

func (m *mockSubjectRepo) Create(_ context.Context, s *entity.Subject) (*entity.Subject, error) {
	return s, nil
}

func (m *mockSubjectRepo) GetByID(_ context.Context, _ uuid.UUID) (*entity.Subject, error) {
	return m.getByID, m.getByIDErr
}

func (m *mockSubjectRepo) GetByCode(_ context.Context, _ string) (*entity.Subject, error) {
	return nil, nil
}

func (m *mockSubjectRepo) List(_ context.Context, _, _ int32) ([]*entity.Subject, error) {
	return m.listAll, m.listAllErr
}

func (m *mockSubjectRepo) ListByDepartment(_ context.Context, _ string, _, _ int32) ([]*entity.Subject, error) {
	return m.listByDept, m.listByDeptErr
}

func (m *mockSubjectRepo) Count(_ context.Context) (int64, error) {
	return m.totalCount, m.countErr
}

func (m *mockSubjectRepo) CountByDepartment(_ context.Context, _ string) (int64, error) {
	return m.deptCount, m.deptCountErr
}

func (m *mockSubjectRepo) Update(_ context.Context, s *entity.Subject) (*entity.Subject, error) {
	return s, nil
}

func (m *mockSubjectRepo) Delete(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockSubjectRepo) ListAllIDs(_ context.Context) ([]uuid.UUID, error) {
	return nil, nil
}
