package formatter

import (
	"fmt"
	"strings"

	// "d3-format" equivalent needed
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"go.uber.org/zap"
)

// MeasureValueFormatter formats measure values
// The code is based on the logic implemented on the frontend
// Type and file names are kept mostly the same for consistency and ease of updating
type MeasureValueFormatter struct {
	formatter
	logger *zap.Logger
}

func (f *MeasureValueFormatter) Format(value any) string {
	str, err := f.stringFormat(value)
	if err == nil {
		return str
	}
	f.logger.Warn("failed to format value, returning non-formatted value", zap.Any("value", value), zap.Error(err))
	return fmt.Sprintf("%v", value)
}

type d3Formatter struct {
	format string
}

func (f *d3Formatter) stringFormat(value any) (string, error) {
	// TODO: Simulate d3 formatter application
	return fmt.Sprintf(f.format, value), nil
}

func NewMeasureValueFormatter(measureSpec *runtimev1.MetricsViewSpec_MeasureV2, useUnabridged bool, logger *zap.Logger) (*MeasureValueFormatter, error) {
	if measureSpec == nil {
		return nil, fmt.Errorf("measureSpec is nil")
	}

	if measureSpec.FormatD3 != "" {
		return &MeasureValueFormatter{formatter: &d3Formatter{format: measureSpec.FormatD3}, logger: logger}, nil
	}

	f, err := presetFormatter(measureSpec.FormatPreset, useUnabridged)
	if err != nil {
		return nil, err
	}
	return &MeasureValueFormatter{f, logger}, nil
}

func NewMeasureValuePresetFormatter(preset string, useUnabridged bool, logger *zap.Logger) (*MeasureValueFormatter, error) {
	f, err := presetFormatter(preset, useUnabridged)
	if err != nil {
		return nil, err
	}
	return &MeasureValueFormatter{f, logger}, nil
}

func presetFormatter(preset string, useUnabridged bool) (formatter, error) {
	if useUnabridged {
		if preset == "humanize" {
			return &intervalExpFormatter{}, nil
		}
		return newNonFormatter(), nil
	}

	switch strings.ToLower(strings.TrimSpace(preset)) {
	case "none":
		return newNonFormatter(), nil
	case "humanize", "":
		return newPerRangeFormatter(defaultGenericNumOptions())
	case "currency_usd":
		return newPerRangeFormatter(defaultCurrencyOptions(numDollar))
	case "currency_eur":
		return newPerRangeFormatter(defaultCurrencyOptions(numEuro))
	case "percentage":
		return newPerRangeFormatter(defaultPercentOptions())
	case "interval_ms":
		return &intervalFormatter{}, nil
	default:
		return newPerRangeFormatter(defaultGenericNumOptions())
	}
}
