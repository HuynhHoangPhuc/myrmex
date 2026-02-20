package query

import (
	"context"
	"fmt"

	"github.com/myrmex-erp/myrmex/services/module-subject/internal/domain/service"
)

// ValidateDAGResult carries the validation outcome.
type ValidateDAGResult struct {
	IsValid   bool
	CyclePath []string
}

// ValidateDAGHandler handles full-graph DAG validation queries.
type ValidateDAGHandler struct {
	dagService *service.DAGService
}

// NewValidateDAGHandler constructs a ValidateDAGHandler.
func NewValidateDAGHandler(dagService *service.DAGService) *ValidateDAGHandler {
	return &ValidateDAGHandler{dagService: dagService}
}

// Handle runs full DAG validation and returns whether the graph is cycle-free.
func (h *ValidateDAGHandler) Handle(ctx context.Context) (*ValidateDAGResult, error) {
	isValid, cyclePath, err := h.dagService.ValidateFullDAG(ctx)
	if err != nil {
		return nil, fmt.Errorf("validate DAG: %w", err)
	}
	return &ValidateDAGResult{IsValid: isValid, CyclePath: cyclePath}, nil
}
