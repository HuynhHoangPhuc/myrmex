package grpc

import (
	"context"

	studentv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/student/v1"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/application/command"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/application/query"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *StudentServer) CreateInviteCode(ctx context.Context, req *studentv1.CreateInviteCodeRequest) (*studentv1.CreateInviteCodeResponse, error) {
	studentID, err := uuid.Parse(req.GetStudentId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid student_id")
	}
	createdBy, err := uuid.Parse(req.GetCreatedBy())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid created_by")
	}

	result, err := s.createInviteCode.Handle(ctx, command.CreateInviteCodeCommand{
		StudentID:    studentID,
		CreatedBy:    createdBy,
		ExpiresHours: req.GetExpiresHours(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create invite code: %v", err)
	}

	return &studentv1.CreateInviteCodeResponse{
		Code:      result.Code,
		StudentId: result.StudentID.String(),
		ExpiresAt: result.ExpiresAt,
	}, nil
}

func (s *StudentServer) ValidateInviteCode(ctx context.Context, req *studentv1.ValidateInviteCodeRequest) (*studentv1.ValidateInviteCodeResponse, error) {
	if req.GetCode() == "" {
		return nil, status.Error(codes.InvalidArgument, "code is required")
	}

	result, err := s.validateInviteCode.Handle(ctx, query.ValidateInviteCodeQuery{Code: req.GetCode()})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "validate invite code: %v", err)
	}

	return &studentv1.ValidateInviteCodeResponse{
		Valid:       result.Valid,
		StudentId:   result.StudentID,
		StudentName: result.StudentName,
		Message:     result.Message,
	}, nil
}

func (s *StudentServer) RedeemInviteCode(ctx context.Context, req *studentv1.RedeemInviteCodeRequest) (*studentv1.RedeemInviteCodeResponse, error) {
	if req.GetCode() == "" {
		return nil, status.Error(codes.InvalidArgument, "code is required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	student, err := s.redeemInviteCode.Handle(ctx, command.RedeemInviteCodeCommand{
		Code:   req.GetCode(),
		UserID: userID,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "redeem invite code: %v", err)
	}

	return &studentv1.RedeemInviteCodeResponse{Student: studentToProto(student)}, nil
}
