package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/domain/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) (*entity.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	List(ctx context.Context, limit, offset int32) ([]*entity.User, error)
	Count(ctx context.Context) (int64, error)
	Update(ctx context.Context, user *entity.User) (*entity.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	// UpdateRole sets role and department_id atomically (admin role management).
	UpdateRole(ctx context.Context, userID uuid.UUID, role string, departmentID *uuid.UUID) (*entity.User, error)
	// GetTeacherIDByUserID looks up teacher record linked to this user (cross-schema, for JWT).
	// Returns ("", nil) if no teacher record exists.
	GetTeacherIDByUserID(ctx context.Context, userID uuid.UUID) (string, error)
	// GetByOAuth finds an existing user by their OAuth provider + subject ID.
	// Returns nil, nil when no matching user exists (not an error).
	GetByOAuth(ctx context.Context, provider, subject string) (*entity.User, error)
	// UpsertOAuthUser creates or links an OAuth user by email.
	// New: inserts with empty password_hash; Existing: links OAuth fields to the account.
	UpsertOAuthUser(ctx context.Context, email, fullName, role, provider, subject, avatarURL string) (*entity.User, error)
}
