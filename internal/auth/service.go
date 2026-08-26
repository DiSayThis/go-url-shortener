package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"go-api/internal/database"
	"go-api/pkg/jwt"
)

const minPasswordLength = 8

// PasswordHasher скрывает конкретный алгоритм от service.
// Следующей реализацией этого интерфейса будет Argon2idPasswordHasher.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(password, encodedHash string) error
}

type RegisterInput struct {
	Email       string
	DisplayName string
	Password    string
}

type LoginInput struct {
	Email     string
	Password  string
	UserAgent string
}

type LoginResult struct {
	User                  *database.User
	AccessToken           string
	AccessTokenExpiresAt  time.Time
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
	RefreshTokenTTL       time.Duration
}

type AuthService interface {
	Register(ctx context.Context, input RegisterInput) (*database.User, error)
	Login(ctx context.Context, input LoginInput) (*LoginResult, error)
}

type AccessTokenIssuer interface {
	Issue(input jwt.AccessTokenInput) (jwt.IssuedAccessToken, error)
}

type ServiceDeps struct {
	UserRepository    UserStore
	RefreshRepository RefreshStore
	Passwords         PasswordHasher
	AccessTokens      AccessTokenIssuer
	RefreshTTL        time.Duration
}

type Service struct {
	userRepository    UserStore
	refreshRepository RefreshStore
	passwords         PasswordHasher
	accessTokens      AccessTokenIssuer
	refreshTTL        time.Duration
}

func NewService(deps ServiceDeps) *Service {
	return &Service{
		userRepository:    deps.UserRepository,
		refreshRepository: deps.RefreshRepository,
		passwords:         deps.Passwords,
		accessTokens:      deps.AccessTokens,
		refreshTTL:        deps.RefreshTTL,
	}
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
	if displayName == "" {
		return nil, ErrInvalidDisplayName
	}
	if len(input.Password) < minPasswordLength {
		return nil, ErrWeakPassword
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
	refreshExpiresAt := time.Now().UTC().Add(service.refreshTTL)

	session, err := service.refreshRepository.CreateRefreshToken(
		ctx,
		CreateRefreshTokenParams{
			UserID:    user.ID,
			TokenHash: refreshToken.Hash,
			ExpiresAt: refreshExpiresAt,
			UserAgent: input.UserAgent,
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

func (service *Service) Authenticate(
	ctx context.Context,
	rawEmail string,
	password string,
) (*database.User, error) {
	email, err := normalizeEmail(rawEmail)
	if err != nil || password == "" {
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

	if err := service.passwords.Compare(password, user.PasswordHash.String); err != nil {
		return nil, ErrInvalidCredentials
	}

	if user.Status != "active" {
		return nil, ErrUserInactive
	}

	if err := service.userRepository.UpdateLastLogin(ctx, user.ID); err != nil {
		return nil, fmt.Errorf("record successful login: %w", err)
	}

	return user, nil
}

func (service *Service) GetUser(
	ctx context.Context,
	userID int64,
) (*database.User, error) {
	if userID <= 0 {
		return nil, ErrUserNotFound
	}

	user, err := service.userRepository.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	return user, nil
}

func normalizeEmail(value string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(value))
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return "", ErrInvalidEmail
	}

	return email, nil
}
