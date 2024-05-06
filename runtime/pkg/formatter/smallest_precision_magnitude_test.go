package formatter

import (
	"fmt"
	"testing"
)

func TestSmallestPrecisionMagnitudeInt(t *testing.T) {
	tests := []struct {
		input    int64
		expected int
	}{
		{0, 0},
		{1, 0},
		{10, 1},
		{100, 2},
		{-100, 2},
		{0, 0},
		{-0, 0},
		{123, 0},
		{1230, 1},
		{123000000000, 9},
		{1239347293742974, 0},
		{6000000000, 9},
		{-123, 0},
		{-1230, 1},
		{-123000000000, 9},
		{-1239347293742974, 0},
		{-6000000000, 9},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.input), func(t *testing.T) {
			if got := smallestPrecisionMagnitude(tt.input); got != tt.expected {
				t.Errorf("smallestPrecisionMagnitude(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSmallestPrecisionMagnitudeFloat(t *testing.T) {
	tests := []struct {
		input    float64
		expected int
	}{
		{0.0, 0},
		{1.0, 0},
		{0.01, -2},
		{0.1, -1},
		{0.000000001, -9},
		{0.000027391, -9},
		{0.973427391, -9},
		{0.1, -1},
		{0.000000001, -9},
		{0.000027391, -9},
		{0.973427391, -9},
		{710.7237956, -7},
		{-710.7237956, -7},
		{79879879710.7237, -4},
		{-79879879710.7237, -4},
		{2.2247239873252e-3523, 0},
		{2.2247239873252e-308, -324},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.input), func(t *testing.T) {
			if got := smallestPrecisionMagnitude(tt.input); got != tt.expected {
				t.Errorf("smallestPrecisionMagnitude(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
