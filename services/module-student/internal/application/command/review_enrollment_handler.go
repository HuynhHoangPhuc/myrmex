package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrEnrollmentNotPending = errors.New("enrollment request is not pending")

// ReviewEnrollmentCommand carries input for enrollment approval or rejection.
type ReviewEnrollmentCommand struct {
	ID         uuid.UUID
	Approve    bool
	AdminNote  string
	ReviewedBy uuid.UUID
}

// ReviewEnrollmentHandler handles enrollment review transitions.
type ReviewEnrollmentHandler struct {
	enrollments repository.EnrollmentRepository
	publisher   EventPublisher
}

func NewReviewEnrollmentHandler(enrollments repository.EnrollmentRepository, publisher EventPublisher) *ReviewEnrollmentHandler {
	return &ReviewEnrollmentHandler{enrollments: enrollments, publisher: publisher}
}

func (h *ReviewEnrollmentHandler) Handle(ctx context.Context, cmd ReviewEnrollmentCommand) (*entity.EnrollmentRequest, error) {
	if h.enrollments == nil {
		return nil, fmt.Errorf("enrollment repository is required")
	}
	if cmd.ReviewedBy == uuid.Nil {
		return nil, fmt.Errorf("reviewed_by is required")
	}

	status := entity.EnrollmentStatusRejected
	eventType := "student.enrollment_rejected"
	if cmd.Approve {
		status = entity.EnrollmentStatusApproved
		eventType = "student.enrollment_approved"
	}

	reviewed, err := h.enrollments.Review(ctx, cmd.ID, status, entity.NormalizeNote(cmd.AdminNote), cmd.ReviewedBy)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			existing, lookupErr := h.enrollments.GetByID(ctx, cmd.ID)
			if lookupErr == nil && existing.Status != entity.EnrollmentStatusPending {
				return nil, ErrEnrollmentNotPending
			}
		}
		return nil, fmt.Errorf("review enrollment: %w", err)
	}

	payload, _ := json.Marshal(map[string]string{
		"enrollment_id": reviewed.ID.String(),
		"student_id":    reviewed.StudentID.String(),
		"reviewed_by":   cmd.ReviewedBy.String(),
		"status":        reviewed.Status,
	})
	_ = h.enrollments.AppendEvent(ctx, reviewed.ID, eventType, payload)
	if h.publisher != nil {
		_ = h.publisher.Publish(ctx, eventType, reviewed)
	}

	return reviewed, nil
}
