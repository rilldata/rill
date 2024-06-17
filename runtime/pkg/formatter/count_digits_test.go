package formatter

import (
	"testing"
)

func TestCountDigits(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"All digits", "1234567890", 10},
		{"With non-digits", "abc123xyz", 3},
		{"Empty string", "", 0},
		{"All non-digits", "abcdef", 0},
		{"Digits and spaces", "12 34 56", 6},
		{"Leading zeros", "000123", 6},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := countDigits(tc.input); got != tc.expected {
				t.Errorf("countDigits(%q) = %d; want %d", tc.input, got, tc.expected)
			}
		})
	}
}

func TestCountNonZeroDigits(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"Normal case", "1234567890", 9},
		{"Include zeros", "1020304050", 5},
		{"Empty string", "", 0},
		{"All zeros", "0000000", 0},
		{"Mixed characters", "abc1d2e3f0", 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := countNonZeroDigits(tc.input); got != tc.expected {
				t.Errorf("countNonZeroDigits(%q) = %d; want %d", tc.input, got, tc.expected)
			}
		})
	}
}
