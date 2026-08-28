package auth

import (
	"context"
	"errors"
	"fmt"
	"go-api/internal/database"
	"go-api/pkg/jwt"
	"strings"
	"unicode/utf8"
)

func (service *Service) Login(
	ctx context.Context,
	input LoginInput,
) (*LoginResult, error) {
	user, err := service.Authenticate(ctx, input.Email, input.Password)
	if err != nil {
		return nil, err
	}
	if !user.PublicID.Valid {
		return nil, fmt.Errorf("authenticated user has invalid public ID")
	}

	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	refreshExpiresAt := service.now().UTC().Add(service.refreshTTL)

	session, err := service.refreshRepository.CreateRefreshToken(
		ctx,
		CreateRefreshTokenParams{
			UserID:    user.ID,
			TokenHash: refreshToken.Hash,
			ExpiresAt: refreshExpiresAt,
			UserAgent: input.UserAgent,
			CreatedIP: input.CreatedIP,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create refresh session: %w", err)
	}

	issuedToken, err := service.accessTokens.Issue(jwt.AccessTokenInput{
		UserID:    user.ID,
		PublicID:  user.PublicID.String(),
		Role:      user.Role,
		SessionID: session.FamilyID.String(),
		Scopes:    nil,
	})
	if err != nil {
		return nil, fmt.Errorf("issue access token: %w", err)
	}

	return &LoginResult{
		User:                  user,
		AccessToken:           issuedToken.Token,
		AccessTokenExpiresAt:  issuedToken.ExpiresAt,
		RefreshToken:          refreshToken.Raw,
		RefreshTokenExpiresAt: refreshExpiresAt,
		RefreshTokenTTL:       service.refreshTTL,
	}, nil
}

func (service *Service) Register(
	ctx context.Context,
	input RegisterInput,
) (*database.User, error) {
	email, err := normalizeEmail(input.Email)
	if err != nil {
		return nil, err
	}

	displayName := strings.TrimSpace(input.DisplayName)
	if displayName == "" ||
		utf8.RuneCountInString(displayName) > maxDisplayNameLength {
		return nil, ErrInvalidDisplayName
	}
	if len(input.Password) < minPasswordLength {
		return nil, ErrWeakPassword
	}

	if len(input.Password) > maxPasswordLength {
		return nil, ErrPasswordTooLong
	}

	passwordHash, err := service.passwords.Hash(input.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := service.userRepository.CreateUser(ctx, CreateUserParams{
		Email:        email,
		DisplayName:  displayName,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return nil, fmt.Errorf("register user: %w", err)
	}

	return user, nil
}

func (service *Service) Authenticate(
	ctx context.Context,
	rawEmail string,
	password string,
) (*database.User, error) {
	email, err := normalizeEmail(rawEmail)
	if err != nil || password == "" || len(password) > maxPasswordLength {
		return nil, ErrInvalidCredentials
	}

	user, err := service.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}

		return nil, fmt.Errorf("authenticate user: %w", err)
	}

	// OAuth-only пользователь может пока не иметь локального пароля.
	// Для клиента это также выглядит как обычные неверные credentials.
	if !user.PasswordHash.Valid {
		return nil, ErrInvalidCredentials
	}

	if err := service.passwords.Compare(
		password,
		user.PasswordHash.String,
	); err != nil {
		if errors.Is(err, ErrPasswordMismatch) {
			return nil, ErrInvalidCredentials
		}

		return nil, fmt.Errorf("verify password hash: %w", err)
	}

	if user.Status != "active" {
		return nil, ErrUserInactive
	}

	if err := service.userRepository.UpdateLastLogin(ctx, user.ID); err != nil {
		return nil, fmt.Errorf("record successful login: %w", err)
	}

	return user, nil
}
