package formatter

import (
	"strings"
)

// Formatter formats numbers, percentages and time intervals
// The code is based on the logic implemented on the frontend
// Type and file names are kept mostly the same for consistency and ease of updating
type Formatter interface {
	StringFormat(x any) (string, error)
}

func NewPresetFormatter(preset string, useUnabridged bool) (Formatter, error) {
	// Default preset is "humanize"
	if preset == "" {
		preset = "humanize"
	}

	if useUnabridged {
		if preset == "humanize" {
			return newIntervalExpFormatter(), nil
		}
		return newNonFormatter(), nil
	}

	switch strings.ToLower(strings.TrimSpace(preset)) {
	case "none":
		return newNonFormatter(), nil
	case "humanize":
		return newPerRangeFormatter(defaultGenericNumOptions())
	case "currency_usd":
		return newPerRangeFormatter(defaultCurrencyOptions(numDollar))
	case "currency_eur":
		return newPerRangeFormatter(defaultCurrencyOptions(numEuro))
	case "percentage":
		return newPerRangeFormatter(defaultPercentOptions())
	case "interval_ms":
		return newIntervalFormatter(), nil
	}

	return newPerRangeFormatter(defaultGenericNumOptions())
}

func NewD3Formatter(useUnabridged bool) (Formatter, error) {
	// TODO: implement d3 formatter
	return newNonFormatter(), nil
}
