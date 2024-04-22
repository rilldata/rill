package formatter

import (
	"errors"
	"fmt"
	"sort"
)

type FormatterRangeSpecsStrategy struct {
	formatterOptionsCommon
	RangeSpecs            []rangeFormatSpec
	DefaultMaxDigitsRight int
}

type PerRangeFormatter struct {
	Options FormatterRangeSpecsStrategy
}

func NewPerRangeFormatter(options FormatterRangeSpecsStrategy) (*PerRangeFormatter, error) {
	// Sort range specs from small to large by lower bound
	sort.Slice(options.RangeSpecs, func(i, j int) bool {
		return options.RangeSpecs[i].MinMag < options.RangeSpecs[j].MinMag
	})

	// Validate range specs
	for i, r := range options.RangeSpecs {
		if r.MinMag >= r.SupMag {
			return nil, fmt.Errorf("invalid range: min %d is not strictly less than sup %d", r.MinMag, r.SupMag)
		}
		if i > 0 && options.RangeSpecs[i-1].SupMag > r.MinMag {
			return nil, errors.New("ranges must not overlap")
		}
		if i > 0 && options.RangeSpecs[i-1].SupMag != r.MinMag {
			return nil, errors.New("gaps are not allowed between formatter ranges")
		}
	}

	return &PerRangeFormatter{Options: options}, nil
}

func (f *PerRangeFormatter) stringFormat(x any) (string, error) {
	parts, _ := f.partsFormat(x)
	return parts.string(), nil
}

func (f *PerRangeFormatter) partsFormat(x any) (numberParts, error) {
	if v, ok := asUnsigned(x); ok {
		return formatPerRange(v, f.Options), nil
	}

	if v, ok := asInteger(x); ok {
		return formatPerRange(v, f.Options), nil
	}

	if v, ok := asFloat(x); ok {
		return formatPerRange(v, f.Options), nil
	}

	return numberParts{}, fmt.Errorf("not a number: %v", x)
}

func formatPerRange[T Number](x T, ops FormatterRangeSpecsStrategy) numberParts {
	rangeSpecs := ops.RangeSpecs
	defaultMaxDigitsRight := ops.DefaultMaxDigitsRight
	// Scale value for percentage if applicable
	if ops.NumberKind == PERCENT {
		x *= 100
	}

	var numParts *numberParts
	if x == 0 {
		numParts = &numberParts{Int: "0"}
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
		np := formatNumWithOrderOfMag(x, magE, defaultMaxDigitsRight, true, false, false)
		if countDigits(np.Int) > 3 {
			np = formatNumWithOrderOfMag(x, magE+3, defaultMaxDigitsRight, true, false, false)
		}
		numParts = &np
	}

	numParts.Suffix = numParts.adjustedSuffix(ops.UpperCaseEForExponent)

	// Adjust currency symbols and percent signs based on number kind
	switch ops.NumberKind {
	case DOLLAR:
		numParts.CurrencySymbol = "$"
	case EURO:
		numParts.CurrencySymbol = "€"
	case PERCENT:
		numParts.Percent = "%"
	}

	return *numParts
}

func formatWithRangeSpec[T Number](x T, spec rangeFormatSpec) numberParts {
	padWithInsignificantZeros := spec.PadWithInsignificantZeros
	useTrailingDot := spec.UseTrailingDot

	return formatNumWithOrderOfMag(x, spec.BaseMagnitude, spec.MaxDigitsRight, padWithInsignificantZeros, useTrailingDot, false)
}

// numberPartsValidForRangeSpec checks if the given number parts are valid according to the specified range spec.
func numberPartsValidForRangeSpec(parts numberParts, spec rangeFormatSpec) bool {
	return countDigits(parts.Int) <= spec.MaxDigitsLeft &&
		countDigits(parts.Frac) <= spec.MaxDigitsRight
}

// numPartsNotZero checks if the number parts represent a non-zero number.
func numPartsNotZero(parts numberParts) bool {
	return countNonZeroDigits(parts.Int) > 0 || countNonZeroDigits(parts.Frac) > 0
}
