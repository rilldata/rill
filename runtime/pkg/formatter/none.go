package formatter

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"
)

type FormatterOptionsNoneStrategy formatterOptionsCommon

type NonFormatter struct {
	Options FormatterOptionsNoneStrategy
}

func NewNonFormatter(options FormatterOptionsNoneStrategy) *NonFormatter {
	return &NonFormatter{Options: options}
}

func (f *NonFormatter) stringFormat(x any) (string, error) {
	return fmt.Sprintf("%v", x), nil
}

func (f *NonFormatter) partsFormat(x any) (numberParts, error) {
	if v, ok := asUnsigned(x); ok {
		return formatNone(v, f.Options)
	}

	if v, ok := asInteger(x); ok {
		return formatNone(v, f.Options)
	}

	if v, ok := asFloat(x); ok {
		return formatNone(v, f.Options)
	}

	return numberParts{}, fmt.Errorf("not a number: %v", x)
}

func formatNone[T Number](x T, ops FormatterOptionsNoneStrategy) (numberParts, error) {
	var numParts numberParts

	if ops.NumberKind == PERCENT {
		x = 100 * x
	}

	if x == 0 {
		numParts = numberParts{Int: "0", Dot: "", Frac: "", Suffix: ""}
	} else {
		p := message.NewPrinter(language.English)
		formatter := number.Decimal(x, number.MaxFractionDigits(20), number.NoSeparator())
		numParts = numStrToParts(p.Sprint(formatter))
	}

	numParts.Suffix = numParts.adjustedSuffix(ops.UpperCaseEForExponent)

	switch ops.NumberKind {
	case DOLLAR:
		numParts.CurrencySymbol = "$"
	case EURO:
		numParts.CurrencySymbol = "€"
	case PERCENT:
		numParts.Percent = "%"
	}

	return numParts, nil
}

func numStrToParts(numStr string) numberParts {
	nonNumRe := regexp.MustCompile(`[a-zA-Z ]`)
	matches := nonNumRe.FindStringIndex(numStr)

	var intPart, fracPart, suffix, dot string

	suffixIndex := len(numStr)
	if matches != nil {
		suffixIndex = matches[0]
		suffix = numStr[suffixIndex:]
	}

	numericPart := numStr[:suffixIndex]

	parts := strings.Split(numericPart, ".")
	intPart = parts[0]
	if len(parts) > 1 {
		fracPart = parts[1]
		dot = "."
	}

	return numberParts{
		Int:    intPart,
		Frac:   fracPart,
		Dot:    dot,
		Suffix: suffix,
	}
}
