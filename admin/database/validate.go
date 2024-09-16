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

	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) && len(verrs) > 0 {
		// Example: "Validation rule 'len=10' failed for field 'Name' ('InsertOptions.Name')"
		verr := verrs[0]
		var param string
		if verr.Param() != "" {
			param = fmt.Sprintf("=%s", verr.Param())
		}
		return NewValidationError(fmt.Sprintf("Validation rule '%s%s' failed for field '%s' ('%s')", verr.Tag(), param, verr.Field(), verr.StructNamespace()))
	}

	return NewValidationError(err.Error())
}

// validate caches parsed validation rules
var validate *validator.Validate

// slugRegexp is used to validate identifying names (e.g. "rill-data", not "Rill Data").
var slugRegexp = regexp.MustCompile("^[_a-zA-Z0-9][-_a-zA-Z0-9]{2,39}$")

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
