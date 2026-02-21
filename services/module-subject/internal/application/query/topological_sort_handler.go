package query

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/service"
)

// TopologicalSortHandler handles topological sort queries over the DAG.
type TopologicalSortHandler struct {
	dagService *service.DAGService
}

// NewTopologicalSortHandler constructs a TopologicalSortHandler.
func NewTopologicalSortHandler(dagService *service.DAGService) *TopologicalSortHandler {
	return &TopologicalSortHandler{dagService: dagService}
}

// Handle returns subject IDs in prerequisite-first topological order.
func (h *TopologicalSortHandler) Handle(ctx context.Context) ([]uuid.UUID, error) {
	ordered, err := h.dagService.TopologicalSort(ctx)
	if err != nil {
		return nil, fmt.Errorf("topological sort: %w", err)
	}
	return ordered, nil
}
