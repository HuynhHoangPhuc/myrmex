package persistence

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/myrmex-erp/myrmex/services/module-hr/internal/domain/entity"
	"github.com/myrmex-erp/myrmex/services/module-hr/internal/infrastructure/persistence/sqlc"
)

// TeacherRepositoryImpl implements domain.TeacherRepository using sqlc-generated queries.
type TeacherRepositoryImpl struct {
	queries *sqlc.Queries
}

func NewTeacherRepository(queries *sqlc.Queries) *TeacherRepositoryImpl {
	return &TeacherRepositoryImpl{queries: queries}
}

func (r *TeacherRepositoryImpl) Create(ctx context.Context, t *entity.Teacher) (*entity.Teacher, error) {
	row, err := r.queries.CreateTeacher(ctx, sqlc.CreateTeacherParams{
		EmployeeCode:    t.EmployeeCode,
		FullName:        t.FullName,
		Email:           t.Email,
		Phone:           pgtype.Text{String: t.Phone, Valid: t.Phone != ""},
		DepartmentID:    uuidOptToPgtype(t.DepartmentID),
		MaxHoursPerWeek: int32(t.MaxHoursPerWeek),
		Title:           pgtype.Text{String: t.Title, Valid: t.Title != ""},
	})
	if err != nil {
		return nil, fmt.Errorf("insert teacher: %w", err)
	}
	return hrTeacherToEntity(row), nil
}

func (r *TeacherRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.Teacher, error) {
	row, err := r.queries.GetTeacherByID(ctx, uuidToPgtype(id))
	if err != nil {
		return nil, fmt.Errorf("get teacher by id: %w", err)
	}
	return hrTeacherToEntity(row), nil
}

func (r *TeacherRepositoryImpl) List(ctx context.Context, limit, offset int32) ([]*entity.Teacher, error) {
	rows, err := r.queries.ListTeachers(ctx, sqlc.ListTeachersParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, fmt.Errorf("list teachers: %w", err)
	}
	return hrTeachersToEntities(rows), nil
}

