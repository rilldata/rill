package driverutil

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
	"unicode/utf8"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/conditionalformat"
	"github.com/rilldata/rill/runtime/pkg/jsonval"
	"github.com/rilldata/rill/runtime/pkg/xlsxutil"
	"github.com/xuri/excelize/v2"
)

const (
	// xlsxSheet is the sheet writeXLSX writes to.
	xlsxSheet = "Sheet1"
	// writeXLSX reads ahead the first xlsxReadAheadRows rows to size the columns,
	// or fewer if their values add up to about xlsxReadAheadBytes.
	xlsxReadAheadRows  = 100
	xlsxReadAheadBytes = 256 << 10
	// Excel's built-in date and date-time formats, which follow the reader's locale.
	xlsxDateNumFmt     = 14
	xlsxDateTimeNumFmt = 22
	// The widths, in characters, that typical numbers, dates and dates with times need in their formats.
	xlsxNumberWidth   = 12
	xlsxDateWidth     = 10
	xlsxDateTimeWidth = 16
	// The color of the header lines below the title.
	xlsxMutedColor = "#6B7280"
)

// xlsxMinTime is the earliest date Excel can hold; earlier dates are written as text.
var xlsxMinTime = time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC)

// XLSXConditionalFormat is a conditional format that an XLSX file applies to the data cells of some of its columns.
type XLSXConditionalFormat struct {
	// Column is the 0-based index of the first column the format applies to.
	Column int
	// Stride, if positive, also applies the format to every Stride-th column after Column,
	// like the columns of a measure in a pivot, which repeat once per pivoted value.
	Stride int
	Format *runtimev1.ExportConditionalFormat
}

// XLSXNumberFormat formats the numbers in some of the data columns of an XLSX file like a measure.
// Columns without one take no number format.
type XLSXNumberFormat struct {
	// Column is the 0-based index of the first column the format applies to.
	Column int
	// Stride, if positive, also applies the format to every Stride-th column after Column, as in XLSXConditionalFormat.
	Stride int
	Format MeasureFormat
}

// MeasureFormat is the format of a measure's numbers:
// a metrics view measure's format_preset, format_d3 and the currency of its format_d3_locale.
// The zero MeasureFormat formats numbers as grouped integers or decimals, depending on their column's type.
// Only integer types count as integers: a DECIMAL column always shows decimals, even a DECIMAL(18,0),
// since the column's type doesn't carry its scale.
type MeasureFormat struct {
	// Preset is a format preset, like "currency_usd" or "percentage".
	Preset string
	// D3 is a d3-format specifier, used unless Preset sets the format.
	D3 string
	// D3Currency is the prefix and suffix that a "$" in D3 stands for; nil means d3's default "$" prefix.
	D3Currency *[2]string
}

// XLSXTimeColumn sets how a timestamp column of an XLSX file is shown, like a time dimension truncated to a time grain.
// A column truncated to a day or a coarser grain shows the day each timestamp falls on in the query's time zone;
// a finer grain shows the timestamp as a UTC date and time, as CSV files do.
// Timestamp columns without one show each value as a date if it is at midnight UTC, and as a date and time otherwise.
type XLSXTimeColumn struct {
	// Column is the 0-based index of the column.
	Column int
	// Date shows the column as dates, without a time of day, like a time dimension truncated to a day or a coarser grain.
	// Otherwise the column shows dates and times.
	Date bool
	// TimeZone is the time zone the dates are days in, like the time zone of the query that truncated the timestamps.
	// Empty means UTC.
	TimeZone string
}

// writeXLSX writes a result as a formatted sheet:
// the header lines, with a bold title; a bold and frozen row of column names;
// and the data as an Excel table, with measures in their number formats, dates in the reader's short date format,
// and columns sized to the first rows.
// Formats depend only on the options, the column types and each value, never on other rows,
// since the rows are streamed.
func writeXLSX(res *drivers.Result, fw io.Writer, opts FileOptions) error {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	sw, err := f.NewStreamWriter(xlsxSheet)
	if err != nil {
		return err
	}
	w := &xlsxWriter{
		res:     res,
		opts:    opts,
		f:       f,
		sw:      sw,
		columns: make([]xlsxColumn, len(res.Schema.Fields)),
		vals:    make([]any, len(res.Schema.Fields)),
		styles:  make(map[xlsxNumFmt]int),
	}
	for i := range w.vals {
		w.vals[i] = new(any)
	}

	if err := w.setColumns(); err != nil {
		return err
	}
	sample, more, err := w.readAhead()
	if err != nil {
		return err
	}
	if err := w.writeHeader(sample); err != nil {
		return err
	}
	for _, values := range sample {
		if err := w.writeRow(values); err != nil {
			return err
		}
	}
	values := make([]any, len(res.Schema.Fields))
	for more && res.Next() {
		if err := w.scan(values); err != nil {
			return err
		}
		if err := w.writeRow(values); err != nil {
			return err
		}
	}
	if res.Err() != nil {
		return res.Err()
	}
	if err := w.finish(); err != nil {
		return err
	}
	return f.Write(fw)
}

