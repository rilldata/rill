package formatter

import (
	"fmt"
	"testing"
)

func TestPerRangeFormatterInvalidRange(t *testing.T) {
	tests := []formatterRangeSpecsStrategy{
		{
			rangeSpecs: []rangeFormatSpec{
				*NewRangeFormatSpec1(3, 3, 0),
				*NewRangeFormatSpec1(-3, 3, 3),
			},
			numKind:               numAny,
			defaultMaxDigitsRight: 2,
		},
		{
			rangeSpecs: []rangeFormatSpec{
				*NewRangeFormatSpec1(3, 2, 0),
				*NewRangeFormatSpec1(-3, 3, 3),
			},
			numKind:               numAny,
			defaultMaxDigitsRight: 2,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("invalid range #%v", i), func(t *testing.T) {
			if _, err := newPerRangeFormatter(tt); err == nil {
				t.Errorf("failed: invalid range #%v", i)
			}
		})
	}
}

func TestPerRangeFormatterOverlappingRanges(t *testing.T) {
	tests := []formatterRangeSpecsStrategy{
		{
			rangeSpecs: []rangeFormatSpec{
				*NewRangeFormatSpec1(2, 6, 0),
				*NewRangeFormatSpec1(-3, 3, 3),
			},
			numKind:               numAny,
			defaultMaxDigitsRight: 2,
		},
		{
			rangeSpecs: []rangeFormatSpec{
				*NewRangeFormatSpec1(2, 6, 0),
				*NewRangeFormatSpec1(-3, 3, 3),
				*NewRangeFormatSpec1(-6, -3, 0),
				*NewRangeFormatSpec1(6, 10, 0),
			},
			numKind:               numAny,
			defaultMaxDigitsRight: 2,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("invalid range #%v", i), func(t *testing.T) {
			if _, err := newPerRangeFormatter(tt); err == nil {
				t.Errorf("failed: overlapping ranges #%v", i)
			}
		})
	}
}

func TestPerRangeFormatterGapInRangeCoverage(t *testing.T) {
	tests := []formatterRangeSpecsStrategy{
		{
			rangeSpecs: []rangeFormatSpec{
				*NewRangeFormatSpec1(6, 9, 0),
				*NewRangeFormatSpec1(-3, 3, 3),
			},
			numKind:               numAny,
			defaultMaxDigitsRight: 2,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("invalid range #%v", i), func(t *testing.T) {
			if _, err := newPerRangeFormatter(tt); err == nil {
				t.Errorf("failed: gap in range coverage #%v", i)
			}
		})
	}
}

