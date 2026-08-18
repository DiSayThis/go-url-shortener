package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"go-api/configs"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AccessTokenService interface {
	Issue(input AccessTokenInput) (IssuedAccessToken, error)
	Parse(rawToken string) (Principal, error)
}

type AccessTokenInput struct {
	UserID    int64
	PublicID  string
	Role      string
	SessionID string
	Scopes    []string
}

type IssuedAccessToken struct {
	Token     string
	ExpiresAt time.Time
}

type JWTAccessTokenService struct {
	config configs.AuthConfig
	now    func() time.Time
}

func NewJWTAccessTokenService(
	config configs.AuthConfig,
) (*JWTAccessTokenService, error) {
	if len(config.Secret) < 32 ||
		config.Issuer == "" ||
		config.Audience == "" ||
		config.TTL <= 0 {
		return nil, ErrAccessTokenConfig
	}
	return &JWTAccessTokenService{
		config: config,
		now:    time.Now,
	}, nil
}

type AccessTokenClaims struct {
	UserID    int64    `json:"uid"`
	Role      string   `json:"role"`
	SessionID string   `json:"sid"`
	Scopes    []string `json:"scopes,omitempty"`
	TokenUse  string   `json:"token_use"`

	jwt.RegisteredClaims
}

func generateTokenID() (string, error) {
	randomBytes := make([]byte, 16)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate token ID: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(randomBytes), nil
}

func (service *JWTAccessTokenService) Issue(
	input AccessTokenInput,
) (IssuedAccessToken, error) {
	if input.UserID <= 0 ||
		input.PublicID == "" ||
		input.SessionID == "" {
		return IssuedAccessToken{}, ErrInvalidTokenSubject
	}
	now := service.now().UTC()
	expiresAt := now.Add(service.config.TTL)
	tokenID, err := generateTokenID()
	if err != nil {
		return IssuedAccessToken{}, err
	}
	claims := AccessTokenClaims{
		UserID:    input.UserID,
		Role:      input.Role,
		SessionID: input.SessionID,
		Scopes:    input.Scopes,
		TokenUse:  "access",

		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    service.config.Issuer,
			Subject:   input.PublicID,
			Audience:  jwt.ClaimStrings{service.config.Audience},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        tokenID,
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	token.Header["typ"] = "at+jwt"
	signedToken, err := token.SignedString(
		service.config.Secret,
	)
	if err != nil {
		return IssuedAccessToken{}, fmt.Errorf(
			"sign access token: %w",
			err,
		)
	}
	return IssuedAccessToken{
		Token:     signedToken,
		ExpiresAt: expiresAt,
	}, nil
}

func (service *JWTAccessTokenService) Parse(
	rawToken string,
) (Principal, error) {
	// Parse получает только JWT-строку без префикса "Bearer ".
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return Principal{}, ErrInvalidAccessToken
	}

	// Передаём указатель, чтобы библиотека могла записать в эту структуру
	// claims, декодированные из payload входящего JWT.
	claims := &AccessTokenClaims{}

	parsedToken, err := jwt.ParseWithClaims(
		rawToken,
		claims,
		// Keyfunc вызывается библиотекой после чтения header.
		// Он проверяет профиль токена и возвращает ключ для проверки подписи.
		func(token *jwt.Token) (any, error) {
			// Не доверяем alg, который прислал клиент: принимаем только HS256.
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, ErrInvalidAccessToken
			}

			// typ="at+jwt" отделяет access token от других видов JWT.
			tokenType, ok := token.Header["typ"].(string)
			if !ok || tokenType != "at+jwt" {
				return nil, ErrInvalidAccessToken
			}

			return service.config.Secret, nil
		},
		// Разрешаем parser использовать только HS256.
		jwt.WithValidMethods([]string{
			jwt.SigningMethodHS256.Alg(),
		}),
		// Проверяем, что token выпущен именно нашим приложением.
		jwt.WithIssuer(service.config.Issuer),
		// Проверяем, что token предназначен именно для нашего API.
		jwt.WithAudience(service.config.Audience),
		// Access token без exp принимать нельзя.
		jwt.WithExpirationRequired(),
		// Включаем проверку времени выпуска iat.
		jwt.WithIssuedAt(),
		// Допускаем небольшое расхождение часов между серверами.
		jwt.WithLeeway(service.config.Leeway),
		// Используем внедрённые часы, чтобы expiration можно было тестировать.
		jwt.WithTimeFunc(service.now),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return Principal{}, ErrAccessTokenExpired
		}

		return Principal{}, ErrInvalidAccessToken
	}

	// При err == nil библиотека обычно уже выставляет Valid=true,
	// но проверяем это явно как security-инвариант.
	if !parsedToken.Valid {
		return Principal{}, ErrInvalidAccessToken
	}

	// Библиотека проверяет стандартные claims, но не знает бизнес-смысл
	// наших uid, sid, role и token_use — их проверяем самостоятельно.
	if claims.TokenUse != "access" ||
		claims.UserID <= 0 ||
		strings.TrimSpace(claims.Subject) == "" ||
		strings.TrimSpace(claims.SessionID) == "" ||
		strings.TrimSpace(claims.Role) == "" ||
		strings.TrimSpace(claims.ID) == "" ||
		claims.IssuedAt == nil {
		return Principal{}, ErrInvalidAccessToken
	}

	// JWT-специфичную структуру превращаем в общий объект безопасности,
	// с которым будут работать middleware и handlers.
	return Principal{
		UserID:    claims.UserID,
		PublicID:  claims.Subject,
		Role:      claims.Role,
		Scopes:    claims.Scopes,
		SessionID: claims.SessionID,
	}, nil
}
