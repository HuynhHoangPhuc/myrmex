package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/entity"
)

type mockPrereqRepo struct {
	edges   []*entity.Prerequisite
	listErr error
	called  int
}

func mustUUID(n int) uuid.UUID {
	b := [16]byte{}
	b[15] = byte(n)
	return uuid.UUID(b)
}

func edge(subjectID, prereqID int) *entity.Prerequisite {
	return &entity.Prerequisite{
		SubjectID:      mustUUID(subjectID),
		PrerequisiteID: mustUUID(prereqID),
	}
}

func (m *mockPrereqRepo) ListAll(_ context.Context) ([]*entity.Prerequisite, error) {
	m.called++
	return m.edges, m.listErr
}

func (m *mockPrereqRepo) Add(_ context.Context, _ *entity.Prerequisite) (*entity.Prerequisite, error) {
	panic("not implemented")
}

func (m *mockPrereqRepo) Remove(_ context.Context, _, _ uuid.UUID) error {
	panic("not implemented")
}

func (m *mockPrereqRepo) ListBySubject(_ context.Context, _ uuid.UUID) ([]*entity.Prerequisite, error) {
	panic("not implemented")
}

func (m *mockPrereqRepo) Get(_ context.Context, _, _ uuid.UUID) (*entity.Prerequisite, error) {
	panic("not implemented")
}
