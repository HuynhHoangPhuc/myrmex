package grpc

import (
	"context"

	"github.com/google/uuid"
	timetablev1 "github.com/HuynhHoangPhuc/myrmex/gen/go/timetable/v1"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/application/command"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/application/query"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/domain/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TimetableServer implements timetablev1.TimetableServiceServer.
type TimetableServer struct {
	timetablev1.UnimplementedTimetableServiceServer

	generateSchedule *command.GenerateScheduleHandler
	manualAssign     *command.ManualAssignHandler
	getSchedule      *query.GetScheduleHandler
	listSchedules    *query.ListSchedulesHandler
	suggestTeachers  *query.SuggestTeachersHandler
	roomRepo         repository.RoomRepository
}

func NewTimetableServer(
	generateSchedule *command.GenerateScheduleHandler,
	manualAssign     *command.ManualAssignHandler,
	getSchedule      *query.GetScheduleHandler,
	listSchedules    *query.ListSchedulesHandler,
	suggestTeachers  *query.SuggestTeachersHandler,
	roomRepo         repository.RoomRepository,
) *TimetableServer {
	return &TimetableServer{
		generateSchedule: generateSchedule,
		manualAssign:     manualAssign,
		getSchedule:      getSchedule,
		listSchedules:    listSchedules,
		suggestTeachers:  suggestTeachers,
		roomRepo:         roomRepo,
	}
}

func (s *TimetableServer) GenerateSchedule(ctx context.Context, req *timetablev1.GenerateScheduleRequest) (*timetablev1.GenerateScheduleResponse, error) {
	semesterID, err := uuid.Parse(req.SemesterId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid semester_id")
	}

	scheduleID, err := s.generateSchedule.Handle(ctx, command.GenerateScheduleCommand{
		SemesterID:     semesterID,
		TimeoutSeconds: int(req.TimeoutSeconds),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "generate schedule: %v", err)
	}

	// Return the newly created (empty) schedule — caller polls via GetSchedule
	result, err := s.getSchedule.Handle(ctx, query.GetScheduleQuery{ScheduleID: scheduleID})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get schedule after generation: %v", err)
	}

	return &timetablev1.GenerateScheduleResponse{
		Schedule:        scheduleToProto(result.Schedule, result.Entries),
		IsPartial:       result.Schedule.Score < 100 && result.Schedule.Score >= 0,
		UnassignedCount: int32(len(result.Entries)),
	}, nil
}

func (s *TimetableServer) GetSchedule(ctx context.Context, req *timetablev1.GetScheduleRequest) (*timetablev1.GetScheduleResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	result, err := s.getSchedule.Handle(ctx, query.GetScheduleQuery{ScheduleID: id})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "schedule not found: %v", err)
	}

	return &timetablev1.GetScheduleResponse{
		Schedule: scheduleToProto(result.Schedule, result.Entries),
	}, nil
}

func (s *TimetableServer) ManualAssign(ctx context.Context, req *timetablev1.ManualAssignRequest) (*timetablev1.ManualAssignResponse, error) {
	scheduleID, err := uuid.Parse(req.ScheduleId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid schedule_id")
	}
	entryID, err := uuid.Parse(req.EntryId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid entry_id")
	}
	teacherID, err := uuid.Parse(req.TeacherId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid teacher_id")
	}

	entry, err := s.manualAssign.Handle(ctx, command.ManualAssignCommand{
		ScheduleID: scheduleID,
		EntryID:    entryID,
		TeacherID:  teacherID,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "manual assign: %v", err)
	}

	return &timetablev1.ManualAssignResponse{
		Entry: entryToProto(entry),
	}, nil
}

