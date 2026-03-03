package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/repository"
)

// RedeemInviteCodeCommand holds the plaintext code and the user claiming it.
type RedeemInviteCodeCommand struct {
	Code   string
	UserID uuid.UUID
}

// RedeemInviteCodeHandler atomically claims an invite code and links the user to the student.
type RedeemInviteCodeHandler struct {
	inviteCodeRepo       repository.InviteCodeRepository
	linkUserToStudent    *LinkUserToStudentHandler
	publisher            EventPublisher
}

func NewRedeemInviteCodeHandler(
	inviteCodeRepo repository.InviteCodeRepository,
	linkUserToStudent *LinkUserToStudentHandler,
	publisher EventPublisher,
) *RedeemInviteCodeHandler {
	return &RedeemInviteCodeHandler{
		inviteCodeRepo:    inviteCodeRepo,
		linkUserToStudent: linkUserToStudent,
		publisher:         publisher,
	}
}

func (h *RedeemInviteCodeHandler) Handle(ctx context.Context, cmd RedeemInviteCodeCommand) (*entity.Student, error) {
	codeHash := HashInviteCode(cmd.Code)

	inviteCode, err := h.inviteCodeRepo.FindByCodeHash(ctx, codeHash)
	if err != nil {
		return nil, fmt.Errorf("invite code not found")
	}

	if !inviteCode.IsValid() {
		if inviteCode.IsExpired() {
			return nil, fmt.Errorf("invite code has expired")
		}
		return nil, fmt.Errorf("invite code has already been used")
	}

	// Atomically mark code as used — WHERE used_at IS NULL prevents TOCTOU race.
	// If another request claimed the code first, MarkUsed returns an error.
	if _, err := h.inviteCodeRepo.MarkUsed(ctx, inviteCode.ID, cmd.UserID); err != nil {
		return nil, fmt.Errorf("invite code already claimed")
	}

	// Link the user account to the student record.
	student, err := h.linkUserToStudent.Handle(ctx, LinkUserToStudentCommand{
		StudentID: inviteCode.StudentID,
		UserID:    cmd.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("link user to student: %w", err)
	}

	_ = h.publisher.Publish(ctx, "student.invite_code_redeemed", student)

	return student, nil
}
