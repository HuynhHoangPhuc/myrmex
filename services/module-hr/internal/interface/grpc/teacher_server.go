package grpc

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	corev1 "github.com/HuynhHoangPhuc/myrmex/gen/go/core/v1"
	hrv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/hr/v1"
	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/application/command"
	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/application/query"
	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/domain/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TeacherServer implements hrv1.TeacherServiceServer.
type TeacherServer struct {
	hrv1.UnimplementedTeacherServiceServer

	createTeacher      *command.CreateTeacherHandler
	updateTeacher      *command.UpdateTeacherHandler
	deleteTeacher      *command.DeleteTeacherHandler
	updateAvailability *command.UpdateAvailabilityHandler
	getTeacher         *query.GetTeacherHandler
	listTeachers       *query.ListTeachersHandler
	getAvailability    *query.GetAvailabilityHandler
}

func NewTeacherServer(
	createTeacher *command.CreateTeacherHandler,
	updateTeacher *command.UpdateTeacherHandler,
	deleteTeacher *command.DeleteTeacherHandler,
	updateAvailability *command.UpdateAvailabilityHandler,
	getTeacher *query.GetTeacherHandler,
	listTeachers *query.ListTeachersHandler,
	getAvailability *query.GetAvailabilityHandler,
) *TeacherServer {
	return &TeacherServer{
		createTeacher:      createTeacher,
		updateTeacher:      updateTeacher,
		deleteTeacher:      deleteTeacher,
		updateAvailability: updateAvailability,
		getTeacher:         getTeacher,
		listTeachers:       listTeachers,
		getAvailability:    getAvailability,
	}
}

func (s *TeacherServer) CreateTeacher(ctx context.Context, req *hrv1.CreateTeacherRequest) (*hrv1.CreateTeacherResponse, error) {
	if req.FullName == "" || req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "full_name and email are required")
	}

	var deptID *uuid.UUID
	if req.DepartmentId != "" {
		id, err := uuid.Parse(req.DepartmentId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid department_id")
		}
		deptID = &id
	}

	// Use provided employee_code or generate a default from email prefix
	employeeCode := req.EmployeeCode
	if employeeCode == "" {
		employeeCode = fmt.Sprintf("TC-%s", req.Email[:min(8, len(req.Email))])
	}
	maxHours := int(req.MaxHoursPerWeek)
	if maxHours <= 0 {
		maxHours = 20
	}

	teacher, err := s.createTeacher.Handle(ctx, command.CreateTeacherCommand{
		EmployeeCode:    employeeCode,
		FullName:        req.FullName,
		Email:           req.Email,
		Phone:           req.Phone,
		Title:           req.Title,
		DepartmentID:    deptID,
		MaxHoursPerWeek: maxHours,
		Specializations: req.Specializations,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create teacher: %v", err)
	}

	return &hrv1.CreateTeacherResponse{Teacher: teacherToProto(teacher)}, nil
}

func (s *TeacherServer) GetTeacher(ctx context.Context, req *hrv1.GetTeacherRequest) (*hrv1.GetTeacherResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	teacher, err := s.getTeacher.Handle(ctx, query.GetTeacherQuery{ID: id})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "teacher not found: %v", err)
	}

	return &hrv1.GetTeacherResponse{Teacher: teacherToProto(teacher)}, nil
}

func (s *TeacherServer) ListTeachers(ctx context.Context, req *hrv1.ListTeachersRequest) (*hrv1.ListTeachersResponse, error) {
	q := query.ListTeachersQuery{}
	if req.Pagination != nil {
		q.Page = req.Pagination.Page
		q.PageSize = req.Pagination.PageSize
	}
	if req.DepartmentId != nil && *req.DepartmentId != "" {
		id, err := uuid.Parse(*req.DepartmentId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid department_id")
		}
		q.DepartmentID = &id
	}

	result, err := s.listTeachers.Handle(ctx, q)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list teachers: %v", err)
	}

	teachers := make([]*hrv1.Teacher, len(result.Teachers))
	for i, t := range result.Teachers {
		teachers[i] = teacherToProto(t)
	}

	pageSize := int32(20)
	if req.Pagination != nil && req.Pagination.PageSize > 0 {
		pageSize = req.Pagination.PageSize
	}
	page := int32(1)
	if req.Pagination != nil && req.Pagination.Page > 0 {
		page = req.Pagination.Page
	}
	totalInt32 := int32(result.Total)

	return &hrv1.ListTeachersResponse{
		Teachers: teachers,
		Pagination: &corev1.PaginationResponse{
			Total:    totalInt32,
			Page:     page,
			PageSize: pageSize,
		},
	}, nil
}

