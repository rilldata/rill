package formatter

import (
	"fmt"
	"math"
	"testing"
)

func TestOrderOfMagnitudeEng(t *testing.T) {
	tests := []struct {
		input    float64
		expected int
	}{
		{0, 0},
		{2.23434, 0},
		{10, 0},
		{210, 0},
		{3210, 3},
		{43210, 3},
		{9_876_543_210, 9},
		{876_543_210, 6},
		{76_543_210, 6},
		{6_543_210, 6},
		{0.1, -3},
		{0.01, -3},
		{0.001, -3},
		{0.000_000_000_001, -12},
		{0.000_000_000_01, -12},
		{0.000_000_000_1, -12},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.input), func(t *testing.T) {
			got := orderOfMagnitudeEng(tt.input)
			if got != tt.expected {
				t.Errorf("orderOfMagnitudeEng(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFormatNumWithOrderOfMag(t *testing.T) {
	testsFloat := []struct {
		input    []interface{}
		expected numberParts
	}{
		{
			[]interface{}{math.Inf(1), 3, 4, true, false, false},
			numberParts{integer: "∞", dot: "", frac: "", suffix: ""},
		},
		{
			[]interface{}{math.Inf(-1), 3, 4, true, false, false},
			numberParts{neg: "-", integer: "∞", dot: "", frac: "", suffix: ""},
		},
		{
			[]interface{}{math.NaN(), 3, 4, true, false, false},
			numberParts{integer: "NaN", dot: "", frac: "", suffix: ""},
		},
		{
			[]interface{}{0.001, 0, 5, false, false, false},
			numberParts{integer: "0", dot: ".", frac: "001", suffix: "E0"},
		},
		{
			[]interface{}{0.001, 0, 5, true, false, false},
			numberParts{integer: "0", dot: ".", frac: "00100", suffix: "E0"},
		},
		{
			[]interface{}{0.001, -3, 5, false, false, false},
			numberParts{integer: "1", dot: "", frac: "", suffix: "E-3"},
		},
		{
			[]interface{}{0.001, -3, 5, true, false, false},
			numberParts{integer: "1", dot: ".", frac: "00000", suffix: "E-3"},
		},
		// yes trailing dot
		{
			[]interface{}{710.272337956, 0, 0, true, true, false},
			numberParts{integer: "710", dot: ".", frac: "", suffix: "E0"},
		},
		{
			[]interface{}{710.272337956, 0, 0, false, true, false},
			numberParts{integer: "710", dot: ".", frac: "", suffix: "E0"},
		},

		// no trailing dot
		{
			[]interface{}{710.272337956, 0, 0, true, false, false},
			numberParts{integer: "710", dot: "", frac: "", suffix: "E0"},
		},
		{
			[]interface{}{710.272337956, 0, 0, false, false, false},
			numberParts{integer: "710", dot: "", frac: "", suffix: "E0"},
		},

		{
			[]interface{}{710.7237956, 0, 2, true, false, false},
			numberParts{integer: "710", dot: ".", frac: "72", suffix: "E0"},
		},
		{
			[]interface{}{710.7237956, 0, 2, false, false, false},
			numberParts{integer: "710", dot: ".", frac: "72", suffix: "E0"},
		},

		// not stripping commas
		{
			[]interface{}{523523710.7237956, 0, 5, true, false, false},
			numberParts{integer: "523,523,710", dot: ".", frac: "72380", suffix: "E0"},
		},
		{
			[]interface{}{523523710.7237956, 0, 5, false, false, false},
			numberParts{integer: "523,523,710", dot: ".", frac: "72380", suffix: "E0"},
		},
		// yes stripping commas
		{
			[]interface{}{523523710.7237956, 0, 5, true, false, true},
			numberParts{integer: "523523710", dot: ".", frac: "72380", suffix: "E0"},
		},
		{
			[]interface{}{523523710.7237956, 0, 5, false, false, true},
			numberParts{integer: "523523710", dot: ".", frac: "72380", suffix: "E0"},
		},

		{
			[]interface{}{0.00087000001, -3, 5, false, false, false},
			numberParts{integer: "0", dot: ".", frac: "87000", suffix: "E-3"},
		},
		{
			[]interface{}{0.00087000001, -3, 5, true, false, false},
			numberParts{integer: "0", dot: ".", frac: "87000", suffix: "E-3"},
		},

		{
			[]interface{}{0.00087, -3, 5, false, false, false},
			numberParts{integer: "0", dot: ".", frac: "87", suffix: "E-3"},
		},
		{
			[]interface{}{0.00087, -3, 5, true, false, false},
			numberParts{integer: "0", dot: ".", frac: "87000", suffix: "E-3"},
		},

		// same as above but negative
		{
			[]interface{}{-0.00087000001, -3, 5, false, false, false},
			numberParts{neg: "-", integer: "0", dot: ".", frac: "87000", suffix: "E-3"},
		},
		{
			[]interface{}{-0.00087000001, -3, 5, true, false, false},
			numberParts{neg: "-", integer: "0", dot: ".", frac: "87000", suffix: "E-3"},
		},

		{
			[]interface{}{-0.00087, -3, 5, false, false, false},
			numberParts{neg: "-", integer: "0", dot: ".", frac: "87", suffix: "E-3"},
		},
		{
			[]interface{}{-0.00087, -3, 5, true, false, false},
			numberParts{neg: "-", integer: "0", dot: ".", frac: "87000", suffix: "E-3"},
		},
	}

	for _, tt := range testsFloat {
		t.Run(fmt.Sprint(tt.input), func(t *testing.T) {
			x := tt.input[0].(float64)
			newOrder := tt.input[1].(int)
			fractionDigits := tt.input[2].(int)
			padInsignificantZeros := tt.input[3].(bool)
			trailingDot := tt.input[4].(bool)
			stripCommas := tt.input[5].(bool)

			got := formatNumWithOrderOfMag(x, newOrder, fractionDigits, padInsignificantZeros, trailingDot, stripCommas)
			if got != tt.expected {
				t.Errorf("formatNumWithOrderOfMag(%+v) = %+v, want %+v", tt.input, got, tt.expected)
			}
		})
	}

	testsInt := []struct {
		input    []interface{}
		expected numberParts
	}{
		{
			[]interface{}{0, 5, 4, false, false, false},
			numberParts{integer: "0", dot: ".", frac: "", suffix: "E5"},
		},
		{
			[]interface{}{0, 5, 4, true, false, false},
			numberParts{integer: "0", dot: ".", frac: "0000", suffix: "E5"},
		},
		{
			[]interface{}{0, -5, 2, true, false, false},
			numberParts{integer: "0", dot: ".", frac: "00", suffix: "E-5"},
		},

		{
			[]interface{}{1, 3, 5, false, false, false},
			numberParts{integer: "0", dot: ".", frac: "001", suffix: "E3"},
		},
		{
			[]interface{}{1, 3, 5, true, false, false},
			numberParts{integer: "0", dot: ".", frac: "00100", suffix: "E3"},
		},

		//  stripCommas = true
		{
			[]interface{}{1, -3, 5, false, false, true},
			numberParts{integer: "1000", dot: "", frac: "", suffix: "E-3"},
		},
		{
			[]interface{}{1, -3, 5, true, false, true},
			numberParts{integer: "1000", dot: ".", frac: "00000", suffix: "E-3"},
		},

		// stripCommas = false (by default)
		{
			[]interface{}{1, -3, 5, false, false, false},
			numberParts{integer: "1,000", dot: "", frac: "", suffix: "E-3"},
		},
		{
			[]interface{}{1, -3, 5, true, false, false},
			numberParts{integer: "1,000", dot: ".", frac: "00000", suffix: "E-3"},
		},
	}

	for _, tt := range testsInt {
		t.Run(fmt.Sprint(tt.input), func(t *testing.T) {
			x := tt.input[0].(int)
			newOrder := tt.input[1].(int)
			fractionDigits := tt.input[2].(int)
			padInsignificantZeros := tt.input[3].(bool)
			trailingDot := tt.input[4].(bool)
			stripCommas := tt.input[5].(bool)

			got := formatNumWithOrderOfMag(x, newOrder, fractionDigits, padInsignificantZeros, trailingDot, stripCommas)
			if got != tt.expected {
				t.Errorf("formatNumWithOrderOfMag(%+v) = %+v, want %+v", tt.input, got, tt.expected)
			}
		})
	}

}
