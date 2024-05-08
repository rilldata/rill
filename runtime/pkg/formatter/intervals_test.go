package formatter

import (
	"fmt"
	"testing"
)

const (
	MS    = 1.0
	SEC   = 1000 * MS
	MIN   = 60 * SEC
	HOUR  = 60 * MIN
	DAY   = 24 * HOUR
	MONTH = 30 * DAY
	YEAR  = 365 * DAY
)

func TestFormatMsIntervalNormalCases(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{1.797, "1.8 ms"},
		{123.7989, "0.12 s"},
		{793.987, "0.79 s"},
		{100.9797, "0.1 s"},
		{1 * SEC, "1 s"},
		{1.4709879 * SEC, "1.5 s"},
		{9.49797 * SEC, "9.5 s"},
		{10 * SEC, "10 s"},
		{59 * SEC, "59 s"},
		{1 * MIN, "60 s"},
		{99.9 * SEC, "1.7 m"},
		{100 * SEC, "1.7 m"},
		{59.23451 * MIN, "59 m"},
		{89.411 * MIN, "89 m"},
		{89.94353 * MIN, "90 m"},
		{99 * MIN, "1.6 h"}, // TOD0: should be 1.7 h
		{99.9 * MIN, "1.7 h"},
		{100 * MIN, "1.7 h"},
		{71.936 * HOUR, "72 h"},
		{72 * HOUR, "3 d"},
		{99 * HOUR, "4.1 d"},
		{89.9 * DAY, "90 d"},
		{90 * DAY, "3 mon"},
		{99 * DAY, "3.3 mon"},
		{7.87978 * MONTH, "7.9 mon"},
		{17.923 * MONTH, "18 mon"},
		{18 * MONTH, "1.5 y"},
		{18.0234234 * MONTH, "1.5 y"},
		{36 * MONTH, "3 y"},
		{3247 * DAY, "8.9 y"},
		{43.34523 * YEAR, "43 y"},
		{99 * YEAR, "99 y"},
		{99*YEAR + 6*SEC, "99 y"},
		{99*YEAR + 6.0004*SEC, "99 y"},
		{99*YEAR + 6.99999*SEC, "99 y"},
		{99.9 * YEAR, "100 y"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.input), func(t *testing.T) {
			if got := formatMsInterval(tt.input); got != tt.expected {
				t.Errorf("formatMsInterval(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
