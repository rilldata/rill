package formatter

import (
	"errors"
	"fmt"
	"sort"
)

type formatterRangeSpecsStrategy struct {
	numKind               numberKind
	upperCaseEForExponent bool
	defaultMaxDigitsRight int
	rangeSpecs            []rangeFormatSpec
}

type perRangeFormatter struct {
	options formatterRangeSpecsStrategy
}

func newPerRangeFormatter(options formatterRangeSpecsStrategy) (*perRangeFormatter, error) {
	// Sort range specs from small to large by lower bound
	sort.Slice(options.rangeSpecs, func(i, j int) bool {
		return options.rangeSpecs[i].minMag < options.rangeSpecs[j].minMag
	})

	// Validate range specs
	for i, r := range options.rangeSpecs {
		if r.minMag >= r.supMag {
			return nil, fmt.Errorf("invalid range: min %d is not strictly less than sup %d", r.minMag, r.supMag)
		}
		if i > 0 && options.rangeSpecs[i-1].supMag > r.minMag {
			return nil, errors.New("ranges must not overlap")
		}
		if i > 0 && options.rangeSpecs[i-1].supMag != r.minMag {
			return nil, errors.New("gaps are not allowed between formatter ranges")
		}
	}

	return &perRangeFormatter{options: options}, nil
}

func (f *perRangeFormatter) StringFormat(x any) (string, error) {
	var numberParts *numberParts
	if v, ok := asUnsigned(x); ok {
		numberParts = partsFormat(v, f.options)
		return numberParts.string(), nil
	}

	if v, ok := asInteger(x); ok {
		numberParts = partsFormat(v, f.options)
		return numberParts.string(), nil
	}

	if v, ok := asFloat(x); ok {
		numberParts = partsFormat(v, f.options)
		return numberParts.string(), nil
	}

	return "", fmt.Errorf("not a number: %v", x)
}

func partsFormat[T number](x T, ops formatterRangeSpecsStrategy) *numberParts {
	rangeSpecs := ops.rangeSpecs
	defaultMaxDigitsRight := ops.defaultMaxDigitsRight
	// Scale value for percentage if applicable
	if ops.numKind == numPercent {
		x *= 100
	}

	var numParts *numberParts
	if x == 0 {
		numParts = &numberParts{integer: "0"}
	}

	if numParts == nil {
		for _, spec := range rangeSpecs {
			np := formatWithRangeSpec(x, spec)
			if numberPartsValidForRangeSpec(np, spec) && numPartsNotZero(np) {
				numParts = &np
				break
			}
		}
	}

	// If no valid format was found, apply default formatting
	if numParts == nil {
		magE := orderOfMagnitudeEng(x)
		np := formatNumWithOrderOfMag(x, formatNumWithOrderOfMagOps{magE, defaultMaxDigitsRight, true, false, false})
		if countDigits(np.integer) > 3 {
			np = formatNumWithOrderOfMag(x, formatNumWithOrderOfMagOps{magE + 3, defaultMaxDigitsRight, true, false, false})
		}
		numParts = &np
	}

	numParts.adjustSuffix(ops.upperCaseEForExponent)

	// Adjust currency symbols and percent signs based on number kind
	switch ops.numKind {
	case numDollar:
		numParts.currencySymbol = "$"
	case numEuro:
		numParts.currencySymbol = "€"
	case numPercent:
		numParts.percent = "%"
	}

	return numParts
}

func formatWithRangeSpec[T number](x T, spec rangeFormatSpec) numberParts {
	padWithInsignificantZeros := spec.padWithInsignificantZeros
	useTrailingDot := spec.useTrailingDot

	ops := formatNumWithOrderOfMagOps{spec.baseMagnitude, spec.maxDigitsRight, padWithInsignificantZeros, useTrailingDot, false}
	return formatNumWithOrderOfMag(x, ops)
}

// numberPartsValidForRangeSpec checks if the given number parts are valid according to the specified range spec.
func numberPartsValidForRangeSpec(parts numberParts, spec rangeFormatSpec) bool {
	return countDigits(parts.integer) <= spec.maxDigitsLeft &&
		countDigits(parts.frac) <= spec.maxDigitsRight
}

// numPartsNotZero checks if the number parts represent a non-zero number.
func numPartsNotZero(parts numberParts) bool {
	return countNonZeroDigits(parts.integer) > 0 || countNonZeroDigits(parts.frac) > 0
}
