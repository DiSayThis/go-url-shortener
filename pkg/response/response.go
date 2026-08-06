package response

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type ErrorDetails struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorDetails `json:"error"`
}

func JsonResponse(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func JsonError(w http.ResponseWriter, status int, message string) {
	JsonResponse(w, ErrorResponse{Error: ErrorDetails{Code: strconv.Itoa(status), Message: message}},
		status,
	)
}
