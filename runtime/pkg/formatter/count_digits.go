package formatter

import "regexp"

// countDigits counts all numeric digits in a string.
func countDigits(numStr string) int {
	re := regexp.MustCompile(`\D`)
	return len(re.ReplaceAllString(numStr, ""))
}

// countNonZeroDigits counts all non-zero numeric digits in a string.
func countNonZeroDigits(numStr string) int {
	re := regexp.MustCompile(`[^1-9]`)
	return len(re.ReplaceAllString(numStr, ""))
}
