package driverutil

import (
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Excel number formats for measures without a more specific one.
const (
	xlsxIntegerFormat = "#,##0"
	xlsxDecimalFormat = "#,##0.00"
)

// d3MaxPrecision is the largest precision d3-format accepts; it clamps larger ones to it.
const d3MaxPrecision = 20

// d3Specifier matches a d3-format specifier: [[fill]align][sign][symbol][0][width][,][.precision][~][type].
var d3Specifier = regexp.MustCompile(`(?i)^(?:(.)?([<>=^]))?([+\-( ])?([$#])?(0)?(\d+)?(,)?(\.\d+)?(~)?([a-z%])?$`)

// excelNumberFormat returns the Excel number format of a measure's values in a column that holds integers or not.
// Presets map directly; d3 format strings are translated where Excel has an equivalent;
// anything else, humanized measures included, falls back to a grouped integer or two-decimal format depending on the column,
// since an abbreviated number can't be summed or filtered.
// An empty format means none (Excel's "General").
// percent reports a percentage, whose format is picked per value by percentNumberFormat.
// It follows the browser's rules for canvas workbooks (web-common/src/features/exports/xlsx/number-format.ts),
// except that the browser looks at all of a column's values, which an export streams instead.
func excelNumberFormat(m MeasureFormat, integers bool) (format string, percent bool) {
	switch m.Preset {
	case "currency_usd":
		return `"$"#,##0.00`, false
	case "currency_eur":
		return `"€"#,##0.00`, false
	case "percentage":
		return "", true
	case "interval_ms", "none":
		return "", false
	}
	if m.D3 != "" {
		if format, ok := d3ToExcelFormat(m.D3, integers, m.D3Currency); ok {
			return format, false
		}
	}
	if integers {
		return xlsxIntegerFormat, false
	}
	return xlsxDecimalFormat, false
}

// percentNumberFormat returns a percentage format with enough decimals to show v with two significant digits,
// as the dashboard's humanized percentages do (0.53%, 6.6%), keeping between one and four decimals.
// A fixed "0.0%" would turn a 0.53% click-through rate into 0.5% and a 0.03% one into 0.0%.
func percentNumberFormat(v float64) string {
	digits := 1
	if v != 0 && !math.IsNaN(v) && !math.IsInf(v, 0) {
		digits = 1 - int(math.Floor(math.Log10(math.Abs(v)*100)))
	}
	return "0." + strings.Repeat("0", min(4, max(1, digits))) + "%"
}

// d3ToExcelFormat translates a d3 format specifier to an Excel number format, if Excel has an equivalent.
// integers tells whether the column holds integers, which a specifier without a precision shows without decimals.
// currency is the prefix and suffix that a "$" in the specifier stands for; when nil, it is d3's default dollar prefix.
func d3ToExcelFormat(spec string, integers bool, currency *[2]string) (string, bool) {
	match := d3Specifier.FindStringSubmatch(strings.TrimSpace(spec))
	if match == nil {
		return "", false
	}
	symbol, comma, precision, typ := match[4], match[7], match[8], match[10]
	digits := -1
	if precision != "" {
		n, err := strconv.Atoi(precision[1:])
		if err != nil {
			return "", false
		}
		digits = min(n, d3MaxPrecision)
	}
	grouped := "0"
	if comma != "" {
		grouped = "#,##0"
	}
	decimals := func(defaultCount int) string {
		count := digits
		if count < 0 {
			count = defaultCount
		}
		if count == 0 {
			return ""
		}
		return "." + strings.Repeat("0", count)
	}

	var format string
	switch strings.ToLower(typ) {
	case "%":
		format = "0" + decimals(0) + "%"
	case "d":
		format = grouped
	case "f":
		format = grouped + decimals(2)
	case "s", "e", "g", "r", "":
		// SI prefixes and significant digits have no Excel equivalent;
		// keep the grouping and show decimals only when the column needs them.
		if integers {
			format = grouped + decimals(0)
		} else {
			format = grouped + decimals(2)
		}
	default:
		return "", false
	}
	if symbol != "$" {
		return format, true
	}
	if currency == nil {
		currency = &[2]string{"$", ""}
	}
	quoted := func(text string) string {
		if text == "" {
			return ""
		}
		return `"` + strings.ReplaceAll(text, `"`, "") + `"`
	}
	return quoted(currency[0]) + format + quoted(currency[1]), true
}
