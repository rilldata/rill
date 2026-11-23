package driverutil

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// The cases follow web-common/src/features/exports/xlsx/number-format.spec.ts, so both writers format measures alike.
// Where the browser looks at a column's values, the export looks at the column's type (integers or not) or at each value.
func TestExcelNumberFormat(t *testing.T) {
	tests := []struct {
		name     string
		format   MeasureFormat
		integers bool
		want     string
		percent  bool
	}{
		{"currency usd", MeasureFormat{Preset: "currency_usd"}, false, `"$"#,##0.00`, false},
		{"currency eur", MeasureFormat{Preset: "currency_eur"}, false, `"€"#,##0.00`, false},
		{"percentage", MeasureFormat{Preset: "percentage"}, false, "", true},
		{"none", MeasureFormat{Preset: "none"}, false, "", false},
		{"interval", MeasureFormat{Preset: "interval_ms"}, true, "", false},

		// Humanized measures are written in full: a grouped integer or decimal format, depending on the column.
		{"humanize integers", MeasureFormat{Preset: "humanize"}, true, "#,##0", false},
		{"humanize decimals", MeasureFormat{Preset: "humanize"}, false, "#,##0.00", false},
		{"no format", MeasureFormat{}, true, "#,##0", false},

		{"d3", MeasureFormat{D3: "$,.2f"}, false, `"$"#,##0.00`, false},
		{"d3 locale prefix", MeasureFormat{D3: "$,.2f", D3Currency: &[2]string{"£", ""}}, false, `"£"#,##0.00`, false},
		{"d3 locale suffix", MeasureFormat{D3: "$,.0f", D3Currency: &[2]string{"", " €"}}, false, `#,##0" €"`, false},
		// A locale without a currency keeps d3's dollar.
		{"d3 locale without currency", MeasureFormat{D3: "$,.2f"}, false, `"$"#,##0.00`, false},
		// A specifier Excel can't express falls back to the column's grouped format.
		{"d3 overflowing precision", MeasureFormat{D3: ".99999999999999999999f"}, false, "#,##0.00", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format, percent := excelNumberFormat(tt.format, tt.integers)
			require.Equal(t, tt.want, format)
			require.Equal(t, tt.percent, percent)
		})
	}
}

// Small percentages keep two significant digits, with one to four decimals.
func TestPercentNumberFormat(t *testing.T) {
	tests := []struct {
		value float64
		want  string
	}{
		{0.56, "0.0%"},
		{0.066, "0.0%"},
		{0.012, "0.0%"},
		{0.005258, "0.00%"},
		{-0.0067, "0.00%"},
		{0.0007, "0.000%"},
		{0.0000001, "0.0000%"},
		{0, "0.0%"},
	}
	for _, tt := range tests {
		require.Equal(t, tt.want, percentNumberFormat(tt.value), "%v", tt.value)
	}
}

func TestD3ToExcelFormat(t *testing.T) {
	tests := []struct {
		spec     string
		integers bool
		want     string
		ok       bool
	}{
		{",", true, "#,##0", true},
		{",", false, "#,##0.00", true},
		{".0f", false, "0", true},
		{",.1f", false, "#,##0.0", true},
		{".2f", false, "0.00", true},
		{"$,.2f", false, `"$"#,##0.00`, true},
		{".1%", false, "0.0%", true},
		{"%", false, "0%", true},
		{",d", false, "#,##0", true},
		{".3s", true, "0.000", true},
		// d3 clamps the precision to 20.
		{".25f", false, "0." + strings.Repeat("0", 20), true},
		// Specifiers it can't express.
		{".99999999999999999999f", false, "", false},
		{"not a format", false, "", false},
		{".2b", false, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.spec, func(t *testing.T) {
			got, ok := d3ToExcelFormat(tt.spec, tt.integers, nil)
			require.Equal(t, tt.ok, ok)
			require.Equal(t, tt.want, got)
		})
	}
}
