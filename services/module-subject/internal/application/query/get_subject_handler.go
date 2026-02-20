package query

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/myrmex-erp/myrmex/services/module-subject/internal/domain/entity"
	"github.com/myrmex-erp/myrmex/services/module-subject/internal/domain/repository"
)

// GetSubjectQuery identifies a subject to fetch.
type GetSubjectQuery struct {
	ID uuid.UUID
}

// GetSubjectHandler handles GetSubjectQuery.
type GetSubjectHandler struct {
	subjectRepo repository.SubjectRepository
}

// NewGetSubjectHandler constructs a GetSubjectHandler.
func NewGetSubjectHandler(subjectRepo repository.SubjectRepository) *GetSubjectHandler {
	return &GetSubjectHandler{subjectRepo: subjectRepo}
}

// Handle returns the subject with the given ID.
func (h *GetSubjectHandler) Handle(ctx context.Context, q GetSubjectQuery) (*entity.Subject, error) {
	subject, err := h.subjectRepo.GetByID(ctx, q.ID)
	if err != nil {
		return nil, fmt.Errorf("get subject: %w", err)
	}
	return subject, nil
}
