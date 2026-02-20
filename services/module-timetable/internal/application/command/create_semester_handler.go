package command

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/myrmex-erp/myrmex/services/module-timetable/internal/domain/entity"
	"github.com/myrmex-erp/myrmex/services/module-timetable/internal/domain/repository"
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
	repo repository.SemesterRepository
}

func NewCreateSemesterHandler(repo repository.SemesterRepository) *CreateSemesterHandler {
	return &CreateSemesterHandler{repo: repo}
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
	return created, nil
}

// semesterEventPayload is the event shape published on semester creation.
type semesterEventPayload struct {
	SemesterID string `json:"semester_id"`
	Name       string `json:"name"`
	Year       int    `json:"year"`
	Term       int    `json:"term"`
}

func semesterCreatedPayload(s *entity.Semester) json.RawMessage {
	b, _ := json.Marshal(semesterEventPayload{
		SemesterID: s.ID.String(),
		Name:       s.Name,
		Year:       s.Year,
		Term:       s.Term,
	})
	return b
}