func (r *TeacherRepositoryImpl) ListByDepartment(ctx context.Context, deptID uuid.UUID, limit, offset int32) ([]*entity.Teacher, error) {
	rows, err := r.queries.ListTeachersByDepartment(ctx, sqlc.ListTeachersByDepartmentParams{
		DepartmentID: uuidToPgtype(deptID),
		Limit:        limit,
		Offset:       offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list teachers by department: %w", err)
	}
	return hrTeachersToEntities(rows), nil
}

func (r *TeacherRepositoryImpl) Count(ctx context.Context) (int64, error) {
	return r.queries.CountTeachers(ctx)
}

func (r *TeacherRepositoryImpl) Update(ctx context.Context, t *entity.Teacher) (*entity.Teacher, error) {
	row, err := r.queries.UpdateTeacher(ctx, sqlc.UpdateTeacherParams{
		ID:              uuidToPgtype(t.ID),
		FullName:        t.FullName,
		Email:           t.Email,
		Phone:           pgtype.Text{String: t.Phone, Valid: t.Phone != ""},
		DepartmentID:    uuidOptToPgtype(t.DepartmentID),
		MaxHoursPerWeek: int32(t.MaxHoursPerWeek),
		Title:           pgtype.Text{String: t.Title, Valid: t.Title != ""},
		IsActive:        t.IsActive,
	})
	if err != nil {
		return nil, fmt.Errorf("update teacher: %w", err)
	}
	return hrTeacherToEntity(row), nil
}

func (r *TeacherRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteTeacher(ctx, uuidToPgtype(id))
}

func (r *TeacherRepositoryImpl) SearchByName(ctx context.Context, name string, limit, offset int32) ([]*entity.Teacher, error) {
	rows, err := r.queries.SearchTeachersByName(ctx, sqlc.SearchTeachersByNameParams{
		Column1: pgtype.Text{String: name, Valid: true},
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, fmt.Errorf("search teachers by name: %w", err)
	}
	return hrTeachersToEntities(rows), nil
}

func (r *TeacherRepositoryImpl) SearchBySpecialization(ctx context.Context, spec string) ([]*entity.Teacher, error) {
	rows, err := r.queries.SearchTeachersBySpecialization(ctx, spec)
	if err != nil {
		return nil, fmt.Errorf("search teachers by specialization: %w", err)
	}
	return hrTeachersToEntities(rows), nil
}

func (r *TeacherRepositoryImpl) AddSpecialization(ctx context.Context, teacherID uuid.UUID, spec string) error {
	return r.queries.AddSpecialization(ctx, sqlc.AddSpecializationParams{
		TeacherID:      uuidToPgtype(teacherID),
		Specialization: spec,
	})
}

func (r *TeacherRepositoryImpl) RemoveSpecialization(ctx context.Context, teacherID uuid.UUID, spec string) error {
	return r.queries.RemoveSpecialization(ctx, sqlc.RemoveSpecializationParams{
		TeacherID:      uuidToPgtype(teacherID),
		Specialization: spec,
	})
}

func (r *TeacherRepositoryImpl) ListSpecializations(ctx context.Context, teacherID uuid.UUID) ([]string, error) {
	return r.queries.ListSpecializations(ctx, uuidToPgtype(teacherID))
}

func (r *TeacherRepositoryImpl) UpsertAvailability(ctx context.Context, slot *entity.Availability) (*entity.Availability, error) {
	row, err := r.queries.UpsertAvailabilitySlot(ctx, sqlc.UpsertAvailabilitySlotParams{
		TeacherID:   uuidToPgtype(slot.TeacherID),
		DayOfWeek:   int32(slot.DayOfWeek),
		StartPeriod: int32(slot.StartPeriod),
		EndPeriod:   int32(slot.EndPeriod),
	})
	if err != nil {
		return nil, fmt.Errorf("upsert availability: %w", err)
	}
	return hrAvailabilityToEntity(row), nil
}

func (r *TeacherRepositoryImpl) DeleteAvailability(ctx context.Context, teacherID uuid.UUID) error {
	return r.queries.DeleteTeacherAvailability(ctx, uuidToPgtype(teacherID))
}

func (r *TeacherRepositoryImpl) ListAvailability(ctx context.Context, teacherID uuid.UUID) ([]*entity.Availability, error) {
	rows, err := r.queries.ListTeacherAvailability(ctx, uuidToPgtype(teacherID))
	if err != nil {
		return nil, fmt.Errorf("list availability: %w", err)
	}
	result := make([]*entity.Availability, len(rows))
	for i, row := range rows {
		result[i] = hrAvailabilityToEntity(row)
	}
	return result, nil
}

// --- mapping helpers ---

func hrTeacherToEntity(t sqlc.HrTeacher) *entity.Teacher {
	e := &entity.Teacher{
		ID:              pgtypeToUUID(t.ID),
		EmployeeCode:    t.EmployeeCode,
		FullName:        t.FullName,
		Email:           t.Email,
		MaxHoursPerWeek: int(t.MaxHoursPerWeek),
		IsActive:        t.IsActive,
		CreatedAt:       t.CreatedAt.Time,
		UpdatedAt:       t.UpdatedAt.Time,
	}
	if t.Phone.Valid {
		e.Phone = t.Phone.String
	}
	if t.Title.Valid {
		e.Title = t.Title.String
	}
	if t.DepartmentID.Valid {
		id := pgtypeToUUID(t.DepartmentID)
		e.DepartmentID = &id
	}
	return e
}

func hrTeachersToEntities(rows []sqlc.HrTeacher) []*entity.Teacher {
	result := make([]*entity.Teacher, len(rows))
	for i, row := range rows {
		result[i] = hrTeacherToEntity(row)
	}
	return result
}

func hrAvailabilityToEntity(a sqlc.HrTeacherAvailability) *entity.Availability {
	return &entity.Availability{
		ID:          pgtypeToUUID(a.ID),
		TeacherID:   pgtypeToUUID(a.TeacherID),
		DayOfWeek:   int(a.DayOfWeek),
		StartPeriod: int(a.StartPeriod),
		EndPeriod:   int(a.EndPeriod),
	}
}

func uuidToPgtype(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func uuidOptToPgtype(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func pgtypeToUUID(id pgtype.UUID) uuid.UUID {
	return uuid.UUID(id.Bytes)
}
