package repository

import (
	"context"

	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/domain/entity"
)

type ModuleRepository interface {
	Register(ctx context.Context, mod *entity.ModuleRegistration) (*entity.ModuleRegistration, error)
	Unregister(ctx context.Context, name string) error
	GetByName(ctx context.Context, name string) (*entity.ModuleRegistration, error)
	List(ctx context.Context) ([]*entity.ModuleRegistration, error)
	UpdateHealth(ctx context.Context, name string, status entity.HealthStatus) error
}
