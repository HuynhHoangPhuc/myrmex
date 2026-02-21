package grpc

import (
	"context"

	corev1 "github.com/HuynhHoangPhuc/myrmex/gen/go/core/v1"
	hrv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/hr/v1"
	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/application/command"
	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/application/query"
	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/domain/entity"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// DepartmentServer implements hrv1.DepartmentServiceServer.
type DepartmentServer struct {
	hrv1.UnimplementedDepartmentServiceServer

	createDepartment *command.CreateDepartmentHandler
	listDepartments  *query.ListDepartmentsHandler
	// getDepartment uses the dept repo directly via the list handler's repo
	deptRepo departmentGetter
}

// departmentGetter is a minimal interface for GetByID without importing the full repo package.
type departmentGetter interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Department, error)
}

func NewDepartmentServer(
	createDepartment *command.CreateDepartmentHandler,
	listDepartments *query.ListDepartmentsHandler,
	deptRepo departmentGetter,
) *DepartmentServer {
	return &DepartmentServer{
		createDepartment: createDepartment,
		listDepartments:  listDepartments,
		deptRepo:         deptRepo,
	}
}

func (s *DepartmentServer) CreateDepartment(ctx context.Context, req *hrv1.CreateDepartmentRequest) (*hrv1.CreateDepartmentResponse, error) {
	if req.Name == "" || req.Code == "" {
		return nil, status.Error(codes.InvalidArgument, "name and code are required")
	}

	dept, err := s.createDepartment.Handle(ctx, command.CreateDepartmentCommand{
		Name: req.Name,
		Code: req.Code,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create department: %v", err)
	}

	return &hrv1.CreateDepartmentResponse{Department: departmentToProto(dept)}, nil
}

func (s *DepartmentServer) GetDepartment(ctx context.Context, req *hrv1.GetDepartmentRequest) (*hrv1.GetDepartmentResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	dept, err := s.deptRepo.GetByID(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "department not found: %v", err)
	}

	return &hrv1.GetDepartmentResponse{Department: departmentToProto(dept)}, nil
}

func (s *DepartmentServer) ListDepartments(ctx context.Context, req *hrv1.ListDepartmentsRequest) (*hrv1.ListDepartmentsResponse, error) {
	q := query.ListDepartmentsQuery{}
	if req.Pagination != nil {
		q.Page = req.Pagination.Page
		q.PageSize = req.Pagination.PageSize
	}

	result, err := s.listDepartments.Handle(ctx, q)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list departments: %v", err)
	}

	depts := make([]*hrv1.Department, len(result.Departments))
	for i, d := range result.Departments {
		depts[i] = departmentToProto(d)
	}

	pageSize := int32(20)
	if req.Pagination != nil && req.Pagination.PageSize > 0 {
		pageSize = req.Pagination.PageSize
	}
	page := int32(1)
	if req.Pagination != nil && req.Pagination.Page > 0 {
		page = req.Pagination.Page
	}

	return &hrv1.ListDepartmentsResponse{
		Departments: depts,
		Pagination: &corev1.PaginationResponse{
			Total:    int32(result.Total),
			Page:     page,
			PageSize: pageSize,
		},
	}, nil
}

func departmentToProto(d *entity.Department) *hrv1.Department {
	return &hrv1.Department{
		Id:        d.ID.String(),
		Name:      d.Name,
		Code:      d.Code,
		CreatedAt: timestamppb.New(d.CreatedAt),
		UpdatedAt: timestamppb.New(d.UpdatedAt),
	}
}
