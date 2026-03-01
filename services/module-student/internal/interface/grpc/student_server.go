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

// StudentServer implements studentv1.StudentServiceServer.
type StudentServer struct {
	studentv1.UnimplementedStudentServiceServer

	createStudent          *command.CreateStudentHandler
	updateStudent          *command.UpdateStudentHandler
	deleteStudent          *command.DeleteStudentHandler
	getStudent             *query.GetStudentHandler
	listStudents           *query.ListStudentsHandler
	requestEnrollment      *command.RequestEnrollmentHandler
	reviewEnrollment       *command.ReviewEnrollmentHandler
	listEnrollmentRequests *query.ListEnrollmentRequestsHandler
	getStudentEnrollments  *query.GetStudentEnrollmentsHandler
	prerequisiteChecker    command.PrerequisiteChecker
	assignGrade            *command.AssignGradeHandler
	updateGrade            *command.UpdateGradeHandler
	getTranscript          *query.GetStudentTranscriptHandler
}

func NewStudentServer(
	createStudent *command.CreateStudentHandler,
	updateStudent *command.UpdateStudentHandler,
	deleteStudent *command.DeleteStudentHandler,
	getStudent *query.GetStudentHandler,
	listStudents *query.ListStudentsHandler,
	requestEnrollment *command.RequestEnrollmentHandler,
	reviewEnrollment *command.ReviewEnrollmentHandler,
	listEnrollmentRequests *query.ListEnrollmentRequestsHandler,
	getStudentEnrollments *query.GetStudentEnrollmentsHandler,
	prerequisiteChecker command.PrerequisiteChecker,
	assignGrade *command.AssignGradeHandler,
	updateGrade *command.UpdateGradeHandler,
	getTranscript *query.GetStudentTranscriptHandler,
) *StudentServer {
	return &StudentServer{
		createStudent:          createStudent,
		updateStudent:          updateStudent,
		deleteStudent:          deleteStudent,
		getStudent:             getStudent,
		listStudents:           listStudents,
		requestEnrollment:      requestEnrollment,
		reviewEnrollment:       reviewEnrollment,
		listEnrollmentRequests: listEnrollmentRequests,
		getStudentEnrollments:  getStudentEnrollments,
		prerequisiteChecker:    prerequisiteChecker,
		assignGrade:            assignGrade,
		updateGrade:            updateGrade,
		getTranscript:          getTranscript,
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

func enrollmentToProto(enrollment *entity.EnrollmentRequest) *studentv1.EnrollmentRequest {
	protoEnrollment := &studentv1.EnrollmentRequest{
		Id:              enrollment.ID.String(),
		StudentId:       enrollment.StudentID.String(),
		SemesterId:      enrollment.SemesterID.String(),
		OfferedSubjectId: enrollment.OfferedSubjectID.String(),
		SubjectId:       enrollment.SubjectID.String(),
		Status:          enrollment.Status,
		RequestNote:     enrollment.RequestNote,
		AdminNote:       enrollment.AdminNote,
		RequestedAt:     timestamppb.New(enrollment.RequestedAt),
	}
	if enrollment.ReviewedAt != nil {
		protoEnrollment.ReviewedAt = timestamppb.New(*enrollment.ReviewedAt)
	}
	if enrollment.ReviewedBy != nil {
		protoEnrollment.ReviewedBy = enrollment.ReviewedBy.String()
	}
	return protoEnrollment
}

func (s *StudentServer) AssignGrade(ctx context.Context, req *studentv1.AssignGradeRequest) (*studentv1.AssignGradeResponse, error) {
	enrollmentID, err := uuid.Parse(req.GetEnrollmentId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid enrollment_id")
	}
	gradedBy, err := uuid.Parse(req.GetGradedBy())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid graded_by")
	}

	grade, err := s.assignGrade.Handle(ctx, command.AssignGradeCommand{
		EnrollmentID: enrollmentID,
		GradeNumeric: req.GetGradeNumeric(),
		GradedBy:     gradedBy,
		Notes:        req.GetNotes(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "assign grade: %v", err)
	}

	return &studentv1.AssignGradeResponse{Grade: gradeToProto(grade)}, nil
}

func (s *StudentServer) UpdateGrade(ctx context.Context, req *studentv1.UpdateGradeRequest) (*studentv1.UpdateGradeResponse, error) {
	gradeID, err := uuid.Parse(req.GetGradeId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid grade_id")
	}
	gradedBy, err := uuid.Parse(req.GetGradedBy())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid graded_by")
	}

	grade, err := s.updateGrade.Handle(ctx, command.UpdateGradeCommand{
		GradeID:      gradeID,
		GradeNumeric: req.GetGradeNumeric(),
		GradedBy:     gradedBy,
		Notes:        req.GetNotes(),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "grade not found")
		}
		return nil, status.Errorf(codes.Internal, "update grade: %v", err)
	}

	return &studentv1.UpdateGradeResponse{Grade: gradeToProto(grade)}, nil
}

func (s *StudentServer) GetStudentTranscript(ctx context.Context, req *studentv1.GetStudentTranscriptRequest) (*studentv1.GetStudentTranscriptResponse, error) {
	studentID, err := uuid.Parse(req.GetStudentId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid student_id")
	}

	transcript, err := s.getTranscript.Handle(ctx, query.GetStudentTranscriptQuery{StudentID: studentID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "student not found")
		}
		return nil, status.Errorf(codes.Internal, "get transcript: %v", err)
	}

	entries := make([]*studentv1.TranscriptEntry, len(transcript.Entries))
	for i, e := range transcript.Entries {
		entry := &studentv1.TranscriptEntry{
			EnrollmentId: e.EnrollmentID,
			SemesterId:   e.SemesterID,
			SubjectId:    e.SubjectID,
			SubjectCode:  e.SubjectCode,
			SubjectName:  e.SubjectName,
			Credits:      e.Credits,
			Status:       e.Status,
			GradeNumeric: e.GradeNumeric,
			GradeLetter:  e.GradeLetter,
		}
		if e.GradedAt != nil {
			entry.GradedAt = timestamppb.New(*e.GradedAt)
		}
		entries[i] = entry
	}

	return &studentv1.GetStudentTranscriptResponse{
		Student:       studentToProto(transcript.Student),
		Entries:       entries,
		Gpa:           transcript.GPA,
		TotalCredits:  transcript.TotalCredits,
		PassedCredits: transcript.PassedCredits,
	}, nil
}

func gradeToProto(grade *entity.Grade) *studentv1.Grade {
	return &studentv1.Grade{
		Id:           grade.ID.String(),
		EnrollmentId: grade.EnrollmentID.String(),
		GradeNumeric: grade.GradeNumeric,
		GradeLetter:  grade.GradeLetter,
		GradedBy:     grade.GradedBy.String(),
		GradedAt:     timestamppb.New(grade.GradedAt),
		Notes:        grade.Notes,
	}
}

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
