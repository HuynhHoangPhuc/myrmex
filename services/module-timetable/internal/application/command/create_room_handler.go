package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/myrmex-erp/myrmex/services/module-timetable/internal/domain/entity"
	"github.com/myrmex-erp/myrmex/services/module-timetable/internal/domain/repository"
)

// CreateRoomCommand carries data for creating a new room.
type CreateRoomCommand struct {
	Name     string
	Capacity int
	Type     string
	Features []string
}

// CreateRoomHandler executes the CreateRoom use case.
type CreateRoomHandler struct {
	repo repository.RoomRepository
}

func NewCreateRoomHandler(repo repository.RoomRepository) *CreateRoomHandler {
	return &CreateRoomHandler{repo: repo}
}

func (h *CreateRoomHandler) Handle(ctx context.Context, cmd CreateRoomCommand) (*entity.Room, error) {
	r := &entity.Room{
		ID:       uuid.New(),
		Name:     cmd.Name,
		Capacity: cmd.Capacity,
		Type:     cmd.Type,
		Features: cmd.Features,
		IsActive: true,
	}
	if err := r.Validate(); err != nil {
		return nil, fmt.Errorf("validate room: %w", err)
	}
	return h.repo.Create(ctx, r)
}
