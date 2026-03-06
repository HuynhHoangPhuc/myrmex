package grpc

import (
	"context"
	"errors"

	corev1 "github.com/HuynhHoangPhuc/myrmex/gen/go/core/v1"
	studentv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/student/v1"
	appservice "github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/application/service"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/application/command"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/application/query"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *StudentServer) RequestEnrollment(ctx context.Context, req *studentv1.RequestEnrollmentRequest) (*studentv1.RequestEnrollmentResponse, error) {
	studentID, err := uuid.Parse(req.GetStudentId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid student_id")
	}
	semesterID, err := uuid.Parse(req.GetSemesterId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid semester_id")
	}
	offeredSubjectID, err := uuid.Parse(req.GetOfferedSubjectId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid offered_subject_id")
	}
	subjectID, err := uuid.Parse(req.GetSubjectId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid subject_id")
	}

	enrollment, err := s.requestEnrollment.Handle(ctx, command.RequestEnrollmentCommand{
		StudentID:        studentID,
		SemesterID:       semesterID,
		OfferedSubjectID: offeredSubjectID,
		SubjectID:        subjectID,
		RequestNote:      req.GetRequestNote(),
	})
	if err != nil {
		var prereqErr *command.PrerequisiteViolationError
		switch {
		case errors.As(err, &prereqErr):
			return nil, status.Error(codes.FailedPrecondition, prereqErr.Error())
		case errors.Is(err, command.ErrDuplicateEnrollmentRequest):
			return nil, status.Error(codes.AlreadyExists, "enrollment request already exists")
		case errors.Is(err, pgx.ErrNoRows):
			return nil, status.Error(codes.NotFound, "student not found")
		default:
			return nil, status.Errorf(codes.Internal, "request enrollment: %v", err)
		}
	}

	return &studentv1.RequestEnrollmentResponse{Enrollment: enrollmentToProto(enrollment)}, nil
}

func (s *StudentServer) ReviewEnrollment(ctx context.Context, req *studentv1.ReviewEnrollmentRequest) (*studentv1.ReviewEnrollmentResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}
	reviewedBy, err := uuid.Parse(req.GetReviewedBy())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid reviewed_by")
	}

	enrollment, err := s.reviewEnrollment.Handle(ctx, command.ReviewEnrollmentCommand{
		ID:         id,
		Approve:    req.GetApprove(),
		AdminNote:  req.GetAdminNote(),
		ReviewedBy: reviewedBy,
	})
	if err != nil {
		switch {
		case errors.Is(err, command.ErrEnrollmentNotPending):
			return nil, status.Error(codes.FailedPrecondition, "enrollment request is not pending")
		case errors.Is(err, pgx.ErrNoRows):
			return nil, status.Error(codes.NotFound, "enrollment request not found")
		default:
			return nil, status.Errorf(codes.Internal, "review enrollment: %v", err)
		}
	}

	return &studentv1.ReviewEnrollmentResponse{Enrollment: enrollmentToProto(enrollment)}, nil
}

func (s *StudentServer) ListEnrollmentRequests(ctx context.Context, req *studentv1.ListEnrollmentRequestsRequest) (*studentv1.ListEnrollmentRequestsResponse, error) {
	q := query.ListEnrollmentRequestsQuery{}
	if req.GetPagination() != nil {
		q.Page = req.GetPagination().GetPage()
		q.PageSize = req.GetPagination().GetPageSize()
	}
	if req.StudentId != nil && req.GetStudentId() != "" {
		studentID, err := uuid.Parse(req.GetStudentId())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid student_id")
		}
		q.StudentID = &studentID
	}
	if req.SemesterId != nil && req.GetSemesterId() != "" {
		semesterID, err := uuid.Parse(req.GetSemesterId())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid semester_id")
		}
		q.SemesterID = &semesterID
	}
	if req.Status != nil && req.GetStatus() != "" {
		statusValue := req.GetStatus()
		q.Status = &statusValue
	}

	result, err := s.listEnrollmentRequests.Handle(ctx, q)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list enrollment requests: %v", err)
	}

	enrollments := make([]*studentv1.EnrollmentRequest, len(result.Enrollments))
	for i, enrollment := range result.Enrollments {
		enrollments[i] = enrollmentToProto(enrollment)
	}

	pageSize := int32(20)
	if req.GetPagination() != nil && req.GetPagination().GetPageSize() > 0 {
		pageSize = req.GetPagination().GetPageSize()
	}
	page := int32(1)
	if req.GetPagination() != nil && req.GetPagination().GetPage() > 0 {
		page = req.GetPagination().GetPage()
	}

	return &studentv1.ListEnrollmentRequestsResponse{
		Enrollments: enrollments,
		Pagination: &corev1.PaginationResponse{
			Total:    int32(result.Total),
			Page:     page,
			PageSize: pageSize,
		},
	}, nil
}

