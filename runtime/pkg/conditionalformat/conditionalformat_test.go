package conditionalformat

import (
	"math"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func colorScale(minColor, midColor, maxColor string) *runtimev1.ExportConditionalFormat {
	return &runtimev1.ExportConditionalFormat{Format: &runtimev1.ExportConditionalFormat_ColorScale{
		ColorScale: &runtimev1.ExportColorScale{MinColor: minColor, MidColor: midColor, MaxColor: maxColor},
	}}
}

func cellRule(op runtimev1.ExportCellRuleOperator, value, upper float64) *runtimev1.ExportConditionalFormat {
	return &runtimev1.ExportConditionalFormat{Format: &runtimev1.ExportConditionalFormat_CellRule{
		CellRule: &runtimev1.ExportCellRule{Operator: op, Value: value, UpperValue: upper, FillColor: "#ffc7ce", TextColor: "#9c0d07"},
	}}
}

func TestValidate(t *testing.T) {
	cases := []struct {
		name   string
		format *runtimev1.ExportConditionalFormat
		want   string
	}{
		{name: "two-color scale", format: colorScale("#ffffff", "", "1a9850")},
		{name: "empty", format: &runtimev1.ExportConditionalFormat{}, want: "has no color scale, data bar or cell rule"},
		{name: "named color", format: colorScale("white", "", "#1a9850"), want: `min_color "white" is not a hex RGB color`},
		{name: "short hex color", format: colorScale("#fff", "", "#1a9850"), want: `min_color "#fff" is not a hex RGB color`},
		{name: "rule without operator", format: cellRule(runtimev1.ExportCellRuleOperator_EXPORT_CELL_RULE_OPERATOR_UNSPECIFIED, 1, 0), want: "no operator"},
		{name: "rule with infinite value", format: cellRule(runtimev1.ExportCellRuleOperator_EXPORT_CELL_RULE_OPERATOR_LESS_THAN, math.Inf(1), 0), want: "finite"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(tc.format)
			if tc.want == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.want)
			}
		})
	}
}

func TestApply(t *testing.T) {
	f := excelize.NewFile()
	defer f.Close()
	// A color scale over two columns, scaled to cells on another sheet, and two-color without a midpoint.
	require.NoError(t, Apply(f, "Sheet1", []string{"B2:B9", "D2:D9"}, "'Data'!$B$2:$B$50", colorScale("#f7fcf5", "#74c476", "#00441b")))
	require.NoError(t, Apply(f, "Sheet1", []string{"F2:F9"}, "", colorScale("#f7fcf5", "", "#00441b")))
	// Rules are written relative to the first cell they cover and stop at the first match.
	require.NoError(t, Apply(f, "Sheet1", []string{"G2:G9", "H2:H9"}, "", cellRule(runtimev1.ExportCellRuleOperator_EXPORT_CELL_RULE_OPERATOR_BETWEEN, -1.5, 1e21)))
	// No ranges, no format.
	require.NoError(t, Apply(f, "Sheet1", nil, "", colorScale("#f7fcf5", "", "#00441b")))

	formats, err := f.GetConditionalFormats("Sheet1")
	require.NoError(t, err)
	require.Len(t, formats, 3)

	scale := formats["B2:B9 D2:D9"][0]
	require.Equal(t, "3_color_scale", scale.Type)
	require.Equal(t, "MIN('Data'!$B$2:$B$50)", scale.MinValue)
	require.Equal(t, "(MIN('Data'!$B$2:$B$50)+MAX('Data'!$B$2:$B$50))/2", scale.MidValue)
	require.Equal(t, "MAX('Data'!$B$2:$B$50)", scale.MaxValue)

	twoColor := formats["F2:F9"][0]
	require.Equal(t, "2_color_scale", twoColor.Type)
	require.Equal(t, "min", twoColor.MinType)
	require.Equal(t, "max", twoColor.MaxType)

	rule := formats["G2:G9 H2:H9"][0]
	require.Equal(t, "formula", rule.Type)
	require.Equal(t, "AND(ISNUMBER(G2),G2>=-1.5,G2<=1E+21)", rule.Criteria)
	require.True(t, rule.StopIfTrue)
	style, err := f.GetConditionalStyle(*rule.Format)
	require.NoError(t, err)
	require.Equal(t, []string{"FFC7CE"}, style.Fill.Color)
	require.Equal(t, "9C0D07", style.Font.Color)
}
