package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/entity"
)

func TestDAGService_SelfLoop_NoRepoCall(t *testing.T) {
	repo := &mockPrereqRepo{}
	svc := NewDAGService(repo)
	idA := mustUUID(1)

	err := svc.ValidateNoCircularDependency(context.Background(), idA, idA)
	if err == nil {
		t.Fatal("expected error for self-loop")
	}
	if repo.called != 0 {
		t.Errorf("expected no repo calls for self-loop, got %d", repo.called)
	}
}

func TestDAGService_ValidateNoCircularDependency(t *testing.T) {
	chainABC := []*entity.Prerequisite{edge(2, 1), edge(3, 2)}

	tests := []struct {
		name    string
		edges   []*entity.Prerequisite
		subj    int
		prereq  int
		wantErr bool
	}{
		{"empty graph any edge", nil, 2, 1, false},
		{"chain extend valid", chainABC, 4, 3, false},
		{"chain creates cycle", chainABC, 1, 3, true},
		{"diamond valid", []*entity.Prerequisite{edge(2, 1), edge(3, 1), edge(4, 2)}, 4, 3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockPrereqRepo{edges: tt.edges}
			svc := NewDAGService(repo)
			err := svc.ValidateNoCircularDependency(context.Background(), mustUUID(tt.subj), mustUUID(tt.prereq))
			if (err != nil) != tt.wantErr {
				t.Errorf("err=%v wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func TestDAGService_ValidateNoCircularDependency_RepoError(t *testing.T) {
	repo := &mockPrereqRepo{listErr: errors.New("db down")}
	svc := NewDAGService(repo)

	err := svc.ValidateNoCircularDependency(context.Background(), mustUUID(1), mustUUID(2))
	if err == nil {
		t.Fatal("expected error from repo")
	}
}

func TestDAGService_TopologicalSort(t *testing.T) {
	t.Run("linear chain prereq before dependent", func(t *testing.T) {
		repo := &mockPrereqRepo{edges: []*entity.Prerequisite{edge(2, 1), edge(3, 2)}}
		result, err := NewDAGService(repo).TopologicalSort(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		pos := indexMap(result)
		if pos[mustUUID(1)] > pos[mustUUID(2)] {
			t.Error("A must come before B")
		}
		if pos[mustUUID(2)] > pos[mustUUID(3)] {
			t.Error("B must come before C")
		}
	})

	t.Run("diamond ordering", func(t *testing.T) {
		edges := []*entity.Prerequisite{edge(2, 1), edge(3, 1), edge(4, 2), edge(4, 3)}
		result, err := NewDAGService(&mockPrereqRepo{edges: edges}).TopologicalSort(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		pos := indexMap(result)
		if pos[mustUUID(1)] > pos[mustUUID(2)] {
			t.Error("A must come before B")
		}
		if pos[mustUUID(1)] > pos[mustUUID(3)] {
			t.Error("A must come before C")
		}
		if pos[mustUUID(2)] > pos[mustUUID(4)] {
			t.Error("B must come before D")
		}
		if pos[mustUUID(3)] > pos[mustUUID(4)] {
			t.Error("C must come before D")
		}
	})

	t.Run("disconnected subgraphs", func(t *testing.T) {
		edges := []*entity.Prerequisite{edge(2, 1), edge(4, 3)}
		result, err := NewDAGService(&mockPrereqRepo{edges: edges}).TopologicalSort(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 4 {
			t.Fatalf("expected 4 nodes, got %d", len(result))
		}
		expected := []uuid.UUID{mustUUID(1), mustUUID(2), mustUUID(3), mustUUID(4)}
		pos := indexMap(result)
		for _, id := range expected {
			if _, ok := pos[id]; !ok {
				t.Errorf("missing node %s", id)
			}
		}
	})

	t.Run("empty graph", func(t *testing.T) {
		result, err := NewDAGService(&mockPrereqRepo{}).TopologicalSort(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("expected empty result, got %d", len(result))
		}
	})

	t.Run("cycle returns error", func(t *testing.T) {
		edges := []*entity.Prerequisite{edge(2, 1), edge(1, 2)}
		_, err := NewDAGService(&mockPrereqRepo{edges: edges}).TopologicalSort(context.Background())
		if err == nil {
			t.Fatal("expected cycle error")
		}
	})
}

func TestDAGService_ValidateFullDAG(t *testing.T) {
	t.Run("valid DAG", func(t *testing.T) {
		repo := &mockPrereqRepo{edges: []*entity.Prerequisite{edge(2, 1), edge(3, 2)}}
		valid, path, err := NewDAGService(repo).ValidateFullDAG(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !valid {
			t.Errorf("expected valid DAG, cyclePath=%v", path)
		}
		if path != nil {
			t.Errorf("expected nil cycle path, got %v", path)
		}
	})

	t.Run("cycle detected", func(t *testing.T) {
		edges := []*entity.Prerequisite{edge(2, 1), edge(1, 2)}
		valid, path, err := NewDAGService(&mockPrereqRepo{edges: edges}).ValidateFullDAG(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if valid {
			t.Error("expected invalid DAG")
		}
		if len(path) < 2 {
			t.Errorf("expected cycle path len>=2, got %v", path)
		}
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &mockPrereqRepo{listErr: errors.New("db")}
		_, _, err := NewDAGService(repo).ValidateFullDAG(context.Background())
		if err == nil {
			t.Fatal("expected error from repo")
		}
	})
}

func indexMap(ids []uuid.UUID) map[uuid.UUID]int {
	m := make(map[uuid.UUID]int, len(ids))
	for i, id := range ids {
		m[id] = i
	}
	return m
}
