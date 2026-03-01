package persistence

import (
	"context"
	"fmt"

	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/repository"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/infrastructure/persistence/sqlc"
	"github.com/google/uuid"
)

// GradeRepositoryImpl implements GradeRepository using sqlc queries.
type GradeRepositoryImpl struct {
	queries *sqlc.Queries
}

func NewGradeRepository(queries *sqlc.Queries) *GradeRepositoryImpl {
	return &GradeRepositoryImpl{queries: queries}
}

func (r *GradeRepositoryImpl) Assign(ctx context.Context, grade *entity.Grade) (*entity.Grade, error) {
	row, err := r.queries.AssignGrade(ctx, sqlc.AssignGradeParams{
		ID:           uuidToPgtype(grade.ID),
		EnrollmentID: uuidToPgtype(grade.EnrollmentID),
		GradeNumeric: grade.GradeNumeric,
		GradedBy:     uuidToPgtype(grade.GradedBy),
		Notes:        grade.Notes,
	})
	if err != nil {
		return nil, fmt.Errorf("assign grade: %w", err)
	}
	return assignGradeRowToEntity(row), nil
}

func (r *GradeRepositoryImpl) Update(ctx context.Context, grade *entity.Grade) (*entity.Grade, error) {
	row, err := r.queries.UpdateGrade(ctx, sqlc.UpdateGradeParams{
		GradeNumeric: grade.GradeNumeric,
		GradedBy:     uuidToPgtype(grade.GradedBy),
		Notes:        grade.Notes,
		ID:           uuidToPgtype(grade.ID),
	})
	if err != nil {
		return nil, fmt.Errorf("update grade: %w", err)
	}
	return updateGradeRowToEntity(row), nil
}

func (r *GradeRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.Grade, error) {
	row, err := r.queries.GetGrade(ctx, uuidToPgtype(id))
	if err != nil {
		return nil, fmt.Errorf("get grade: %w", err)
	}
	return getGradeRowToEntity(row), nil
}

func (r *GradeRepositoryImpl) MarkEnrollmentCompleted(ctx context.Context, enrollmentID uuid.UUID) error {
	if err := r.queries.MarkEnrollmentCompleted(ctx, uuidToPgtype(enrollmentID)); err != nil {
		return fmt.Errorf("mark enrollment completed: %w", err)
	}
	return nil
}

func (r *GradeRepositoryImpl) GetStudentTranscript(ctx context.Context, studentID uuid.UUID) ([]repository.TranscriptRecord, error) {
	rows, err := r.queries.GetStudentTranscript(ctx, uuidToPgtype(studentID))
	if err != nil {
		return nil, fmt.Errorf("get student transcript: %w", err)
	}
	result := make([]repository.TranscriptRecord, len(rows))
	for i, row := range rows {
		record := repository.TranscriptRecord{
			EnrollmentID:     pgtypeToUUID(row.EnrollmentID),
			StudentID:        pgtypeToUUID(row.StudentID),
			SemesterID:       pgtypeToUUID(row.SemesterID),
			SubjectID:        pgtypeToUUID(row.SubjectID),
			OfferedSubjectID: pgtypeToUUID(row.OfferedSubjectID),
			Status:           row.Status,
			RequestedAt:      row.RequestedAt.Time,
		}
		if value, ok := row.GradeNumeric.(float64); ok {
			gradeNumeric := value
			record.GradeNumeric = &gradeNumeric
		}
		if row.GradeLetter.Valid {
			gradeLetter := row.GradeLetter.String
			record.GradeLetter = &gradeLetter
		}
		if row.GradedAt.Valid {
			gradedAt := row.GradedAt.Time
			record.GradedAt = &gradedAt
		}
		result[i] = record
	}
	return result, nil
}

func assignGradeRowToEntity(row sqlc.AssignGradeRow) *entity.Grade {
	return &entity.Grade{
		ID:           pgtypeToUUID(row.ID),
		EnrollmentID: pgtypeToUUID(row.EnrollmentID),
		GradeNumeric: row.GradeNumeric,
		GradeLetter:  row.GradeLetter.String,
		GradedBy:     pgtypeToUUID(row.GradedBy),
		GradedAt:     row.GradedAt.Time,
		Notes:        row.Notes,
	}
}

func updateGradeRowToEntity(row sqlc.UpdateGradeRow) *entity.Grade {
	return &entity.Grade{
		ID:           pgtypeToUUID(row.ID),
		EnrollmentID: pgtypeToUUID(row.EnrollmentID),
		GradeNumeric: row.GradeNumeric,
		GradeLetter:  row.GradeLetter.String,
		GradedBy:     pgtypeToUUID(row.GradedBy),
		GradedAt:     row.GradedAt.Time,
		Notes:        row.Notes,
	}
}

func getGradeRowToEntity(row sqlc.GetGradeRow) *entity.Grade {
	return &entity.Grade{
		ID:           pgtypeToUUID(row.ID),
		EnrollmentID: pgtypeToUUID(row.EnrollmentID),
		GradeNumeric: row.GradeNumeric,
		GradeLetter:  row.GradeLetter.String,
		GradedBy:     pgtypeToUUID(row.GradedBy),
		GradedAt:     row.GradedAt.Time,
		Notes:        row.Notes,
	}
}
