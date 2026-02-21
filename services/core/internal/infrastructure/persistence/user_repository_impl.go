package persistence

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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
	row, err := r.queries.UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:       uuidToPgtype(user.ID),
		FullName: pgtype.Text{String: user.FullName, Valid: user.FullName != ""},
		Email:    pgtype.Text{String: user.Email, Valid: user.Email != ""},
		Role:     pgtype.Text{String: string(user.Role), Valid: user.Role != ""},
		IsActive: pgtype.Bool{Bool: user.IsActive, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return coreUserToEntity(row), nil
}

func (r *UserRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteUser(ctx, uuidToPgtype(id))
}

func coreUserToEntity(u sqlc.CoreUser) *entity.User {
	return &entity.User{
		ID:           pgtypeToUUID(u.ID),
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		FullName:     u.FullName,
		Role:         valueobject.Role(u.Role),
		IsActive:     u.IsActive,
		CreatedAt:    u.CreatedAt.Time,
		UpdatedAt:    u.UpdatedAt.Time,
	}
}

func uuidToPgtype(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func pgtypeToUUID(id pgtype.UUID) uuid.UUID {
	return uuid.UUID(id.Bytes)
}
