package env

import (
	"fmt"
	"regexp"
)

var re = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_.]*$`)

func ValidateName(name string) error {
	if !re.MatchString(name) {
		return fmt.Errorf("invalid variable name: %q. Must start with a letter or underscore, can contain letters, digits, underscores and dots", name)
	}
	return nil
}
