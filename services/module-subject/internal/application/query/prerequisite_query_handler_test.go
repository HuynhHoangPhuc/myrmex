package query

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/repository"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/service"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/valueobject"
)

// mockPrereqRepo implements repository.PrerequisiteRepository for query tests.
type mockPrereqRepo struct {
	listBySubject    []*entity.Prerequisite
	listBySubjectErr error
	listAll          []*entity.Prerequisite
	listAllErr       error
}

var _ repository.PrerequisiteRepository = (*mockPrereqRepo)(nil)

func (m *mockPrereqRepo) Add(_ context.Context, p *entity.Prerequisite) (*entity.Prerequisite, error) {
	return p, nil
}

func (m *mockPrereqRepo) Remove(_ context.Context, _, _ uuid.UUID) error { return nil }

func (m *mockPrereqRepo) ListBySubject(_ context.Context, _ uuid.UUID) ([]*entity.Prerequisite, error) {
	return m.listBySubject, m.listBySubjectErr
}

func (m *mockPrereqRepo) ListAll(_ context.Context) ([]*entity.Prerequisite, error) {
	return m.listAll, m.listAllErr
}

func (m *mockPrereqRepo) Get(_ context.Context, _, _ uuid.UUID) (*entity.Prerequisite, error) {
	return nil, nil
}

// --- ListPrerequisitesHandler tests ---

func TestListPrerequisitesHandler_Success(t *testing.T) {
	subjectID := uuid.New()
	prereqID := uuid.New()
	prereq := &entity.Prerequisite{
		SubjectID:      subjectID,
		PrerequisiteID: prereqID,
		Type:           valueobject.PrerequisiteTypeHard,
		Priority:       1,
	}
	repo := &mockPrereqRepo{listBySubject: []*entity.Prerequisite{prereq}}
	h := NewListPrerequisitesHandler(repo)

	result, err := h.Handle(context.Background(), ListPrerequisitesQuery{SubjectID: subjectID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 prerequisite, got %d", len(result))
	}
}

func TestListPrerequisitesHandler_Empty(t *testing.T) {
	repo := &mockPrereqRepo{listBySubject: []*entity.Prerequisite{}}
	h := NewListPrerequisitesHandler(repo)

	result, err := h.Handle(context.Background(), ListPrerequisitesQuery{SubjectID: uuid.New()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected 0 prerequisites, got %d", len(result))
	}
}

func TestListPrerequisitesHandler_RepoError(t *testing.T) {
	repo := &mockPrereqRepo{listBySubjectErr: errors.New("db error")}
	h := NewListPrerequisitesHandler(repo)

	_, err := h.Handle(context.Background(), ListPrerequisitesQuery{SubjectID: uuid.New()})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}

// --- ValidateDAGHandler tests ---

func TestValidateDAGHandler_ValidGraph(t *testing.T) {
	// A -> B -> C (no cycle)
	a, b, c := uuid.New(), uuid.New(), uuid.New()
	edges := []*entity.Prerequisite{
		{SubjectID: a, PrerequisiteID: b, Type: valueobject.PrerequisiteTypeHard, Priority: 1},
		{SubjectID: b, PrerequisiteID: c, Type: valueobject.PrerequisiteTypeHard, Priority: 1},
	}
	prereqRepo := &mockPrereqRepo{listAll: edges}
	dagSvc := service.NewDAGService(prereqRepo)
	h := NewValidateDAGHandler(dagSvc)

	result, err := h.Handle(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsValid {
		t.Fatal("expected valid DAG")
	}
	if len(result.CyclePath) != 0 {
		t.Fatalf("expected empty cycle path, got %v", result.CyclePath)
	}
}

func TestValidateDAGHandler_RepoError(t *testing.T) {
	prereqRepo := &mockPrereqRepo{listAllErr: errors.New("db error")}
	dagSvc := service.NewDAGService(prereqRepo)
	h := NewValidateDAGHandler(dagSvc)

	_, err := h.Handle(context.Background())
	if err == nil {
		t.Fatal("expected error from repo")
	}
}

// --- TopologicalSortHandler tests ---

func TestTopologicalSortHandler_Success(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	edges := []*entity.Prerequisite{
		{SubjectID: a, PrerequisiteID: b, Type: valueobject.PrerequisiteTypeHard, Priority: 1},
	}
	prereqRepo := &mockPrereqRepo{listAll: edges}
	dagSvc := service.NewDAGService(prereqRepo)
	h := NewTopologicalSortHandler(dagSvc)

	result, err := h.Handle(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) < 2 {
		t.Fatalf("expected at least 2 IDs in sort, got %d", len(result))
	}
}

func TestTopologicalSortHandler_Empty(t *testing.T) {
	prereqRepo := &mockPrereqRepo{listAll: nil}
	dagSvc := service.NewDAGService(prereqRepo)
	h := NewTopologicalSortHandler(dagSvc)

	result, err := h.Handle(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected 0 IDs, got %d", len(result))
	}
}

func TestTopologicalSortHandler_RepoError(t *testing.T) {
	prereqRepo := &mockPrereqRepo{listAllErr: errors.New("db error")}
	dagSvc := service.NewDAGService(prereqRepo)
	h := NewTopologicalSortHandler(dagSvc)

	_, err := h.Handle(context.Background())
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
