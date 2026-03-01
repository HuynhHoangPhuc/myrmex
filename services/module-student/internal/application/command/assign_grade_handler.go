package command

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/HuynhHoangPhuc/myrmex/pkg/cache"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/repository"
	"github.com/google/uuid"
)

func transcriptCacheKey(studentID uuid.UUID) string {
	return fmt.Sprintf("student:transcript:%s", studentID)
}

// AssignGradeCommand carries input for assigning a grade.
type AssignGradeCommand struct {
	EnrollmentID uuid.UUID
	GradeNumeric float64
	GradedBy     uuid.UUID
	Notes        string
}

// AssignGradeHandler writes a grade and completes the enrollment.
type AssignGradeHandler struct {
	enrollments repository.EnrollmentRepository
	grades      repository.GradeRepository
	cache       cache.Cache
	publisher   EventPublisher
}

func NewAssignGradeHandler(
	enrollments repository.EnrollmentRepository,
	grades repository.GradeRepository,
	cache cache.Cache,
	publisher EventPublisher,
) *AssignGradeHandler {
	return &AssignGradeHandler{
		enrollments: enrollments,
		grades:      grades,
		cache:       cache,
		publisher:   publisher,
	}
}

func (h *AssignGradeHandler) Handle(ctx context.Context, cmd AssignGradeCommand) (*entity.Grade, error) {
	if h.enrollments == nil || h.grades == nil {
		return nil, fmt.Errorf("repositories are required")
	}

	enrollment, err := h.enrollments.GetByID(ctx, cmd.EnrollmentID)
	if err != nil {
		return nil, fmt.Errorf("get enrollment: %w", err)
	}
	if enrollment.Status != entity.EnrollmentStatusApproved && enrollment.Status != entity.EnrollmentStatusCompleted {
		return nil, fmt.Errorf("enrollment must be approved before grading")
	}

	grade := &entity.Grade{
		ID:           uuid.New(),
		EnrollmentID: cmd.EnrollmentID,
		GradeNumeric: cmd.GradeNumeric,
		GradedBy:     cmd.GradedBy,
		Notes:        cmd.Notes,
	}
	if err := grade.Validate(); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	created, err := h.grades.Assign(ctx, grade)
	if err != nil {
		return nil, fmt.Errorf("assign grade: %w", err)
	}
	if err := h.grades.MarkEnrollmentCompleted(ctx, cmd.EnrollmentID); err != nil {
		return nil, fmt.Errorf("mark enrollment completed: %w", err)
	}
	if h.cache != nil {
		_ = h.cache.Delete(ctx, transcriptCacheKey(enrollment.StudentID))
	}

	payload, _ := json.Marshal(map[string]string{
		"grade_id":       created.ID.String(),
		"enrollment_id":  created.EnrollmentID.String(),
		"student_id":     enrollment.StudentID.String(),
		"grade_letter":   created.GradeLetter,
	})
	_ = h.enrollments.AppendEvent(ctx, created.ID, "student.grade_assigned", payload)
	if h.publisher != nil {
		_ = h.publisher.Publish(ctx, "student.grade_assigned", created)
	}
	return created, nil
}
