// SPDX-License-Identifier: AGPL-3.0-or-later

package httputils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

func DecodeJSON[T any](r *http.Request) (*T, error) {
	var v T

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&v); err != nil {
		return nil, errors.New("invalid request body")
	}

	if err := validator.New(validator.WithRequiredStructEnabled()).Struct(&v); err != nil {
		return nil, formatValidationErr(err)
	}

	return &v, nil
}

type TrimmedString string

func (s *TrimmedString) UnmarshalJSON(b []byte) error {
	var raw string
	if err := json.Unmarshal(b, &raw); err != nil {
		//nolint:wrapcheck
		return err
	}

	*s = TrimmedString(strings.TrimSpace(raw))

	return nil
}

func (s *TrimmedString) String() string {
	return string(*s)
}

func formatValidationErr(err error) error {
	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) {
		return errors.New("invalid request body")
	}

	msgs := make([]string, 0, len(verrs))
	for _, fe := range verrs {
		msgs = append(msgs, fieldErrMsg(fe))
	}

	return errors.New(strings.Join(msgs, ", "))
}

func fieldErrMsg(fe validator.FieldError) string {
	field := fe.Field()

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		if fe.Kind() == reflect.String {
			return fmt.Sprintf("%s must be at least %s characters", field, fe.Param())
		}

		return fmt.Sprintf("%s must be at least %s", field, fe.Param())
	case "max":
		if fe.Kind() == reflect.String {
			return fmt.Sprintf("%s must be at most %s characters", field, fe.Param())
		}

		return fmt.Sprintf("%s must be at most %s", field, fe.Param())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, strings.ReplaceAll(fe.Param(), " ", ", "))
	case "eqfield":
		return fmt.Sprintf("%s must match %s", field, fe.Param())
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}
