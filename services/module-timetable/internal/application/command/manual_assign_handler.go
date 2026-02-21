package command

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/domain/repository"
)

// ManualAssignCommand requests assigning a specific teacher to an existing entry.
type ManualAssignCommand struct {
	ScheduleID uuid.UUID
	EntryID    uuid.UUID
	TeacherID  uuid.UUID
}

// EventPublisher is the minimal interface the handler needs for publishing.
type EventPublisher interface {
	Publish(ctx context.Context, subject string, payload any) error
}

// ManualAssignHandler executes the ManualAssign use case.
type ManualAssignHandler struct {
	repo      repository.ScheduleRepository
	publisher EventPublisher
}

func NewManualAssignHandler(repo repository.ScheduleRepository, publisher EventPublisher) *ManualAssignHandler {
	return &ManualAssignHandler{repo: repo, publisher: publisher}
}

func (h *ManualAssignHandler) Handle(ctx context.Context, cmd ManualAssignCommand) (*entity.ScheduleEntry, error) {
	// 1. Load existing entry
	entry, err := h.repo.GetEntry(ctx, cmd.EntryID)
	if err != nil {
		return nil, fmt.Errorf("get entry: %w", err)
	}
	if entry.ScheduleID != cmd.ScheduleID {
		return nil, fmt.Errorf("entry %s does not belong to schedule %s", cmd.EntryID, cmd.ScheduleID)
	}

	// 2. Apply override
	entry.TeacherID = cmd.TeacherID
	entry.IsManualOverride = true

	// 3. Persist
	updated, err := h.repo.UpdateEntry(ctx, entry)
	if err != nil {
		return nil, fmt.Errorf("update entry: %w", err)
	}

	// 4. Append event to event store
	payload, _ := json.Marshal(map[string]string{
		"schedule_id": cmd.ScheduleID.String(),
		"entry_id":    cmd.EntryID.String(),
		"teacher_id":  cmd.TeacherID.String(),
	})
	_ = h.repo.AppendEvent(ctx, cmd.ScheduleID, "Schedule", "ManualAssignment", payload)

	// 5. Publish to NATS (non-blocking — don't fail the operation on publish error)
	_ = h.publisher.Publish(ctx, "timetable.entry.assigned", map[string]string{
		"schedule_id": cmd.ScheduleID.String(),
		"entry_id":    cmd.EntryID.String(),
		"teacher_id":  cmd.TeacherID.String(),
	})

	return updated, nil
}
