package query

import (
	"context"
	"fmt"

	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/repository"
)

// FullDAGResult holds all nodes and edges for the prerequisite DAG.
type FullDAGResult struct {
	Subjects     []*entity.Subject
	Prerequisites []*entity.Prerequisite
}

// GetFullDAGHandler returns all subjects and prerequisite edges in one call.
// Avoids N+1 queries on the frontend.
type GetFullDAGHandler struct {
	subjectRepo repository.SubjectRepository
	prereqRepo  repository.PrerequisiteRepository
}

// NewGetFullDAGHandler constructs a GetFullDAGHandler.
func NewGetFullDAGHandler(subjectRepo repository.SubjectRepository, prereqRepo repository.PrerequisiteRepository) *GetFullDAGHandler {
	return &GetFullDAGHandler{subjectRepo: subjectRepo, prereqRepo: prereqRepo}
}

// Handle fetches all subjects and all prerequisite edges concurrently.
func (h *GetFullDAGHandler) Handle(ctx context.Context) (*FullDAGResult, error) {
	// Use a practical upper limit; the system supports ~100-500 subjects.
	subjects, err := h.subjectRepo.List(ctx, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("list subjects: %w", err)
	}

	prereqs, err := h.prereqRepo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("list prerequisites: %w", err)
	}

	return &FullDAGResult{Subjects: subjects, Prerequisites: prereqs}, nil
}
