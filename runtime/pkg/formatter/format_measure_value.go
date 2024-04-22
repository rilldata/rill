package formatter

import (
	"fmt"
	"strings"

	// "d3-format" equivalent needed
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

type MeasureValueFormatter interface {
	Format(value any) string
}

type HumanReadableFormatter struct {
	formatter formatter
}

func (f *HumanReadableFormatter) Format(value any) string {
	if res, err := f.formatter.stringFormat(value); err == nil {
		return res
	}
	return fallBackFormatter.Format(value)
}

type D3Formatter struct {
	format string
}

func (f *D3Formatter) Format(value any) string {
	// TODO: Simulate d3 formatter application
	return fmt.Sprintf(f.format, value)
}

type FallBackFormatter struct{}

func (f *FallBackFormatter) Format(value any) string {
	return fmt.Sprintf("%v", value)
}

var fallBackFormatter MeasureValueFormatter = &FallBackFormatter{}

func NewMeasureValueFormatter(measureSpec *runtimev1.MetricsViewSpec_MeasureV2, useUnabridged bool) (MeasureValueFormatter, error) {
	if measureSpec == nil {
		return nil, fmt.Errorf("measureSpec is nil")
	}

	if measureSpec.FormatD3 != "" {
		return &D3Formatter{format: measureSpec.FormatD3}, nil
	}

	formatter, err := presetFormatter(measureSpec.FormatPreset, useUnabridged)
	if err != nil {
		return nil, err
	}
	return &HumanReadableFormatter{formatter}, nil
}

func presetFormatter(preset string, useUnabridged bool) (formatter, error) {
	if useUnabridged {
		if preset == "humanize" {
			return &IntervalExpFormatter{}, nil
		}
		return NewNonFormatter(defaultNoneOptions()), nil
	}

	switch strings.ToLower(strings.TrimSpace(preset)) {
	case "none":
		return NewNonFormatter(defaultNoneOptions()), nil
	case "humanize", "":
		return NewPerRangeFormatter(defaultGenericNumOptions())
	case "currency_usd":
		return NewPerRangeFormatter(defaultCurrencyOptions(DOLLAR))
	case "currency_eur":
		return NewPerRangeFormatter(defaultCurrencyOptions(EURO))
	case "percentage":
		return NewPerRangeFormatter(defaultPercentOptions())
	case "interval_ms":
		return &IntervalFormatter{}, nil
	default:
		return NewPerRangeFormatter(defaultGenericNumOptions())
	}
}
