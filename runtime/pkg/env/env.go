package env

import (
	"fmt"
	"regexp"
)

var re = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_.]*$`)

func ValidateName(name string) error {
	if !re.MatchString(name) {
		return fmt.Errorf("invalid variable name %q: must start with a letter or underscore and contain only letters, digits, underscores and dots", name)
	}
	return nil
}

func ValidateVariables(variables map[string]string) error {
	for name := range variables {
		if err := ValidateName(name); err != nil {
			return err
		}
	}
	return nil
}
