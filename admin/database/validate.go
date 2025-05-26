package database

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/go-playground/validator/v10"
)

// Validate validates a struct based on struct tags and other custom rules registered
func Validate(v any) error {
	err := validate.Struct(v)
	if err == nil {
		return nil
	}

	// Attempt to unwrap the first field error.
	var fe validator.FieldError
	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) && len(verrs) > 0 {
		fe = verrs[0]
	}

	// If the error is not a field error, return it as is.
	if fe == nil {
		return NewValidationError(err.Error())
	}

	// Custom error messages
	if fe.Tag() == "slug" {
		return NewValidationError(fmt.Sprintf("Validation failed for field '%s': must use only letters, numbers, underscores and dashes", fe.Field()))
	}
	if fe.Tag() == "required" {
		return NewValidationError(fmt.Sprintf("Validation failed for field '%s': must be set", fe.Field()))
	}
	if fe.Tag() == "min" {
		return NewValidationError(fmt.Sprintf("Validation failed for field '%s': must be at least %s characters", fe.Field(), fe.Param()))
	}
	if fe.Tag() == "max" {
		return NewValidationError(fmt.Sprintf("Validation failed for field '%s': must be at most %s characters", fe.Field(), fe.Param()))
	}

	// Fallback to a generic error message.
	// Example: "Validation rule 'len=10' failed for field 'Name' ('InsertOptions.Name')"
	var param string
	if fe.Param() != "" {
		param = fmt.Sprintf("=%s", fe.Param())
	}
	return NewValidationError(fmt.Sprintf("Validation rule '%s%s' failed for field '%s' ('%s')", fe.Tag(), param, fe.Field(), fe.StructNamespace()))
}

// validate caches parsed validation rules
var validate *validator.Validate

// slugRegexp is used to validate identifying names (e.g. "rill-data", not "Rill Data").
var slugRegexp = regexp.MustCompile("^[_a-zA-Z0-9][-_a-zA-Z0-9]*$")

func init() {
	validate = validator.New()

	// Register "slug" validation rule
	err := validate.RegisterValidation("slug", func(fl validator.FieldLevel) bool {
		return slugRegexp.MatchString(fl.Field().String())
	})
	if err != nil {
		panic(err)
	}
}
