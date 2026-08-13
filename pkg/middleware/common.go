package middleware

import "net/http"

type WrapperWriter struct {
	http.ResponseWriter
	StatusCode int
}

func (w *WrapperWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
