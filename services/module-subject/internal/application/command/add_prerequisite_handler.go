package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/repository"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/service"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/valueobject"
)

// AddPrerequisiteCommand carries the data to add a prerequisite edge.
type AddPrerequisiteCommand struct {
	SubjectID      uuid.UUID
	PrerequisiteID uuid.UUID
	// Type is "hard" or "soft"; defaults to "hard" if empty.
	Type     string
	Priority int32
}

// AddPrerequisiteHandler handles AddPrerequisiteCommand.
// CRITICAL: validates no circular dependency via DAGService before persisting.
type AddPrerequisiteHandler struct {
	prereqRepo  repository.PrerequisiteRepository
	subjectRepo repository.SubjectRepository
	dagService  *service.DAGService
}

// NewAddPrerequisiteHandler constructs an AddPrerequisiteHandler.
func NewAddPrerequisiteHandler(
	prereqRepo repository.PrerequisiteRepository,
	subjectRepo repository.SubjectRepository,
	dagService *service.DAGService,
) *AddPrerequisiteHandler {
	return &AddPrerequisiteHandler{
		prereqRepo:  prereqRepo,
		subjectRepo: subjectRepo,
		dagService:  dagService,
	}
}

// Handle executes the add prerequisite use case with DAG cycle validation.
func (h *AddPrerequisiteHandler) Handle(ctx context.Context, cmd AddPrerequisiteCommand) (*entity.Prerequisite, error) {
	// Default type to "hard" when not specified.
	prereqType := cmd.Type
	if prereqType == "" {
		prereqType = string(valueobject.PrerequisiteTypeHard)
	}

	pType, err := valueobject.ParsePrerequisiteType(prereqType)
	if err != nil {
		return nil, err
	}

	// Default priority to 1 when not specified.
	priority := cmd.Priority
	if priority == 0 {
		priority = 1
	}

	prereq := &entity.Prerequisite{
		SubjectID:      cmd.SubjectID,
		PrerequisiteID: cmd.PrerequisiteID,
		Type:           pType,
		Priority:       priority,
	}

	if err := prereq.Validate(); err != nil {
		return nil, fmt.Errorf("validate prerequisite: %w", err)
	}

	// Verify both subjects exist.
	if _, err := h.subjectRepo.GetByID(ctx, cmd.SubjectID); err != nil {
		return nil, fmt.Errorf("subject not found: %w", err)
	}
	if _, err := h.subjectRepo.GetByID(ctx, cmd.PrerequisiteID); err != nil {
		return nil, fmt.Errorf("prerequisite subject not found: %w", err)
	}

	// CRITICAL: validate no circular dependency before inserting.
	if err := h.dagService.ValidateNoCircularDependency(ctx, cmd.SubjectID, cmd.PrerequisiteID); err != nil {
		return nil, fmt.Errorf("circular dependency check failed: %w", err)
	}

	added, err := h.prereqRepo.Add(ctx, prereq)
	if err != nil {
		return nil, fmt.Errorf("add prerequisite: %w", err)
	}
	return added, nil
}
