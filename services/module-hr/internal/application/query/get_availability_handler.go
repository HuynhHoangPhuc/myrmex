package query

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/myrmex-erp/myrmex/services/module-hr/internal/domain/entity"
	"github.com/myrmex-erp/myrmex/services/module-hr/internal/domain/repository"
)

// GetAvailabilityQuery fetches all availability slots for a teacher.
type GetAvailabilityQuery struct {
	TeacherID uuid.UUID
}

// GetAvailabilityHandler returns the weekly slots for a teacher.
type GetAvailabilityHandler struct {
	repo repository.TeacherRepository
}

func NewGetAvailabilityHandler(repo repository.TeacherRepository) *GetAvailabilityHandler {
	return &GetAvailabilityHandler{repo: repo}
}

func (h *GetAvailabilityHandler) Handle(ctx context.Context, q GetAvailabilityQuery) ([]*entity.Availability, error) {
	slots, err := h.repo.ListAvailability(ctx, q.TeacherID)
	if err != nil {
		return nil, fmt.Errorf("list availability: %w", err)
	}
	return slots, nil
}
