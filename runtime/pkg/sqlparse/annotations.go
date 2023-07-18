package sqlparse

import (
	"regexp"
)

var annotationsRegex = regexp.MustCompile(`(?m)^--[ \t]*@([a-zA-Z0-9_\-.]*)[ \t]*(?::[ \t]*(.*?))?\s*$`)

// ExtractAnnotations extracts annotations from comments prefixed with '@', and optionally a value after a ':'.
// Examples: "-- @materialize" and "-- @materialize: true".
func ExtractAnnotations(sql string) map[string]string {
	annotations := map[string]string{}
	subMatches := annotationsRegex.FindAllStringSubmatch(sql, -1)
	for _, subMatch := range subMatches {
		k := subMatch[1]
		v := ""
		if len(subMatch) > 2 {
			v = subMatch[2]
		}
		annotations[k] = v
	}
	return annotations
}
