package formatter

import "regexp"

var (
	digits        = regexp.MustCompile(`\D`)
	nonZeroDigits = regexp.MustCompile(`[^1-9]`)
)

// countDigits counts all numeric digits in a string.
func countDigits(numStr string) int {
	return len(digits.ReplaceAllString(numStr, ""))
}

// countNonZeroDigits counts all non-zero numeric digits in a string.
func countNonZeroDigits(numStr string) int {
	return len(nonZeroDigits.ReplaceAllString(numStr, ""))
}
