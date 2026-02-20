package command

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	infragrpc "github.com/myrmex-erp/myrmex/services/module-timetable/internal/infrastructure/grpc"

	"github.com/myrmex-erp/myrmex/services/module-timetable/internal/domain/entity"
	"github.com/myrmex-erp/myrmex/services/module-timetable/internal/domain/repository"
	"github.com/myrmex-erp/myrmex/services/module-timetable/internal/domain/service"
	"github.com/myrmex-erp/myrmex/services/module-timetable/internal/domain/valueobject"
)

// GenerateScheduleCommand requests async schedule generation for a semester.
type GenerateScheduleCommand struct {
	SemesterID     uuid.UUID
	TimeoutSeconds int // solver wall-clock timeout (default 30)
}

// GenerateScheduleResult holds the outcome of an async generation run.
type GenerateScheduleResult struct {
	ScheduleID     uuid.UUID
	HardViolations int
	SoftPenalty    float64
	Score          float64
	IsPartial      bool
	UnassignedCount int
}

// GenerateScheduleHandler orchestrates async schedule generation via the CSP solver.
type GenerateScheduleHandler struct {
	semesterRepo repository.SemesterRepository
	scheduleRepo repository.ScheduleRepository
	roomRepo     repository.RoomRepository
	hrClient     *infragrpc.HRClient
	subjectClient *infragrpc.SubjectClient
	publisher    EventPublisher
}

func NewGenerateScheduleHandler(
	semesterRepo repository.SemesterRepository,
	scheduleRepo repository.ScheduleRepository,
	roomRepo     repository.RoomRepository,
	hrClient     *infragrpc.HRClient,
	subjectClient *infragrpc.SubjectClient,
	publisher    EventPublisher,
) *GenerateScheduleHandler {
	return &GenerateScheduleHandler{
		semesterRepo:  semesterRepo,
		scheduleRepo:  scheduleRepo,
		roomRepo:      roomRepo,
		hrClient:      hrClient,
		subjectClient: subjectClient,
		publisher:     publisher,
	}
}

// Handle creates a draft schedule record, runs the CSP solver in a goroutine,
// and returns the schedule ID immediately.  The caller can poll GetSchedule
// to observe the final status.
func (h *GenerateScheduleHandler) Handle(ctx context.Context, cmd GenerateScheduleCommand) (uuid.UUID, error) {
	// 1. Validate semester
	semester, err := h.semesterRepo.GetByID(ctx, cmd.SemesterID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("get semester: %w", err)
	}

	// 2. Create a draft schedule row immediately so callers can track it
	schedule := &entity.Schedule{
		SemesterID: semester.ID,
		Name:       fmt.Sprintf("Auto-%s-%d-%d", semester.Name, semester.Year, semester.Term),
		Status:     valueobject.ScheduleStatusDraft,
	}
	created, err := h.scheduleRepo.Create(ctx, schedule)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create schedule record: %w", err)
	}

	timeout := cmd.TimeoutSeconds
	if timeout <= 0 {
		timeout = 30
	}

	// 3. Run solver asynchronously — do not block the gRPC call
	go h.runSolver(created.ID, semester, timeout)

	return created.ID, nil
}

