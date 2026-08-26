package auth

import (
	"go-api/pkg/jwt"
	"net/http"
)

func RequirePrincipal(w http.ResponseWriter, req *http.Request) (jwt.Principal, bool) {
	principal, ok := jwt.PrincipalFromContext(req.Context())
	if !ok {
		jwt.WriteUnauthorized(w)
		return jwt.Principal{}, false
	}

	return principal, true
}
