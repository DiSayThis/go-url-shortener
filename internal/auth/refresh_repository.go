package auth

import (
	"context"
	"errors"
	"fmt"
	"go-api/internal/database"
	"net/netip"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshRepository struct {
	pool    *pgxpool.Pool
	queries *database.Queries
}
type RevokeUserSessionParams struct {
	UserID        int64
	FamilyID      string
	RevokedReason refreshTokenRevokeReason
}

type RevokeAllUserSessionsParams struct {
	UserID        int64
	RevokedReason refreshTokenRevokeReason
}
type RefreshStore interface {
	CreateRefreshToken(ctx context.Context, params CreateRefreshTokenParams) (*database.RefreshToken, error)
	RotateRefreshToken(ctx context.Context, params RotateRefreshTokenParams) (*RotateRefreshTokenResult, error)
	ListActiveSessions(ctx context.Context, userID int64) ([]Session, error)
	RevokeUserSession(ctx context.Context, params RevokeUserSessionParams) (int64, error)
	RevokeAllUserSessions(ctx context.Context, params RevokeAllUserSessionsParams) (int64, error)
}

type CreateRefreshTokenParams struct {
	UserID    int64
	TokenHash []byte
	ExpiresAt time.Time
	UserAgent string
	CreatedIP *netip.Addr
}

type refreshTokenRevokeReason string

const (
	refreshTokenRevokeReasonReuse        refreshTokenRevokeReason = "refresh_token_reuse"
	refreshTokenRevokeReasonUserNotFound refreshTokenRevokeReason = "user_not_found"
	refreshTokenRevokeReasonUserInactive refreshTokenRevokeReason = "user_inactive"

	refreshTokenRevokeReasonLogout        refreshTokenRevokeReason = "user_logout"
	refreshTokenRevokeReasonSessionDelete refreshTokenRevokeReason = "user_revoked_session"
	refreshTokenRevokeReasonLogoutAll     refreshTokenRevokeReason = "user_logout_all"
)

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
			CreatedIp: params.CreatedIP,
			UserAgent: pgtype.Text{
				String: params.UserAgent,
				Valid:  params.UserAgent != "",
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("insert initial refresh token: %w", err)
	}

	return &token, nil
}

type RotateRefreshTokenParams struct {
	CurrentTokenHash []byte
	NewTokenHash     []byte
	UserAgent        string
	ClientIP         *netip.Addr
}
type RotateRefreshTokenResult struct {
	RefreshToken database.RefreshToken
	User         database.User
}

func (repo *RefreshRepository) RotateRefreshToken(ctx context.Context, params RotateRefreshTokenParams) (*RotateRefreshTokenResult, error) {
	const op = "rotate refresh token"
	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: begin transaction: %w", op, err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()
	txQueries := repo.queries.WithTx(tx)
	current, err := txQueries.GetRefreshTokenForUpdate(ctx, params.CurrentTokenHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrInvalidRefreshToken
	}
	if err != nil {
		return nil, fmt.Errorf("%s: lock current token: %w", op, err)
	}

	revokeFamily := func(reason refreshTokenRevokeReason) error {
		_, err := txQueries.RevokeRefreshTokenFamily(
			ctx,
			database.RevokeRefreshTokenFamilyParams{
				FamilyID: current.FamilyID,
				RevokedReason: pgtype.Text{
					String: string(reason),
					Valid:  true,
				},
			},
		)
		return err
	}

	if current.UsedAt.Valid {
		err := revokeFamily(refreshTokenRevokeReasonReuse)
		if err != nil {
			return nil, fmt.Errorf(
				"%s: revoke token family (refresh_token_reuse): %w",
				op,
				err,
			)
		}

		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf(
				"%s: commit token family revocation (refresh_token_reuse): %w",
				op,
				err,
			)
		}

		return nil, ErrRefreshTokenReused
	}

	if current.RevokedAt.Valid {
		return nil, ErrRefreshTokenRevoked
	}

	user, err := txQueries.GetUserByID(ctx, current.UserID)
	if errors.Is(err, pgx.ErrNoRows) {
		err := revokeFamily(refreshTokenRevokeReasonUserNotFound)
		if err != nil {
			return nil, fmt.Errorf(
				"%s: revoke token family (user_not_found): %w",
				op,
				err,
			)
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf(
				"%s: commit token family revocation (user_not_found): %w",
				op,
				err,
			)
		}
		return nil, ErrInvalidRefreshSession
	}
	if err != nil {
		return nil, fmt.Errorf("%s: get session user: %w", op, err)
	}
	if user.Status != "active" {
		err := revokeFamily(refreshTokenRevokeReasonUserInactive)
		if err != nil {
			return nil, fmt.Errorf(
				"%s: revoke token family (user_inactive): %w",
				op,
				err,
			)
		}

		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf(
				"%s: commit token family revocation (user_inactive): %w",
				op,
				err,
			)
		}
		return nil, ErrInvalidRefreshSession
	}

	_, err = txQueries.MarkRefreshTokenUsed(
		ctx,
		database.MarkRefreshTokenUsedParams{
			ID:         current.ID,
			LastUsedIp: params.ClientIP,
		},
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRefreshTokenExpired
	}
	if err != nil {
		return nil, fmt.Errorf("%s: mark current token as used: %w", op, err)
	}
	rotated, err := txQueries.CreateRotatedRefreshToken(
		ctx, database.CreateRotatedRefreshTokenParams{
			FamilyID:  current.FamilyID,
			UserID:    current.UserID,
			TokenHash: params.NewTokenHash,
			ParentID:  current.ID,
			ExpiresAt: current.ExpiresAt,
			CreatedIp: params.ClientIP,
			UserAgent: pgtype.Text{
				String: params.UserAgent,
				Valid:  params.UserAgent != "",
			},
		})
	if err != nil {
		return nil, fmt.Errorf("%s: insert rotated token: %w", op, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%s: commit transaction: %w", op, err)
	}

	return &RotateRefreshTokenResult{
		RefreshToken: rotated,
		User:         user,
	}, nil
}

func (repo *RefreshRepository) ListActiveSessions(
	ctx context.Context,
	userID int64,
) ([]Session, error) {
	return nil, ErrSessionOperationNotImplemented
}

func (repo *RefreshRepository) RevokeUserSession(
	ctx context.Context,
	params RevokeUserSessionParams,
) (int64, error) {
	return 0, ErrSessionOperationNotImplemented
}

func (repo *RefreshRepository) RevokeAllUserSessions(
	ctx context.Context,
	params RevokeAllUserSessionsParams,
) (int64, error) {
	return 0, ErrSessionOperationNotImplemented
}
