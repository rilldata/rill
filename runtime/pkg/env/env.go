package env

import (
	"fmt"
	"regexp"
	"strings"
)

var re = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_.]*$`)

func ValidateName(name string) error {
	if !re.MatchString(name) {
		return fmt.Errorf("invalid variable name %q: must start with a letter or underscore and contain only letters, digits, underscores and dots", name)
	}
	return nil
}

func ParseKeyVal(s string) (string, string, error) {
	parts := strings.Split(s, "=")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid key-value pair %q: must be in the form key=value", s)
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
}

func ParseAndValidate(s string) (string, string, error) {
	key, value, err := ParseKeyVal(s)
	if err != nil {
		return "", "", err
	}
	if err := ValidateName(key); err != nil {
		return "", "", err
	}
	return key, value, nil
}
