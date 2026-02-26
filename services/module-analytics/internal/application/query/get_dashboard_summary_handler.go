package query

import (
	"context"
	"fmt"

	"github.com/HuynhHoangPhuc/myrmex/services/module-analytics/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-analytics/internal/infrastructure/persistence"
)

// GetDashboardSummaryHandler returns aggregate dashboard counts.
type GetDashboardSummaryHandler struct {
	repo *persistence.AnalyticsRepository
}

// NewGetDashboardSummaryHandler creates a GetDashboardSummaryHandler.
func NewGetDashboardSummaryHandler(repo *persistence.AnalyticsRepository) *GetDashboardSummaryHandler {
	return &GetDashboardSummaryHandler{repo: repo}
}

// Handle executes the dashboard summary query and returns results.
func (h *GetDashboardSummaryHandler) Handle(ctx context.Context) (entity.DashboardSummary, error) {
	summary, err := h.repo.GetDashboardSummary(ctx)
	if err != nil {
		return entity.DashboardSummary{}, fmt.Errorf("get dashboard summary: %w", err)
	}
	return summary, nil
}
