package query

import (
	"context"

	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/application/command"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/repository"
)

// ValidateInviteCodeQuery holds the plaintext code to validate.
type ValidateInviteCodeQuery struct {
	Code string
}

// ValidateInviteCodeResult reports whether a code is valid and the associated student info.
type ValidateInviteCodeResult struct {
	Valid       bool
	StudentID   string
	StudentName string
	Message     string // reason when invalid
}

// ValidateInviteCodeHandler checks whether an invite code is valid without consuming it.
type ValidateInviteCodeHandler struct {
	inviteCodeRepo repository.InviteCodeRepository
	studentRepo    repository.StudentRepository
}

func NewValidateInviteCodeHandler(
	inviteCodeRepo repository.InviteCodeRepository,
	studentRepo repository.StudentRepository,
) *ValidateInviteCodeHandler {
	return &ValidateInviteCodeHandler{
		inviteCodeRepo: inviteCodeRepo,
		studentRepo:    studentRepo,
	}
}

func (h *ValidateInviteCodeHandler) Handle(ctx context.Context, q ValidateInviteCodeQuery) (*ValidateInviteCodeResult, error) {
	codeHash := command.HashInviteCode(q.Code)

	inviteCode, err := h.inviteCodeRepo.FindByCodeHash(ctx, codeHash)
	if err != nil {
		return &ValidateInviteCodeResult{Valid: false, Message: "invalid invite code"}, nil
	}

	if inviteCode.IsExpired() {
		return &ValidateInviteCodeResult{Valid: false, Message: "invite code has expired"}, nil
	}
	if inviteCode.IsUsed() {
		return &ValidateInviteCodeResult{Valid: false, Message: "invite code has already been used"}, nil
	}

	student, err := h.studentRepo.GetByID(ctx, inviteCode.StudentID)
	if err != nil {
		return &ValidateInviteCodeResult{Valid: false, Message: "associated student not found"}, nil
	}

	return &ValidateInviteCodeResult{
		Valid:       true,
		StudentID:   student.ID.String(),
		StudentName: student.FullName,
	}, nil
}
