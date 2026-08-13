package auth

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	// Ошибки регистрации.
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidDisplayName = errors.New("invalid display name")
	ErrWeakPassword       = errors.New("weak password")
	ErrEmailAlreadyExists = errors.New("email already exists")

	// Одинаковая ошибка для неизвестного email и неправильного пароля.
	// Это не позволяет клиенту выяснять, зарегистрирован ли конкретный email.
	ErrInvalidCredentials = errors.New("invalid credentials")

	ErrUserNotFound = errors.New("user not found")
	ErrUserInactive = errors.New("user is not active")

	ErrPasswordMismatch    = errors.New("password does not match")
	ErrInvalidPasswordHash = errors.New("invalid password hash")
)

func isEmailCollision(err error) bool {
	var pgErr *pgconn.PgError

	return errors.As(err, &pgErr) &&
		pgErr.Code == "23505" &&
		pgErr.ConstraintName == "uq_users_email"
}