func (s *TeacherServer) UpdateTeacher(ctx context.Context, req *hrv1.UpdateTeacherRequest) (*hrv1.UpdateTeacherResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	// Fetch current to merge partial updates
	current, err := s.getTeacher.Handle(ctx, query.GetTeacherQuery{ID: id})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "teacher not found: %v", err)
	}

	cmd := command.UpdateTeacherCommand{
		ID:              id,
		FullName:        current.FullName,
		Email:           current.Email,
		Phone:           current.Phone,
		Title:           current.Title,
		DepartmentID:    current.DepartmentID,
		MaxHoursPerWeek: current.MaxHoursPerWeek,
		IsActive:        current.IsActive,
	}
	if req.FullName != nil {
		cmd.FullName = req.GetFullName()
	}
	if req.Email != nil {
		cmd.Email = req.GetEmail()
	}
	if req.Title != nil {
		cmd.Title = req.GetTitle()
	}
	if req.IsActive != nil {
		cmd.IsActive = req.GetIsActive()
	}
	if req.DepartmentId != nil {
		if req.GetDepartmentId() == "" {
			cmd.DepartmentID = nil
		} else {
			deptID, err := uuid.Parse(req.GetDepartmentId())
			if err != nil {
				return nil, status.Error(codes.InvalidArgument, "invalid department_id")
			}
			cmd.DepartmentID = &deptID
		}
	}

	updated, err := s.updateTeacher.Handle(ctx, cmd)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "update teacher: %v", err)
	}

	return &hrv1.UpdateTeacherResponse{Teacher: teacherToProto(updated)}, nil
}

func (s *TeacherServer) DeleteTeacher(ctx context.Context, req *hrv1.DeleteTeacherRequest) (*hrv1.DeleteTeacherResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	if err := s.deleteTeacher.Handle(ctx, command.DeleteTeacherCommand{ID: id}); err != nil {
		return nil, status.Errorf(codes.Internal, "delete teacher: %v", err)
	}

	return &hrv1.DeleteTeacherResponse{}, nil
}

func (s *TeacherServer) ListTeacherAvailability(ctx context.Context, req *hrv1.ListTeacherAvailabilityRequest) (*hrv1.ListTeacherAvailabilityResponse, error) {
	teacherID, err := uuid.Parse(req.TeacherId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid teacher_id")
	}

	slots, err := s.getAvailability.Handle(ctx, query.GetAvailabilityQuery{TeacherID: teacherID})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list availability: %v", err)
	}

	protoSlots := make([]*hrv1.TimeSlot, len(slots))
	for i, slot := range slots {
		protoSlots[i] = availabilityToProto(slot)
	}

	return &hrv1.ListTeacherAvailabilityResponse{
		TeacherId:      req.TeacherId,
		AvailableSlots: protoSlots,
	}, nil
}

func (s *TeacherServer) UpdateTeacherAvailability(ctx context.Context, req *hrv1.UpdateTeacherAvailabilityRequest) (*hrv1.UpdateTeacherAvailabilityResponse, error) {
	teacherID, err := uuid.Parse(req.TeacherId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid teacher_id")
	}

	slots := make([]command.AvailabilitySlot, len(req.AvailableSlots))
	for i, s := range req.AvailableSlots {
		slots[i] = command.AvailabilitySlot{
			DayOfWeek:   int(s.DayOfWeek),
			StartPeriod: int(s.StartPeriod),
			EndPeriod:   int(s.EndPeriod),
		}
	}

	result, err := s.updateAvailability.Handle(ctx, command.UpdateAvailabilityCommand{
		TeacherID: teacherID,
		Slots:     slots,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "update availability: %v", err)
	}

	protoSlots := make([]*hrv1.TimeSlot, len(result))
	for i, slot := range result {
		protoSlots[i] = availabilityToProto(slot)
	}

	return &hrv1.UpdateTeacherAvailabilityResponse{
		TeacherId:      req.TeacherId,
		AvailableSlots: protoSlots,
	}, nil
}

// --- proto mapping helpers ---

func teacherToProto(t *entity.Teacher) *hrv1.Teacher {
	p := &hrv1.Teacher{
		Id:              t.ID.String(),
		FullName:        t.FullName,
		Email:           t.Email,
		Phone:           t.Phone,
		Title:           t.Title,
		IsActive:        t.IsActive,
		EmployeeCode:    t.EmployeeCode,
		MaxHoursPerWeek: int32(t.MaxHoursPerWeek),
		Specializations: t.Specializations,
		CreatedAt:       timestamppb.New(t.CreatedAt),
		UpdatedAt:       timestamppb.New(t.UpdatedAt),
	}
	if t.DepartmentID != nil {
		p.DepartmentId = t.DepartmentID.String()
	}
	return p
}

func availabilityToProto(a *entity.Availability) *hrv1.TimeSlot {
	return &hrv1.TimeSlot{
		DayOfWeek:   int32(a.DayOfWeek),
		StartPeriod: int32(a.StartPeriod),
		EndPeriod:   int32(a.EndPeriod),
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
