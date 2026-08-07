package response

import (
	"encoding/json"
	"net/http"
)

type ErrorDetails struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorDetails `json:"error"`
}

func JsonResponse(w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func JsonError(w http.ResponseWriter, status int, code string, message string) {
	JsonResponse(w, ErrorResponse{Error: ErrorDetails{Code: code, Message: message}},
		status,
	)
}
