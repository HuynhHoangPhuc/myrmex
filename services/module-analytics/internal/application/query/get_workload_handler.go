package query

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-analytics/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-analytics/internal/infrastructure/persistence"
)

// GetWorkloadQuery is the input for the workload query.
type GetWorkloadQuery struct {
	// SemesterID filters results; zero value means all semesters.
	SemesterID uuid.UUID
}

// GetWorkloadHandler returns teacher workload stats.
type GetWorkloadHandler struct {
	repo *persistence.AnalyticsRepository
}

// NewGetWorkloadHandler creates a GetWorkloadHandler.
func NewGetWorkloadHandler(repo *persistence.AnalyticsRepository) *GetWorkloadHandler {
	return &GetWorkloadHandler{repo: repo}
}

// Handle executes the workload query and returns results.
func (h *GetWorkloadHandler) Handle(ctx context.Context, q GetWorkloadQuery) ([]entity.WorkloadStat, error) {
	stats, err := h.repo.GetWorkloadStats(ctx, q.SemesterID)
	if err != nil {
		return nil, fmt.Errorf("get workload stats: %w", err)
	}
	return stats, nil
}
