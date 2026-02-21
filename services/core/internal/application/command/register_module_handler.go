package command

import (
	"context"
	"fmt"

	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/domain/repository"
)

type RegisterModuleCommand struct {
	Name        string
	Version     string
	GRPCAddress string
}

type RegisterModuleHandler struct {
	moduleRepo repository.ModuleRepository
}

func NewRegisterModuleHandler(moduleRepo repository.ModuleRepository) *RegisterModuleHandler {
	return &RegisterModuleHandler{moduleRepo: moduleRepo}
}

func (h *RegisterModuleHandler) Handle(ctx context.Context, cmd RegisterModuleCommand) (*entity.ModuleRegistration, error) {
	mod := &entity.ModuleRegistration{
		Name:        cmd.Name,
		Version:     cmd.Version,
		GRPCAddress: cmd.GRPCAddress,
	}
	if err := mod.Validate(); err != nil {
		return nil, fmt.Errorf("validate module: %w", err)
	}
	return h.moduleRepo.Register(ctx, mod)
}
