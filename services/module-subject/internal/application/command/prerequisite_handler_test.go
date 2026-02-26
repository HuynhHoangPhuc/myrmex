package command

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

// mockPrereqRepo implements repository.PrerequisiteRepository for command tests.
type mockPrereqRepo struct {
	added     *entity.Prerequisite
	addErr    error
	removeErr error
	listAll   []*entity.Prerequisite
	listErr   error
}

var _ repository.PrerequisiteRepository = (*mockPrereqRepo)(nil)

func (m *mockPrereqRepo) Add(_ context.Context, p *entity.Prerequisite) (*entity.Prerequisite, error) {
	if m.addErr != nil {
		return nil, m.addErr
	}
	m.added = p
	return p, nil
}

func (m *mockPrereqRepo) Remove(_ context.Context, _, _ uuid.UUID) error {
	return m.removeErr
}

func (m *mockPrereqRepo) ListBySubject(_ context.Context, _ uuid.UUID) ([]*entity.Prerequisite, error) {
	return nil, nil
}

func (m *mockPrereqRepo) ListAll(_ context.Context) ([]*entity.Prerequisite, error) {
	return m.listAll, m.listErr
}

func (m *mockPrereqRepo) Get(_ context.Context, _, _ uuid.UUID) (*entity.Prerequisite, error) {
	return nil, nil
}

// --- AddPrerequisiteHandler tests ---

func TestAddPrerequisiteHandler_Success(t *testing.T) {
	subjectID := uuid.New()
	prereqID := uuid.New()

	subjectRepo := &mockSubjectRepo{
		getByID: newTestSubject("CS101", "Intro", 3),
	}
	prereqRepo := &mockPrereqRepo{listAll: nil}
	dagSvc := service.NewDAGService(prereqRepo)
	h := NewAddPrerequisiteHandler(prereqRepo, subjectRepo, dagSvc, NewNoopPublisher())

	result, err := h.Handle(context.Background(), AddPrerequisiteCommand{
		SubjectID:      subjectID,
		PrerequisiteID: prereqID,
		Type:           "hard",
		Priority:       1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

func TestAddPrerequisiteHandler_DefaultType(t *testing.T) {
	subjectID := uuid.New()
	prereqID := uuid.New()

	subjectRepo := &mockSubjectRepo{getByID: newTestSubject("CS101", "Intro", 3)}
	prereqRepo := &mockPrereqRepo{}
	dagSvc := service.NewDAGService(prereqRepo)
	h := NewAddPrerequisiteHandler(prereqRepo, subjectRepo, dagSvc, NewNoopPublisher())

	result, err := h.Handle(context.Background(), AddPrerequisiteCommand{
		SubjectID:      subjectID,
		PrerequisiteID: prereqID,
		Type:           "", // should default to "hard"
		Priority:       0, // should default to 1
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Type != valueobject.PrerequisiteTypeHard {
		t.Fatalf("expected hard type, got %s", result.Type)
	}
	if result.Priority != 1 {
		t.Fatalf("expected priority=1, got %d", result.Priority)
	}
}

func TestAddPrerequisiteHandler_InvalidType(t *testing.T) {
	subjectRepo := &mockSubjectRepo{getByID: newTestSubject("CS101", "Intro", 3)}
	prereqRepo := &mockPrereqRepo{}
	dagSvc := service.NewDAGService(prereqRepo)
	h := NewAddPrerequisiteHandler(prereqRepo, subjectRepo, dagSvc, NewNoopPublisher())

	_, err := h.Handle(context.Background(), AddPrerequisiteCommand{
		SubjectID:      uuid.New(),
		PrerequisiteID: uuid.New(),
		Type:           "invalid",
		Priority:       1,
	})
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestAddPrerequisiteHandler_SelfReference(t *testing.T) {
	id := uuid.New()
	subjectRepo := &mockSubjectRepo{getByID: newTestSubject("CS101", "Intro", 3)}
	prereqRepo := &mockPrereqRepo{}
	dagSvc := service.NewDAGService(prereqRepo)
	h := NewAddPrerequisiteHandler(prereqRepo, subjectRepo, dagSvc, NewNoopPublisher())

	_, err := h.Handle(context.Background(), AddPrerequisiteCommand{
		SubjectID:      id,
		PrerequisiteID: id,
		Type:           "hard",
		Priority:       1,
	})
	if err == nil {
		t.Fatal("expected error for self-reference")
	}
}

func TestAddPrerequisiteHandler_SubjectNotFound(t *testing.T) {
	subjectRepo := &mockSubjectRepo{getByIDErr: errors.New("not found")}
	prereqRepo := &mockPrereqRepo{}
	dagSvc := service.NewDAGService(prereqRepo)
	h := NewAddPrerequisiteHandler(prereqRepo, subjectRepo, dagSvc, NewNoopPublisher())

	_, err := h.Handle(context.Background(), AddPrerequisiteCommand{
		SubjectID:      uuid.New(),
		PrerequisiteID: uuid.New(),
		Type:           "hard",
		Priority:       1,
	})
	if err == nil {
		t.Fatal("expected not-found error")
	}
}

func TestAddPrerequisiteHandler_RepoError(t *testing.T) {
	subjectRepo := &mockSubjectRepo{getByID: newTestSubject("CS101", "Intro", 3)}
	prereqRepo := &mockPrereqRepo{addErr: errors.New("db error")}
	dagSvc := service.NewDAGService(prereqRepo)
	h := NewAddPrerequisiteHandler(prereqRepo, subjectRepo, dagSvc, NewNoopPublisher())

	_, err := h.Handle(context.Background(), AddPrerequisiteCommand{
		SubjectID:      uuid.New(),
		PrerequisiteID: uuid.New(),
		Type:           "hard",
		Priority:       1,
	})
	if err == nil {
		t.Fatal("expected repo error")
	}
}

// --- RemovePrerequisiteHandler tests ---

func TestRemovePrerequisiteHandler_Success(t *testing.T) {
	repo := &mockPrereqRepo{}
	h := NewRemovePrerequisiteHandler(repo, NewNoopPublisher())

	err := h.Handle(context.Background(), RemovePrerequisiteCommand{
		SubjectID:      uuid.New(),
		PrerequisiteID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRemovePrerequisiteHandler_RepoError(t *testing.T) {
	repo := &mockPrereqRepo{removeErr: errors.New("db error")}
	h := NewRemovePrerequisiteHandler(repo, NewNoopPublisher())

	err := h.Handle(context.Background(), RemovePrerequisiteCommand{
		SubjectID:      uuid.New(),
		PrerequisiteID: uuid.New(),
	})
	if err == nil {
		t.Fatal("expected repo error")
	}
}