// xlsxWriter writes a result to a sheet with excelize's stream writer.
type xlsxWriter struct {
	res     *drivers.Result
	opts    FileOptions
	f       *excelize.File
	sw      *excelize.StreamWriter
	columns []xlsxColumn
	// vals are the destinations of res.Scan.
	vals []any
	// styles caches the number format styles registered with f.
	styles map[xlsxNumFmt]int
	// headerRow is the row of column names, and nextRow the row writeRow writes next (1-based).
	headerRow int
	nextRow   int
	// The columns of each conditional format, and the range of the numbers in them, which sets the ends of its color scale or data bar.
	formatColumns [][]int
	domains       []conditionalformat.Domain
}

// xlsxNumFmt is a number format: an Excel built-in format ID, or a custom format if set.
type xlsxNumFmt struct {
	id     int
	custom string
}

// xlsxColumn is how the values of a column are written.
type xlsxColumn struct {
	// numFmt is the number format of the column's values, if any.
	numFmt xlsxNumFmt
	// percent picks the number format of each value for its decimals (see percentNumberFormat).
	percent bool
	// timestamps picks a date or a date and time format for each timestamp, for timestamps without a time grain.
	timestamps bool
	// dateIn writes each timestamp as a date: the day it falls on in this location.
	dateIn *time.Location
	// minWidth is the width, in characters, the column's values need at least.
	minWidth int
}

// setColumns sets how each column is written, from its type and the options.
func (w *xlsxWriter) setColumns() error {
	fields := w.res.Schema.Fields
	for i, f := range fields {
		switch f.Type.GetCode() {
		case runtimev1.Type_CODE_DATE:
			w.columns[i].numFmt.id = xlsxDateNumFmt
			w.columns[i].minWidth = xlsxDateWidth
		case runtimev1.Type_CODE_TIMESTAMP:
			w.columns[i].timestamps = true
			w.columns[i].minWidth = xlsxDateTimeWidth
		}
	}

	for _, tc := range w.opts.TimeColumns {
		if tc.Column < 0 || tc.Column >= len(fields) || fields[tc.Column].Type.GetCode() != runtimev1.Type_CODE_TIMESTAMP {
			continue
		}
		c := &w.columns[tc.Column]
		c.timestamps = false
		if !tc.Date {
			c.numFmt.id = xlsxDateTimeNumFmt
			continue
		}
		loc, err := time.LoadLocation(tc.TimeZone)
		if err != nil {
			return err
		}
		c.numFmt.id = xlsxDateNumFmt
		c.dateIn = loc
		c.minWidth = xlsxDateWidth
	}

	for _, nf := range w.opts.NumberFormats {
		for _, col := range xlsxColumns(nf.Column, nf.Stride, len(fields)) {
			c := &w.columns[col]
			var integers bool
			switch fields[col].Type.GetCode() {
			case runtimev1.Type_CODE_INT8, runtimev1.Type_CODE_INT16, runtimev1.Type_CODE_INT32, runtimev1.Type_CODE_INT64, runtimev1.Type_CODE_INT128, runtimev1.Type_CODE_INT256,
				runtimev1.Type_CODE_UINT8, runtimev1.Type_CODE_UINT16, runtimev1.Type_CODE_UINT32, runtimev1.Type_CODE_UINT64, runtimev1.Type_CODE_UINT128, runtimev1.Type_CODE_UINT256:
				integers = true
			}
			c.numFmt.custom, c.percent = excelNumberFormat(nf.Format, integers)
			c.minWidth = xlsxNumberWidth
		}
	}

	w.formatColumns = make([][]int, len(w.opts.ConditionalFormats))
	w.domains = make([]conditionalformat.Domain, len(w.opts.ConditionalFormats))
	for i, cf := range w.opts.ConditionalFormats {
		w.formatColumns[i] = xlsxColumns(cf.Column, cf.Stride, len(fields))
	}
	return nil
}