// runSolver executes the full CSP pipeline and updates the DB with results.
func (h *GenerateScheduleHandler) runSolver(scheduleID uuid.UUID, semester *entity.Semester, timeoutSec int) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	// --- Pre-fetch all data before solver begins (no IO during search) ---

	// 4a. Fetch offered subjects
	subjects, err := h.subjectClient.ListSubjectsByIDs(ctx, semester.OfferedSubjectIDs)
	if err != nil {
		h.markFailed(scheduleID, fmt.Sprintf("fetch subjects: %v", err))
		return
	}

	// 4b. Fetch available rooms
	rooms, err := h.roomRepo.List(ctx, 200, 0)
	if err != nil {
		h.markFailed(scheduleID, fmt.Sprintf("fetch rooms: %v", err))
		return
	}

	// 4c. Fetch time slots for the semester
	slots, err := h.semesterRepo.ListTimeSlots(ctx, semester.ID)
	if err != nil {
		h.markFailed(scheduleID, fmt.Sprintf("fetch slots: %v", err))
		return
	}

	// 4d. Fetch teachers from HR module
	teachers, err := h.hrClient.ListTeachers(ctx)
	if err != nil {
		h.markFailed(scheduleID, fmt.Sprintf("fetch teachers: %v", err))
		return
	}

	// 4e. Fetch availability for each teacher
	teacherAvailability := make(map[uuid.UUID][]*entity.TimeSlot, len(teachers))
	for _, t := range teachers {
		avail, err := h.hrClient.GetTeacherAvailability(ctx, t.ID)
		if err == nil {
			teacherAvailability[t.ID] = avail
		}
	}

	// 5. Build CSP inputs
	teacherSpecs := make(map[uuid.UUID]map[string]bool, len(teachers))
	for _, t := range teachers {
		specSet := make(map[string]bool, len(t.Specializations))
		for _, s := range t.Specializations {
			specSet[s] = true
		}
		teacherSpecs[t.ID] = specSet
	}

	teacherMaxHours := make(map[uuid.UUID]int, len(teachers))
	for _, t := range teachers {
		teacherMaxHours[t.ID] = t.MaxHoursPerWeek
	}

	subjectSpecs := make(map[uuid.UUID][]string, len(subjects))
	for _, s := range subjects {
		subjectSpecs[s.ID] = s.RequiredSpecializations
	}

	checker := service.NewConstraintChecker(teacherAvailability, teacherSpecs, subjectSpecs, teacherMaxHours)

	// 6. Build slot lookup map
	slotMap := make(map[uuid.UUID]*entity.TimeSlot, len(slots))
	for _, sl := range slots {
		slotMap[sl.ID] = sl
	}

	// 7. Build variables and initial domains
	variables := make([]service.ScheduleVariable, 0, len(subjects))
	for _, s := range subjects {
		variables = append(variables, service.ScheduleVariable{
			SubjectID:               s.ID,
			SubjectCode:             s.Code,
			WeeklyHours:             s.Credits,
			RequiredSpecializations: s.RequiredSpecializations,
		})
	}

	domains := buildDomains(variables, teachers, rooms, slots)

	// 8. Solve
	solver := service.NewCSPSolver(variables, domains, slotMap, checker)
	result, err := solver.Solve(ctx)
	if err != nil {
		h.markFailed(scheduleID, fmt.Sprintf("solver: %v", err))
		return
	}

	// 9. Persist entries
	for _, entry := range result.Entries {
		entry.ScheduleID = scheduleID
		_, _ = h.scheduleRepo.CreateEntry(context.Background(), entry)
	}

	// 10. Update schedule with results
	_, _ = h.scheduleRepo.UpdateResult(context.Background(), scheduleID,
		result.Score, result.HardViolations, result.SoftPenalty)

	status := valueobject.ScheduleStatusDraft
	_, _ = h.scheduleRepo.UpdateStatus(context.Background(), scheduleID, status)

	// 11. Append event + publish
	payload, _ := json.Marshal(map[string]interface{}{
		"schedule_id": scheduleID.String(),
		"score":       result.Score,
		"is_partial":  result.IsPartial,
	})
	_ = h.scheduleRepo.AppendEvent(context.Background(), scheduleID, "Schedule", "ScheduleGenerated", payload)
	_ = h.publisher.Publish(context.Background(), "timetable.schedule.generated", map[string]interface{}{
		"schedule_id": scheduleID.String(),
		"score":       result.Score,
		"is_partial":  result.IsPartial,
	})
}

func (h *GenerateScheduleHandler) markFailed(scheduleID uuid.UUID, reason string) {
	// Mark schedule as failed by setting a negative score sentinel
	_, _ = h.scheduleRepo.UpdateResult(context.Background(), scheduleID, -1, 0, 0)
	payload, _ := json.Marshal(map[string]string{
		"schedule_id": scheduleID.String(),
		"error":       reason,
	})
	_ = h.scheduleRepo.AppendEvent(context.Background(), scheduleID, "Schedule", "ScheduleGenerationFailed", payload)
}

// buildDomains creates the initial domain (all valid assignments) for each subject variable.
// Each assignment is a (teacher, room, slot) triple; hard constraints are NOT checked here —
// that is the solver's job.  We only enumerate structurally possible combinations.
func buildDomains(
	variables []service.ScheduleVariable,
	teachers []service.TeacherInfo,
	rooms []*entity.Room,
	slots []*entity.TimeSlot,
) map[string][]service.Assignment {
	domains := make(map[string][]service.Assignment, len(variables))
	for _, v := range variables {
		var combos []service.Assignment
		for _, t := range teachers {
			for _, r := range rooms {
				for _, sl := range slots {
					combos = append(combos, service.Assignment{
						TeacherID: t.ID,
						RoomID:    r.ID,
						SlotID:    sl.ID,
					})
				}
			}
		}
		domains[v.SubjectID.String()] = combos
	}
	return domains
}
