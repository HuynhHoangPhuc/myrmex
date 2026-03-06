package grpc

import (
	"context"
	"errors"

	corev1 "github.com/HuynhHoangPhuc/myrmex/gen/go/core/v1"
	studentv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/student/v1"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/application/command"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/application/query"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// StudentServer implements studentv1.StudentServiceServer.
type StudentServer struct {
	studentv1.UnimplementedStudentServiceServer

	createStudent          *command.CreateStudentHandler
	updateStudent          *command.UpdateStudentHandler
	deleteStudent          *command.DeleteStudentHandler
	linkUserToStudent      *command.LinkUserToStudentHandler
	getStudent             *query.GetStudentHandler
	getStudentByUserID     *query.GetStudentByUserIDHandler
	listStudents           *query.ListStudentsHandler
	requestEnrollment      *command.RequestEnrollmentHandler
	reviewEnrollment       *command.ReviewEnrollmentHandler
	listEnrollmentRequests *query.ListEnrollmentRequestsHandler
	getStudentEnrollments  *query.GetStudentEnrollmentsHandler
	prerequisiteChecker    command.PrerequisiteChecker
	assignGrade            *command.AssignGradeHandler
	updateGrade            *command.UpdateGradeHandler
	getTranscript          *query.GetStudentTranscriptHandler
	createInviteCode       *command.CreateInviteCodeHandler
	validateInviteCode     *query.ValidateInviteCodeHandler
	redeemInviteCode       *command.RedeemInviteCodeHandler
}

func NewStudentServer(
	createStudent *command.CreateStudentHandler,
	updateStudent *command.UpdateStudentHandler,
	deleteStudent *command.DeleteStudentHandler,
	linkUserToStudent *command.LinkUserToStudentHandler,
	getStudent *query.GetStudentHandler,
	getStudentByUserID *query.GetStudentByUserIDHandler,
	listStudents *query.ListStudentsHandler,
	requestEnrollment *command.RequestEnrollmentHandler,
	reviewEnrollment *command.ReviewEnrollmentHandler,
	listEnrollmentRequests *query.ListEnrollmentRequestsHandler,
	getStudentEnrollments *query.GetStudentEnrollmentsHandler,
	prerequisiteChecker command.PrerequisiteChecker,
	assignGrade *command.AssignGradeHandler,
	updateGrade *command.UpdateGradeHandler,
	getTranscript *query.GetStudentTranscriptHandler,
	createInviteCode *command.CreateInviteCodeHandler,
	validateInviteCode *query.ValidateInviteCodeHandler,
	redeemInviteCode *command.RedeemInviteCodeHandler,
) *StudentServer {
	return &StudentServer{
		createStudent:          createStudent,
		updateStudent:          updateStudent,
		deleteStudent:          deleteStudent,
		linkUserToStudent:      linkUserToStudent,
		getStudent:             getStudent,
		getStudentByUserID:     getStudentByUserID,
		listStudents:           listStudents,
		requestEnrollment:      requestEnrollment,
		reviewEnrollment:       reviewEnrollment,
		listEnrollmentRequests: listEnrollmentRequests,
		getStudentEnrollments:  getStudentEnrollments,
		prerequisiteChecker:    prerequisiteChecker,
		assignGrade:            assignGrade,
		updateGrade:            updateGrade,
		getTranscript:          getTranscript,
		createInviteCode:       createInviteCode,
		validateInviteCode:     validateInviteCode,
		redeemInviteCode:       redeemInviteCode,
	}
}

func (s *StudentServer) CreateStudent(ctx context.Context, req *studentv1.CreateStudentRequest) (*studentv1.CreateStudentResponse, error) {
	if req.GetStudentCode() == "" || req.GetFullName() == "" || req.GetEmail() == "" || req.GetDepartmentId() == "" {
		return nil, status.Error(codes.InvalidArgument, "student_code, full_name, email, and department_id are required")
	}

	departmentID, err := uuid.Parse(req.GetDepartmentId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid department_id")
	}

	student, err := s.createStudent.Handle(ctx, command.CreateStudentCommand{
		StudentCode:    req.GetStudentCode(),
		FullName:       req.GetFullName(),
		Email:          req.GetEmail(),
		DepartmentID:   departmentID,
		EnrollmentYear: int(req.GetEnrollmentYear()),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create student: %v", err)
	}

	return &studentv1.CreateStudentResponse{Student: studentToProto(student)}, nil
}

func (s *StudentServer) GetStudent(ctx context.Context, req *studentv1.GetStudentRequest) (*studentv1.GetStudentResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	student, err := s.getStudent.Handle(ctx, query.GetStudentQuery{ID: id})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "student not found")
		}
		return nil, status.Errorf(codes.Internal, "get student: %v", err)
	}

	return &studentv1.GetStudentResponse{Student: studentToProto(student)}, nil
}

