package repository

import (
    "context"
    "time"

    "github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
    "github.com/google/uuid"
)

// TranscriptRecord is the transcript read model assembled from enrollments + grades.
type TranscriptRecord struct {
    EnrollmentID     uuid.UUID
    StudentID        uuid.UUID
    SemesterID       uuid.UUID
    SubjectID        uuid.UUID
    OfferedSubjectID uuid.UUID
    Status           string
    RequestedAt      time.Time
    GradeNumeric     *float64
    GradeLetter      *string
    GradedAt         *time.Time
}

// GradeRepository persists grades and transcript queries.
type GradeRepository interface {
    Assign(ctx context.Context, grade *entity.Grade) (*entity.Grade, error)
    Update(ctx context.Context, grade *entity.Grade) (*entity.Grade, error)
    GetByID(ctx context.Context, id uuid.UUID) (*entity.Grade, error)
    MarkEnrollmentCompleted(ctx context.Context, enrollmentID uuid.UUID) error
    GetStudentTranscript(ctx context.Context, studentID uuid.UUID) ([]TranscriptRecord, error)
}