// readAhead reads the first rows, which size the columns.
// more reports whether the result may have more rows.
func (w *xlsxWriter) readAhead() (sample [][]any, more bool, err error) {
	size := 0
	for len(sample) < xlsxReadAheadRows && size < xlsxReadAheadBytes {
		if !w.res.Next() {
			return sample, false, w.res.Err()
		}
		values := make([]any, len(w.columns))
		if err := w.scan(values); err != nil {
			return nil, false, err
		}
		for _, v := range values {
			if s, ok := v.(string); ok {
				size += len(s)
			} else {
				size += 8
			}
		}
		sample = append(sample, values)
	}
	return sample, true, nil
}

// writeHeader sizes the columns, freezes the rows above the data and writes the header lines and the column names.
// The stream writer needs the column widths and the panes before the first row.
func (w *xlsxWriter) writeHeader(sample [][]any) error {
	names := make([]string, len(w.columns))
	for i, f := range w.res.Schema.Fields {
		names[i] = f.Name
	}
	names = xlsxutil.UniqueNames(names)

	for i, c := range w.columns {
		// Two more characters for the table's filter button.
		width := max(utf8.RuneCountInString(names[i])+2, c.minWidth)
		for _, values := range sample {
			// Dates and times are as wide as their format.
			if _, ok := values[i].(time.Time); !ok {
				width = max(width, xlsxutil.CellWidth(values[i]))
			}
		}
		// The header lines are left out: they spill over the empty cells to their right.
		if err := w.sw.SetColWidth(i+1, i+1, xlsxutil.ColumnWidth(width)); err != nil {
			return err
		}
	}

	w.headerRow = len(w.opts.Headers) + 1
	w.nextRow = w.headerRow + 1
	topLeft, err := excelize.CoordinatesToCellName(1, w.nextRow)
	if err != nil {
		return err
	}
	if err := w.sw.SetPanes(&excelize.Panes{Freeze: true, YSplit: w.headerRow, TopLeftCell: topLeft, ActivePane: "bottomLeft"}); err != nil {
		return err
	}

	titleStyle, err := w.f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 12}})
	if err != nil {
		return err
	}
	mutedStyle, err := w.f.NewStyle(&excelize.Style{Font: &excelize.Font{Color: xlsxMutedColor}})
	if err != nil {
		return err
	}
	for i, line := range w.opts.Headers {
		if line == "" {
			continue
		}
		id := mutedStyle
		if i == 0 {
			id = titleStyle
		}
		cell, err := excelize.CoordinatesToCellName(1, i+1)
		if err != nil {
			return err
		}
		if err := w.sw.SetRow(cell, []any{excelize.Cell{StyleID: id, Value: line}}); err != nil {
			return err
		}
	}

	// The table laid over the data fills the column names; they are only bold.
	id, err := w.f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	if err != nil {
		return err
	}
	row := make([]any, len(names))
	for i, name := range names {
		row[i] = excelize.Cell{StyleID: id, Value: name}
	}
	cell, err := excelize.CoordinatesToCellName(1, w.headerRow)
	if err != nil {
		return err
	}
	return w.sw.SetRow(cell, row, excelize.RowOpts{Height: 25})
}

// writeRow writes a row of values, each in its column's format.
func (w *xlsxWriter) writeRow(values []any) error {
	for i, cols := range w.formatColumns {
		if w.opts.ConditionalFormats[i].Format.GetCellRule() != nil {
			// Apply doesn't read a cell rule's domain.
			continue
		}
		for _, col := range cols {
			if v, ok := xlsxNumber(values[col]); ok {
				w.domains[i].Add(v)
			}
		}
	}

	row := make([]any, len(values))
	for i, v := range values {
		row[i] = v
		c := w.columns[i]
		numFmt := c.numFmt
		switch v := v.(type) {
		case nil, string:
			// Text, like a date Excel can't hold, keeps no format.
			continue
		case time.Time:
			if c.timestamps {
				numFmt.id = xlsxDateTimeNumFmt
				if v.Hour() == 0 && v.Minute() == 0 && v.Second() == 0 && v.Nanosecond() == 0 {
					numFmt.id = xlsxDateNumFmt
				}
			}
		default:
			if n, ok := xlsxNumber(v); ok && c.percent {
				numFmt.custom = percentNumberFormat(n)
			}
		}
		if numFmt == (xlsxNumFmt{}) {
			continue
		}
		id, err := w.style(numFmt)
		if err != nil {
			return err
		}
		row[i] = excelize.Cell{StyleID: id, Value: v}
	}

	cell, err := excelize.CoordinatesToCellName(1, w.nextRow)
	if err != nil {
		return err
	}
	if err := w.sw.SetRow(cell, row); err != nil {
		return err
	}
	w.nextRow++
	return nil
}

