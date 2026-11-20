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
	var domain Domain
	for _, v := range []float64{-150, 20, math.NaN(), 80} {
		domain.Add(v)
	}
	// A color scale over two columns, and two-color without a midpoint; the ends are the domain's, written as numbers.
	require.NoError(t, Apply(f, "Sheet1", []string{"B2:B9", "D2:D9"}, domain, colorScale("#f7fcf5", "#74c476", "#00441b")))
	require.NoError(t, Apply(f, "Sheet1", []string{"F2:F9"}, domain, colorScale("#f7fcf5", "", "#00441b")))
	// The midpoint of ends whose difference is past the largest float doesn't overflow.
	var wide Domain
	wide.Add(-8e307)
	wide.Add(1.6e308)
	require.NoError(t, Apply(f, "Sheet1", []string{"L2:L9"}, wide, colorScale("#f7fcf5", "#74c476", "#00441b")))
	// Rules are written relative to the first cell they cover and stop at the first match.
	require.NoError(t, Apply(f, "Sheet1", []string{"G2:G9", "H2:H9"}, domain, cellRule(runtimev1.ExportCellRuleOperator_EXPORT_CELL_RULE_OPERATOR_BETWEEN, -1.5, 1e21)))
	// Data bars run from zero to the largest absolute value.
	dataBar := &runtimev1.ExportConditionalFormat{Format: &runtimev1.ExportConditionalFormat_DataBar{DataBar: &runtimev1.ExportDataBar{Color: "#638ec6"}}}
	require.NoError(t, Apply(f, "Sheet1", []string{"I2:I9", "J2:J9"}, domain, dataBar))
	// Without numbers, a data bar or color scale has nothing to color.
	require.NoError(t, Apply(f, "Sheet1", []string{"K2:K9"}, Domain{}, dataBar))
	// No ranges, no format.
	require.NoError(t, Apply(f, "Sheet1", nil, domain, colorScale("#f7fcf5", "", "#00441b")))

	formats, err := f.GetConditionalFormats("Sheet1")
	require.NoError(t, err)
	require.Len(t, formats, 5)

	scale := formats["B2:B9 D2:D9"][0]
	require.Equal(t, "3_color_scale", scale.Type)
	require.Equal(t, []string{"num", "num", "num"}, []string{scale.MinType, scale.MidType, scale.MaxType})
	require.Equal(t, []string{"-150", "-35", "80"}, []string{scale.MinValue, scale.MidValue, scale.MaxValue})

	require.Equal(t, "4E+307", formats["L2:L9"][0].MidValue)

	twoColor := formats["F2:F9"][0]
	require.Equal(t, "2_color_scale", twoColor.Type)
	require.Equal(t, []string{"-150", "80"}, []string{twoColor.MinValue, twoColor.MaxValue})

	bar := formats["I2:I9 J2:J9"][0]
	require.Equal(t, "data_bar", bar.Type)
	require.Equal(t, []string{"num", "0", "num", "150"}, []string{bar.MinType, bar.MinValue, bar.MaxType, bar.MaxValue})

	rule := formats["G2:G9 H2:H9"][0]
	require.Equal(t, "formula", rule.Type)
	require.Equal(t, "AND(ISNUMBER(G2),G2>=-1.5,G2<=1E+21)", rule.Criteria)
	require.True(t, rule.StopIfTrue)
	style, err := f.GetConditionalStyle(*rule.Format)
	require.NoError(t, err)
	require.Equal(t, []string{"FFC7CE"}, style.Fill.Color)
	require.Equal(t, "9C0D07", style.Font.Color)
}
