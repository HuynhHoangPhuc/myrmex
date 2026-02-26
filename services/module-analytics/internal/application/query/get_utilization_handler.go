package query

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-analytics/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-analytics/internal/infrastructure/persistence"
)

// GetUtilizationQuery is the input for the utilization query.
type GetUtilizationQuery struct {
	// SemesterID filters results; zero value means all semesters.
	SemesterID uuid.UUID
}

// GetUtilizationHandler returns room/department utilization stats.
type GetUtilizationHandler struct {
	repo *persistence.AnalyticsRepository
}

// NewGetUtilizationHandler creates a GetUtilizationHandler.
func NewGetUtilizationHandler(repo *persistence.AnalyticsRepository) *GetUtilizationHandler {
	return &GetUtilizationHandler{repo: repo}
}

// Handle executes the utilization query and returns results.
func (h *GetUtilizationHandler) Handle(ctx context.Context, q GetUtilizationQuery) ([]entity.UtilizationStat, error) {
	stats, err := h.repo.GetUtilizationStats(ctx, q.SemesterID)
	if err != nil {
		return nil, fmt.Errorf("get utilization stats: %w", err)
	}
	return stats, nil
}
