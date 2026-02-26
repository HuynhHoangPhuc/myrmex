package command

import (
	"context"
	"fmt"

	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/domain/repository"
)

// CreateDepartmentCommand carries input for creating a department.
type CreateDepartmentCommand struct {
	Name string
	Code string
}

// CreateDepartmentHandler handles department creation.
type CreateDepartmentHandler struct {
	repo      repository.DepartmentRepository
	publisher EventPublisher
}

func NewCreateDepartmentHandler(repo repository.DepartmentRepository, publisher EventPublisher) *CreateDepartmentHandler {
	return &CreateDepartmentHandler{repo: repo, publisher: publisher}
}

func (h *CreateDepartmentHandler) Handle(ctx context.Context, cmd CreateDepartmentCommand) (*entity.Department, error) {
	dept := &entity.Department{
		Name: cmd.Name,
		Code: cmd.Code,
	}
	if err := dept.Validate(); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}
	created, err := h.repo.Create(ctx, dept)
	if err != nil {
		return nil, fmt.Errorf("create department: %w", err)
	}
	_ = h.publisher.Publish(ctx, "hr.department.created", created)
	return created, nil
}