func (s *TimetableServer) SuggestTeachers(ctx context.Context, req *timetablev1.SuggestTeachersRequest) (*timetablev1.SuggestTeachersResponse, error) {
	subjectID, err := uuid.Parse(req.SubjectId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid subject_id")
	}

	ranks, err := s.suggestTeachers.Handle(ctx, query.SuggestTeachersQuery{
		SubjectID:   subjectID,
		DayOfWeek:   int(req.DayOfWeek),
		StartPeriod: int(req.StartPeriod),
		EndPeriod:   int(req.EndPeriod),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "suggest teachers: %v", err)
	}

	suggestions := make([]*timetablev1.TeacherSuggestion, len(ranks))
	for i, r := range ranks {
		suggestions[i] = &timetablev1.TeacherSuggestion{
			TeacherId:   r.TeacherID.String(),
			TeacherName: r.Name,
			Score:       float32(r.Score),
		}
	}

	return &timetablev1.SuggestTeachersResponse{Suggestions: suggestions}, nil
}

func (s *TimetableServer) ListSchedules(ctx context.Context, req *timetablev1.ListSchedulesRequest) (*timetablev1.ListSchedulesResponse, error) {
	var semesterID uuid.UUID
	if req.SemesterId != "" {
		var err error
		semesterID, err = uuid.Parse(req.SemesterId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid semester_id")
		}
	}

	result, err := s.listSchedules.Handle(ctx, query.ListSchedulesQuery{
		SemesterID: semesterID,
		Page:       req.Page,
		PageSize:   req.PageSize,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list schedules: %v", err)
	}

	schedules := make([]*timetablev1.Schedule, len(result.Schedules))
	for i, s := range result.Schedules {
		schedules[i] = scheduleToProto(s, nil)
	}
	return &timetablev1.ListSchedulesResponse{
		Schedules: schedules,
		Total:     int32(result.Total),
		Page:      result.Page,
		PageSize:  result.PageSize,
	}, nil
}

func (s *TimetableServer) ListRooms(ctx context.Context, _ *timetablev1.ListRoomsRequest) (*timetablev1.ListRoomsResponse, error) {
	rooms, err := s.roomRepo.List(ctx, 200, 0)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list rooms: %v", err)
	}
	protos := make([]*timetablev1.Room, len(rooms))
	for i, r := range rooms {
		protos[i] = &timetablev1.Room{
			Id:       r.ID.String(),
			Name:     r.Name,
			Capacity: int32(r.Capacity),
			RoomType: r.Type,
		}
	}
	return &timetablev1.ListRoomsResponse{Rooms: protos}, nil
}

func (s *TimetableServer) UpdateScheduleEntry(ctx context.Context, req *timetablev1.UpdateScheduleEntryRequest) (*timetablev1.UpdateScheduleEntryResponse, error) {
	// Minimal implementation — field updates delegated to ManualAssign for teacher changes.
	return nil, status.Error(codes.Unimplemented, "use ManualAssign for teacher changes")
}

// --- proto mapping helpers ---

func scheduleToProto(s *entity.Schedule, entries []*entity.ScheduleEntry) *timetablev1.Schedule {
	p := &timetablev1.Schedule{
		Id:             s.ID.String(),
		SemesterId:     s.SemesterID.String(),
		Status:         s.Status.String(),
		Score:          s.Score,
		HardViolations: int32(s.HardViolations),
		SoftViolations: s.SoftPenalty,
		CreatedAt:      timestamppb.New(s.CreatedAt),
	}
	for _, e := range entries {
		p.Entries = append(p.Entries, entryToProto(e))
	}
	return p
}

func entryToProto(e *entity.ScheduleEntry) *timetablev1.ScheduleEntry {
	return &timetablev1.ScheduleEntry{
		Id:               e.ID.String(),
		SubjectId:        e.SubjectID.String(),
		TeacherId:        e.TeacherID.String(),
		Room:             e.RoomID.String(),
		DayOfWeek:        int32(e.DayOfWeek),
		StartPeriod:      int32(e.StartPeriod),
		EndPeriod:        int32(e.EndPeriod),
		SubjectName:      e.SubjectName,
		SubjectCode:      e.SubjectCode,
		TeacherName:      e.TeacherName,
		RoomName:         e.RoomName,
		IsManualOverride: e.IsManualOverride,
		DepartmentId:     e.DepartmentID.String(),
	}
}
