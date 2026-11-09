package workbook

import (
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/xuri/excelize/v2"
)

// Default number formats for typed columns that don't carry their own.
const (
	defaultDateFormat      = "yyyy-mm-dd"
	defaultTimestampFormat = "yyyy-mm-dd hh:mm:ss"
	percentFormat          = "0.0%"
)

const (
	mutedColor = "#6B7280"
	linkColor  = "#2563EB"
	// The table previews on dashboard sheets match the data sheets' table style
	// (TableStyleMedium2: a blue header with white text and light-blue banded rows).
	tableHeaderFill = "#4472C4"
	tableBandFill   = "#D9E1F2"
)

// styleSet holds the cell styles a workbook uses. Styles are registered once and shared by every sheet.
type styleSet struct {
	f *excelize.File
	// Dashboard chrome.
	title    int
	subtitle int
	muted    int
	note     int
	link     int
	label    int
	wrap     int
	// Data sheets and table previews.
	header        int
	previewHeader int
	// Styles that carry a number format, cached by format code (and banding for previews).
	numFmt        map[string]int
	totals        map[string]int
	kpiValue      map[string]int
	kpiNumber     map[string]int
	preview       map[string]int
	previewTotals map[string]int
}

func newStyleSet(f *excelize.File) (*styleSet, error) {
	s := &styleSet{
		f:             f,
		numFmt:        map[string]int{},
		totals:        map[string]int{},
		kpiValue:      map[string]int{},
		kpiNumber:     map[string]int{},
		preview:       map[string]int{},
		previewTotals: map[string]int{},
	}
	var err error
	if s.title, err = f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 16}}); err != nil {
		return nil, err
	}
	if s.subtitle, err = f.NewStyle(&excelize.Style{Font: &excelize.Font{Size: 11, Color: mutedColor}}); err != nil {
		return nil, err
	}
	if s.muted, err = f.NewStyle(&excelize.Style{Font: &excelize.Font{Size: 10, Color: mutedColor}}); err != nil {
		return nil, err
	}
	if s.note, err = f.NewStyle(&excelize.Style{Font: &excelize.Font{Size: 10, Italic: true, Color: mutedColor}}); err != nil {
		return nil, err
	}
	if s.link, err = f.NewStyle(&excelize.Style{Font: &excelize.Font{Size: 10, Color: linkColor, Underline: "single"}}); err != nil {
		return nil, err
	}
	if s.label, err = f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 11}}); err != nil {
		return nil, err
	}
	if s.wrap, err = f.NewStyle(&excelize.Style{Alignment: &excelize.Alignment{WrapText: true, Vertical: "top"}}); err != nil {
		return nil, err
	}
	// Data sheet headers are only bold: the Excel table laid over the sheet supplies the fill.
	if s.header, err = f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}}); err != nil {
		return nil, err
	}
	if s.previewHeader, err = f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{tableHeaderFill}},
	}); err != nil {
		return nil, err
	}
	return s, nil
}

// numFmtStyle returns a plain style with the given number format, or 0 (no style) for an empty format.
func (s *styleSet) numFmtStyle(format string) (int, error) {
	if format == "" {
		return 0, nil
	}
	if id, ok := s.numFmt[format]; ok {
		return id, nil
	}
	id, err := s.f.NewStyle(&excelize.Style{CustomNumFmt: &format})
	if err != nil {
		return 0, err
	}
	s.numFmt[format] = id
	return id, nil
}

// totalsStyle returns the style of a data sheet's totals row: bold, with a rule above it, and the given number format.
func (s *styleSet) totalsStyle(format string) (int, error) {
	return s.cachedStyle(s.totals, format, format, &excelize.Style{
		Font:   &excelize.Font{Bold: true},
		Border: []excelize.Border{{Type: "top", Color: tableHeaderFill, Style: 1}},
	})
}

// kpiValueStyle returns the large bold style of a KPI value with the given number format.
// KPI numbers are left-aligned under their labels rather than right-aligned like a column of figures.
func (s *styleSet) kpiValueStyle(format string) (int, error) {
	return s.cachedStyle(s.kpiValue, format, format, &excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
}

// kpiNumberStyle returns the style of a KPI's comparison numbers with the given number format.
func (s *styleSet) kpiNumberStyle(format string) (int, error) {
	return s.cachedStyle(s.kpiNumber, format, format, &excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "left"},
	})
}

// cachedStyle registers style with the given number format once per key.
func (s *styleSet) cachedStyle(cache map[string]int, key, format string, style *excelize.Style) (int, error) {
	if id, ok := cache[key]; ok {
		return id, nil
	}
	if format != "" {
		style.CustomNumFmt = &format
	}
	id, err := s.f.NewStyle(style)
	if err != nil {
		return 0, err
	}
	cache[key] = id
	return id, nil
}

// previewStyle returns the style of a column's cells in a table preview on a dashboard sheet:
// numbers are shown compactly and every other row is banded like the data sheets' tables.
func (s *styleSet) previewStyle(col *runtimev1.WorkbookColumn, banded bool) (int, error) {
	format := compactNumberFormat(columnFormat(col))
	key := format
	style := &excelize.Style{}
	if banded {
		key += "|banded"
		style.Fill = excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{tableBandFill}}
	}
	return s.cachedStyle(s.preview, key, format, style)
}

// previewTotalsStyle returns the style of a column's cell in the totals row of a table preview: bold, compact numbers.
func (s *styleSet) previewTotalsStyle(col *runtimev1.WorkbookColumn) (int, error) {
	format := compactNumberFormat(columnFormat(col))
	return s.cachedStyle(s.previewTotals, format, format, &excelize.Style{Font: &excelize.Font{Bold: true}})
}

// compactNumberFormat abbreviates thousands and millions ("1.5M", "$347.1k") the way the dashboard does,
// so values fit the narrow cells of a dashboard sheet. Percentages, dates and unknown formats are kept.
// Only positive values are abbreviated: Excel number formats allow two conditions, and large negatives are rare.
func compactNumberFormat(format string) string {
	if format == "" || strings.Contains(format, "%") {
		return format
	}
	// A quoted currency symbol before or after the number, e.g. "$" in "$"#,##0.00 or "€" in #,##0.00"€".
	prefix, suffix := "", ""
	if strings.HasPrefix(format, `"`) {
		if end := strings.Index(format[1:], `"`); end >= 0 {
			prefix = format[:end+2]
		}
	}
	base := strings.TrimPrefix(format, prefix)
	if strings.HasSuffix(base, `"`) {
		if start := strings.LastIndex(strings.TrimSuffix(base, `"`), `"`); start >= 0 {
			suffix = base[start:]
			base = base[:start]
		}
	}
	if base != "General" && !strings.HasPrefix(base, "#") && !strings.HasPrefix(base, "0") {
		return format
	}
	return fmt.Sprintf(`[>=1000000]%s#,##0.0,,"M"%s;[>=1000]%s#,##0.0,"k"%s;%s`, prefix, suffix, prefix, suffix, format)
}

// columnFormat returns the number format of a column, falling back to a default for dates and timestamps.
func columnFormat(col *runtimev1.WorkbookColumn) string {
	if col.NumberFormat != "" {
		return col.NumberFormat
	}
	switch col.Type {
	case runtimev1.WorkbookColumnType_WORKBOOK_COLUMN_TYPE_DATE:
		return defaultDateFormat
	case runtimev1.WorkbookColumnType_WORKBOOK_COLUMN_TYPE_TIMESTAMP:
		return defaultTimestampFormat
	default:
		return ""
	}
}
