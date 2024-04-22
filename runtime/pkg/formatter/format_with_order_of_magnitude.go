package formatter

import (
	"fmt"
	"math"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"
)

// orderOfMagnitudeEng rounds the order of magnitude to the nearest number divisible by 3.
func orderOfMagnitudeEng[T Number](x T) int {
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

// formatNumWithOrderOfMag formats a number according to the given order of magnitude and formatting rules.
func formatNumWithOrderOfMag[T Number](x T,
	newOrder int,
	fractionDigits int,
	// Set to true to pad with insignificant zeros.
	// Integers will be padded with zeros if this is set.
	padInsignificantZeros bool,
	// Set to `true` to leave a trailing "." in the case
	// of non-integers formatted to e0 with 0 fraction digits.
	// Even if this is `true` integers WILL NOT be formatted with a trailing "."
	trailingDot bool,
	// strip commas from output?
	stripCommas bool,
) numberParts {
	if isFloat(x) {
		if math.IsInf(float64(x), 1) {
			return numberParts{Int: "∞"}
		}
		if math.IsInf(float64(x), -1) {
			return numberParts{Neg: "-", Int: "∞"}
		}
		if math.IsNaN(float64(x)) {
			return numberParts{Int: "NaN"}
		}
	}

	suffix := fmt.Sprintf("E%d", newOrder)
	if x == 0 {
		frac := ""
		if padInsignificantZeros {
			frac = strings.Repeat("0", fractionDigits)
		}
		return numberParts{Int: "0", Dot: ".", Frac: frac, Suffix: suffix}
	}

	if !padInsignificantZeros {
		spm := smallestPrecisionMagnitude(x)
		newSpm := spm - newOrder
		if newSpm < 0 {
			fractionDigits = min(-newSpm, fractionDigits)
		} else {
			fractionDigits = 0
		}
	}

	// Adjust number by its new order of magnitude
	adjustedNumber := float64(x) / math.Pow(10, float64(newOrder))

	// Format number with potential trailing dot
	p := message.NewPrinter(language.English)
	formatter := number.Decimal(
		adjustedNumber,
		number.Scale(fractionDigits),
	)
	formatted := p.Sprint(formatter)
	intPart, fracPart := splitFormattedNumber(formatted)

	if stripCommas {
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
	if fracPart != "" || (fractionDigits == 0 && trailingDot && !noFraction(float64(x))) {
		dot = "."
	}

	return numberParts{Neg: neg, Int: intPart, Dot: dot, Frac: fracPart, Suffix: suffix}
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
