package middleware

import (
	"go-api/pkg/response"
	"net/http"
)

func CORS(allowedOrigins []string) func(http.Handler) http.Handler {

	allowed := make(map[string]struct{}, len(allowedOrigins))
	allowAllOrigins := false
	for _, origin := range allowedOrigins {
		if origin == "*" {
			allowAllOrigins = true
			continue
		}
		allowed[origin] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				origin := r.Header.Get("Origin")
				header := w.Header()
				if !allowAllOrigins {
					header.Add("Vary", "Origin")
				}
				if origin == "" {
					next.ServeHTTP(w, r)
					return
				}
				if allowAllOrigins {
					header.Set("Access-Control-Allow-Origin", "*")
				} else {
					_, ok := allowed[origin]
					if !ok {
						response.JsonError(
							w,
							http.StatusForbidden,
							"ORIGIN_NOT_ALLOWED",
							"origin is not allowed",
						)
						return
					}
					header.Set("Access-Control-Allow-Origin", origin)
					header.Set("Access-Control-Allow-Credentials", "true")
				}
				if r.Method == http.MethodOptions {
					header.Add("Vary", "Access-Control-Request-Method")
					header.Add("Vary", "Access-Control-Request-Headers")
					header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS")
					header.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
					header.Set("Access-Control-Max-Age", "86400")
					w.WriteHeader(http.StatusNoContent)
					return
				}
				next.ServeHTTP(w, r)
			})
	}
}
