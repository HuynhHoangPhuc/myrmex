package query

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/myrmex-erp/myrmex/services/module-timetable/internal/domain/entity"
	"github.com/myrmex-erp/myrmex/services/module-timetable/internal/domain/repository"
)

// GetScheduleQuery requests a single schedule with all its entries.
type GetScheduleQuery struct {
	ScheduleID uuid.UUID
}

// GetScheduleResult holds a schedule and its entries.
type GetScheduleResult struct {
	Schedule *entity.Schedule
	Entries  []*entity.ScheduleEntry
}

// GetScheduleHandler executes the GetSchedule read use case.
type GetScheduleHandler struct {
	repo repository.ScheduleRepository
}

func NewGetScheduleHandler(repo repository.ScheduleRepository) *GetScheduleHandler {
	return &GetScheduleHandler{repo: repo}
}

func (h *GetScheduleHandler) Handle(ctx context.Context, q GetScheduleQuery) (*GetScheduleResult, error) {
	schedule, err := h.repo.GetByID(ctx, q.ScheduleID)
	if err != nil {
		return nil, fmt.Errorf("get schedule %s: %w", q.ScheduleID, err)
	}
	entries, err := h.repo.ListEntries(ctx, q.ScheduleID)
	if err != nil {
		return nil, fmt.Errorf("list entries for schedule %s: %w", q.ScheduleID, err)
	}
	return &GetScheduleResult{Schedule: schedule, Entries: entries}, nil
}
