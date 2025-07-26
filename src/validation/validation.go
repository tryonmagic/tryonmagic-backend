package validation

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type SortOrder string

const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
)

type BasePaginationQueryParams struct {
	// Note: Fiber's QueryParser does not support struct tag defaults.
	// You must set defaults in code, not via struct tags.
	Page         int       `query:"page" validate:"omitempty,number" default:"1"`
	PageSize     int       `query:"pageSize" validate:"omitempty,number,max=200" default:"10"`
	Search       *string   `query:"search" validate:"omitempty"`
	SortBy       *string   `query:"sortBy" validate:"omitempty"`
	SortOrder    SortOrder `query:"sortOrder" validate:"omitempty,oneof=asc desc" default:"desc"`
	SearchFields *[]string `query:"searchFields" validate:"omitempty"`
}

var customMessages = map[string]string{
	"required": "Field %s must be filled",
	"email":    "Invalid email address for field %s",
	"min":      "Field %s must have a minimum length of %s characters",
	"max":      "Field %s must have a maximum length of %s characters",
	"len":      "Field %s must be exactly %s characters long",
	"number":   "Field %s must be a number",
	"positive": "Field %s must be a positive number",
	"alphanum": "Field %s must contain only alphanumeric characters",
	"oneof":    "Invalid value for field %s",
	"password": "Field %s must contain at least 1 letter and 1 number",
}

func CustomErrorMessages(err error) map[string]string {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		return generateErrorMessages(validationErrors)
	}
	return nil
}

func generateErrorMessages(validationErrors validator.ValidationErrors) map[string]string {
	errorsMap := make(map[string]string)
	for _, err := range validationErrors {
		fieldName := err.StructNamespace()
		tag := err.Tag()

		customMessage := customMessages[tag]
		if customMessage != "" {
			errorsMap[fieldName] = formatErrorMessage(customMessage, err, tag)
		} else {
			errorsMap[fieldName] = defaultErrorMessage(err)
		}
	}
	return errorsMap
}

func formatErrorMessage(customMessage string, err validator.FieldError, tag string) string {
	if tag == "min" || tag == "max" || tag == "len" {
		return fmt.Sprintf(customMessage, err.Field(), err.Param())
	}
	return fmt.Sprintf(customMessage, err.Field())
}

func defaultErrorMessage(err validator.FieldError) string {
	return fmt.Sprintf("Field validation for '%s' failed on the '%s' tag", err.Field(), err.Tag())
}

func Validator() *validator.Validate {
	validate := validator.New()

	if err := validate.RegisterValidation("password", Password); err != nil {
		return nil
	}

	return validate
}
