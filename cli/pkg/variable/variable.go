package variable

import (
	"fmt"
	"strings"
)

func Parse(variables []string) (map[string]string, error) {
	vars := make(map[string]string, len(variables))
	for _, v := range variables {
		// split into key value pairs
		key, value, found := strings.Cut(v, "=")
		// key can't be empty value can be
		if !found || key == "" {
			return nil, fmt.Errorf("invalid token %q", v)
		}
		vars[key] = value
	}
	return vars, nil
}

func Serialize(variables map[string]string) []string {
	result := make([]string, len(variables))
	i := 0
	for k, v := range variables {
		result[i] = fmt.Sprintf("%v=%v", k, v)
	}
	return result
}
