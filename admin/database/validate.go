package database

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// Validate validates a struct based on struct tags and other custom rules registered
func Validate(v any) error {
	return validate.Struct(v)
}

// validate caches parsed validation rules
var validate *validator.Validate

// slugRegexp is used to validate identifying names (e.g. "rill-data", not "Rill Data").
var slugRegexp = regexp.MustCompile("^[_a-zA-Z0-9][-_a-zA-Z0-9]{2,39}$")

// Regular expression to match a domain name
var domainRegex = regexp.MustCompile(`^[a-zA-Z0-9]+([-.][a-zA-Z0-9]+)*\.[a-zA-Z]{2,}$`)

func init() {
	validate = validator.New()

	// Register "slug" validation rule
	err := validate.RegisterValidation("slug", func(fl validator.FieldLevel) bool {
		return slugRegexp.MatchString(fl.Field().String())
	})
	if err != nil {
		panic(err)
	}

	// Register "domain" validation rule
	err = validate.RegisterValidation("domain", func(fl validator.FieldLevel) bool {
		return domainRegex.MatchString(fl.Field().String())
	})
	if err != nil {
		panic(err)
	}
}
