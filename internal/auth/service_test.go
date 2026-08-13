package auth

import (
	"context"
	"errors"
	"testing"

	"go-api/internal/database"

	"github.com/jackc/pgx/v5/pgtype"
)

type authRepositoryStub struct {
	createUser      func(context.Context, CreateUserParams) (*database.User, error)
	getUserByEmail  func(context.Context, string) (*database.User, error)
	getUserByID     func(context.Context, int64) (*database.User, error)
	updateLastLogin func(context.Context, int64) error
}

func (stub authRepositoryStub) CreateUser(
	ctx context.Context,
	params CreateUserParams,
) (*database.User, error) {
	return stub.createUser(ctx, params)
}

func (stub authRepositoryStub) GetUserByEmail(
	ctx context.Context,
	email string,
) (*database.User, error) {
	return stub.getUserByEmail(ctx, email)
}

func (stub authRepositoryStub) GetUserByID(
	ctx context.Context,
	userID int64,
) (*database.User, error) {
	return stub.getUserByID(ctx, userID)
}

func (stub authRepositoryStub) UpdateLastLogin(ctx context.Context, userID int64) error {
	return stub.updateLastLogin(ctx, userID)
}

type passwordHasherStub struct {
	hash    func(string) (string, error)
	compare func(string, string) error
}

func (stub passwordHasherStub) Hash(password string) (string, error) {
	return stub.hash(password)
}

func (stub passwordHasherStub) Compare(password, encodedHash string) error {
	return stub.compare(password, encodedHash)
}

func TestServiceRegisterNormalizesInputAndHashesPassword(t *testing.T) {
	repository := authRepositoryStub{
		createUser: func(_ context.Context, params CreateUserParams) (*database.User, error) {
			if params.Email != "user@example.com" {
				t.Fatalf("email = %q, want normalized email", params.Email)
			}
			if params.DisplayName != "Alex" {
				t.Fatalf("display name = %q, want %q", params.DisplayName, "Alex")
			}
			if params.PasswordHash != "encoded-hash" {
				t.Fatalf("password hash = %q, want encoded-hash", params.PasswordHash)
			}

			return &database.User{ID: 42, Email: params.Email}, nil
		},
	}
	passwords := passwordHasherStub{
		hash: func(password string) (string, error) {
			if password != "long-password" {
				t.Fatalf("password = %q, want original password", password)
			}
			return "encoded-hash", nil
		},
	}

	service := NewService(repository, passwords)
	user, err := service.Register(context.Background(), RegisterInput{
		Email:       " User@Example.com ",
		DisplayName: " Alex ",
		Password:    "long-password",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if user.ID != 42 {
		t.Errorf("user ID = %d, want 42", user.ID)
	}
}

func TestServiceAuthenticateHidesUnknownEmail(t *testing.T) {
	repository := authRepositoryStub{
		getUserByEmail: func(context.Context, string) (*database.User, error) {
			return nil, ErrUserNotFound
		},
	}
	passwords := passwordHasherStub{
		compare: func(string, string) error {
			t.Fatal("Compare() must not be called for an unknown user")
			return nil
		},
	}

	service := NewService(repository, passwords)
	_, err := service.Authenticate(
		context.Background(),
		"missing@example.com",
		"long-password",
	)
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Authenticate() error = %v, want ErrInvalidCredentials", err)
	}
}

func TestServiceAuthenticateRejectsInactiveUser(t *testing.T) {
	repository := authRepositoryStub{
		getUserByEmail: func(context.Context, string) (*database.User, error) {
			return &database.User{
				ID:     42,
				Status: "blocked",
				PasswordHash: pgtype.Text{
					String: "encoded-hash",
					Valid:  true,
				},
			}, nil
		},
		updateLastLogin: func(context.Context, int64) error {
			t.Fatal("UpdateLastLogin() must not be called for an inactive user")
			return nil
		},
	}
	passwords := passwordHasherStub{
		compare: func(password, hash string) error {
			return nil
		},
	}

	service := NewService(repository, passwords)
	_, err := service.Authenticate(
		context.Background(),
		"user@example.com",
		"long-password",
	)
	if !errors.Is(err, ErrUserInactive) {
		t.Fatalf("Authenticate() error = %v, want ErrUserInactive", err)
	}
}
