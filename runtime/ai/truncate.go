package ai

import "unicode/utf8"

// truncateUTF8 returns a prefix of s at most maxLen bytes long, never splitting a multi-byte UTF-8
// rune: if the byte cut lands inside a rune, it backs off to the previous rune boundary. Callers
// append their own ellipsis. Byte-indexed slicing (s[:maxLen]) is not safe for this: a dangling lead
// byte followed by an appended "..." is invalid UTF-8, which protobuf refuses to marshal in a string field.
func truncateUTF8(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	cut := maxLen
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut]
}
