package request

import (
	"go-api/pkg/response"
	"net/http"
)

func HandleBody[T any](w http.ResponseWriter, req *http.Request) (*T, error) {

	body, err := Decode[T](req.Body)
	if err != nil {
		response.JsonError(
			w,
			http.StatusBadRequest,
			"INVALID_REQUEST_BODY",
			"Request body is invalid",
		)
		return nil, err
	}

	err = IsValid(body)
	if err != nil {
		response.JsonError(
			w,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"Request validation failed",
		)
		return nil, err
	}
	return &body, nil
}
