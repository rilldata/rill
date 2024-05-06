package formatter

import (
	"fmt"
	"math"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
	xnumber "golang.org/x/text/number"
)

// orderOfMagnitudeEng rounds the order of magnitude to the nearest number divisible by 3.
func orderOfMagnitudeEng[T number](x T) int {
	return int(math.Floor(float64(orderOfMagnitude(float64(x)))/3) * 3)
}

// orderOfMagnitude computes the order of magnitude of a number.
func orderOfMagnitude(x float64) int {
	if x == 0 {
		return 0
	}
	mag := math.Floor(math.Log10(math.Abs(x)))
	// having found the order of magnitude, if we divide it
	// out of the number, we should get a number between 1 and 10.
	// However, because of floating point errors, if we get a number
	// very just less than 10, we may have a floating point error,
	// in which we want to bump the order of magnitude up by one.
	//
	// Ex: 0.0009999999999999 has magnitude -4, but if multiply away
	// the magnitude, we get:
	// 0.0009999999999999 * 10**4 = 9.999999999999999
	// -- just less than 10, so we want to bump the magnitude up to -3
	// so that we can format this as e.g. "1.0e-3"
	if 10-math.Abs(x)*math.Pow(10, -mag) < 1e-8 {
		mag++
	}
	return int(mag)
}

type formatNumWithOrderOfMagOps struct {
	newOrder       int
	fractionDigits int
	// Set to true to pad with insignificant zeros.
	// Integers will be padded with zeros if this is set.
	padInsignificantZeros bool
	// Set to `true` to leave a trailing "." in the case
	// of non-integers formatted to e0 with 0 fraction digits.
	// Even if this is `true` integers WILL NOT be formatted with a trailing "."
	trailingDot bool
	// strip commas from output?
	stripCommas bool
}

// formatNumWithOrderOfMag formats a number according to the given order of magnitude and formatting rules.
func formatNumWithOrderOfMag[T number](x T, ops formatNumWithOrderOfMagOps) numberParts {
	if isFloat(x) {
		if math.IsInf(float64(x), 1) {
			return numberParts{integer: "∞"}
		}
		if math.IsInf(float64(x), -1) {
			return numberParts{neg: "-", integer: "∞"}
		}
		if math.IsNaN(float64(x)) {
			return numberParts{integer: "NaN"}
		}
	}

	suffix := fmt.Sprintf("E%d", ops.newOrder)
	if x == 0 {
		frac := ""
		if ops.padInsignificantZeros {
			frac = strings.Repeat("0", ops.fractionDigits)
		}
		return numberParts{integer: "0", dot: ".", frac: frac, suffix: suffix}
	}

	if !ops.padInsignificantZeros {
		spm := smallestPrecisionMagnitude(x)
		newSpm := spm - ops.newOrder
		if newSpm < 0 {
			ops.fractionDigits = min(-newSpm, ops.fractionDigits)
		} else {
			ops.fractionDigits = 0
		}
	}

	// Adjust number by its new order of magnitude
	adjustedNumber := float64(x) / math.Pow(10, float64(ops.newOrder))

	// Format number with potential trailing dot
	p := message.NewPrinter(language.English)
	nf := xnumber.Decimal(
		adjustedNumber,
		xnumber.Scale(ops.fractionDigits),
	)
	formatted := p.Sprint(nf)
	intPart, fracPart := splitFormattedNumber(formatted)

	if ops.stripCommas {
		intPart = strings.ReplaceAll(intPart, ",", "")
	}

	// Detect negatives
	neg := ""
	if strings.HasPrefix(intPart, "-") {
		neg = "-"
		intPart = strings.TrimPrefix(intPart, "-")
	}

	var dot string
	// Decide on dot presence
	if fracPart != "" || (ops.fractionDigits == 0 && ops.trailingDot && !noFraction(float64(x))) {
		dot = "."
	}

	return numberParts{neg: neg, integer: intPart, dot: dot, frac: fracPart, suffix: suffix}
}

func splitFormattedNumber(formatted string) (intPart, fracPart string) {
	parts := strings.Split(formatted, ".")
	intPart = parts[0]
	if len(parts) > 1 {
		fracPart = parts[1]
	}
	return
}

func noFraction(x float64) bool {
	return math.Floor(x) == x
}
