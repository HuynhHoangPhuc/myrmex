package query

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	infragrpc "github.com/myrmex-erp/myrmex/services/module-timetable/internal/infrastructure/grpc"
	"github.com/myrmex-erp/myrmex/services/module-timetable/internal/domain/entity"
	"github.com/myrmex-erp/myrmex/services/module-timetable/internal/domain/service"
)

// SuggestTeachersQuery requests a ranked list of teachers for a subject/slot.
type SuggestTeachersQuery struct {
	SubjectID   uuid.UUID
	DayOfWeek   int
	StartPeriod int
	EndPeriod   int
}

// SuggestTeachersHandler fetches teachers from HR and ranks them via TeacherRanker.
type SuggestTeachersHandler struct {
	hrClient *infragrpc.HRClient
	ranker   *service.TeacherRanker
}

func NewSuggestTeachersHandler(hrClient *infragrpc.HRClient, ranker *service.TeacherRanker) *SuggestTeachersHandler {
	return &SuggestTeachersHandler{hrClient: hrClient, ranker: ranker}
}

func (h *SuggestTeachersHandler) Handle(ctx context.Context, q SuggestTeachersQuery) ([]service.TeacherRank, error) {
	teachers, err := h.hrClient.ListTeachers(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch teachers: %w", err)
	}

	// Filter to teachers available for the requested slot
	slot := &entity.TimeSlot{
		DayOfWeek:   q.DayOfWeek,
		StartPeriod: q.StartPeriod,
		EndPeriod:   q.EndPeriod,
	}

	var available []service.TeacherInfo
	for _, t := range teachers {
		avail, err := h.hrClient.GetTeacherAvailability(ctx, t.ID)
		if err != nil {
			continue
		}
		for _, a := range avail {
			if a.DayOfWeek == slot.DayOfWeek &&
				a.StartPeriod <= slot.StartPeriod &&
				a.EndPeriod >= slot.EndPeriod {
				available = append(available, t)
				break
			}
		}
	}

	if len(available) == 0 {
		// Fall back to all teachers if availability data is sparse
		available = teachers
	}

	ranked := h.ranker.RankForSubject(q.SubjectID, available, nil)
	return ranked, nil
}
