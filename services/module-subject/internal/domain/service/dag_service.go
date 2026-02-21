package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/repository"
)

// DAGService manages the directed acyclic graph of subject prerequisites.
// All graph operations load edges from the repository to ensure consistency.
type DAGService struct {
	prereqRepo repository.PrerequisiteRepository
}

// NewDAGService constructs a DAGService with the given prerequisite repository.
func NewDAGService(prereqRepo repository.PrerequisiteRepository) *DAGService {
	return &DAGService{prereqRepo: prereqRepo}
}

// adjacencyList builds a map of subjectID -> list of prerequisiteIDs from edges.
// Edge direction: subjectID depends on prerequisiteID (prereq must come first).
func buildAdjacencyList(edges []*entity.Prerequisite) map[uuid.UUID][]uuid.UUID {
	adj := make(map[uuid.UUID][]uuid.UUID, len(edges))
	for _, e := range edges {
		adj[e.SubjectID] = append(adj[e.SubjectID], e.PrerequisiteID)
		// Ensure prerequisite node exists in map even if it has no deps.
		if _, ok := adj[e.PrerequisiteID]; !ok {
			adj[e.PrerequisiteID] = nil
		}
	}
	return adj
}

// ValidateNoCircularDependency checks that adding the edge (subjectID -> prerequisiteID)
// would not create a cycle in the existing prerequisite graph.
// Uses DFS-based cycle detection on the augmented graph.
func (d *DAGService) ValidateNoCircularDependency(
	ctx context.Context,
	subjectID, prerequisiteID uuid.UUID,
) error {
	// Self-reference is always invalid.
	if subjectID == prerequisiteID {
		return fmt.Errorf("cycle detected: subject cannot be its own prerequisite")
	}

	edges, err := d.prereqRepo.ListAll(ctx)
	if err != nil {
		return fmt.Errorf("load prerequisites: %w", err)
	}

	// Build adjacency list and tentatively add the proposed edge.
	adj := buildAdjacencyList(edges)
	adj[subjectID] = append(adj[subjectID], prerequisiteID)

	// DFS cycle detection using three-color marking:
	//   0 = unvisited, 1 = in-stack (gray), 2 = done (black)
	color := make(map[uuid.UUID]int, len(adj))

	var dfs func(node uuid.UUID) ([]uuid.UUID, bool)
	dfs = func(node uuid.UUID) ([]uuid.UUID, bool) {
		color[node] = 1 // gray: currently being explored
		for _, neighbor := range adj[node] {
			if color[neighbor] == 1 {
				// Back-edge found -> cycle.
				return []uuid.UUID{neighbor, node}, true
			}
			if color[neighbor] == 0 {
				if path, found := dfs(neighbor); found {
					return append(path, node), true
				}
			}
		}
		color[node] = 2 // black: fully explored
		return nil, false
	}

	for node := range adj {
		if color[node] == 0 {
			if path, found := dfs(node); found {
				return fmt.Errorf("cycle detected in prerequisite graph: %v", formatCyclePath(path))
			}
		}
	}
	return nil
}

// TopologicalSort returns all subject IDs in prerequisite-first order (Kahn / DFS post-order).
// Prerequisites appear before subjects that depend on them.
// Returns an error if a cycle exists (graph is not a DAG).
func (d *DAGService) TopologicalSort(ctx context.Context) ([]uuid.UUID, error) {
	edges, err := d.prereqRepo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("load prerequisites: %w", err)
	}

	adj := buildAdjacencyList(edges)

	// DFS post-order topological sort.
	// color: 0=unvisited, 1=in-stack, 2=done
	color := make(map[uuid.UUID]int, len(adj))
	result := make([]uuid.UUID, 0, len(adj))

	var dfs func(node uuid.UUID) error
	dfs = func(node uuid.UUID) error {
		color[node] = 1
		for _, neighbor := range adj[node] {
			if color[neighbor] == 1 {
				return fmt.Errorf("cycle detected involving subject %s", neighbor)
			}
			if color[neighbor] == 0 {
				if err := dfs(neighbor); err != nil {
					return err
				}
			}
		}
		color[node] = 2
		// Post-order: append after all dependencies are processed.
		result = append(result, node)
		return nil
	}

	for node := range adj {
		if color[node] == 0 {
			if err := dfs(node); err != nil {
				return nil, err
			}
		}
	}

	// result is in reverse topological order (leaves first via post-order DFS).
	// Prerequisites appear earlier, which is what we want (prereq-first).
	return result, nil
}

// ValidateFullDAG checks the entire prerequisite graph for cycles.
// Returns (true, nil cycle path) if the graph is a valid DAG.
func (d *DAGService) ValidateFullDAG(ctx context.Context) (isValid bool, cyclePath []string, err error) {
	edges, err := d.prereqRepo.ListAll(ctx)
	if err != nil {
		return false, nil, fmt.Errorf("load prerequisites: %w", err)
	}

	adj := buildAdjacencyList(edges)
	color := make(map[uuid.UUID]int, len(adj))

	var dfs func(node uuid.UUID) ([]uuid.UUID, bool)
	dfs = func(node uuid.UUID) ([]uuid.UUID, bool) {
		color[node] = 1
		for _, neighbor := range adj[node] {
			if color[neighbor] == 1 {
				return []uuid.UUID{neighbor, node}, true
			}
			if color[neighbor] == 0 {
				if path, found := dfs(neighbor); found {
					return append(path, node), true
				}
			}
		}
		color[node] = 2
		return nil, false
	}

	for node := range adj {
		if color[node] == 0 {
			if path, found := dfs(node); found {
				strPath := formatCyclePath(path)
				return false, strPath, nil
			}
		}
	}
	return true, nil, nil
}

// formatCyclePath converts a UUID path to string slice for error reporting.
func formatCyclePath(path []uuid.UUID) []string {
	strs := make([]string, len(path))
	for i, id := range path {
		strs[i] = id.String()
	}
	return strs
}
