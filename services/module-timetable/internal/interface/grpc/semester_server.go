package grpc

import (
	"context"

	"github.com/google/uuid"
	corev1 "github.com/HuynhHoangPhuc/myrmex/gen/go/core/v1"
	timetablev1 "github.com/HuynhHoangPhuc/myrmex/gen/go/timetable/v1"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/application/command"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/application/query"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/domain/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SemesterServer implements timetablev1.SemesterServiceServer.
type SemesterServer struct {
	timetablev1.UnimplementedSemesterServiceServer

	createSemester *command.CreateSemesterHandler
	listSchedules  *query.ListSchedulesHandler
	semesterRepo   semesterReader
}

// semesterReader is the minimal read interface needed by this server.
type semesterReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Semester, error)
	List(ctx context.Context, limit, offset int32) ([]*entity.Semester, error)
	Count(ctx context.Context) (int64, error)
	AddOfferedSubject(ctx context.Context, semesterID, subjectID uuid.UUID) (*entity.Semester, error)
	RemoveOfferedSubject(ctx context.Context, semesterID, subjectID uuid.UUID) (*entity.Semester, error)
	ListTimeSlots(ctx context.Context, semesterID uuid.UUID) ([]*entity.TimeSlot, error)
	CreateTimeSlot(ctx context.Context, ts *entity.TimeSlot) (*entity.TimeSlot, error)
	DeleteTimeSlot(ctx context.Context, slotID uuid.UUID) error
	DeleteTimeSlotsBySemester(ctx context.Context, semesterID uuid.UUID) error
}

func NewSemesterServer(
	createSemester *command.CreateSemesterHandler,
	listSchedules  *query.ListSchedulesHandler,
	semesterRepo   semesterReader,
) *SemesterServer {
	return &SemesterServer{
		createSemester: createSemester,
		listSchedules:  listSchedules,
		semesterRepo:   semesterRepo,
	}
}

func (s *SemesterServer) CreateSemester(ctx context.Context, req *timetablev1.CreateSemesterRequest) (*timetablev1.CreateSemesterResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	var startDate, endDate interface{ AsTime() interface{ IsZero() bool } }
	_ = startDate
	_ = endDate

	sem, err := s.createSemester.Handle(ctx, command.CreateSemesterCommand{
		Name:      req.Name,
		Year:      int(req.Year),
		Term:      int(req.Term),
		StartDate: req.StartDate.AsTime(),
		EndDate:   req.EndDate.AsTime(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create semester: %v", err)
	}

	return &timetablev1.CreateSemesterResponse{Semester: semesterToProto(sem)}, nil
}

func (s *SemesterServer) GetSemester(ctx context.Context, req *timetablev1.GetSemesterRequest) (*timetablev1.GetSemesterResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}
	sem, err := s.semesterRepo.GetByID(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "semester not found: %v", err)
	}
	return &timetablev1.GetSemesterResponse{Semester: semesterToProto(sem)}, nil
}

func (s *SemesterServer) ListSemesters(ctx context.Context, req *timetablev1.ListSemestersRequest) (*timetablev1.ListSemestersResponse, error) {
	page := int32(1)
	pageSize := int32(20)
	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			page = req.Pagination.Page
		}
		if req.Pagination.PageSize > 0 {
			pageSize = req.Pagination.PageSize
		}
	}
	offset := (page - 1) * pageSize

	semesters, err := s.semesterRepo.List(ctx, pageSize, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list semesters: %v", err)
	}
	total, _ := s.semesterRepo.Count(ctx)

	proto := make([]*timetablev1.Semester, len(semesters))
	for i, sem := range semesters {
		proto[i] = semesterToProto(sem)
	}

	return &timetablev1.ListSemestersResponse{
		Semesters: proto,
		Pagination: &corev1.PaginationResponse{
			Total:    int32(total),
			Page:     page,
			PageSize: pageSize,
		},
	}, nil
}

func (s *SemesterServer) AddOfferedSubject(ctx context.Context, req *timetablev1.AddOfferedSubjectRequest) (*timetablev1.AddOfferedSubjectResponse, error) {
	semesterID, err := uuid.Parse(req.SemesterId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid semester_id")
	}
	subjectID, err := uuid.Parse(req.SubjectId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid subject_id")
	}
	sem, err := s.semesterRepo.AddOfferedSubject(ctx, semesterID, subjectID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "add offered subject: %v", err)
	}
	return &timetablev1.AddOfferedSubjectResponse{Semester: semesterToProto(sem)}, nil
}

func (s *SemesterServer) RemoveOfferedSubject(ctx context.Context, req *timetablev1.RemoveOfferedSubjectRequest) (*timetablev1.RemoveOfferedSubjectResponse, error) {
	semesterID, err := uuid.Parse(req.SemesterId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid semester_id")
	}
	subjectID, err := uuid.Parse(req.SubjectId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid subject_id")
	}
	sem, err := s.semesterRepo.RemoveOfferedSubject(ctx, semesterID, subjectID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "remove offered subject: %v", err)
	}
	return &timetablev1.RemoveOfferedSubjectResponse{Semester: semesterToProto(sem)}, nil
}

func (s *SemesterServer) ListTimeSlots(ctx context.Context, req *timetablev1.ListTimeSlotsRequest) (*timetablev1.ListTimeSlotsResponse, error) {
	semesterID, err := uuid.Parse(req.SemesterId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid semester_id")
	}
	slots, err := s.semesterRepo.ListTimeSlots(ctx, semesterID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list time slots: %v", err)
	}
	protos := make([]*timetablev1.TimeSlot, len(slots))
	for i, sl := range slots {
		protos[i] = &timetablev1.TimeSlot{
			Id:          sl.ID.String(),
			SemesterId:  sl.SemesterID.String(),
			DayOfWeek:   int32(sl.DayOfWeek),
			StartPeriod: int32(sl.StartPeriod),
			EndPeriod:   int32(sl.EndPeriod),
		}
	}
	return &timetablev1.ListTimeSlotsResponse{TimeSlots: protos}, nil
}

