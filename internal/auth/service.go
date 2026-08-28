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
	Clock             Clock
}

type Service struct {
	userRepository    UserStore
	refreshRepository RefreshStore
	passwords         PasswordHasher
	accessTokens      AccessTokenIssuer
	refreshTTL        time.Duration
	now               Clock
}

type Clock func() time.Time

const (
	minPasswordLength    = 8
	maxPasswordLength    = 1024 // байты
	maxEmailLength       = 254  // байты
	maxDisplayNameLength = 100  // Unicode-символы
)

func NewService(deps ServiceDeps) *Service {
	now := deps.Clock
	if now == nil {
		now = time.Now
	}

	return &Service{
		userRepository:    deps.UserRepository,
		refreshRepository: deps.RefreshRepository,
		passwords:         deps.Passwords,
		accessTokens:      deps.AccessTokens,
		refreshTTL:        deps.RefreshTTL,
		now:               now,
	}
}

func normalizeEmail(value string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(value))
	if email == "" || len(email) > maxEmailLength {
		return "", ErrInvalidEmail
	}

	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return "", ErrInvalidEmail
	}

	return email, nil
}
