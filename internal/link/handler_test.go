package link

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerHandleError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{
			name:       "invalid URL",
			err:        ErrInvalidURL,
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_URL",
		},
		{
			name:       "link not found through wrapped error",
			err:        errors.Join(errors.New("service error"), ErrLinkNotFound),
			wantStatus: http.StatusNotFound,
			wantCode:   "LINK_NOT_FOUND",
		},
		{
			name:       "deadline exceeded",
			err:        context.DeadlineExceeded,
			wantStatus: http.StatusGatewayTimeout,
			wantCode:   "REQUEST_TIMEOUT",
		},
		{
			name:       "unexpected internal error",
			err:        errors.New("database password must not reach the client"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   "INTERNAL_ERROR",
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := &Handler{logger: logger}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/test", nil)

			handler.handleError(recorder, request, tt.err)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}

			var body struct {
				Error struct {
					Code string `json:"code"`
				} `json:"error"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if body.Error.Code != tt.wantCode {
				t.Errorf("error code = %q, want %q", body.Error.Code, tt.wantCode)
			}
			if strings.Contains(recorder.Body.String(), "database password") {
				t.Error("response exposed an internal error")
			}
		})
	}
}
