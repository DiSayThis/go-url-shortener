package auth

import (
	"go-api/pkg/response"
	"net/http"
)

func RequirePrincipal(w http.ResponseWriter, req *http.Request) (Principal, bool) {
	principal, ok := PrincipalFromContext(req.Context())
	if !ok {
		WriteUnauthorized(w)
		return Principal{}, false
	}

	return principal, true
}

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
