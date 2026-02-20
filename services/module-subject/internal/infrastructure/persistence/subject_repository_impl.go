package persistence

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/myrmex-erp/myrmex/services/module-subject/internal/domain/entity"
	"github.com/myrmex-erp/myrmex/services/module-subject/internal/infrastructure/persistence/sqlc"
)

// SubjectRepositoryImpl implements repository.SubjectRepository using sqlc.
type SubjectRepositoryImpl struct {
	queries *sqlc.Queries
}

// NewSubjectRepository constructs a SubjectRepositoryImpl.
func NewSubjectRepository(queries *sqlc.Queries) *SubjectRepositoryImpl {
	return &SubjectRepositoryImpl{queries: queries}
}

func (r *SubjectRepositoryImpl) Create(ctx context.Context, s *entity.Subject) (*entity.Subject, error) {
	row, err := r.queries.CreateSubject(ctx, sqlc.CreateSubjectParams{
		Code:         s.Code,
		Name:         s.Name,
		Credits:      s.Credits,
		Description:  s.Description,
		DepartmentID: s.DepartmentID,
		WeeklyHours:  s.WeeklyHours,
		IsActive:     s.IsActive,
	})
	if err != nil {
		return nil, fmt.Errorf("create subject: %w", err)
	}
	return subjectRowToEntity(row), nil
}

func (r *SubjectRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.Subject, error) {
	row, err := r.queries.GetSubjectByID(ctx, uuidToPgtype(id))
	if err != nil {
		return nil, fmt.Errorf("get subject by id: %w", err)
	}
	return subjectRowToEntity(row), nil
}

func (r *SubjectRepositoryImpl) GetByCode(ctx context.Context, code string) (*entity.Subject, error) {
	row, err := r.queries.GetSubjectByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("get subject by code: %w", err)
	}
	return subjectRowToEntity(row), nil
}

func (r *SubjectRepositoryImpl) List(ctx context.Context, limit, offset int32) ([]*entity.Subject, error) {
	rows, err := r.queries.ListSubjects(ctx, sqlc.ListSubjectsParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, fmt.Errorf("list subjects: %w", err)
	}
	return subjectRowsToEntities(rows), nil
}

func (r *SubjectRepositoryImpl) ListByDepartment(ctx context.Context, deptID string, limit, offset int32) ([]*entity.Subject, error) {
	rows, err := r.queries.ListSubjectsByDepartment(ctx, sqlc.ListSubjectsByDepartmentParams{
		DepartmentID: deptID, Limit: limit, Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list subjects by department: %w", err)
	}
	return subjectRowsToEntities(rows), nil
}

func (r *SubjectRepositoryImpl) Count(ctx context.Context) (int64, error) {
	return r.queries.CountSubjects(ctx)
}

func (r *SubjectRepositoryImpl) CountByDepartment(ctx context.Context, deptID string) (int64, error) {
	return r.queries.CountSubjectsByDepartment(ctx, deptID)
}

func (r *SubjectRepositoryImpl) Update(ctx context.Context, s *entity.Subject) (*entity.Subject, error) {
	row, err := r.queries.UpdateSubject(ctx, sqlc.UpdateSubjectParams{
		ID:           uuidToPgtype(s.ID),
		Code:         pgtype.Text{String: s.Code, Valid: s.Code != ""},
		Name:         pgtype.Text{String: s.Name, Valid: s.Name != ""},
		Credits:      pgtype.Int4{Int32: s.Credits, Valid: true},
		Description:  pgtype.Text{String: s.Description, Valid: true},
		DepartmentID: pgtype.Text{String: s.DepartmentID, Valid: s.DepartmentID != ""},
		WeeklyHours:  pgtype.Int4{Int32: s.WeeklyHours, Valid: true},
		IsActive:     pgtype.Bool{Bool: s.IsActive, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("update subject: %w", err)
	}
	return subjectRowToEntity(row), nil
}

func (r *SubjectRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteSubject(ctx, uuidToPgtype(id))
}

func (r *SubjectRepositoryImpl) ListAllIDs(ctx context.Context) ([]uuid.UUID, error) {
	pgIDs, err := r.queries.ListAllSubjectIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all subject ids: %w", err)
	}
	ids := make([]uuid.UUID, len(pgIDs))
	for i, pgID := range pgIDs {
		ids[i] = pgtypeToUUID(pgID)
	}
	return ids, nil
}

// --- helpers ---

func subjectRowToEntity(row sqlc.SubjectSubject) *entity.Subject {
	return &entity.Subject{
		ID:           pgtypeToUUID(row.ID),
		Code:         row.Code,
		Name:         row.Name,
		Credits:      row.Credits,
		Description:  row.Description,
		DepartmentID: row.DepartmentID,
		WeeklyHours:  row.WeeklyHours,
		IsActive:     row.IsActive,
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}
}

func subjectRowsToEntities(rows []sqlc.SubjectSubject) []*entity.Subject {
	result := make([]*entity.Subject, len(rows))
	for i, row := range rows {
		result[i] = subjectRowToEntity(row)
	}
	return result
}

func uuidToPgtype(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func pgtypeToUUID(id pgtype.UUID) uuid.UUID {
	return uuid.UUID(id.Bytes)
}
