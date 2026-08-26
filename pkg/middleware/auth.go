package middleware

import (
	"go-api/pkg/jwt"
	"net/http"
	"strings"
)

type AuthenticatedHandlerFunc func(
	w http.ResponseWriter,
	req *http.Request,
)

func RequireAuth(next AuthenticatedHandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_, ok := jwt.PrincipalFromContext(req.Context())
		if !ok {
			jwt.WriteUnauthorized(w)
			return
		}
		next(w, req)
	})
}

func FindToken(tokenService *jwt.JWTAccessTokenService) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				bearerToken := strings.TrimSpace(r.Header.Get("Authorization"))
				if bearerToken == "" {
					next.ServeHTTP(w, r)
					return
				}

				parts := strings.Fields(bearerToken)
				if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
					next.ServeHTTP(w, r)
					return
				}

				principal, err := tokenService.Parse(parts[1])
				if err != nil {
					next.ServeHTTP(w, r)
					return
				}
				r = r.WithContext(jwt.WithPrincipal(r.Context(), principal))
				next.ServeHTTP(w, r)
			})
	}
}
