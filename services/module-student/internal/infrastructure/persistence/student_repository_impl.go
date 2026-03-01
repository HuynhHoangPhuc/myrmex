package persistence

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/infrastructure/persistence/sqlc"
)

// StudentRepositoryImpl implements StudentRepository using sqlc queries.
type StudentRepositoryImpl struct {
	queries *sqlc.Queries
}

func NewStudentRepository(queries *sqlc.Queries) *StudentRepositoryImpl {
	return &StudentRepositoryImpl{queries: queries}
}

func (r *StudentRepositoryImpl) Create(ctx context.Context, student *entity.Student) (*entity.Student, error) {
	row, err := r.queries.CreateStudent(ctx, sqlc.CreateStudentParams{
		StudentCode:    student.StudentCode,
		UserID:         uuidOptToPgtype(student.UserID),
		FullName:       student.FullName,
		Email:          student.Email,
		DepartmentID:   uuidToPgtype(student.DepartmentID),
		EnrollmentYear: int32(student.EnrollmentYear),
		Status:         student.Status,
		IsActive:       student.IsActive,
	})
	if err != nil {
		return nil, fmt.Errorf("insert student: %w", err)
	}
	return studentRowToEntity(row), nil
}

func (r *StudentRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.Student, error) {
	row, err := r.queries.GetStudentByID(ctx, uuidToPgtype(id))
	if err != nil {
		return nil, fmt.Errorf("get student by id: %w", err)
	}
	return studentRowToEntity(row), nil
}

func (r *StudentRepositoryImpl) List(ctx context.Context, departmentID *uuid.UUID, status *string, limit, offset int32) ([]*entity.Student, error) {
	arg := sqlc.ListStudentsParams{
		DepartmentID: uuidOptToPgtype(departmentID),
		Status:       textOpt(status),
		OffsetCount:  offset,
		LimitCount:   limit,
	}
	rows, err := r.queries.ListStudents(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("list students: %w", err)
	}
	result := make([]*entity.Student, len(rows))
	for i, row := range rows {
		result[i] = studentRowToEntity(row)
	}
	return result, nil
}

func (r *StudentRepositoryImpl) Count(ctx context.Context, departmentID *uuid.UUID, status *string) (int64, error) {
	return r.queries.CountStudents(ctx, sqlc.CountStudentsParams{
		DepartmentID: uuidOptToPgtype(departmentID),
		Status:       textOpt(status),
	})
}

func (r *StudentRepositoryImpl) Update(ctx context.Context, student *entity.Student) (*entity.Student, error) {
	row, err := r.queries.UpdateStudent(ctx, sqlc.UpdateStudentParams{
		ID:             uuidToPgtype(student.ID),
		UserID:         uuidOptToPgtype(student.UserID),
		FullName:       student.FullName,
		Email:          student.Email,
		DepartmentID:   uuidToPgtype(student.DepartmentID),
		EnrollmentYear: int32(student.EnrollmentYear),
		Status:         student.Status,
		IsActive:       student.IsActive,
	})
	if err != nil {
		return nil, fmt.Errorf("update student: %w", err)
	}
	return studentRowToEntity(row), nil
}

func (r *StudentRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := r.queries.DeleteStudent(ctx, uuidToPgtype(id)); err != nil {
		if err == pgx.ErrNoRows {
			return err
		}
		return fmt.Errorf("delete student: %w", err)
	}
	return nil
}

func studentRowToEntity(row sqlc.StudentStudent) *entity.Student {
	student := &entity.Student{
		ID:             pgtypeToUUID(row.ID),
		StudentCode:    row.StudentCode,
		FullName:       row.FullName,
		Email:          row.Email,
		DepartmentID:   pgtypeToUUID(row.DepartmentID),
		EnrollmentYear: int(row.EnrollmentYear),
		Status:         row.Status,
		IsActive:       row.IsActive,
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
	}
	if row.UserID.Valid {
		userID := pgtypeToUUID(row.UserID)
		student.UserID = &userID
	}
	return student
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

func textOpt(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *value, Valid: true}
}
