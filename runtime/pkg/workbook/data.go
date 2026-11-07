package workbook

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"
	"unicode/utf8"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/xuri/excelize/v2"
	"google.golang.org/protobuf/types/known/structpb"
)

// Column widths of data sheets, in characters.
const (
	minDataColWidth = 10
	maxDataColWidth = 60
)

// writeDataSheet fills a data sheet: a header row of display names, typed rows with their number formats,
// a frozen header and an Excel table over the data so structured references work.
// The rows are streamed to a temporary file rather than held in excelize's in-memory sheet model,
// so a large table doesn't multiply the runtime's memory use.
func (w *writer) writeDataSheet(name string, ds *runtimev1.WorkbookDataSheet) error {
	headers := uniqueHeaders(ds.Columns)
	// The stream writer needs the column widths before the first row,
	// so the values are converted twice: once to size the columns and once to write them.
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = utf8.RuneCountInString(h)
	}
	for i, r := range ds.Rows {
		for j, col := range ds.Columns {
			cell, err := cellValue(col, rowValue(r, j))
			if err != nil {
				return validationErrorf("data sheet %q, row %d, column %q: %v", ds.Title, i+1, col.Name, err)
			}
			if l := cellWidth(cell); l > widths[j] {
				widths[j] = l
			}
		}
	}

	sw, err := w.f.NewStreamWriter(name)
	if err != nil {
		return err
	}
	styles := make([]int, len(ds.Columns))
	for j, col := range ds.Columns {
		width := math.Min(math.Max(float64(widths[j])+2, minDataColWidth), maxDataColWidth)
		if err := sw.SetColWidth(j+1, j+1, width); err != nil {
			return err
		}
		if styles[j], err = w.styles.columnStyle(col); err != nil {
			return err
		}
	}
	if err := sw.SetPanes(&excelize.Panes{Freeze: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"}); err != nil {
		return err
	}

	row := make([]any, len(ds.Columns))
	for j, h := range headers {
		row[j] = excelize.Cell{StyleID: w.styles.header, Value: h}
	}
	if err := sw.SetRow("A1", row); err != nil {
		return err
	}
	for i, r := range ds.Rows {
		for j, col := range ds.Columns {
			// Validated by the sizing pass above.
			cell, _ := cellValue(col, rowValue(r, j))
			switch {
			case cell == nil:
				row[j] = nil
			case styles[j] != 0:
				row[j] = excelize.Cell{StyleID: styles[j], Value: cell}
			default:
				row[j] = cell
			}
		}
		ref, err := excelize.CoordinatesToCellName(1, i+2)
		if err != nil {
			return err
		}
		if err := sw.SetRow(ref, row); err != nil {
			return err
		}
	}

	// An Excel table needs at least one data row.
	if n := len(ds.Rows); n > 0 {
		lastCell, err := excelize.CoordinatesToCellName(len(ds.Columns), n+1)
		if err != nil {
			return err
		}
		showRowStripes := true
		if err := sw.AddTable(&excelize.Table{
			Range:          "A1:" + lastCell,
			Name:           w.names.tableName(ds.Title),
			StyleName:      "TableStyleMedium2",
			ShowRowStripes: &showRowStripes,
		}); err != nil {
			return err
		}
	}
	return sw.Flush()
}

// rowValue returns the value of a row in the given column, or nil when the row is shorter.
func rowValue(r *structpb.ListValue, j int) *structpb.Value {
	if r == nil || j >= len(r.Values) {
		return nil
	}
	return r.Values[j]
}

// uniqueHeaders returns the display names of the columns, made unique and non-empty: Excel table headers must be both.
func uniqueHeaders(columns []*runtimev1.WorkbookColumn) []string {
	headers := make([]string, len(columns))
	used := make(map[string]bool, len(columns))
	for i, c := range columns {
		h := c.DisplayName
		if h == "" {
			h = c.Name
		}
		if h == "" {
			h = fmt.Sprintf("Column %d", i+1)
		}
		candidate := h
		for n := 2; used[candidate]; n++ {
			candidate = fmt.Sprintf("%s (%d)", h, n)
		}
		used[candidate] = true
		headers[i] = candidate
	}
	return headers
}

