package formatter

import (
	"strings"
)

type numberKind string

const (
	numDollar  numberKind = "dollar"
	numEuro    numberKind = "euro"
	numPercent numberKind = "percent"
	numAny     numberKind = "any"
)

type numberParts struct {
	neg            string
	currencySymbol string
	integer        string
	dot            string
	frac           string
	suffix         string
	percent        string
}

func (np *numberParts) string() string {
	return np.neg + np.currencySymbol + np.integer + np.dot + np.frac + np.suffix + np.percent
}

func (np *numberParts) adjustSuffix(upperCaseExp bool) {
	suffix := shortScaleSuffixIfAvailableForStr(np.suffix)
	if !upperCaseExp {
		suffix = strings.ReplaceAll(suffix, "E", "e")
	}
	np.suffix = suffix
}

type rangeFormatSpec struct {
	/**
	 * Minimum order of magnitude for this range.
	 * Target number must have OoM >= minMag.
	 */
	minMag int
	/**
	 * Supremum order of magnitude for this range.
	 * Target number must have OoM OoM < supMag.
	 */
	supMag int
	/**
	 *Max number of digits left of decimal point.
	 * If undefined, default is 3 digits
	 */
	maxDigitsLeft int // TODO: should be 3 by default
	/**
	 * Max number of digits right of decimal point.
	 */
	maxDigitsRight int
	/**
	 * If set, this will be used as the order of magnitude
	 * for formatting numbers in this range.
	 * For example, if baseMagnitude=3, then we'd have:
	 * - 1,000,000 => 1,000k
	 * - 100 => .1k
	 * If this is set to 0, numbers in this range
	 * will be rendered as plain numbers (no suffix).
	 * If not set, the engineering magnitude of `min` is used by default. (orderOfMagnitudeEng(float64(minMag)))
	 */
	baseMagnitude int
	/**
	 * Whether or not to pad numbers with insignificant zeros. If undefined, treated as true
	 */
	padWithInsignificantZeros bool
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
	useTrailingDot bool
}

func newRangeFormatSpec(minMag, supMag, maxDigitsLeft, maxDigitsRight, baseMagnitude int, padWithInsignificantZeros bool) *rangeFormatSpec {
	return &rangeFormatSpec{minMag: minMag, supMag: supMag, maxDigitsLeft: maxDigitsLeft, maxDigitsRight: maxDigitsRight, baseMagnitude: baseMagnitude, padWithInsignificantZeros: padWithInsignificantZeros, useTrailingDot: true}
}

type formatter interface {
	stringFormat(x any) (string, error)
}
