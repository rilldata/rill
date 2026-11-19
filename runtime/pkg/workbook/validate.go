package workbook

import (
	"bytes"
	"errors"
	"fmt"
	"image/png"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// Caps on the size of a request.
// The browser builds a request from data it already holds under its own claims,
// so the caps protect the runtime from oversized workbooks rather than from untrusted content.
const (
	MaxSheets          = 200
	MaxRowsPerSheet    = 100_000
	MaxColumnsPerSheet = 200
	// Cells across all data sheets: the request is parsed into memory before the sheets are streamed out.
	MaxCells      = 1_000_000
	MaxImageBytes = 32 << 20
	// Blocks wider or taller than this are clamped so a bad geometry can't produce a huge drawing.
	maxBlockSidePx = 10_000
	// How far right and down a block may start: about 1,500 grid columns and 100,000 grid rows, well inside a sheet.
	// Positions are rejected rather than clamped, since clamping would stack the blocks of a long canvas on top of each other.
	maxBlockXPx = 100_000
	maxBlockYPx = 2_000_000
)

var pngMagic = []byte("\x89PNG\r\n\x1a\n")

// ValidationError is returned when a request can't be rendered as a workbook. Handlers should map it to a client error.
type ValidationError struct {
	msg string
}

func (e *ValidationError) Error() string {
	return e.msg
}

// IsValidationError reports whether err is a ValidationError.
func IsValidationError(err error) bool {
	var ve *ValidationError
	return errors.As(err, &ve)
}

func validationErrorf(format string, args ...any) error {
	return &ValidationError{msg: fmt.Sprintf(format, args...)}
}

// validate checks the caps and the internal references of a request.
func validate(req *runtimev1.CanvasWorkbookExportRequest) error {
	if req == nil {
		return validationErrorf("request is empty")
	}
	if len(req.DashboardSheets) == 0 {
		return validationErrorf("the workbook needs at least one dashboard sheet")
	}
	if n := len(req.DashboardSheets) + len(req.DataSheets); n > MaxSheets {
		return validationErrorf("the workbook has %d sheets, which exceeds the limit of %d", n, MaxSheets)
	}

	dataSheets := make(map[string]*runtimev1.WorkbookDataSheet, len(req.DataSheets))
	cells := 0
	for _, ds := range req.DataSheets {
		if ds.Id == "" {
			return validationErrorf("data sheet %q has no id", ds.Title)
		}
		if _, ok := dataSheets[ds.Id]; ok {
			return validationErrorf("data sheet id %q is used more than once", ds.Id)
		}
		dataSheets[ds.Id] = ds
		if len(ds.Columns) == 0 {
			return validationErrorf("data sheet %q has no columns", ds.Title)
		}
		if len(ds.Columns) > MaxColumnsPerSheet {
			return validationErrorf("data sheet %q has %d columns, which exceeds the limit of %d", ds.Title, len(ds.Columns), MaxColumnsPerSheet)
		}
		names := make(map[string]bool, len(ds.Columns))
		for _, c := range ds.Columns {
			if c.Name == "" {
				return validationErrorf("data sheet %q has a column without a name", ds.Title)
			}
			if names[c.Name] {
				return validationErrorf("data sheet %q has more than one column named %q", ds.Title, c.Name)
			}
			names[c.Name] = true
		}
		if len(ds.Rows) > MaxRowsPerSheet {
			return validationErrorf("data sheet %q has %d rows, which exceeds the limit of %d", ds.Title, len(ds.Rows), MaxRowsPerSheet)
		}
		for i, r := range ds.Rows {
			if r != nil && len(r.Values) > len(ds.Columns) {
				return validationErrorf("row %d of data sheet %q has %d values but the sheet has %d columns", i+1, ds.Title, len(r.Values), len(ds.Columns))
			}
		}
		if ds.Totals != nil && len(ds.Totals.Values) > len(ds.Columns) {
			return validationErrorf("the totals row of data sheet %q has %d values but the sheet has %d columns", ds.Title, len(ds.Totals.Values), len(ds.Columns))
		}
		if len(ds.RowNumberFormats) > len(ds.Rows) {
			return validationErrorf("data sheet %q has %d row number formats but %d rows", ds.Title, len(ds.RowNumberFormats), len(ds.Rows))
		}
		cells += len(ds.Rows) * len(ds.Columns)
		if ds.Totals != nil {
			cells += len(ds.Columns)
		}
		if cells > MaxCells {
			return validationErrorf("the data sheets hold more than %d cells", MaxCells)
		}
	}

	var imageBytes int
	for _, sheet := range req.DashboardSheets {
		for _, b := range sheet.Blocks {
			if b.Geometry == nil || b.Geometry.WidthPx == 0 || b.Geometry.HeightPx == 0 {
				return validationErrorf("block %q has no size", b.Component)
			}
			if b.Geometry.XPx > maxBlockXPx || b.Geometry.YPx > maxBlockYPx {
				return validationErrorf("block %q is placed beyond the edge of the sheet", b.Component)
			}
			switch c := b.Content.(type) {
			case *runtimev1.WorkbookBlock_Chart:
				ds, ok := dataSheets[c.Chart.DataSheetId]
				if !ok {
					return validationErrorf("chart %q references unknown data sheet %q", b.Component, c.Chart.DataSheetId)
				}
				if c.Chart.Type == runtimev1.WorkbookChartType_WORKBOOK_CHART_TYPE_UNSPECIFIED {
					return validationErrorf("chart %q has no chart type", b.Component)
				}
				if columnIndex(ds, c.Chart.CategoryColumn) == 0 {
					return validationErrorf("chart %q references unknown category column %q", b.Component, c.Chart.CategoryColumn)
				}
				if len(c.Chart.Series) == 0 {
					return validationErrorf("chart %q has no series", b.Component)
				}
				for _, s := range c.Chart.Series {
					if columnIndex(ds, s.Column) == 0 {
						return validationErrorf("chart %q references unknown series column %q", b.Component, s.Column)
					}
					if c.Chart.Type == runtimev1.WorkbookChartType_WORKBOOK_CHART_TYPE_COMBO && s.Mark == runtimev1.WorkbookSeriesMark_WORKBOOK_SERIES_MARK_UNSPECIFIED {
						return validationErrorf("combo chart %q has a series without a mark", b.Component)
					}
				}
			case *runtimev1.WorkbookBlock_Image:
				if !bytes.HasPrefix(c.Image.Png, pngMagic) {
					return validationErrorf("image %q is not a PNG", b.Component)
				}
				if _, err := png.DecodeConfig(bytes.NewReader(c.Image.Png)); err != nil {
					return validationErrorf("image %q is not a valid PNG: %v", b.Component, err)
				}
				imageBytes += len(c.Image.Png)
				if imageBytes > MaxImageBytes {
					return validationErrorf("the images exceed the limit of %d bytes", MaxImageBytes)
				}
			case *runtimev1.WorkbookBlock_Table:
				if len(c.Table.DataSheetIds) == 0 {
					return validationErrorf("table %q references no data sheet", b.Component)
				}
				for _, id := range c.Table.DataSheetIds {
					if _, ok := dataSheets[id]; !ok {
						return validationErrorf("table %q references unknown data sheet %q", b.Component, id)
					}
				}
			case *runtimev1.WorkbookBlock_Kpi:
				if c.Kpi.TrendDataSheetId == "" {
					break
				}
				ds, ok := dataSheets[c.Kpi.TrendDataSheetId]
				if !ok {
					return validationErrorf("KPI block %q references unknown trend data sheet %q", b.Component, c.Kpi.TrendDataSheetId)
				}
				for _, v := range c.Kpi.Values {
					if v.TrendColumn != "" && columnIndex(ds, v.TrendColumn) == 0 {
						return validationErrorf("KPI %q of block %q references unknown trend column %q", v.Label, b.Component, v.TrendColumn)
					}
				}
			case *runtimev1.WorkbookBlock_Text, *runtimev1.WorkbookBlock_Note:
			default:
				return validationErrorf("block %q has no content", b.Component)
			}
		}
	}
	return nil
}

// columnIndex returns the 1-based index of the named column, or 0 if the sheet has no such column.
func columnIndex(ds *runtimev1.WorkbookDataSheet, name string) int {
	for i, c := range ds.Columns {
		if c.Name == name {
			return i + 1
		}
	}
	return 0
}