// cellValue converts a JSON value to the Go value excelize writes for a column of the given type.
// Null and missing values become empty cells. Values that don't fit the column type are written as text rather than rejected,
// since the data comes from the browser's own rendering of the dashboard.
func cellValue(col *runtimev1.WorkbookColumn, v *structpb.Value) (any, error) {
	if v == nil {
		return nil, nil
	}
	switch k := v.Kind.(type) {
	case *structpb.Value_NullValue:
		return nil, nil
	case *structpb.Value_NumberValue:
		if math.IsNaN(k.NumberValue) || math.IsInf(k.NumberValue, 0) {
			return nil, nil
		}
		switch col.Type {
		case runtimev1.WorkbookColumnType_WORKBOOK_COLUMN_TYPE_STRING:
			return strconv.FormatFloat(k.NumberValue, 'f', -1, 64), nil
		case runtimev1.WorkbookColumnType_WORKBOOK_COLUMN_TYPE_BOOL:
			return k.NumberValue != 0, nil
		default:
			return k.NumberValue, nil
		}
	case *structpb.Value_BoolValue:
		switch col.Type {
		case runtimev1.WorkbookColumnType_WORKBOOK_COLUMN_TYPE_NUMBER:
			if k.BoolValue {
				return 1.0, nil
			}
			return 0.0, nil
		case runtimev1.WorkbookColumnType_WORKBOOK_COLUMN_TYPE_STRING:
			return strconv.FormatBool(k.BoolValue), nil
		default:
			return k.BoolValue, nil
		}
	case *structpb.Value_StringValue:
		switch col.Type {
		case runtimev1.WorkbookColumnType_WORKBOOK_COLUMN_TYPE_NUMBER:
			if f, err := strconv.ParseFloat(k.StringValue, 64); err == nil && !math.IsNaN(f) && !math.IsInf(f, 0) {
				return f, nil
			}
			return k.StringValue, nil
		case runtimev1.WorkbookColumnType_WORKBOOK_COLUMN_TYPE_BOOL:
			if b, err := strconv.ParseBool(k.StringValue); err == nil {
				return b, nil
			}
			return k.StringValue, nil
		case runtimev1.WorkbookColumnType_WORKBOOK_COLUMN_TYPE_DATE, runtimev1.WorkbookColumnType_WORKBOOK_COLUMN_TYPE_TIMESTAMP:
			if t, ok := parseTime(k.StringValue); ok {
				return t, nil
			}
			return k.StringValue, nil
		default:
			return k.StringValue, nil
		}
	case *structpb.Value_ListValue, *structpb.Value_StructValue:
		data, err := json.Marshal(v.AsInterface())
		if err != nil {
			return nil, err
		}
		return string(data), nil
	default:
		return nil, fmt.Errorf("unsupported value %T", v.Kind)
	}
}

// parseTime parses the time formats the browser sends: RFC 3339, a local date-time without a zone, or a plain date.
// The wall-clock time is kept as-is, since Excel has no notion of time zones.
func parseTime(s string) (time.Time, bool) {
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02T15:04:05", "2006-01-02T15:04", time.DateOnly} {
		if t, err := time.Parse(layout, s); err == nil {
			return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC), true
		}
	}
	return time.Time{}, false
}

// cellWidth estimates the width of a value in characters, for sizing columns.
func cellWidth(v any) int {
	switch t := v.(type) {
	case nil:
		return 0
	case string:
		return utf8.RuneCountInString(t)
	case float64:
		// Digits, two decimals, thousands separators and a currency symbol.
		digits := len(strconv.FormatFloat(t, 'f', 2, 64))
		return digits + digits/3 + 1
	case bool:
		return 5
	case time.Time:
		return 19
	default:
		return 10
	}
}
