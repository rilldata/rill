// Package conditionalformat writes the conditional formatting of a dashboard table or pivot into an Excel file
// as Excel's own conditional formats: color scales, data bars and cell rules.
// The formats are described by runtimev1.ExportConditionalFormat, whose colors the browser resolves from the dashboard's theme.
package conditionalformat

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/xuri/excelize/v2"
)

// MaxFormats caps the conditional formats of one export.
// A dashboard has a few per measure, so the cap only stops requests that would make the writer add rules without bound.
const MaxFormats = 1000

var hexColorRegexp = regexp.MustCompile(`^#?[0-9a-fA-F]{6}$`)

// Validate checks that a format is complete, with hex RGB colors and finite numbers.
func Validate(format *runtimev1.ExportConditionalFormat) error {
	switch f := format.GetFormat().(type) {
	case *runtimev1.ExportConditionalFormat_ColorScale:
		if err := validateColor("min_color", f.ColorScale.MinColor); err != nil {
			return err
		}
		if f.ColorScale.MidColor != "" {
			if err := validateColor("mid_color", f.ColorScale.MidColor); err != nil {
				return err
			}
		}
		return validateColor("max_color", f.ColorScale.MaxColor)
	case *runtimev1.ExportConditionalFormat_DataBar:
		return validateColor("color", f.DataBar.Color)
	case *runtimev1.ExportConditionalFormat_CellRule:
		r := f.CellRule
		if r.Operator == runtimev1.ExportCellRuleOperator_EXPORT_CELL_RULE_OPERATOR_UNSPECIFIED {
			return errors.New("cell rule has no operator")
		}
		if math.IsNaN(r.Value) || math.IsInf(r.Value, 0) || math.IsNaN(r.UpperValue) || math.IsInf(r.UpperValue, 0) {
			return errors.New("cell rule values must be finite numbers")
		}
		if err := validateColor("fill_color", r.FillColor); err != nil {
			return err
		}
		if r.TextColor != "" {
			return validateColor("text_color", r.TextColor)
		}
		return nil
	default:
		return errors.New("conditional format has no color scale, data bar or cell rule")
	}
}

func validateColor(field, color string) error {
	if !hexColorRegexp.MatchString(color) {
		return fmt.Errorf("conditional format %s %q is not a hex RGB color like #1a9850", field, color)
	}
	return nil
}

// Apply adds a format to a sheet as an Excel conditional format over the cell ranges, like "B2:B10".
// The ranges must span the same rows and be listed left to right, since a cell rule's formula is written relative to the first cell of the first range.
// A color scale or a data bar spans all the ranges, so it compares their values.
// domain optionally references the cells whose lowest and highest values set the ends of a color scale or data bar,
// for cells that show some of the values of other cells, like a table preview; it defaults to the ranges themselves.
// When several formats cover a cell, the one applied first takes priority.
// A stream writer writes the sheet's conditional formats when it is flushed, so Apply must be called before that.
func Apply(f *excelize.File, sheet string, ranges []string, domain string, format *runtimev1.ExportConditionalFormat) error {
	if len(ranges) == 0 {
		return nil
	}
	if err := Validate(format); err != nil {
		return err
	}
	rangeRef := strings.Join(ranges, " ")

	var opts excelize.ConditionalFormatOptions
	switch t := format.Format.(type) {
	case *runtimev1.ExportConditionalFormat_ColorScale:
		opts = excelize.ConditionalFormatOptions{
			Type:     "2_color_scale",
			Criteria: "=",
			MinType:  "min",
			MaxType:  "max",
			MinColor: t.ColorScale.MinColor,
			MaxColor: t.ColorScale.MaxColor,
		}
		if t.ColorScale.MidColor != "" {
			// The midpoint of the values, not their median, so the colors follow the values linearly as on the dashboard.
			opts.Type = "3_color_scale"
			opts.MidType = "percent"
			opts.MidValue = "50"
			opts.MidColor = t.ColorScale.MidColor
		}
		if domain != "" {
			opts.MinType, opts.MinValue = "formula", fmt.Sprintf("MIN(%s)", domain)
			opts.MaxType, opts.MaxValue = "formula", fmt.Sprintf("MAX(%s)", domain)
			if opts.MidColor != "" {
				opts.MidType, opts.MidValue = "formula", fmt.Sprintf("(MIN(%[1]s)+MAX(%[1]s))/2", domain)
			}
		}
	case *runtimev1.ExportConditionalFormat_DataBar:
		// Bars start at zero, so their lengths are proportional to the values as on the dashboard.
		// A solid fill would need an extension list that stream writers drop, so the bars keep Excel's default gradient.
		opts = excelize.ConditionalFormatOptions{
			Type:     "data_bar",
			Criteria: "=",
			MinType:  "num",
			MinValue: "0",
			MaxType:  "max",
			BarColor: t.DataBar.Color,
		}
		if domain != "" {
			opts.MaxType, opts.MaxValue = "formula", fmt.Sprintf("MAX(%s)", domain)
		}
	case *runtimev1.ExportConditionalFormat_CellRule:
		style := &excelize.Style{Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{t.CellRule.FillColor}}}
		if t.CellRule.TextColor != "" {
			style.Font = &excelize.Font{Color: t.CellRule.TextColor}
		}
		dxf, err := f.NewConditionalStyle(style)
		if err != nil {
			return err
		}
		// A formula rather than a "cell value" rule, because Excel compares an empty cell as zero.
		cell, _, _ := strings.Cut(ranges[0], ":")
		opts = excelize.ConditionalFormatOptions{
			Type:       "formula",
			Criteria:   cellRuleFormula(cell, t.CellRule),
			Format:     &dxf,
			StopIfTrue: true,
		}
	}
	return f.SetConditionalFormat(sheet, rangeRef, []excelize.ConditionalFormatOptions{opts})
}

// cellRuleFormula returns the formula of a cell rule for cell, the first cell it applies to.
func cellRuleFormula(cell string, r *runtimev1.ExportCellRule) string {
	value := formatNumber(r.Value)
	var cond string
	switch r.Operator {
	case runtimev1.ExportCellRuleOperator_EXPORT_CELL_RULE_OPERATOR_GREATER_THAN:
		cond = cell + ">" + value
	case runtimev1.ExportCellRuleOperator_EXPORT_CELL_RULE_OPERATOR_GREATER_THAN_OR_EQUAL:
		cond = cell + ">=" + value
	case runtimev1.ExportCellRuleOperator_EXPORT_CELL_RULE_OPERATOR_LESS_THAN:
		cond = cell + "<" + value
	case runtimev1.ExportCellRuleOperator_EXPORT_CELL_RULE_OPERATOR_LESS_THAN_OR_EQUAL:
		cond = cell + "<=" + value
	case runtimev1.ExportCellRuleOperator_EXPORT_CELL_RULE_OPERATOR_EQUAL:
		cond = cell + "=" + value
	case runtimev1.ExportCellRuleOperator_EXPORT_CELL_RULE_OPERATOR_BETWEEN:
		cond = cell + ">=" + value + "," + cell + "<=" + formatNumber(r.UpperValue)
	}
	return fmt.Sprintf("AND(ISNUMBER(%s),%s)", cell, cond)
}

// formatNumber formats a number for a formula; Excel reads exponents like 1E+21.
func formatNumber(v float64) string {
	return strconv.FormatFloat(v, 'G', -1, 64)
}
