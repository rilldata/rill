// Package urlutils provides helpers for producing URL-safe strings.
package urlutils

import "strings"

// Slugify converts a label into a lowercase, URL-safe identifier. Runs of
// non-alphanumeric characters collapse into a single dash, and leading and
// trailing dashes are trimmed.
func Slugify(s string) string {
	var b strings.Builder
	prevDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash && b.Len() > 0 {
				b.WriteRune('-')
				prevDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}
