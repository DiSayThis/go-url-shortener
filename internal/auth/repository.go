package auth

import (
	"context"
	"errors"
	"fmt"

	"go-api/internal/database"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// CreateUserParams принадлежит auth-модулю.
// Благодаря этому service не зависит от сгенерированного sqlc-типа параметров.
type CreateUserParams struct {
	Email        string
	DisplayName  string
	PasswordHash string
}

type AuthRepository interface {
	CreateUser(ctx context.Context, params CreateUserParams) (*database.User, error)
	GetUserByEmail(ctx context.Context, email string) (*database.User, error)
	GetUserByID(ctx context.Context, userID int64) (*database.User, error)
	UpdateLastLogin(ctx context.Context, userID int64) error
}

type Repository struct {
	queries database.Querier
}

func NewRepository(queries database.Querier) *Repository {
	return &Repository{queries: queries}
}

func (repo *Repository) CreateUser(
	ctx context.Context,
	params CreateUserParams,
) (*database.User, error) {
	user, err := repo.queries.CreateUser(ctx, database.CreateUserParams{
		Email:       params.Email,
		DisplayName: params.DisplayName,
		PasswordHash: pgtype.Text{
			String: params.PasswordHash,
			Valid:  true,
		},
	})
	if err != nil {
		if isEmailCollision(err) {
			return nil, ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("insert user: %w", err)
	}

	return &user, nil
}

func (repo *Repository) GetUserByEmail(
	ctx context.Context,
	email string,
) (*database.User, error) {
	user, err := repo.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("select user by email: %w", err)
	}

	return &user, nil
}

func (repo *Repository) GetUserByID(
	ctx context.Context,
	userID int64,
) (*database.User, error) {
	user, err := repo.queries.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("select user by ID: %w", err)
	}

	return &user, nil
}

func (repo *Repository) UpdateLastLogin(ctx context.Context, userID int64) error {
	if err := repo.queries.UpdateUserLastLogin(ctx, userID); err != nil {
		return fmt.Errorf("update user last login: %w", err)
	}

	return nil
}
