package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/myrmex-erp/myrmex/services/module-timetable/internal/domain/entity"
)

// RoomRepository defines persistence operations for Room entities.
type RoomRepository interface {
	Create(ctx context.Context, r *entity.Room) (*entity.Room, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Room, error)
	List(ctx context.Context, limit, offset int32) ([]*entity.Room, error)
	Count(ctx context.Context) (int64, error)
}