// finish applies the conditional formats, lays a table over the data and flushes the sheet.
func (w *xlsxWriter) finish() error {
	firstDataRow, lastDataRow := w.headerRow+1, w.nextRow-1
	if lastDataRow >= firstDataRow && len(w.columns) > 0 {
		// The stream writer writes the sheet's conditional formats when it is flushed.
		for i, cf := range w.opts.ConditionalFormats {
			ranges := make([]string, len(w.formatColumns[i]))
			for j, col := range w.formatColumns[i] {
				name, err := excelize.ColumnNumberToName(col + 1)
				if err != nil {
					return err
				}
				ranges[j] = fmt.Sprintf("%s%d:%s%d", name, firstDataRow, name, lastDataRow)
			}
			if err := conditionalformat.Apply(w.f, xlsxSheet, ranges, w.domains[i], cf.Format); err != nil {
				return err
			}
		}

		// An Excel table needs at least one data row.
		firstCell, err := excelize.CoordinatesToCellName(1, w.headerRow)
		if err != nil {
			return err
		}
		lastCell, err := excelize.CoordinatesToCellName(len(w.columns), lastDataRow)
		if err != nil {
			return err
		}
		showRowStripes := true
		if err := w.sw.AddTable(&excelize.Table{
			Range:          firstCell + ":" + lastCell,
			StyleName:      "TableStyleMedium2",
			ShowRowStripes: &showRowStripes,
		}); err != nil {
			return err
		}
	}
	return w.sw.Flush()
}

// style returns the ID of a style with the given number format, registering it the first time.
func (w *xlsxWriter) style(numFmt xlsxNumFmt) (int, error) {
	if id, ok := w.styles[numFmt]; ok {
		return id, nil
	}
	style := &excelize.Style{NumFmt: numFmt.id}
	if numFmt.custom != "" {
		style.CustomNumFmt = &numFmt.custom
	}
	id, err := w.f.NewStyle(style)
	if err != nil {
		return 0, err
	}
	w.styles[numFmt] = id
	return id, nil
}

// scan scans the next row of the result into dst, converted to the values written to the sheet.
func (w *xlsxWriter) scan(dst []any) error {
	if err := w.res.Scan(w.vals...); err != nil {
		return err
	}
	for i, v := range w.vals {
		t := w.res.Schema.Fields[i].Type
		jval, err := jsonval.ToValue(*(v.(*any)), t)
		if err != nil {
			return fmt.Errorf("failed to convert to JSON value: %w", err)
		}

		var s any
		switch tval := jval.(type) {
		case nil:
		case time.Time:
			// jsonval.ToValue turns DATE values into text, handled below.
			switch {
			case t.GetCode() == runtimev1.Type_CODE_TIME:
				s = tval.In(time.UTC).Format(time.TimeOnly)
			case w.columns[i].dateIn != nil:
				local := tval.In(w.columns[i].dateIn)
				s = xlsxTime(time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC), true)
			default:
				s = xlsxTime(tval, false)
			}
		case string:
			s = tval
			// Dates arrive as text; written as dates, Excel shows them in the reader's format and sorts them.
			if t.GetCode() == runtimev1.Type_CODE_DATE {
				if d, err := time.Parse(time.DateOnly, tval); err == nil {
					s = xlsxTime(d, true)
				}
			}
		case []any, map[string]any:
			mres, err := json.Marshal(tval)
			if err != nil {
				return fmt.Errorf("failed to marshal JSON value: %w", err)
			}
			s = jsonval.TrimQuotes(string(mres))
		default:
			s = tval
		}
		dst[i] = s
	}
	return nil
}

// xlsxTime returns a time as written to a sheet: as is, or as text if Excel can't hold it, like a date before 1900.
func xlsxTime(t time.Time, date bool) any {
	if !t.Before(xlsxMinTime) {
		return t
	}
	if date {
		return t.Format(time.DateOnly)
	}
	return t.In(time.UTC).Format(time.DateTime)
}

// xlsxColumns returns the 0-based indexes of the columns, of n, that start at column and repeat every stride columns, if stride is positive.
func xlsxColumns(column, stride, n int) []int {
	if stride <= 0 {
		// Only the first column.
		stride = n
	}
	var res []int
	for col := column; col < n; col += stride {
		res = append(res, col)
	}
	return res
}

// xlsxNumber returns a cell value as a float64, if it is a number.
func xlsxNumber(v any) (float64, bool) {
	switch v := v.(type) {
	case int64:
		return float64(v), true
	case uint64:
		return float64(v), true
	case float64:
		return v, true
	}
	return 0, false
}
