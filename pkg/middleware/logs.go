package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {

			startedAt := time.Now()
			slog.Info(
				"HTTP request started",
				"method", req.Method,
				"address", req.Header.Get("Origin"),
				"path", req.URL.Path,
			)
			wrapper := &WrapperWriter{
				ResponseWriter: w,
				StatusCode:     http.StatusOK,
			}
			next.ServeHTTP(wrapper, req)
			duration := time.Since(startedAt)
			slog.Info(
				"HTTP request completed",
				"status", wrapper.StatusCode,
				"method", req.Method,
				"address", req.Header.Get("Origin"),
				"path", req.URL.Path,
				"duration", duration,
			)
		},
	)
}
