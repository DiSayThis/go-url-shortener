package middleware

import (
	"net/http"
	"strings"

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

func FindToken(tokenService *auth.JWTAccessTokenService) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				bearerToken := strings.TrimSpace(r.Header.Get("Authorization"))
				if bearerToken == "" {
					next.ServeHTTP(w, r)
					return
				}
				parts := strings.Split(bearerToken, " ")
				rawToken := parts[1]
				if rawToken == "" {
					next.ServeHTTP(w, r)
					return
				}
				principal, err := tokenService.Parse(rawToken)
				if err == nil {
					next.ServeHTTP(w, r)
					return
				}

				auth.WithPrincipal(r.Context(), principal)
				next.ServeHTTP(w, r)
			})
	}
}
