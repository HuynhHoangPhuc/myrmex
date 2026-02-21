package persistence

import (
	"context"
	"fmt"

	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/persistence/sqlc"
)

type ModuleRepositoryImpl struct {
	queries *sqlc.Queries
}

func NewModuleRepository(queries *sqlc.Queries) *ModuleRepositoryImpl {
	return &ModuleRepositoryImpl{queries: queries}
}

func (r *ModuleRepositoryImpl) Register(ctx context.Context, mod *entity.ModuleRegistration) (*entity.ModuleRegistration, error) {
	row, err := r.queries.RegisterModule(ctx, sqlc.RegisterModuleParams{
		Name:        mod.Name,
		Version:     mod.Version,
		GrpcAddress: mod.GRPCAddress,
	})
	if err != nil {
		return nil, fmt.Errorf("register module: %w", err)
	}
	return moduleRowToEntity(row), nil
}

func (r *ModuleRepositoryImpl) Unregister(ctx context.Context, name string) error {
	return r.queries.UnregisterModule(ctx, name)
}

func (r *ModuleRepositoryImpl) GetByName(ctx context.Context, name string) (*entity.ModuleRegistration, error) {
	row, err := r.queries.GetModuleByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("get module: %w", err)
	}
	return moduleRowToEntity(row), nil
}

func (r *ModuleRepositoryImpl) List(ctx context.Context) ([]*entity.ModuleRegistration, error) {
	rows, err := r.queries.ListModules(ctx)
	if err != nil {
		return nil, fmt.Errorf("list modules: %w", err)
	}
	mods := make([]*entity.ModuleRegistration, len(rows))
	for i, row := range rows {
		mods[i] = moduleRowToEntity(row)
	}
	return mods, nil
}

func (r *ModuleRepositoryImpl) UpdateHealth(ctx context.Context, name string, status entity.HealthStatus) error {
	return r.queries.UpdateModuleHealth(ctx, sqlc.UpdateModuleHealthParams{
		Name:         name,
		HealthStatus: string(status),
	})
}

func moduleRowToEntity(row sqlc.CoreModuleRegistry) *entity.ModuleRegistration {
	mod := &entity.ModuleRegistration{
		ID:           pgtypeToUUID(row.ID),
		Name:         row.Name,
		Version:      row.Version,
		GRPCAddress:  row.GrpcAddress,
		HealthStatus: entity.HealthStatus(row.HealthStatus),
		RegisteredAt: row.RegisteredAt.Time,
	}
	if row.LastHealthCheck.Valid {
		t := row.LastHealthCheck.Time
		mod.LastHealthCheck = &t
	}
	return mod
}
