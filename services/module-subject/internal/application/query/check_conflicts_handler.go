package query

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/repository"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/service"
)

// ConflictResult describes one subject with unmet hard prerequisites.
type ConflictResult struct {
	Subject  *entity.Subject
	Missing  []*entity.Subject
}

// CheckConflictsQuery carries the subject IDs to validate.
type CheckConflictsQuery struct {
	SubjectIDs []uuid.UUID
}

// CheckConflictsHandler finds subjects with missing hard prerequisites.
type CheckConflictsHandler struct {
	dagService  *service.DAGService
	subjectRepo repository.SubjectRepository
}

// NewCheckConflictsHandler constructs a CheckConflictsHandler.
func NewCheckConflictsHandler(dagService *service.DAGService, subjectRepo repository.SubjectRepository) *CheckConflictsHandler {
	return &CheckConflictsHandler{dagService: dagService, subjectRepo: subjectRepo}
}

// Handle returns conflicts: subjects in the given set that have hard prerequisites NOT in the set.
func (h *CheckConflictsHandler) Handle(ctx context.Context, q CheckConflictsQuery) ([]*ConflictResult, error) {
	if len(q.SubjectIDs) == 0 {
		return nil, nil
	}

	// Detect which subjects have missing hard prerequisites.
	conflictMap, err := h.dagService.CheckConflicts(ctx, q.SubjectIDs)
	if err != nil {
		return nil, fmt.Errorf("check conflicts: %w", err)
	}
	if len(conflictMap) == 0 {
		return nil, nil
	}

	// Build a lookup map for the input subjects.
	subjectMap := make(map[uuid.UUID]*entity.Subject, len(q.SubjectIDs))
	for _, id := range q.SubjectIDs {
		s, err := h.subjectRepo.GetByID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("get subject %s: %w", id, err)
		}
		subjectMap[id] = s
	}

	// Collect all unique missing prerequisite IDs and fetch them.
	missingIDSet := make(map[uuid.UUID]struct{})
	for _, missingIDs := range conflictMap {
		for _, mid := range missingIDs {
			missingIDSet[mid] = struct{}{}
		}
	}
	missingMap := make(map[uuid.UUID]*entity.Subject, len(missingIDSet))
	for mid := range missingIDSet {
		s, err := h.subjectRepo.GetByID(ctx, mid)
		if err != nil {
			// Missing prereq may not exist in subjects table; use placeholder.
			missingMap[mid] = &entity.Subject{ID: mid, Code: mid.String(), Name: "Unknown"}
			continue
		}
		missingMap[mid] = s
	}

	// Build results.
	results := make([]*ConflictResult, 0, len(conflictMap))
	for subjectID, missingIDs := range conflictMap {
		subject, ok := subjectMap[subjectID]
		if !ok {
			continue
		}
		missing := make([]*entity.Subject, 0, len(missingIDs))
		for _, mid := range missingIDs {
			if s, ok := missingMap[mid]; ok {
				missing = append(missing, s)
			}
		}
		results = append(results, &ConflictResult{Subject: subject, Missing: missing})
	}
	return results, nil
}