func (s *StudentServer) ListStudents(ctx context.Context, req *studentv1.ListStudentsRequest) (*studentv1.ListStudentsResponse, error) {
	q := query.ListStudentsQuery{}
	if req.GetPagination() != nil {
		q.Page = req.GetPagination().GetPage()
		q.PageSize = req.GetPagination().GetPageSize()
	}
	if req.DepartmentId != nil && req.GetDepartmentId() != "" {
		departmentID, err := uuid.Parse(req.GetDepartmentId())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid department_id")
		}
		q.DepartmentID = &departmentID
	}
	if req.Status != nil && req.GetStatus() != "" {
		statusValue := req.GetStatus()
		q.Status = &statusValue
	}

	result, err := s.listStudents.Handle(ctx, q)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list students: %v", err)
	}

	students := make([]*studentv1.Student, len(result.Students))
	for i, student := range result.Students {
		students[i] = studentToProto(student)
	}

	pageSize := int32(20)
	if req.GetPagination() != nil && req.GetPagination().GetPageSize() > 0 {
		pageSize = req.GetPagination().GetPageSize()
	}
	page := int32(1)
	if req.GetPagination() != nil && req.GetPagination().GetPage() > 0 {
		page = req.GetPagination().GetPage()
	}

	return &studentv1.ListStudentsResponse{
		Students: students,
		Pagination: &corev1.PaginationResponse{
			Total:    int32(result.Total),
			Page:     page,
			PageSize: pageSize,
		},
	}, nil
}

func (s *StudentServer) UpdateStudent(ctx context.Context, req *studentv1.UpdateStudentRequest) (*studentv1.UpdateStudentResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	current, err := s.getStudent.Handle(ctx, query.GetStudentQuery{ID: id})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "student not found")
		}
		return nil, status.Errorf(codes.Internal, "get student: %v", err)
	}

	cmd := command.UpdateStudentCommand{
		ID:             id,
		StudentCode:    current.StudentCode,
		UserID:         current.UserID,
		FullName:       current.FullName,
		Email:          current.Email,
		DepartmentID:   current.DepartmentID,
		EnrollmentYear: current.EnrollmentYear,
		Status:         current.Status,
		IsActive:       current.IsActive,
	}
	if req.FullName != nil {
		cmd.FullName = req.GetFullName()
	}
	if req.Email != nil {
		cmd.Email = req.GetEmail()
	}
	if req.DepartmentId != nil {
		departmentID, err := uuid.Parse(req.GetDepartmentId())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid department_id")
		}
		cmd.DepartmentID = departmentID
	}
	if req.Status != nil {
		cmd.Status = req.GetStatus()
	}
	if req.IsActive != nil {
		cmd.IsActive = req.GetIsActive()
	}

	updated, err := s.updateStudent.Handle(ctx, cmd)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "update student: %v", err)
	}

	return &studentv1.UpdateStudentResponse{Student: studentToProto(updated)}, nil
}

func (s *StudentServer) DeleteStudent(ctx context.Context, req *studentv1.DeleteStudentRequest) (*studentv1.DeleteStudentResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	if err := s.deleteStudent.Handle(ctx, command.DeleteStudentCommand{ID: id}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "student not found")
		}
		return nil, status.Errorf(codes.Internal, "delete student: %v", err)
	}

	return &studentv1.DeleteStudentResponse{Success: true}, nil
}

func (s *StudentServer) GetStudentByUserID(ctx context.Context, req *studentv1.GetStudentByUserIDRequest) (*studentv1.GetStudentByUserIDResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	student, err := s.getStudentByUserID.Handle(ctx, query.GetStudentByUserIDQuery{UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "student not found for user_id")
		}
		return nil, status.Errorf(codes.Internal, "get student by user_id: %v", err)
	}

	return &studentv1.GetStudentByUserIDResponse{Student: studentToProto(student)}, nil
}

func (s *StudentServer) LinkUserToStudent(ctx context.Context, req *studentv1.LinkUserToStudentRequest) (*studentv1.LinkUserToStudentResponse, error) {
	studentID, err := uuid.Parse(req.GetStudentId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid student_id")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	student, err := s.linkUserToStudent.Handle(ctx, command.LinkUserToStudentCommand{
		StudentID: studentID,
		UserID:    userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "student not found")
		}
		return nil, status.Errorf(codes.Internal, "link user to student: %v", err)
	}

	return &studentv1.LinkUserToStudentResponse{Student: studentToProto(student)}, nil
}

// studentToProto converts a domain Student entity to its proto representation.
func studentToProto(student *entity.Student) *studentv1.Student {
	protoStudent := &studentv1.Student{
		Id:             student.ID.String(),
		StudentCode:    student.StudentCode,
		FullName:       student.FullName,
		Email:          student.Email,
		DepartmentId:   student.DepartmentID.String(),
		EnrollmentYear: int32(student.EnrollmentYear),
		Status:         student.Status,
		IsActive:       student.IsActive,
		CreatedAt:      timestamppb.New(student.CreatedAt),
		UpdatedAt:      timestamppb.New(student.UpdatedAt),
	}
	if student.UserID != nil {
		protoStudent.UserId = student.UserID.String()
	}
	return protoStudent
}
