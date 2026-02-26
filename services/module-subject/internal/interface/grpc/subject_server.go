package grpc

import (
	"context"

	"github.com/google/uuid"
	corev1 "github.com/HuynhHoangPhuc/myrmex/gen/go/core/v1"
	subjectv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/subject/v1"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/application/command"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/application/query"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SubjectServer implements subjectv1.SubjectServiceServer.
type SubjectServer struct {
	subjectv1.UnimplementedSubjectServiceServer

	createHandler *command.CreateSubjectHandler
	updateHandler *command.UpdateSubjectHandler
	deleteHandler *command.DeleteSubjectHandler
	getHandler    *query.GetSubjectHandler
	listHandler   *query.ListSubjectsHandler
}

// NewSubjectServer constructs a SubjectServer with all required handlers.
func NewSubjectServer(
	createHandler *command.CreateSubjectHandler,
	updateHandler *command.UpdateSubjectHandler,
	deleteHandler *command.DeleteSubjectHandler,
	getHandler *query.GetSubjectHandler,
	listHandler *query.ListSubjectsHandler,
) *SubjectServer {
	return &SubjectServer{
		createHandler: createHandler,
		updateHandler: updateHandler,
		deleteHandler: deleteHandler,
		getHandler:    getHandler,
		listHandler:   listHandler,
	}
}

// CreateSubject handles subject creation.
func (s *SubjectServer) CreateSubject(ctx context.Context, req *subjectv1.CreateSubjectRequest) (*subjectv1.CreateSubjectResponse, error) {
	subject, err := s.createHandler.Handle(ctx, command.CreateSubjectCommand{
		Code:         req.GetCode(),
		Name:         req.GetName(),
		Credits:      req.GetCredits(),
		Description:  req.GetDescription(),
		DepartmentID: req.GetDepartmentId(),
		WeeklyHours:  req.GetWeeklyHours(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create subject: %v", err)
	}
	return &subjectv1.CreateSubjectResponse{Subject: subjectToProto(subject)}, nil
}

// GetSubject fetches a subject by ID.
func (s *SubjectServer) GetSubject(ctx context.Context, req *subjectv1.GetSubjectRequest) (*subjectv1.GetSubjectResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid subject id: %v", err)
	}
	subject, err := s.getHandler.Handle(ctx, query.GetSubjectQuery{ID: id})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "get subject: %v", err)
	}
	return &subjectv1.GetSubjectResponse{Subject: subjectToProto(subject)}, nil
}

// ListSubjects returns a paginated list of subjects.
func (s *SubjectServer) ListSubjects(ctx context.Context, req *subjectv1.ListSubjectsRequest) (*subjectv1.ListSubjectsResponse, error) {
	pg := req.GetPagination()
	var limit, offset int32 = 20, 0
	if pg != nil {
		if pg.GetPageSize() > 0 {
			limit = pg.GetPageSize()
		}
		if pg.GetPage() > 1 {
			offset = (pg.GetPage() - 1) * limit
		}
	}

	result, err := s.listHandler.Handle(ctx, query.ListSubjectsQuery{
		Limit:        limit,
		Offset:       offset,
		DepartmentID: req.GetDepartmentId(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list subjects: %v", err)
	}

	protos := make([]*subjectv1.Subject, len(result.Subjects))
	for i, s := range result.Subjects {
		protos[i] = subjectToProto(s)
	}

	page := int32(1)
	if pg != nil && pg.GetPage() > 1 {
		page = pg.GetPage()
	}

	return &subjectv1.ListSubjectsResponse{
		Subjects: protos,
		Pagination: &corev1.PaginationResponse{
			Total:    int32(result.Total),
			Page:     page,
			PageSize: limit,
		},
	}, nil
}

// UpdateSubject applies partial updates to a subject.
func (s *SubjectServer) UpdateSubject(ctx context.Context, req *subjectv1.UpdateSubjectRequest) (*subjectv1.UpdateSubjectResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid subject id: %v", err)
	}

	cmd := command.UpdateSubjectCommand{ID: id}
	if req.Code != nil {
		v := req.GetCode()
		cmd.Code = &v
	}
	if req.Name != nil {
		v := req.GetName()
		cmd.Name = &v
	}
	if req.Credits != nil {
		v := req.GetCredits()
		cmd.Credits = &v
	}
	if req.Description != nil {
		v := req.GetDescription()
		cmd.Description = &v
	}
	if req.DepartmentId != nil {
		v := req.GetDepartmentId()
		cmd.DepartmentID = &v
	}
	if req.WeeklyHours != nil {
		v := req.GetWeeklyHours()
		cmd.WeeklyHours = &v
	}

	subject, err := s.updateHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "update subject: %v", err)
	}
	return &subjectv1.UpdateSubjectResponse{Subject: subjectToProto(subject)}, nil
}

// DeleteSubject removes a subject by ID.
func (s *SubjectServer) DeleteSubject(ctx context.Context, req *subjectv1.DeleteSubjectRequest) (*subjectv1.DeleteSubjectResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid subject id: %v", err)
	}
	if err := s.deleteHandler.Handle(ctx, command.DeleteSubjectCommand{ID: id}); err != nil {
		return nil, status.Errorf(codes.Internal, "delete subject: %v", err)
	}
	return &subjectv1.DeleteSubjectResponse{}, nil
}

// subjectToProto maps a domain Subject entity to its protobuf representation.
func subjectToProto(s *entity.Subject) *subjectv1.Subject {
	return &subjectv1.Subject{
		Id:           s.ID.String(),
		Code:         s.Code,
		Name:         s.Name,
		Credits:      s.Credits,
		DepartmentId: s.DepartmentID,
		Description:  s.Description,
		WeeklyHours:  s.WeeklyHours,
		IsActive:     s.IsActive,
		CreatedAt:    timestamppb.New(s.CreatedAt),
		UpdatedAt:    timestamppb.New(s.UpdatedAt),
	}
}
