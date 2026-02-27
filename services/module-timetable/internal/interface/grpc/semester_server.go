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
