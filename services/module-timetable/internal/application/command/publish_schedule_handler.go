package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/myrmex-erp/myrmex/services/module-timetable/internal/domain/repository"
	"github.com/myrmex-erp/myrmex/services/module-timetable/internal/domain/valueobject"
)

// PublishScheduleCommand requests transitioning a draft schedule to published.
type PublishScheduleCommand struct {
	ScheduleID uuid.UUID
}

// PublishScheduleHandler executes the PublishSchedule use case.
type PublishScheduleHandler struct {
	repo repository.ScheduleRepository
}

func NewPublishScheduleHandler(repo repository.ScheduleRepository) *PublishScheduleHandler {
	return &PublishScheduleHandler{repo: repo}
}

func (h *PublishScheduleHandler) Handle(ctx context.Context, cmd PublishScheduleCommand) error {
	schedule, err := h.repo.GetByID(ctx, cmd.ScheduleID)
	if err != nil {
		return fmt.Errorf("get schedule: %w", err)
	}
	if err := schedule.Publish(); err != nil {
		return fmt.Errorf("publish schedule: %w", err)
	}
	if _, err := h.repo.UpdateStatus(ctx, cmd.ScheduleID, valueobject.ScheduleStatusPublished); err != nil {
		return fmt.Errorf("persist published status: %w", err)
	}
	return nil
}
