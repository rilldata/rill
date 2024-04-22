package formatter

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"golang.org/x/exp/constraints"
)

// returns the smallest order of magnitude to which a number has precision --
// basically, the smallest OoM that has a non-zero digit.
func smallestPrecisionMagnitude[T Number](x T) int {
	if isUnsigned(x) {
		return smallestPrecisionMagnitudeInt(uint64(x))
	}
	if isInteger(x) {
		return smallestPrecisionMagnitudeInt(int64(x))
	}
	return smallestPrecisionMagnitudeFloat(float64(x))
}

func smallestPrecisionMagnitudeInt[T int64 | uint64](x T) int {
	if x == 0 {
		return 0
	}

	var e T = 10
	p := 0
	for x/e*e == x {
		e *= 10
		p++
	}
	return p
}

func smallestPrecisionMagnitudeFloat(x float64) int {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return 0 // This should never happen
	}

	if x == 0 {
		return 0
	}

	if math.Abs(x) > 1e280 {
		return smallestPrecisionMagnitudeLargeNumber(x)
	}

	e := 1.0
	p := 0
	for math.Round(x*e)/e != x && p < 324 {
		e *= 10
		p++
	}
	return -p
}

func smallestPrecisionMagnitudeLargeNumber(n float64) int {
	s := fmt.Sprintf("%e", n)
	eIndex := strings.Index(s, "e")
	dotIndex := strings.Index(s, ".")
	exp, _ := strconv.Atoi(s[eIndex+1:])
	digitsAfterDot := eIndex - dotIndex - 1
	i := exp - digitsAfterDot
	return i
}

type Number interface {
	constraints.Integer | constraints.Float
}

func asNumber[T Number](x any) (T, bool) {
	switch x := x.(type) {
	case int:
		return T(x), true
	case int8:
		return T(x), true
	case int16:
		return T(x), true
	case int32:
		return T(x), true
	case int64:
		return T(x), true
	case uint:
		return T(x), true
	case uint8:
		return T(x), true
	case uint16:
		return T(x), true
	case uint32:
		return T(x), true
	case uint64:
		return T(x), true
	case float32:
		return T(x), true
	case float64:
		return T(x), true
	}
	return 0, false
}

func isUnsigned(x any) bool {
	switch x.(type) {
	case uint, uint8, uint16, uint32, uint64:
		return true
	}
	return false
}

func asUnsigned(x any) (uint64, bool) {
	switch x := x.(type) {
	case uint:
		return uint64(x), true
	case uint8:
		return uint64(x), true
	case uint16:
		return uint64(x), true
	case uint32:
		return uint64(x), true
	case uint64:
		return x, true
	}
	return 0, false
}

func isInteger(x any) bool {
	switch x.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return true
	}
	return false
}

func asInteger(x any) (int64, bool) {
	switch x := x.(type) {
	case int:
		return int64(x), true
	case int8:
		return int64(x), true
	case int16:
		return int64(x), true
	case int32:
		return int64(x), true
	case int64:
		return x, true
	}
	return 0, false
}

func isFloat(x any) bool {
	switch x.(type) {
	case float32, float64:
		return true
	}
	return false
}

func asFloat(x any) (float64, bool) {
	switch x := x.(type) {
	case float32:
		return float64(x), true
	case float64:
		return x, true
	}
	return 0, false
}
