package pathutil

import "strings"

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
