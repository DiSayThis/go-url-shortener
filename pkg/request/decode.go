package request

import (
	"encoding/json"
	"errors"
	"io"
)

func Decode[T any](body io.ReadCloser) (T, error) {
	var payload T

	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&payload); err != nil {
		return payload, err
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return payload, errors.New(
				"request body must contain one JSON value",
			)
		}

		return payload, err
	}

	return payload, nil
}