func (s *StudentServer) GetStudentEnrollments(ctx context.Context, req *studentv1.GetStudentEnrollmentsRequest) (*studentv1.GetStudentEnrollmentsResponse, error) {
	studentID, err := uuid.Parse(req.GetStudentId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid student_id")
	}

	q := query.GetStudentEnrollmentsQuery{StudentID: studentID}
	if req.SemesterId != nil && req.GetSemesterId() != "" {
		semesterID, err := uuid.Parse(req.GetSemesterId())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid semester_id")
		}
		q.SemesterID = &semesterID
	}

	enrollments, err := s.getStudentEnrollments.Handle(ctx, q)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get student enrollments: %v", err)
	}

	protos := make([]*studentv1.EnrollmentRequest, len(enrollments))
	for i, enrollment := range enrollments {
		protos[i] = enrollmentToProto(enrollment)
	}

	return &studentv1.GetStudentEnrollmentsResponse{Enrollments: protos}, nil
}

func (s *StudentServer) CheckPrerequisites(ctx context.Context, req *studentv1.CheckPrerequisitesRequest) (*studentv1.CheckPrerequisitesResponse, error) {
	if s.prerequisiteChecker == nil {
		return nil, status.Error(codes.Unavailable, "prerequisite checker unavailable")
	}

	studentID, err := uuid.Parse(req.GetStudentId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid student_id")
	}
	subjectID, err := uuid.Parse(req.GetSubjectId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid subject_id")
	}

	missing, err := s.prerequisiteChecker.Check(ctx, studentID, subjectID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "check prerequisites: %v", err)
	}

	return &studentv1.CheckPrerequisitesResponse{
		CanEnroll: len(missing) == 0,
		Missing:   missingPrerequisitesToProto(missing),
	}, nil
}

// enrollmentToProto converts a domain EnrollmentRequest entity to its proto representation.
func enrollmentToProto(enrollment *entity.EnrollmentRequest) *studentv1.EnrollmentRequest {
	protoEnrollment := &studentv1.EnrollmentRequest{
		Id:               enrollment.ID.String(),
		StudentId:        enrollment.StudentID.String(),
		SemesterId:       enrollment.SemesterID.String(),
		OfferedSubjectId: enrollment.OfferedSubjectID.String(),
		SubjectId:        enrollment.SubjectID.String(),
		Status:           enrollment.Status,
		RequestNote:      enrollment.RequestNote,
		AdminNote:        enrollment.AdminNote,
		RequestedAt:      timestamppb.New(enrollment.RequestedAt),
	}
	if enrollment.ReviewedAt != nil {
		protoEnrollment.ReviewedAt = timestamppb.New(*enrollment.ReviewedAt)
	}
	if enrollment.ReviewedBy != nil {
		protoEnrollment.ReviewedBy = enrollment.ReviewedBy.String()
	}
	return protoEnrollment
}

// missingPrerequisitesToProto converts a slice of MissingPrerequisite to proto.
func missingPrerequisitesToProto(missing []appservice.MissingPrerequisite) []*studentv1.MissingPrerequisite {
	protos := make([]*studentv1.MissingPrerequisite, len(missing))
	for i, item := range missing {
		protos[i] = &studentv1.MissingPrerequisite{
			SubjectId:   item.SubjectID.String(),
			SubjectCode: item.SubjectCode,
			SubjectName: item.SubjectName,
			Type:        item.Type,
		}
	}
	return protos
}
