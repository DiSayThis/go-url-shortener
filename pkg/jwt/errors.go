package jwt

import (
	"errors"
	"go-api/pkg/response"
	"net/http"
)

var (
	ErrAccessTokenConfig   = errors.New("invalid token config")
	ErrInvalidTokenSubject = errors.New("invalid token subject")
	ErrInvalidAccessToken  = errors.New("invalid access token")
	ErrAccessTokenExpired  = errors.New("access token expired")
)

func WriteUnauthorized(w http.ResponseWriter) {
	w.Header().Set(
		"WWW-Authenticate",
		`Bearer realm="api"`,
	)

	response.JsonError(
		w,
		http.StatusUnauthorized,
		"UNAUTHORIZED",
		"Authentication is required",
	)
}
