package pathutil

import (
	"fmt"
	"strings"
)

func CommonPrefix(a, b string) string {
	if a == "" || b == "" {
		return ""
	}

	i := 0
	for {
		if i < len(a) && i < len(b) && a[i] == b[i] {
			i++
			continue
		}

		if i == len(a) && i == len(b) { // nolint:gocritic // More readable
			return a
		}

		if i == len(a) && b[i] == '/' {
			return a
		}

		if i == len(b) && a[i] == '/' {
			return b
		}

		// a and b differ at i. Find the last '/' before i.
		idx := strings.LastIndex(a[:i], "/")
		if idx == -1 {
			return ""
		}
		return a[:idx]
	}
}

// GetPath retrieves a value from a nested map using a dot-separated path.
// It supports accessing array elements using numeric indices in the path.
// For example, given the map:
//
//	{
//	  "user": {
//	    "name": "Alice",
//	    "addresses": [
//	      {"city": "New York"},
//	      {"city": "Los Angeles"}
//	    ]
//	  }
//	}
//
// The path "user.name" would return "Alice", and the path "user.addresses.0.city" would return "New York".
func GetPath(m map[string]any, path string) (any, bool) {
	parts := strings.Split(path, ".")
	var cur any = m

	for _, p := range parts {
		switch node := cur.(type) {
		case map[string]any:
			val, ok := node[p]
			if !ok {
				return nil, false
			}
			cur = val
		case []any:
			// allow array indices like "items.0.name"
			// parse p as index
			var idx int
			_, err := fmt.Sscanf(p, "%d", &idx)
			if err != nil || idx < 0 || idx >= len(node) {
				return nil, false
			}
			cur = node[idx]
		default:
			return nil, false
		}
	}
	return cur, true
}
