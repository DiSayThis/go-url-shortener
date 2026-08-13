package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"go-api/internal/database"
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

type AuthService interface {
	Register(ctx context.Context, input RegisterInput) (*database.User, error)
	Authenticate(ctx context.Context, email, password string) (*database.User, error)
	GetUser(ctx context.Context, userID int64) (*database.User, error)
}

type Service struct {
	repository AuthRepository
	passwords  PasswordHasher
}

func NewService(repository AuthRepository, passwords PasswordHasher) *Service {
	return &Service{
		repository: repository,
		passwords:  passwords,
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

	user, err := service.repository.CreateUser(ctx, CreateUserParams{
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
	if err != nil || password == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := service.repository.GetUserByEmail(ctx, email)
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

	if err := service.repository.UpdateLastLogin(ctx, user.ID); err != nil {
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

	user, err := service.repository.GetUserByID(ctx, userID)
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
