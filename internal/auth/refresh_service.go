package auth

import (
	"context"
	"fmt"
	"net/netip"
	"time"

	"go-api/pkg/jwt"
)

type RefreshInput struct {
	RefreshToken string
	UserAgent    string
	ClientIP     *netip.Addr
}

type RefreshResult struct {
	AccessToken           string
	AccessTokenExpiresAt  time.Time
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
	RefreshTokenTTL       time.Duration
}

func (service *Service) Refresh(
	ctx context.Context,
	input RefreshInput,
) (*RefreshResult, error) {
	const op = "refresh authentication tokens"

	currentTokenHash, err := HashRefreshToken(input.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("%s: validate current refresh token: %w", op, err)
	}

	newRefreshToken, err := GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("%s: generate replacement refresh token: %w", op, err)
	}

	rotation, err := service.refreshRepository.RotateRefreshToken(
		ctx,
		RotateRefreshTokenParams{
			CurrentTokenHash: currentTokenHash,
			NewTokenHash:     newRefreshToken.Hash,
			UserAgent:        input.UserAgent,
			ClientIP:         input.ClientIP,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%s: rotate refresh session: %w", op, err)
	}

	user := rotation.User
	session := rotation.RefreshToken

	// Эти поля являются инвариантами БД, но проверяем их перед выпуском JWT,
	// потому что пустые claims сделают access token некорректным.
	if !user.PublicID.Valid ||
		!session.FamilyID.Valid ||
		!session.ExpiresAt.Valid {
		return nil, fmt.Errorf("%s: repository returned invalid session state", op)
	}

	issuedAccessToken, err := service.accessTokens.Issue(
		jwt.AccessTokenInput{
			UserID:    user.ID,
			PublicID:  user.PublicID.String(),
			Role:      user.Role,
			SessionID: session.FamilyID.String(),
			Scopes:    nil,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%s: issue access token: %w", op, err)
	}

	refreshExpiresAt := session.ExpiresAt.Time.UTC()
	remainingRefreshTTL := time.Until(refreshExpiresAt)

	if remainingRefreshTTL <= 0 {
		return nil, ErrRefreshTokenExpired
	}

	return &RefreshResult{
		AccessToken:           issuedAccessToken.Token,
		AccessTokenExpiresAt:  issuedAccessToken.ExpiresAt,
		RefreshToken:          newRefreshToken.Raw,
		RefreshTokenExpiresAt: refreshExpiresAt,
		RefreshTokenTTL:       remainingRefreshTTL,
	}, nil
}
