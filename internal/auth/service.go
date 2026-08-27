package auth

import (
	"context"
	"net/mail"
	"net/netip"
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
	CreatedIP *netip.Addr
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
	Refresh(ctx context.Context, input RefreshInput) (*RefreshResult, error)
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

func normalizeEmail(value string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(value))
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return "", ErrInvalidEmail
	}

	return email, nil
}
