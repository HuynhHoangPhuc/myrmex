package grpc

import (
	"context"
	"errors"

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

// gradeToProto converts a domain Grade entity to its proto representation.
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
