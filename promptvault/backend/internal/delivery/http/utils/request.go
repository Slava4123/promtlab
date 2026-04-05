package utils

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

func DecodeJSON(r *http.Request, dst any) error {
	defer func() { _ = r.Body.Close() }()
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return fmt.Errorf("invalid request body: %w", err)
	}
	return nil
}

// DecodeAndValidate decodes the request body as JSON into T and validates it.
func DecodeAndValidate[T any](r *http.Request, v *validator.Validate) (T, error) {
	var req T
	if err := DecodeJSON(r, &req); err != nil {
		return req, err
	}
	if err := v.Struct(req); err != nil {
		return req, err
	}
	return req, nil
}

