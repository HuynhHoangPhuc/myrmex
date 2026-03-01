package query

import (
	"context"
	"fmt"

	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/repository"
	"github.com/google/uuid"
)

// GetStudentEnrollmentsQuery filters enrollments for one student.
type GetStudentEnrollmentsQuery struct {
	StudentID  uuid.UUID
	SemesterID *uuid.UUID
}

// GetStudentEnrollmentsHandler returns enrollments for one student.
type GetStudentEnrollmentsHandler struct {
	repo repository.EnrollmentRepository
}

func NewGetStudentEnrollmentsHandler(repo repository.EnrollmentRepository) *GetStudentEnrollmentsHandler {
	return &GetStudentEnrollmentsHandler{repo: repo}
}

func (h *GetStudentEnrollmentsHandler) Handle(ctx context.Context, q GetStudentEnrollmentsQuery) ([]*entity.EnrollmentRequest, error) {
	enrollments, err := h.repo.ListByStudent(ctx, q.StudentID, q.SemesterID)
	if err != nil {
		return nil, fmt.Errorf("get student enrollments: %w", err)
	}
	return enrollments, nil
}
