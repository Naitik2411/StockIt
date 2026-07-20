package validation

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	errorss "github.com/Naitik2411/stockit/internal/errors"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

type Validatable interface {
	Validate() error
}

type CustomValidationError struct {
	Field   string
	Message string
}

type CustomValidationErrors []CustomValidationError

func (c CustomValidationErrors) Error() string {
	return "Validation failed"
}

func BindAndValidate(c *echo.Context, payload Validatable) error {
	if err := c.Bind(payload); err != nil {
		message := strings.Split(strings.Split(err.Error(), ",")[1], "message=")[1]
		return errorss.NewBadRequestError(message, false, nil, nil, nil)
	}

	if msg, fieldErrors := validateStruct(payload); fieldErrors != nil {
		return errorss.NewBadRequestError(msg, true, nil, fieldErrors, nil)
	}

	return nil
}

func validateStruct(v Validatable) (string, []errorss.FieldError) {
	if err := v.Validate(); err != nil {
		return extractValidationErrors(err)
	}
	return "", nil
}

func extractValidationErrors(err error) (string, []errorss.FieldError) {
	var fieldErrors []errorss.FieldError

	var customValidationErrors CustomValidationErrors
	if errors.As(err, &customValidationErrors) {
		for _, cerr := range customValidationErrors {
			fieldErrors = append(fieldErrors, errorss.FieldError{
				Field: cerr.Field,
				Error: cerr.Message,
			})
		}
		return "Validation failed", fieldErrors
	}

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return err.Error(), fieldErrors
	}

	for _, verr := range validationErrors {
		field := strings.ToLower(verr.Field())
		var msg string

		switch verr.Tag() {
		case "required":
			msg = "is required"
		case "min":
			if verr.Type().Kind() == reflect.String {
				msg = fmt.Sprintf("must be at least %s characters", verr.Param())
			} else {
				msg = fmt.Sprintf("must be at least %s", verr.Param())
			}
		case "max":
			if verr.Type().Kind() == reflect.String {
				msg = fmt.Sprintf("must not exceed %s characters", verr.Param())
			} else {
				msg = fmt.Sprintf("must not exceed %s", verr.Param())
			}
		case "oneof":
			msg = fmt.Sprintf("must be one of: %s", verr.Param())
		case "email":
			msg = "must be a valid email address"
		case "e164":
			msg = "must be a valid phone number with country code"
		case "uuid":
			msg = "must be a valid UUID"
		case "uuidList":
			msg = "must be a comma-separated list of valid UUIDs"
		case "dive":
			msg = "some items are invalid"
		default:
			if verr.Param() != "" {
				msg = fmt.Sprintf("%s: %s:%s", field, verr.Tag(), verr.Param())
			} else {
				msg = fmt.Sprintf("%s: %s", field, verr.Tag())
			}
		}

		fieldErrors = append(fieldErrors, errorss.FieldError{
			Field: strings.ToLower(verr.Field()),
			Error: msg,
		})
	}

	return "Validation failed", fieldErrors
}

var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func IsValidUUID(uuid string) bool {
	return uuidRegex.MatchString(uuid)
}
