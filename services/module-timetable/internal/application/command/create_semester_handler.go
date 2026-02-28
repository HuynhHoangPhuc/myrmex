package command

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/domain/repository"
)

// CreateSemesterCommand carries the data needed to create a new semester.
type CreateSemesterCommand struct {
	Name      string
	Year      int
	Term      int
	StartDate time.Time
	EndDate   time.Time
}

// CreateSemesterHandler executes the CreateSemester use case.
type CreateSemesterHandler struct {
	repo      repository.SemesterRepository
	publisher EventPublisher
}

func NewCreateSemesterHandler(repo repository.SemesterRepository, publisher EventPublisher) *CreateSemesterHandler {
	return &CreateSemesterHandler{repo: repo, publisher: publisher}
}

func (h *CreateSemesterHandler) Handle(ctx context.Context, cmd CreateSemesterCommand) (*entity.Semester, error) {
	s := &entity.Semester{
		ID:        uuid.New(),
		Name:      cmd.Name,
		Year:      cmd.Year,
		Term:      cmd.Term,
		StartDate: cmd.StartDate,
		EndDate:   cmd.EndDate,
	}
	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("validate semester: %w", err)
	}
	created, err := h.repo.Create(ctx, s)
	if err != nil {
		return nil, fmt.Errorf("persist semester: %w", err)
	}

	// Publish semester created event for analytics consumer
	_ = h.publisher.Publish(ctx, "timetable.semester.created", semesterEventPayload{
		SemesterID: created.ID.String(),
		Name:       created.Name,
		Year:       created.Year,
		Term:       created.Term,
		StartDate:  created.StartDate.Format(time.DateOnly),
		EndDate:    created.EndDate.Format(time.DateOnly),
	})

	return created, nil
}

// semesterEventPayload is the event shape published on semester creation.
type semesterEventPayload struct {
	SemesterID string `json:"semester_id"`
	Name       string `json:"name"`
	Year       int    `json:"year"`
	Term       int    `json:"term"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
}
