package rillv1

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// toDisplayName converts a snake_case name to a display name by replacing underscores and dashes with spaces and capitalizing the first letter.
func ToDisplayName(name string) string {
	// Don't transform names that start with an underscore (since it's probably internal).
	if name != "" && name[0] == '_' {
		return name
	}

	// Replace underscores and dashes with spaces.
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, "-", " ")

	// Capitalize the first letter.
	name = cases.Title(language.English).String(name)

	return name
}