func (s *SemesterServer) CreateTimeSlot(ctx context.Context, req *timetablev1.CreateTimeSlotRequest) (*timetablev1.CreateTimeSlotResponse, error) {
	semesterID, err := uuid.Parse(req.SemesterId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid semester_id")
	}
	ts := &entity.TimeSlot{
		SemesterID:  semesterID,
		DayOfWeek:   int(req.DayOfWeek),
		StartPeriod: int(req.StartPeriod),
		EndPeriod:   int(req.EndPeriod),
	}
	if err := ts.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid time slot: %v", err)
	}
	created, err := s.semesterRepo.CreateTimeSlot(ctx, ts)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create time slot: %v", err)
	}
	return &timetablev1.CreateTimeSlotResponse{
		TimeSlot: &timetablev1.TimeSlot{
			Id:          created.ID.String(),
			SemesterId:  created.SemesterID.String(),
			DayOfWeek:   int32(created.DayOfWeek),
			StartPeriod: int32(created.StartPeriod),
			EndPeriod:   int32(created.EndPeriod),
		},
	}, nil
}

func (s *SemesterServer) DeleteTimeSlot(ctx context.Context, req *timetablev1.DeleteTimeSlotRequest) (*timetablev1.DeleteTimeSlotResponse, error) {
	slotID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}
	if err := s.semesterRepo.DeleteTimeSlot(ctx, slotID); err != nil {
		return nil, status.Errorf(codes.Internal, "delete time slot: %v", err)
	}
	return &timetablev1.DeleteTimeSlotResponse{}, nil
}

// standardPresetSlots returns the standard Mon-Sat × 3 periods preset (18 slots).
var standardPresetSlots = []struct{ day, start, end int }{
	{0, 1, 2}, {0, 3, 4}, {0, 5, 6},
	{1, 1, 2}, {1, 3, 4}, {1, 5, 6},
	{2, 1, 2}, {2, 3, 4}, {2, 5, 6},
	{3, 1, 2}, {3, 3, 4}, {3, 5, 6},
	{4, 1, 2}, {4, 3, 4}, {4, 5, 6},
	{5, 1, 2}, {5, 3, 4}, {5, 5, 6},
}

// mwfPresetSlots returns Mon/Wed/Fri × 4 periods preset.
var mwfPresetSlots = []struct{ day, start, end int }{
	{0, 1, 2}, {0, 3, 4}, {0, 5, 6}, {0, 7, 8},
	{2, 1, 2}, {2, 3, 4}, {2, 5, 6}, {2, 7, 8},
	{4, 1, 2}, {4, 3, 4}, {4, 5, 6}, {4, 7, 8},
}

// tuthPresetSlots returns Tue/Thu × 4 periods preset.
var tuthPresetSlots = []struct{ day, start, end int }{
	{1, 1, 2}, {1, 3, 4}, {1, 5, 6}, {1, 7, 8},
	{3, 1, 2}, {3, 3, 4}, {3, 5, 6}, {3, 7, 8},
}

func (s *SemesterServer) ApplyTimeSlotPreset(ctx context.Context, req *timetablev1.ApplyTimeSlotPresetRequest) (*timetablev1.ApplyTimeSlotPresetResponse, error) {
	semesterID, err := uuid.Parse(req.SemesterId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid semester_id")
	}

	var preset []struct{ day, start, end int }
	switch req.Preset {
	case "standard":
		preset = standardPresetSlots
	case "mwf":
		preset = mwfPresetSlots
	case "tuth":
		preset = tuthPresetSlots
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unknown preset %q; use standard|mwf|tuth", req.Preset)
	}

	// Clear existing slots first
	if err := s.semesterRepo.DeleteTimeSlotsBySemester(ctx, semesterID); err != nil {
		return nil, status.Errorf(codes.Internal, "clear existing slots: %v", err)
	}

	created := make([]*timetablev1.TimeSlot, 0, len(preset))
	for _, p := range preset {
		ts, err := s.semesterRepo.CreateTimeSlot(ctx, &entity.TimeSlot{
			SemesterID:  semesterID,
			DayOfWeek:   p.day,
			StartPeriod: p.start,
			EndPeriod:   p.end,
		})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "create preset slot: %v", err)
		}
		created = append(created, &timetablev1.TimeSlot{
			Id:          ts.ID.String(),
			SemesterId:  ts.SemesterID.String(),
			DayOfWeek:   int32(ts.DayOfWeek),
			StartPeriod: int32(ts.StartPeriod),
			EndPeriod:   int32(ts.EndPeriod),
		})
	}
	return &timetablev1.ApplyTimeSlotPresetResponse{TimeSlots: created}, nil
}

// --- proto helpers ---

func semesterToProto(s *entity.Semester) *timetablev1.Semester {
	p := &timetablev1.Semester{
		Id:        s.ID.String(),
		Name:      s.Name,
		Year:      int32(s.Year),
		Term:      int32(s.Term),
		StartDate: timestamppb.New(s.StartDate),
		EndDate:   timestamppb.New(s.EndDate),
		CreatedAt: timestamppb.New(s.CreatedAt),
	}
	for _, id := range s.OfferedSubjectIDs {
		p.OfferedSubjectIds = append(p.OfferedSubjectIds, id.String())
	}
	return p
}
