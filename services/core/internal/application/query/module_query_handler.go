package query

import (
	"context"

	"github.com/myrmex-erp/myrmex/services/core/internal/domain/entity"
	"github.com/myrmex-erp/myrmex/services/core/internal/domain/repository"
)

type ListModulesHandler struct {
	moduleRepo repository.ModuleRepository
}

func NewListModulesHandler(moduleRepo repository.ModuleRepository) *ListModulesHandler {
	return &ListModulesHandler{moduleRepo: moduleRepo}
}

func (h *ListModulesHandler) Handle(ctx context.Context) ([]*entity.ModuleRegistration, error) {
	return h.moduleRepo.List(ctx)
}