func TestPerRangeFormatter1(t *testing.T) {
	ops := formatterRangeSpecsStrategy{
		rangeSpecs: []rangeFormatSpec{
			*newRangeFormatSpec(3, 6, 6, 0, 0, true),
			*newRangeFormatSpec(-3, 3, 3, 3, 0, true),
		},
		numKind:               numAny,
		defaultMaxDigitsRight: 2,
	}
	tests := []struct {
		input    any
		expected string
	}{
		// integers
		{999_999_999, "1.00B"},
		{12_345_789, "12.35M"},
		{2_345_789, "2.35M"},
		{999_999, "999,999"},
		{345_789, "345,789"},
		{45_789, "45,789"},
		{5_789, "5,789"},
		{999, "999.000"},
		{789, "789.000"},
		{89, "89.000"},
		{9, "9.000"},
		{0, "0"},
		{-999_999_999, "-1.00B"},
		{-12_345_789, "-12.35M"},
		{-2_345_789, "-2.35M"},
		{-999_999, "-999,999"},
		{-345_789, "-345,789"},
		{-45_789, "-45,789"},
		{-5_789, "-5,789"},
		{-999, "-999.000"},
		{-789, "-789.000"},
		{-89, "-89.000"},
		{-9, "-9.000"},
		{-0, "0"},

		// non integers
		{999_999_999.1234686, "1.00B"},
		{12_345_789.1234686, "12.35M"},
		{2_345_789.1234686, "2.35M"},
		{999_999.1234686, "999,999."},
		{345_789.1234686, "345,789."},
		{45_789.1234686, "45,789."},
		{5_789.1234686, "5,789."},
		{999.1234686, "999.123"},
		{789.1234686, "789.123"},
		{89.1234686, "89.123"},
		{9.1234686, "9.123"},
		{0.1234686, "0.123"},
		{-999_999_999.1234686, "-1.00B"},
		{-12_345_789.1234686, "-12.35M"},
		{-2_345_789.1234686, "-2.35M"},
		{-999_999.1234686, "-999,999."},
		{-345_789.1234686, "-345,789."},
		{-45_789.1234686, "-45,789."},
		{-5_789.1234686, "-5,789."},
		{-999.1234686, "-999.123"},
		{-789.1234686, "-789.123"},
		{-89.1234686, "-89.123"},
		{-9.1234686, "-9.123"},
		{-0.1234686, "-0.123"},

		// infinitesimals
		{0.00095, "0.001"},
		{0.000999999, "0.001"},
		{0.00012335234, "123.35e-6"},
		{0.000_000_999999, "1.00e-6"},
		{0.000_000_02341253, "23.41e-9"},
		{0.000_000_000_999999, "1.00e-9"},

		// padding with insignificant zeros
		{9.1, "9.100"},
		{9.12, "9.120"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.input), func(t *testing.T) {
			formatter, err := newPerRangeFormatter(ops)
			if err != nil {
				t.Errorf("failed: %v", err)
			}
			if got, _ := formatter.stringFormat(tt.input); got != tt.expected {
				t.Errorf("perRangeFormatter.stringFormat(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestPerRangeFormatter2(t *testing.T) {
	ops := formatterRangeSpecsStrategy{
		rangeSpecs: []rangeFormatSpec{
			*newRangeFormatSpec(3, 6, 6, 0, 0, false),
			*newRangeFormatSpec(-3, 3, 3, 3, 0, false),
		},
		numKind:               numAny,
		defaultMaxDigitsRight: 2,
	}
	tests := []struct {
		input    any
		expected string
	}{
		// integers
		{999_999_999, "1.00B"},
		{12_345_789, "12.35M"},
		{2_345_789, "2.35M"},
		{999_999, "999,999"},
		{345_789, "345,789"},
		{45_789, "45,789"},
		{5_789, "5,789"},
		{999, "999"},
		{789, "789"},
		{89, "89"},
		{9, "9"},
		{0, "0"},
		{-999_999_999, "-1.00B"},
		{-12_345_789, "-12.35M"},
		{-2_345_789, "-2.35M"},
		{-999_999, "-999,999"},
		{-345_789, "-345,789"},
		{-45_789, "-45,789"},
		{-5_789, "-5,789"},
		{-999, "-999"},
		{-789, "-789"},
		{-89, "-89"},
		{-9, "-9"},
		{-0, "0"},

		// non integers
		{9.1, "9.1"},
		{9.12, "9.12"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.input), func(t *testing.T) {
			formatter, err := newPerRangeFormatter(ops)
			if err != nil {
				t.Errorf("failed: %v", err)
			}
			if got, _ := formatter.stringFormat(tt.input); got != tt.expected {
				t.Errorf("perRangeFormatter.stringFormat(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestPerRangeFormatter3(t *testing.T) {
	ops := formatterRangeSpecsStrategy{
		rangeSpecs: []rangeFormatSpec{
			*NewRangeFormatSpec3(3, 6, 6, 0, 0, true, false),
			*newRangeFormatSpec(-3, 3, 3, 3, 0, true),
		},
		numKind:               numAny,
		defaultMaxDigitsRight: 2,
	}
	tests := []struct {
		input    any
		expected string
	}{
		{999_999.1234686, "999,999"},
		{345_789.1234686, "345,789"},

		{45_789.1234686, "45,789"},
		{5_789.1234686, "5,789"},
		{-999_999.1234686, "-999,999"},
		{-345_789.1234686, "-345,789"},
		{-45_789.1234686, "-45,789"},
		{-5_789.1234686, "-5,789"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.input), func(t *testing.T) {
			formatter, err := newPerRangeFormatter(ops)
			if err != nil {
				t.Errorf("failed: %v", err)
			}
			if got, _ := formatter.stringFormat(tt.input); got != tt.expected {
				t.Errorf("perRangeFormatter.stringFormat(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func NewRangeFormatSpec1(minMag, supMag, maxDigitsRight int) *rangeFormatSpec {
	return &rangeFormatSpec{minMag: minMag, supMag: supMag, maxDigitsLeft: 3, maxDigitsRight: maxDigitsRight, baseMagnitude: orderOfMagnitudeEng(float64(minMag)), padWithInsignificantZeros: true, useTrailingDot: true}
}

func NewRangeFormatSpec3(minMag, supMag, maxDigitsLeft, maxDigitsRight, baseMagnitude int, padWithInsignificantZeros, useTrailingDot bool) *rangeFormatSpec {
	return &rangeFormatSpec{minMag: minMag, supMag: supMag, maxDigitsLeft: maxDigitsLeft, maxDigitsRight: maxDigitsRight, baseMagnitude: baseMagnitude, padWithInsignificantZeros: padWithInsignificantZeros, useTrailingDot: useTrailingDot}
}
