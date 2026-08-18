package auth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthHandlerHandleError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{name: "invalid email", err: ErrInvalidEmail, wantStatus: http.StatusBadRequest, wantCode: "INVALID_EMAIL"},
		{name: "invalid name", err: ErrInvalidDisplayName, wantStatus: http.StatusBadRequest, wantCode: "INVALID_NAME"},
		{name: "weak password", err: ErrWeakPassword, wantStatus: http.StatusBadRequest, wantCode: "WEAK_PASSWORD"},
		{name: "email exists", err: ErrEmailAlreadyExists, wantStatus: http.StatusConflict, wantCode: "EMAIL_ALREADY_EXISTS"},
		{name: "invalid credentials", err: ErrInvalidCredentials, wantStatus: http.StatusUnauthorized, wantCode: "INVALID_CREDENTIALS"},
		{name: "password mismatch", err: ErrPasswordMismatch, wantStatus: http.StatusUnauthorized, wantCode: "INVALID_CREDENTIALS"},
		{name: "user not found", err: ErrUserNotFound, wantStatus: http.StatusNotFound, wantCode: "USER_NOT_FOUND"},
		{name: "inactive user", err: ErrUserInactive, wantStatus: http.StatusForbidden, wantCode: "USER_INACTIVE"},
		{name: "deadline", err: context.DeadlineExceeded, wantStatus: http.StatusGatewayTimeout, wantCode: "REQUEST_TIMEOUT"},
		{name: "invalid password hash", err: ErrInvalidPasswordHash, wantStatus: http.StatusInternalServerError, wantCode: "INTERNAL_ERROR"},
		{name: "token config", err: ErrAccessTokenConfig, wantStatus: http.StatusInternalServerError, wantCode: "INTERNAL_ERROR"},
		{name: "wrapped internal error", err: errors.New("database unavailable"), wantStatus: http.StatusInternalServerError, wantCode: "INTERNAL_ERROR"},
	}

	handler := &AuthHandler{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/auth/login", nil)

			handler.handleError(recorder, request, tt.err)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}

			var body struct {
				Error struct {
					Code string `json:"code"`
				} `json:"error"`
			}
			if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if body.Error.Code != tt.wantCode {
				t.Errorf("error code = %q, want %q", body.Error.Code, tt.wantCode)
			}
		})
	}
}

func TestAuthHandlerHandleErrorIgnoresCanceledRequest(t *testing.T) {
	handler := &AuthHandler{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/auth/login", nil)

	handler.handleError(recorder, request, context.Canceled)

	if recorder.Code != http.StatusOK || recorder.Body.Len() != 0 {
		t.Fatalf("canceled request wrote status %d and body %q", recorder.Code, recorder.Body.String())
	}
}
