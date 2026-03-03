package command

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/repository"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/infrastructure/persistence"
)

// CreateInviteCodeCommand holds data for invite code creation.
type CreateInviteCodeCommand struct {
	StudentID    uuid.UUID
	CreatedBy    uuid.UUID
	ExpiresHours int32 // default 48
}

// CreateInviteCodeResult holds the plaintext code (shown only once) and metadata.
type CreateInviteCodeResult struct {
	Code      string // plaintext — caller must store securely, not persisted
	StudentID uuid.UUID
	ExpiresAt string // ISO 8601
}

// CreateInviteCodeHandler generates a single-use invite code for a student.
type CreateInviteCodeHandler struct {
	studentRepo    repository.StudentRepository
	inviteCodeRepo repository.InviteCodeRepository
	publisher      EventPublisher
}

func NewCreateInviteCodeHandler(
	studentRepo repository.StudentRepository,
	inviteCodeRepo repository.InviteCodeRepository,
	publisher EventPublisher,
) *CreateInviteCodeHandler {
	return &CreateInviteCodeHandler{
		studentRepo:    studentRepo,
		inviteCodeRepo: inviteCodeRepo,
		publisher:      publisher,
	}
}

func (h *CreateInviteCodeHandler) Handle(ctx context.Context, cmd CreateInviteCodeCommand) (*CreateInviteCodeResult, error) {
	// Guard: reject if student is already linked to a user account.
	student, err := h.studentRepo.GetByID(ctx, cmd.StudentID)
	if err != nil {
		return nil, fmt.Errorf("get student: %w", err)
	}
	if student.UserID != nil {
		return nil, fmt.Errorf("student is already linked to a user account")
	}

	// Expire all existing active codes for this student (prevents multiple valid codes).
	if err := h.inviteCodeRepo.InvalidateStudentCodes(ctx, cmd.StudentID); err != nil {
		return nil, fmt.Errorf("invalidate existing codes: %w", err)
	}

	// Generate 16 cryptographically random bytes → 32-char hex plaintext code.
	rawBytes := make([]byte, 16)
	if _, err := rand.Read(rawBytes); err != nil {
		return nil, fmt.Errorf("generate random code: %w", err)
	}
	plaintext := hex.EncodeToString(rawBytes) // 32 hex chars

	// SHA-256 hash for storage.
	hashBytes := sha256.Sum256([]byte(plaintext))
	codeHash := hex.EncodeToString(hashBytes[:])

	expiresAt := persistence.InviteCodeExpiresAt(cmd.ExpiresHours)

	code := &entity.InviteCode{
		CodeHash:  codeHash,
		StudentID: cmd.StudentID,
		ExpiresAt: expiresAt,
		CreatedBy: cmd.CreatedBy,
	}

	created, err := h.inviteCodeRepo.Create(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("persist invite code: %w", err)
	}

	_ = h.publisher.Publish(ctx, "student.invite_code_created", created)

	return &CreateInviteCodeResult{
		Code:      plaintext,
		StudentID: created.StudentID,
		ExpiresAt: created.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z"),
	}, nil
}

// HashInviteCode computes the SHA-256 hex hash of a plaintext code.
// Used by validate and redeem handlers to look up by hash.
func HashInviteCode(plaintext string) string {
	h := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(h[:])
}
