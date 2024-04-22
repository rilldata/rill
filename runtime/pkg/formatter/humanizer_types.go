package formatter

import (
	"strings"
)

type numberKind string

const (
	DOLLAR  numberKind = "DOLLAR"
	EURO    numberKind = "EURO"
	PERCENT numberKind = "PERCENT"
	ANY     numberKind = "ANY"
)

type numberParts struct {
	Neg            string
	CurrencySymbol string
	Int            string
	Dot            string
	Frac           string
	Suffix         string
	Percent        string
	ApproxZero     bool
}

func (np numberParts) string() string {
	return np.Neg + np.CurrencySymbol + np.Int + np.Dot + np.Frac + np.Suffix + np.Percent
}

func (np numberParts) adjustedSuffix(upperCaseExp bool) string {
	suffix := shortScaleSuffixIfAvailableForStr(np.Suffix)
	if !upperCaseExp {
		suffix = strings.ReplaceAll(suffix, "E", "e")
	}
	return suffix
}

type rangeFormatSpec struct {
	/**
	 * Minimum order of magnitude for this range.
	 * Target number must have OoM >= minMag.
	 */
	MinMag int
	/**
	 * Supremum order of magnitude for this range.
	 * Target number must have OoM OoM < supMag.
	 */
	SupMag int
	/**
	 *Max number of digits left of decimal point.
	 * If undefined, default is 3 digits
	 */
	MaxDigitsLeft int // TODO: should be 3 by default
	/**
	 * Max number of digits right of decimal point.
	 */
	MaxDigitsRight int
	/**
	 * If set, this will be used as the order of magnitude
	 * for formatting numbers in this range.
	 * For example, if baseMagnitude=3, then we'd have:
	 * - 1,000,000 => 1,000k
	 * - 100 => .1k
	 * If this is set to 0, numbers in this range
	 * will be rendered as plain numbers (no suffix).
	 * If not set, the engineering magnitude of `min` is used by default. (orderOfMagnitudeEng(float64(MinMag)))
	 */
	BaseMagnitude int
	/**
	 * Whether or not to pad numbers with insignificant zeros. If undefined, treated as true
	 */
	PadWithInsignificantZeros bool
	/**
	 * For a range with `maxDigitsRight=0`, by default a trailling
	 * "." will be added if formatting causes some of a number's
	 * true precision to be lost. For example, `123.234` with
	 * `baseMagnitude=0` and `maxDigitsRight=0` will be rendered as
	 * "123.", with the trailing "." retained to indicate that there
	 * is additional precision that is not shown.
	 *
	 * If this is not desired, then setting `useTrailingDot=false` will
	 * remove this decimal point--e.g., in the example above, `123.234`
	 * will be formatted as just "123", with no decimal point.
	 */
	UseTrailingDot bool
}

func newRangeFormatSpec(minMag, supMag, maxDigitsLeft, maxDigitsRight, baseMagnitude int, padWithInsignificantZeros bool) *rangeFormatSpec {
	return &rangeFormatSpec{MinMag: minMag, SupMag: supMag, MaxDigitsLeft: maxDigitsLeft, MaxDigitsRight: maxDigitsRight, BaseMagnitude: baseMagnitude, PadWithInsignificantZeros: padWithInsignificantZeros, UseTrailingDot: true}
}

type formatterOptionsCommon struct {
	NumberKind            numberKind
	UpperCaseEForExponent bool
}

type formatter interface {
	stringFormat(x any) (string, error)
	partsFormat(x any) (numberParts, error)
}
