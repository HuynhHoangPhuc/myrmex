package persistence

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/infrastructure/persistence/sqlc"
)

// RoomRepositoryImpl implements domain/repository.RoomRepository.
type RoomRepositoryImpl struct {
	q *sqlc.Queries
}

func NewRoomRepository(q *sqlc.Queries) *RoomRepositoryImpl {
	return &RoomRepositoryImpl{q: q}
}

func (r *RoomRepositoryImpl) Create(ctx context.Context, room *entity.Room) (*entity.Room, error) {
	row, err := r.q.CreateRoom(ctx, sqlc.CreateRoomParams{
		Name:     room.Name,
		Capacity: int32(room.Capacity),
		Type:     room.Type,
		Features: room.Features,
	})
	if err != nil {
		return nil, fmt.Errorf("create room: %w", err)
	}
	return roomToEntity(row), nil
}

func (r *RoomRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.Room, error) {
	row, err := r.q.GetRoomByID(ctx, uuidToPg(id))
	if err != nil {
		return nil, fmt.Errorf("get room: %w", err)
	}
	return roomToEntity(row), nil
}

func (r *RoomRepositoryImpl) List(ctx context.Context, limit, offset int32) ([]*entity.Room, error) {
	rows, err := r.q.ListRooms(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list rooms: %w", err)
	}
	result := make([]*entity.Room, len(rows))
	for i, row := range rows {
		result[i] = roomToEntity(row)
	}
	return result, nil
}

func (r *RoomRepositoryImpl) Count(ctx context.Context) (int64, error) {
	return r.q.CountRooms(ctx)
}

func roomToEntity(r sqlc.TimetableRoom) *entity.Room {
	return &entity.Room{
		ID:       pgToUUID(r.ID),
		Name:     r.Name,
		Capacity: int(r.Capacity),
		Type:     r.Type,
		Features: r.Features,
		IsActive: r.IsActive,
	}
}
