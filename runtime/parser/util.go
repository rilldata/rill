package parser

import (
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// sanitizeYAMLMap recursively converts time.Time values in a map to YYYY-MM-DD strings.
// YAML decoders automatically parse unquoted date-like values (e.g. 2025-02-20) into time.Time,
// but connector properties expect plain strings.
func sanitizeYAMLMap(m map[string]any) {
	for k, v := range m {
		switch val := v.(type) {
		case time.Time:
			m[k] = val.Format(time.DateOnly)
		case map[string]any:
			sanitizeYAMLMap(val)
		case []any:
			sanitizeYAMLSlice(val)
		}
	}
}

func sanitizeYAMLSlice(s []any) {
	for i, v := range s {
		switch val := v.(type) {
		case time.Time:
			s[i] = val.Format(time.DateOnly)
		case map[string]any:
			sanitizeYAMLMap(val)
		case []any:
			sanitizeYAMLSlice(val)
		}
	}
}

// toDisplayName converts a snake_case name to a display name by replacing underscores and dashes with spaces and capitalizing every word.
func ToDisplayName(name string) string {
	// Don't transform names that start with an underscore (since it's probably internal).
	if name != "" && name[0] == '_' {
		return name
	}

	// Replace underscores and dashes with spaces.
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, "-", " ")

	// Replace colons with colon-space.
	name = strings.ReplaceAll(name, ":", ": ")

	// Capitalize the first letter.
	name = cases.Title(language.English).String(name)

	return name
}
