package request

import (
	"go-api/pkg/response"
	"net/http"
)

func HandleBody[T any](w http.ResponseWriter, req *http.Request) (*T, error) {

	body, err := Decode[T](req.Body)
	if err != nil {
		response.JsonError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return nil, err
	}

	err = IsValid(body)
	if err != nil {
		response.JsonError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return nil, err
	}
	return &body, nil
}
