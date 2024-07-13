package rest

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"strings"
)

type httpError struct {
	Status     int               `json:"-"`
	Message    string            `json:"message"`
	Details    string            `json:"details,omitempty"`
	Validation map[string]string `json:"validator,omitempty"`
}

func (e *httpError) Error() string {
	var (
		msg  strings.Builder
		args = make([]any, 0, 4)
	)

	msg.WriteString("message: %s")
	args = append(args, e.Message)

	if e.Details != "" {
		args = append(args, e.Details)
		msg.WriteString(", details: %s")
	}

	if e.Validation != nil {
		args = append(args, fmt.Sprintf("%s", e.Validation))

		msg.WriteString(", validation: %s")
	}

	return fmt.Sprintf(msg.String(), args...)
}

func (e *httpError) withDetails(err error) *httpError {
	newErr := *e
	newErr.Details = err.Error()

	return &newErr
}

func (e *httpError) withValidator(err error) *httpError {
	newErr := *e

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		newErr.Validation = make(map[string]string, len(validationErrors))

		for _, ve := range validationErrors {
			if ve.Param() == "" {
				newErr.Validation[ve.Field()] = ve.Tag()
			} else {
				newErr.Validation[ve.Field()] = fmt.Sprintf("%s=%s", ve.Tag(), ve.Param())
			}
		}

		return &newErr
	}

	newErr.Details = err.Error()

	return &newErr
}

func newHTTPError(status int, msg string) *httpError {
	return &httpError{
		Status:  status,
		Message: msg,
	}
}
