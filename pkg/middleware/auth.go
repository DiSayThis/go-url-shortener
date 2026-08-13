package middleware

import (
	"net/http"

	"go-api/internal/auth"
)

type AuthenticatedHandlerFunc func(
	w http.ResponseWriter,
	req *http.Request,
)

func RequireAuth(next AuthenticatedHandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_, ok := auth.PrincipalFromContext(req.Context())
		if !ok {
			auth.WriteUnauthorized(w)
			return
		}
		next(w, req)
	})
}
