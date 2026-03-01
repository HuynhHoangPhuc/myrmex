package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	appservice "github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/application/service"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrDuplicateEnrollmentRequest = errors.New("duplicate enrollment request")

// PrerequisiteChecker validates enrollment prerequisites.
type PrerequisiteChecker interface {
	Check(ctx context.Context, studentID, subjectID uuid.UUID) ([]appservice.MissingPrerequisite, error)
}

// PrerequisiteViolationError reports missing prerequisite subjects.
type PrerequisiteViolationError struct {
	Missing []appservice.MissingPrerequisite
}

func (e *PrerequisiteViolationError) Error() string {
	if len(e.Missing) == 0 {
		return "missing_prerequisites"
	}
	ids := make([]string, len(e.Missing))
	for i, item := range e.Missing {
		ids[i] = item.SubjectID.String()
	}
	return "missing_prerequisites: " + strings.Join(ids, ",")
}

// RequestEnrollmentCommand carries input for creating an enrollment request.
type RequestEnrollmentCommand struct {
	StudentID        uuid.UUID
	SemesterID       uuid.UUID
	OfferedSubjectID uuid.UUID
	SubjectID        uuid.UUID
	RequestNote      string
}

// RequestEnrollmentHandler handles enrollment request creation.
type RequestEnrollmentHandler struct {
	students    repository.StudentRepository
	enrollments repository.EnrollmentRepository
	checker     PrerequisiteChecker
	publisher   EventPublisher
}

func NewRequestEnrollmentHandler(
	students repository.StudentRepository,
	enrollments repository.EnrollmentRepository,
	checker PrerequisiteChecker,
	publisher EventPublisher,
) *RequestEnrollmentHandler {
	return &RequestEnrollmentHandler{
		students:    students,
		enrollments: enrollments,
		checker:     checker,
		publisher:   publisher,
	}
}

func (h *RequestEnrollmentHandler) Handle(ctx context.Context, cmd RequestEnrollmentCommand) (*entity.EnrollmentRequest, error) {
	if h.students == nil || h.enrollments == nil {
		return nil, fmt.Errorf("repositories are required")
	}

	if _, err := h.students.GetByID(ctx, cmd.StudentID); err != nil {
		return nil, fmt.Errorf("get student: %w", err)
	}

	if h.checker != nil {
		missing, err := h.checker.Check(ctx, cmd.StudentID, cmd.SubjectID)
		if err != nil {
			return nil, fmt.Errorf("check prerequisites: %w", err)
		}
		if len(missing) > 0 {
			return nil, &PrerequisiteViolationError{Missing: missing}
		}
	}

	enrollment := &entity.EnrollmentRequest{
		ID:               uuid.New(),
		StudentID:        cmd.StudentID,
		SemesterID:       cmd.SemesterID,
		OfferedSubjectID: cmd.OfferedSubjectID,
		SubjectID:        cmd.SubjectID,
		Status:           entity.EnrollmentStatusPending,
		RequestNote:      entity.NormalizeNote(cmd.RequestNote),
	}
	if err := enrollment.Validate(); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	created, err := h.enrollments.Create(ctx, enrollment)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("%w: %v", ErrDuplicateEnrollmentRequest, err)
		}
		return nil, fmt.Errorf("create enrollment request: %w", err)
	}

	payload, _ := json.Marshal(map[string]string{
		"enrollment_id": created.ID.String(),
		"student_id":    created.StudentID.String(),
		"semester_id":   created.SemesterID.String(),
		"subject_id":    created.SubjectID.String(),
	})
	_ = h.enrollments.AppendEvent(ctx, created.ID, "student.enrollment_requested", payload)
	if h.publisher != nil {
		_ = h.publisher.Publish(ctx, "student.enrollment_requested", created)
	}

	return created, nil
}
