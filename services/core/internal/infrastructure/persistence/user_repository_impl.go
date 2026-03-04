package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/domain/valueobject"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/persistence/sqlc"
)

type UserRepositoryImpl struct {
	queries *sqlc.Queries
}

func NewUserRepository(queries *sqlc.Queries) *UserRepositoryImpl {
	return &UserRepositoryImpl{queries: queries}
}

func (r *UserRepositoryImpl) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	row, err := r.queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		FullName:     user.FullName,
		Role:         string(user.Role),
	})
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return coreUserToEntity(row), nil
}

func (r *UserRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	row, err := r.queries.GetUserByID(ctx, uuidToPgtype(id))
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return coreUserToEntity(row), nil
}

func (r *UserRepositoryImpl) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return coreUserToEntity(row), nil
}

func (r *UserRepositoryImpl) List(ctx context.Context, limit, offset int32) ([]*entity.User, error) {
	rows, err := r.queries.ListUsers(ctx, sqlc.ListUsersParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	users := make([]*entity.User, len(rows))
	for i, row := range rows {
		users[i] = coreUserToEntity(row)
	}
	return users, nil
}

func (r *UserRepositoryImpl) Count(ctx context.Context) (int64, error) {
	return r.queries.CountUsers(ctx)
}

func (r *UserRepositoryImpl) Update(ctx context.Context, user *entity.User) (*entity.User, error) {
	params := sqlc.UpdateUserParams{
		ID:       uuidToPgtype(user.ID),
		FullName: pgtype.Text{String: user.FullName, Valid: user.FullName != ""},
		Email:    pgtype.Text{String: user.Email, Valid: user.Email != ""},
		Role:     pgtype.Text{String: string(user.Role), Valid: user.Role != ""},
		IsActive: pgtype.Bool{Bool: user.IsActive, Valid: true},
	}
	if user.DepartmentID != nil {
		params.DepartmentID = pgtype.UUID{Bytes: *user.DepartmentID, Valid: true}
	}
	row, err := r.queries.UpdateUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return coreUserToEntity(row), nil
}

func (r *UserRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteUser(ctx, uuidToPgtype(id))
}

func (r *UserRepositoryImpl) UpdateRole(ctx context.Context, userID uuid.UUID, role string, departmentID *uuid.UUID) (*entity.User, error) {
	params := sqlc.UpdateUserRoleParams{
		ID:   uuidToPgtype(userID),
		Role: role,
	}
	if departmentID != nil {
		params.DepartmentID = pgtype.UUID{Bytes: *departmentID, Valid: true}
	}
	row, err := r.queries.UpdateUserRole(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("update user role: %w", err)
	}
	return coreUserToEntity(row), nil
}

// GetTeacherIDByUserID looks up the teacher record linked to this user.
// Returns ("", nil) when no teacher record exists (user is not a teacher).
func (r *UserRepositoryImpl) GetTeacherIDByUserID(ctx context.Context, userID uuid.UUID) (string, error) {
	teacherID, err := r.queries.GetTeacherIDByUserID(ctx, uuidToPgtype(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("get teacher id by user id: %w", err)
	}
	if !teacherID.Valid {
		return "", nil
	}
	return uuid.UUID(teacherID.Bytes).String(), nil
}

// GetByOAuth finds a user by their provider+subject. Returns nil, nil if not found.
func (r *UserRepositoryImpl) GetByOAuth(ctx context.Context, provider, subject string) (*entity.User, error) {
	row, err := r.queries.GetUserByOAuth(ctx, sqlc.GetUserByOAuthParams{
		OauthProvider: pgtype.Text{String: provider, Valid: true},
		OauthSubject:  pgtype.Text{String: subject, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by oauth: %w", err)
	}
	return coreUserToEntity(row), nil
}

// UpsertOAuthUser creates or links an OAuth user by email.
func (r *UserRepositoryImpl) UpsertOAuthUser(ctx context.Context, email, fullName, role, provider, subject, avatarURL string) (*entity.User, error) {
	row, err := r.queries.UpsertOAuthUser(ctx, sqlc.UpsertOAuthUserParams{
		Email:         email,
		FullName:      fullName,
		Role:          role,
		OauthProvider: pgtype.Text{String: provider, Valid: true},
		OauthSubject:  pgtype.Text{String: subject, Valid: true},
		AvatarUrl:     pgtype.Text{String: avatarURL, Valid: avatarURL != ""},
	})
	if err != nil {
		return nil, fmt.Errorf("upsert oauth user: %w", err)
	}
	return coreUserToEntity(row), nil
}

func coreUserToEntity(u sqlc.CoreUser) *entity.User {
	e := &entity.User{
		ID:           pgtypeToUUID(u.ID),
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		FullName:     u.FullName,
		Role:         valueobject.Role(u.Role),
		IsActive:     u.IsActive,
		CreatedAt:    u.CreatedAt.Time,
		UpdatedAt:    u.UpdatedAt.Time,
	}
	if u.DepartmentID.Valid {
		id := uuid.UUID(u.DepartmentID.Bytes)
		e.DepartmentID = &id
	}
	if u.OauthProvider.Valid {
		e.OAuthProvider = u.OauthProvider.String
	}
	if u.OauthSubject.Valid {
		e.OAuthSubject = u.OauthSubject.String
	}
	if u.AvatarUrl.Valid {
		e.AvatarURL = u.AvatarUrl.String
	}
	return e
}

func uuidToPgtype(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func pgtypeToUUID(id pgtype.UUID) uuid.UUID {
	return uuid.UUID(id.Bytes)
}
