package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/myrmex-erp/myrmex/services/module-hr/internal/domain/entity"
	"github.com/myrmex-erp/myrmex/services/module-hr/internal/domain/repository"
)

// AvailabilitySlot represents a single weekly time slot input.
type AvailabilitySlot struct {
	DayOfWeek   int
	StartPeriod int
	EndPeriod   int
}

// UpdateAvailabilityCommand replaces all availability slots for a teacher.
type UpdateAvailabilityCommand struct {
	TeacherID uuid.UUID
	Slots     []AvailabilitySlot
}

// UpdateAvailabilityHandler replaces teacher availability slots atomically.
type UpdateAvailabilityHandler struct {
	repo repository.TeacherRepository
}

func NewUpdateAvailabilityHandler(repo repository.TeacherRepository) *UpdateAvailabilityHandler {
	return &UpdateAvailabilityHandler{repo: repo}
}

func (h *UpdateAvailabilityHandler) Handle(ctx context.Context, cmd UpdateAvailabilityCommand) ([]*entity.Availability, error) {
	// Validate all slots first
	for _, s := range cmd.Slots {
		slot := &entity.Availability{
			TeacherID:   cmd.TeacherID,
			DayOfWeek:   s.DayOfWeek,
			StartPeriod: s.StartPeriod,
			EndPeriod:   s.EndPeriod,
		}
		if err := slot.Validate(); err != nil {
			return nil, fmt.Errorf("invalid slot: %w", err)
		}
	}

	// Clear existing slots then upsert new ones
	if err := h.repo.DeleteAvailability(ctx, cmd.TeacherID); err != nil {
		return nil, fmt.Errorf("clear availability: %w", err)
	}

	result := make([]*entity.Availability, 0, len(cmd.Slots))
	for _, s := range cmd.Slots {
		slot := &entity.Availability{
			TeacherID:   cmd.TeacherID,
			DayOfWeek:   s.DayOfWeek,
			StartPeriod: s.StartPeriod,
			EndPeriod:   s.EndPeriod,
		}
		created, err := h.repo.UpsertAvailability(ctx, slot)
		if err != nil {
			return nil, fmt.Errorf("upsert slot: %w", err)
		}
		result = append(result, created)
	}
	return result, nil
}
