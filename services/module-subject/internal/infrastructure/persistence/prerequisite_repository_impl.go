package persistence

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/myrmex-erp/myrmex/services/module-subject/internal/domain/entity"
	"github.com/myrmex-erp/myrmex/services/module-subject/internal/domain/valueobject"
	"github.com/myrmex-erp/myrmex/services/module-subject/internal/infrastructure/persistence/sqlc"
)

// PrerequisiteRepositoryImpl implements repository.PrerequisiteRepository using sqlc.
type PrerequisiteRepositoryImpl struct {
	queries *sqlc.Queries
}

// NewPrerequisiteRepository constructs a PrerequisiteRepositoryImpl.
func NewPrerequisiteRepository(queries *sqlc.Queries) *PrerequisiteRepositoryImpl {
	return &PrerequisiteRepositoryImpl{queries: queries}
}

func (r *PrerequisiteRepositoryImpl) Add(ctx context.Context, p *entity.Prerequisite) (*entity.Prerequisite, error) {
	row, err := r.queries.AddPrerequisite(ctx, sqlc.AddPrerequisiteParams{
		SubjectID:      uuidToPgtype(p.SubjectID),
		PrerequisiteID: uuidToPgtype(p.PrerequisiteID),
		Type:           p.Type.String(),
		Priority:       p.Priority,
	})
	if err != nil {
		return nil, fmt.Errorf("add prerequisite: %w", err)
	}
	return prereqRowToEntity(row), nil
}

func (r *PrerequisiteRepositoryImpl) Remove(ctx context.Context, subjectID, prerequisiteID uuid.UUID) error {
	return r.queries.RemovePrerequisite(ctx, uuidToPgtype(subjectID), uuidToPgtype(prerequisiteID))
}

func (r *PrerequisiteRepositoryImpl) ListBySubject(ctx context.Context, subjectID uuid.UUID) ([]*entity.Prerequisite, error) {
	rows, err := r.queries.ListPrerequisitesBySubject(ctx, uuidToPgtype(subjectID))
	if err != nil {
		return nil, fmt.Errorf("list prerequisites by subject: %w", err)
	}
	return prereqRowsToEntities(rows), nil
}

func (r *PrerequisiteRepositoryImpl) ListAll(ctx context.Context) ([]*entity.Prerequisite, error) {
	rows, err := r.queries.ListAllPrerequisites(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all prerequisites: %w", err)
	}
	return prereqRowsToEntities(rows), nil
}

func (r *PrerequisiteRepositoryImpl) Get(ctx context.Context, subjectID, prerequisiteID uuid.UUID) (*entity.Prerequisite, error) {
	row, err := r.queries.GetPrerequisite(ctx, uuidToPgtype(subjectID), uuidToPgtype(prerequisiteID))
	if err != nil {
		return nil, fmt.Errorf("get prerequisite: %w", err)
	}
	return prereqRowToEntity(row), nil
}

// --- helpers ---

func prereqRowToEntity(row sqlc.SubjectPrerequisite) *entity.Prerequisite {
	// Safe parse: type is constrained by DB check constraint.
	pType, _ := valueobject.ParsePrerequisiteType(row.Type)
	return &entity.Prerequisite{
		SubjectID:      pgtypeToUUID(row.SubjectID),
		PrerequisiteID: pgtypeToUUID(row.PrerequisiteID),
		Type:           pType,
		Priority:       row.Priority,
	}
}

func prereqRowsToEntities(rows []sqlc.SubjectPrerequisite) []*entity.Prerequisite {
	result := make([]*entity.Prerequisite, len(rows))
	for i, row := range rows {
		result[i] = prereqRowToEntity(row)
	}
	return result
}
