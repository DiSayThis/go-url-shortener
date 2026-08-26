package auth

import (
	"context"
	"fmt"
	"go-api/internal/database"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshRepository struct {
	pool    *pgxpool.Pool
	queries *database.Queries
}
type RefreshStore interface {
	CreateRefreshToken(ctx context.Context, params CreateRefreshTokenParams) (*database.RefreshToken, error)
}

type CreateRefreshTokenParams struct {
	UserID    int64
	TokenHash []byte
	ExpiresAt time.Time
	UserAgent string
}

func NewRefreshRepository(pool *pgxpool.Pool) *RefreshRepository {
	return &RefreshRepository{
		pool:    pool,
		queries: database.New(pool),
	}
}

func (repo *RefreshRepository) CreateRefreshToken(ctx context.Context, params CreateRefreshTokenParams) (*database.RefreshToken, error) {
	token, err := repo.queries.CreateRefreshToken(
		ctx,
		database.CreateRefreshTokenParams{
			UserID:    params.UserID,
			TokenHash: params.TokenHash,
			ExpiresAt: pgtype.Timestamptz{
				Time:  params.ExpiresAt.UTC(),
				Valid: true,
			},
			CreatedIp: nil,
			UserAgent: pgtype.Text{
				String: params.UserAgent,
				Valid:  params.UserAgent != "",
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("insert refresh token: %w", err)
	}

	return &token, nil
}

type RotateRefreshTokenParams struct {
	CurrentTokenHash []byte
	NewTokenHash     []byte
	ExpiresAt        time.Time
	UserAgent        string
}
