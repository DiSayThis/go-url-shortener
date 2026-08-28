package auth

import (
	"context"
	"errors"
	"go-api/pkg/response"
	"net/http"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	// Регистрация.
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidDisplayName = errors.New("invalid display name")
	ErrWeakPassword       = errors.New("weak password")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrPasswordTooLong    = errors.New("password is too long")

	// Аутентификация по email/password.
	// Одинаковая ошибка скрывает существование пользователя.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// Пользователь.
	ErrUserNotFound = errors.New("user not found")
	ErrUserInactive = errors.New("user is not active")

	// Password hashing.
	ErrPasswordMismatch    = errors.New("password does not match")
	ErrInvalidPasswordHash = errors.New("invalid password hash")

	// Refresh token/session.
	ErrInvalidRefreshToken   = errors.New("invalid refresh token")
	ErrRefreshTokenExpired   = errors.New("refresh token expired")
	ErrRefreshTokenRevoked   = errors.New("refresh token revoked")
	ErrRefreshTokenReused    = errors.New("refresh token reuse detected")
	ErrInvalidRefreshSession = errors.New("invalid refresh session")
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
	case errors.Is(err, ErrInvalidPasswordHash):
		handler.writeInternalError(w, req, err)
	case errors.Is(err, ErrPasswordTooLong):
		response.JsonError(
			w,
			http.StatusBadRequest,
			"PASSWORD_TOO_LONG",
			"Password is too long",
		)
	case errors.Is(err, ErrInvalidRefreshToken),
		errors.Is(err, ErrRefreshTokenExpired),
		errors.Is(err, ErrRefreshTokenRevoked),
		errors.Is(err, ErrRefreshTokenReused),
		errors.Is(err, ErrInvalidRefreshSession):
		handler.clearRefreshCookie(w)
		if errors.Is(err, ErrRefreshTokenReused) {
			handler.Logger.WarnContext(
				req.Context(),
				"refresh token reuse detected",
				"method", req.Method,
				"path", req.URL.Path,
			)
		}
		response.JsonError(
			w,
			http.StatusUnauthorized,
			"INVALID_SESSION",
			"Refresh session is invalid",
		)
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
