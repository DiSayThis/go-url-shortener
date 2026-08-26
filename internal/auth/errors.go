package auth

import (
	"context"
	"errors"
	"go-api/pkg/response"
	"net/http"

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
	ErrAccessTokenConfig   = errors.New("invalid token config")
)

func isEmailCollision(err error) bool {
	var pgErr *pgconn.PgError

	return errors.As(err, &pgErr) &&
		pgErr.Code == "23505" &&
		pgErr.ConstraintName == "uq_users_email"
}

func (handler *AuthHandler) handleError(w http.ResponseWriter, req *http.Request, err error) {
	switch {
	case errors.Is(err, ErrInvalidEmail):
		response.JsonError(
			w, http.StatusBadRequest,
			"INVALID_EMAIL",
			"Email is invalid",
		)
	case errors.Is(err, ErrInvalidDisplayName):
		response.JsonError(
			w, http.StatusBadRequest,
			"INVALID_NAME",
			"Name is invalid",
		)
	case errors.Is(err, ErrWeakPassword):
		response.JsonError(
			w, http.StatusBadRequest,
			"WEAK_PASSWORD",
			"Password is too weak",
		)
	case errors.Is(err, ErrEmailAlreadyExists):
		response.JsonError(
			w, http.StatusConflict,
			"EMAIL_ALREADY_EXISTS",
			"Email is already registered",
		)
	case errors.Is(err, ErrInvalidCredentials),
		errors.Is(err, ErrPasswordMismatch):
		response.JsonError(
			w, http.StatusUnauthorized,
			"INVALID_CREDENTIALS",
			"Email or password is incorrect",
		)
	case errors.Is(err, ErrUserNotFound):
		response.JsonError(
			w, http.StatusNotFound,
			"USER_NOT_FOUND",
			"User not found",
		)
	case errors.Is(err, ErrUserInactive):
		response.JsonError(
			w, http.StatusForbidden,
			"USER_INACTIVE",
			"User account is not active",
		)
	case errors.Is(err, context.DeadlineExceeded):
		handler.Logger.WarnContext(
			req.Context(),
			"auth request deadline exceeded",
			"method", req.Method,
			"path", req.URL.Path,
		)
		response.JsonError(
			w, http.StatusGatewayTimeout,
			"REQUEST_TIMEOUT",
			"Request took too long",
		)
	case errors.Is(err, context.Canceled):
		return
	case errors.Is(err, ErrInvalidPasswordHash),
		errors.Is(err, ErrAccessTokenConfig):
		handler.writeInternalError(w, req, err)
	default:
		handler.writeInternalError(w, req, err)
	}
}

func (handler *AuthHandler) writeInternalError(
	w http.ResponseWriter,
	req *http.Request,
	err error,
) {
	handler.Logger.ErrorContext(
		req.Context(),
		"auth request failed",
		"error", err,
		"method", req.Method,
		"path", req.URL.Path,
	)
	response.JsonError(
		w, http.StatusInternalServerError,
		"INTERNAL_ERROR",
		"Internal server error",
	)
}
