package workbook

import (
	"bytes"
	"cmp"
	"fmt"
	"image/png"
	"slices"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/conditionalformat"
	"github.com/xuri/excelize/v2"
	"google.golang.org/protobuf/types/known/structpb"
)

// The dashboard sheets use a uniform grid of cells that the pixel geometry of the canvas maps onto.
// With the default font, a column of 8.43 characters is 64 px wide and a row of 15 points is 20 px tall.
const (
	cellWidthPx      = 64
	cellHeightPx     = 20
	gridColWidth     = 8.43
	gridRowHeightPt  = 15
	titleRowHeightPt = 24
	// Blocks are drawn inset from their canvas geometry so neighbours don't touch.
	blockInsetPx = 10
	// A KPI value occupies at least two grid columns (more when the block is wide enough)
	// and five rows: label, value, comparison, change and percent change.
	kpiMinCols = 2
	kpiRows    = 5
	// Rows a KPI's sparkline needs below its value; shorter blocks get no sparkline.
	sparklineMinRows = 3
	sparklineColor   = "#2563EB"
	// Characters of a muted label that fit in one grid column.
	labelCharsPerCol = 9
	// Grid columns a table preview gives a text column and a number column.
	previewTextCols   = 3
	previewNumberCols = 2
	// Characters of wrapping text that fit in one grid column.
	textCharsPerCol = 8
	// Rows a table preview takes besides the table's own: its header and the link to the data sheet.
	previewExtraRows = 2
	// Rows of a table that a text block always previews, growing if it must.
	textTableMinRows = 5
)

// writeDashboardSheet writes a dashboard sheet: the header lines, then every block at the cell its geometry maps to.
func (w *writer) writeDashboardSheet(name string, sheet *runtimev1.WorkbookDashboardSheet) error {
	colWidth, rowHeight, customHeight, showGridLines := gridColWidth, float64(gridRowHeightPt), true, false
	if err := w.f.SetSheetProps(name, &excelize.SheetPropsOptions{DefaultColWidth: &colWidth, DefaultRowHeight: &rowHeight, CustomHeight: &customHeight}); err != nil {
		return err
	}
	if err := w.f.SetSheetView(name, 0, &excelize.ViewOptions{ShowGridLines: &showGridLines}); err != nil {
		return err
	}

	bodyRow, err := w.writeHeader(name, sheet)
	if err != nil {
		return err
	}

	geometries := w.blockGeometries(sheet.Blocks)
	for i, b := range sheet.Blocks {
		g := geometries[i]
		col := g.x/cellWidthPx + 1
		row := bodyRow + g.y/cellHeightPx
		offsetX := g.x % cellWidthPx
		offsetY := g.y % cellHeightPx
		cell, err := excelize.CoordinatesToCellName(col, row)
		if err != nil {
			return err
		}
		switch c := b.Content.(type) {
		case *runtimev1.WorkbookBlock_Chart:
			err = w.addChart(name, b.Title, c.Chart, cell, g, offsetX, offsetY)
		case *runtimev1.WorkbookBlock_Image:
			err = w.addImage(name, b.Title, c.Image, cell, g, offsetX, offsetY)
		case *runtimev1.WorkbookBlock_Kpi:
			err = w.writeKpi(name, b, c.Kpi, col, row, g)
		case *runtimev1.WorkbookBlock_Table:
			err = w.writeTablePreview(name, b, c.Table, col, row, g)
		case *runtimev1.WorkbookBlock_Text:
			err = w.writeText(name, b, c.Text, col, row, g)
		case *runtimev1.WorkbookBlock_Note:
			err = w.writeNote(name, b, c.Note.Text, col, row, g)
		}
		if err != nil {
			return fmt.Errorf("block %q: %w", b.Component, err)
		}
	}
	return nil
}

type geometry struct {
	x, y, width, height int
}

// clampGeometry caps a block's size and insets it, so adjacent blocks keep a visible gap.
// The position is kept as sent (validate bounds it): moving a block would put it on top of another one.
func clampGeometry(g *runtimev1.WorkbookGeometry) geometry {
	return geometry{
		x:      int(g.XPx) + blockInsetPx,
		y:      int(g.YPx) + blockInsetPx,
		width:  max(1, min(int(g.WidthPx), maxBlockSidePx)-2*blockInsetPx),
		height: max(1, min(int(g.HeightPx), maxBlockSidePx)-2*blockInsetPx),
	}
}

// blockGeometries returns the geometry of each block on a dashboard sheet.
// A text block whose content needs more rows than its height gives is made taller, and the blocks below it move down by
// as much: unlike the dashboard's component, a cell can't scroll, so text that doesn't fit would be cut off.
// Content that sits side by side in the component (cards, say) is stacked here, so this is common.
// The block gets every row writeText fills, however many: the cap on a block's side bounds drawings, not rows of text,
// and a block that stopped short of its text would have the blocks below it written over that text.
// A block moves down as far as the bottom of a block above it moved, so text blocks side by side that both grow
// move the blocks below them by the larger growth, not by the sum.
func (w *writer) blockGeometries(blocks []*runtimev1.WorkbookBlock) []geometry {
	geometries := make([]geometry, len(blocks))
	grows := make([]int, len(blocks))
	for i, b := range blocks {
		geometries[i] = clampGeometry(b.Geometry)
		if b.GetText() == nil {
			continue
		}
		need := 0
		if b.Title != "" {
			need++
		}
		if b.Subtitle != "" {
			need++
		}
		for _, p := range w.textParts(b.GetText(), max(1, geometries[i].width/cellWidthPx)) {
			need += p.minRows
		}
		grows[i] = max(0, need*cellHeightPx-geometries[i].height)
	}
	// A block above another one starts higher, so going top to bottom finds the shifts of the blocks above a block before its own.
	order := make([]int, len(blocks))
	for i := range order {
		order[i] = i
	}
	slices.SortStableFunc(order, func(a, b int) int { return cmp.Compare(geometries[a].y, geometries[b].y) })
	shifts := make([]int, len(blocks))
	for k, j := range order {
		for _, i := range order[:k] {
			if geometries[i].y+geometries[i].height <= geometries[j].y {
				shifts[j] = max(shifts[j], shifts[i]+grows[i])
			}
		}
	}
	for i := range geometries {
		geometries[i].y += shifts[i]
		geometries[i].height += grows[i]
	}
	return geometries
}

// writeHeader writes the title, period, filters, generation time, link and notes at the top of a dashboard sheet.
// It returns the first row of the body, which follows a blank row.
func (w *writer) writeHeader(name string, sheet *runtimev1.WorkbookDashboardSheet) (int, error) {
	meta := w.req.Meta
	if meta == nil {
		meta = &runtimev1.WorkbookMeta{}
	}
	title := sheet.Title
	if title == "" {
		title = meta.Title
	}
	if err := w.setCell(name, 1, 1, title, w.styles.title); err != nil {
		return 0, err
	}
	if err := w.f.SetRowHeight(name, 1, titleRowHeightPt); err != nil {
		return 0, err
	}
	row := 2
	if meta.TimeRangeText != "" {
		period := meta.TimeRangeText
		if meta.ComparisonText != "" {
			period += " vs " + meta.ComparisonText
		}
		if meta.TimeZone != "" {
			period += " · " + meta.TimeZone
		}
		if err := w.setCell(name, 1, row, period, w.styles.subtitle); err != nil {
			return 0, err
		}
		row++
	}
	for _, line := range meta.FilterLines {
		if err := w.setCell(name, 1, row, line, w.styles.muted); err != nil {
			return 0, err
		}
		row++
	}
	if meta.GeneratedAt != "" {
		if err := w.setCell(name, 1, row, "Generated "+meta.GeneratedAt, w.styles.muted); err != nil {
			return 0, err
		}
		row++
	}
	if meta.DashboardUrl != "" {
		if err := w.setLink(name, 1, row, "Open in Rill", meta.DashboardUrl, "External"); err != nil {
			return 0, err
		}
		row++
	}
	for _, note := range sheet.Notes {
		if err := w.setCell(name, 1, row, note, w.styles.note); err != nil {
			return 0, err
		}
		row++
	}
	return row + 1, nil
}

func (w *writer) addImage(sheet, title string, img *runtimev1.WorkbookImageBlock, cell string, g geometry, offsetX, offsetY int) error {
	cfg, err := png.DecodeConfig(bytes.NewReader(img.Png))
	if err != nil {
		return err
	}
	// Snapshots are captured at a higher pixel ratio than their CSS size; scale them back to the block's size.
	scaleX, scaleY := 1.0, 1.0
	if cfg.Width > 0 {
		scaleX = float64(g.width) / float64(cfg.Width)
	}
	if cfg.Height > 0 {
		scaleY = float64(g.height) / float64(cfg.Height)
	}
	return w.f.AddPictureFromBytes(sheet, cell, &excelize.Picture{
		Extension: ".png",
		File:      img.Png,
		Format: &excelize.GraphicOptions{
			AltText: title,
			OffsetX: offsetX,
			OffsetY: offsetY,
			ScaleX:  scaleX,
			ScaleY:  scaleY,
		},
	})
}

// writeKpi lays the KPI values out left to right, each over a share of the block's width,
// wrapping to a new band when they don't all fit. Labels and values are merged across the
// KPI's columns so long names and large numbers show in full.
func (w *writer) writeKpi(sheet string, b *runtimev1.WorkbookBlock, kpi *runtimev1.WorkbookKpiBlock, col, row int, g geometry) error {
	bottom := row + max(1, g.height/cellHeightPx) - 1
	widthCols := max(kpiMinCols, g.width/cellWidthPx)
	used, err := w.writeHeading(sheet, b, col, row, widthCols)
	if err != nil {
		return err
	}
	row += used
	perRow := max(1, min(len(kpi.Values), widthCols/kpiMinCols))
	cols := max(kpiMinCols, widthCols/perRow)
	bands := (len(kpi.Values) + perRow - 1) / perRow
	// Rows each band of KPIs gets: the block's remaining height shared between the bands.
	bandRows := max(kpiRows, (bottom-row+1)/max(1, bands))
	for i, v := range kpi.Values {
		c := col + (i%perRow)*cols
		r := row + (i/perRow)*bandRows
		if err := w.setMergedCell(sheet, c, r, cols, v.Label, w.styles.muted); err != nil {
			return err
		}
		valueStyle, err := w.styles.kpiValueStyle(compactNumberFormat(v.NumberFormat))
		if err != nil {
			return err
		}
		if err := w.setMergedNumber(sheet, c, r+1, cols, v.Value, valueStyle, "No data"); err != nil {
			return err
		}
		next := r + 2
		if isNumber(v.ComparisonValue) || isNumber(v.Delta) || isNumber(v.DeltaPercent) {
			if err := w.writeKpiComparison(sheet, v, c, next, cols); err != nil {
				return err
			}
			next += 3
		}
		// The sparkline fills the rest of the band.
		sparkRows := r + bandRows - next
		if kpi.TrendDataSheetId != "" && v.TrendColumn != "" && sparkRows >= sparklineMinRows {
			if err := w.addSparkline(sheet, kpi.TrendDataSheetId, v.TrendColumn, c, next, cols, sparkRows); err != nil {
				return err
			}
		}
	}
	return nil
}

// writeKpiComparison writes the comparison value, change and percent change rows of a KPI.
func (w *writer) writeKpiComparison(sheet string, v *runtimev1.WorkbookKpiValue, col, row, cols int) error {
	numStyle, err := w.styles.kpiNumberStyle(compactNumberFormat(v.NumberFormat))
	if err != nil {
		return err
	}
	pctStyle, err := w.styles.kpiNumberStyle(percentFormat)
	if err != nil {
		return err
	}
	// The labels share the KPI's width with the numbers, which take its last column.
	// A comparison period named at length ("Previous quarter") becomes "Previous" when it wouldn't fit.
	comparisonLabel := v.ComparisonLabel
	if comparisonLabel == "" || utf8.RuneCountInString(comparisonLabel) > (cols-1)*labelCharsPerCol {
		comparisonLabel = "Previous"
	}
	rows := []struct {
		label string
		value *structpb.Value
		style int
	}{
		{comparisonLabel, v.ComparisonValue, numStyle},
		{"Change", v.Delta, numStyle},
		{"% change", v.DeltaPercent, pctStyle},
	}
	for j, line := range rows {
		if err := w.setMergedCell(sheet, col, row+j, cols-1, line.label, w.styles.muted); err != nil {
			return err
		}
		if err := w.setNumber(sheet, col+cols-1, row+j, line.value, line.style, ""); err != nil {
			return err
		}
	}
	return nil
}

// addSparkline draws a KPI's trend as an Excel sparkline over a merged area of cols by rows cells,
// bound to the column of the trend data sheet that holds its time series.
func (w *writer) addSparkline(sheet, trendSheetID, column string, col, row, cols, rows int) error {
	ds := w.dataSheets[trendSheetID]
	index := columnIndex(ds, column)
	if index == 0 || len(ds.Rows) == 0 {
		return nil
	}
	from, err := excelize.CoordinatesToCellName(col, row)
	if err != nil {
		return err
	}
	to, err := excelize.CoordinatesToCellName(col+cols-1, row+rows-1)
	if err != nil {
		return err
	}
	if err := w.f.MergeCell(sheet, from, to); err != nil {
		return err
	}
	values, err := columnRange(w.dataSheetNames[trendSheetID], index, len(ds.Rows))
	if err != nil {
		return err
	}
	return w.f.AddSparkline(sheet, &excelize.SparklineOptions{
		Location:    []string{from},
		Range:       []string{values},
		Type:        "line",
		Style:       1,
		SeriesColor: sparklineColor,
		Weight:      1.5,
		EmptyCells:  "gap",
	})
}

// writeTablePreview writes the first rows of each of the block's data sheets (one per leaderboard dimension)
// side by side within the block: as many rows as fit its height and, per sheet, as many columns as fit its
// share of the width, each followed by a link to the sheet. Text columns span three grid columns and number
// columns two, numbers are shown compactly, and the header and banded rows match the data sheets' table style.
func (w *writer) writeTablePreview(sheet string, b *runtimev1.WorkbookBlock, t *runtimev1.WorkbookTableBlock, col, row int, g geometry) error {
	top := row
	widthCols := max(previewTextCols, g.width/cellWidthPx)
	used, err := w.writeHeading(sheet, b, col, row, widthCols)
	if err != nil {
		return err
	}
	if used > 0 {
		// A blank row separates the heading from the table.
		used++
	}
	row += used
	// Rows available for data: the block's height minus the heading, the header row and the link row.
	rows := max(1, g.height/cellHeightPx-(row-top)-previewExtraRows)
	// Every data sheet gets an equal share of the block's width, with a gap column between them.
	n := len(t.DataSheetIds)
	share := max(1, (widthCols-(n-1))/n)
	for i, id := range t.DataSheetIds {
		if err := w.writePreviewTable(sheet, id, col+i*(share+1), row, share, rows); err != nil {
			return err
		}
	}
	return nil
}

// writePreviewTable writes the preview of one data sheet at col and row: at most rows rows within width grid columns.
func (w *writer) writePreviewTable(sheet, id string, col, row, width, rows int) error {
	ds := w.dataSheets[id]
	dsName := w.dataSheetNames[id]
	headers := uniqueHeaders(ds.Columns)

	// The columns that fit the width, with their spans. A column never spans more than the width,
	// so the first one always fits and a preview never overlaps its neighbour.
	// When a number column follows the first one (the label) and wouldn't fit beside it at full span,
	// the label narrows, so a narrow preview still shows a number next to each label.
	var columns, spans []int
	used := 0
	for j, column := range ds.Columns {
		span := min(previewColumnSpan(column), width)
		if j == 0 && len(ds.Columns) > 1 && previewColumnSpan(ds.Columns[1]) == previewNumberCols {
			if rest := width - previewNumberCols; rest >= 1 && rest < span {
				span = rest
			}
		}
		if used+span > width {
			break
		}
		columns = append(columns, j)
		spans = append(spans, span)
		used += span
	}

	c := col
	for k, j := range columns {
		if err := w.setMergedCell(sheet, c, row, spans[k], headers[j], w.styles.previewHeader); err != nil {
			return err
		}
		c += spans[k]
	}
	// The totals row keeps one of the preview's rows, before or after the data rows as the dashboard shows it.
	type line struct {
		values *structpb.ListValue
		// The row's place on the data sheet, which its number format may depend on.
		index    int
		isTotals bool
	}
	var lines []line
	dataRows := ds.Rows[:min(len(ds.Rows), rows)]
	if ds.Totals != nil {
		dataRows = dataRows[:min(len(dataRows), max(0, rows-1))]
	}
	if ds.Totals != nil && !ds.TotalsLast {
		lines = append(lines, line{values: ds.Totals, index: len(ds.Rows), isTotals: true})
	}
	for i, r := range dataRows {
		lines = append(lines, line{values: r, index: i})
	}
	if ds.Totals != nil && ds.TotalsLast {
		lines = append(lines, line{values: ds.Totals, index: len(ds.Rows), isTotals: true})
	}
	banded := true
	for i, l := range lines {
		c := col
		for k, j := range columns {
			column := ds.Columns[j]
			value, err := cellValue(column, rowValue(l.values, j))
			if err != nil {
				return err
			}
			format := cellFormat(ds, column, l.index)
			var style int
			if l.isTotals {
				style, err = w.styles.previewTotalsStyle(format)
			} else {
				style, err = w.styles.previewStyle(format, banded)
			}
			if err != nil {
				return err
			}
			if err := w.setMergedCell(sheet, c, row+1+i, spans[k], value, style); err != nil {
				return err
			}
			c += spans[k]
		}
		if !l.isTotals {
			banded = !banded
		}
	}
	firstDataRow := row + 1
	if ds.Totals != nil && !ds.TotalsLast {
		firstDataRow++
	}
	if err := w.applyPreviewFormats(sheet, ds, columns, spans, col, firstDataRow, len(dataRows)); err != nil {
		return err
	}
	link := fmt.Sprintf("Full data: %s (%d rows)", dsName, len(ds.Rows))
	if ds.MoreRows {
		link = fmt.Sprintf("Data: %s (first %d rows)", dsName, len(ds.Rows))
	}
	switch hidden := len(ds.Columns) - len(columns); {
	case hidden == 1:
		link += ", 1 more column"
	case hidden > 1:
		link += fmt.Sprintf(", %d more columns", hidden)
	}
	return w.setLink(sheet, col, row+1+len(lines), link, quoteSheet(dsName)+"!A1", "Location")
}

// applyPreviewFormats applies a data sheet's conditional formats to its preview, whose data rows start at firstRow.
// columns are the data sheet columns the preview shows, laid out from grid column col with the given spans.
// A merged cell takes the format of its first cell. The ends of the color scales and data bars are those of the full
// data sheet, so the preview's colors match the sheet's.
func (w *writer) applyPreviewFormats(sheet string, ds *runtimev1.WorkbookDataSheet, columns, spans []int, col, firstRow, rows int) error {
	if rows == 0 || len(ds.Rows) == 0 {
		return nil
	}
	for _, cf := range ds.ConditionalFormats {
		var ranges []string
		formatted := formatColumns(ds, cf)
		for _, j := range formatted {
			c := col
			for k, shown := range columns {
				if shown == j {
					r, err := cellRange(c, firstRow, rows, false)
					if err != nil {
						return err
					}
					ranges = append(ranges, r)
					break
				}
				c += spans[k]
			}
		}
		if err := conditionalformat.Apply(w.f, sheet, ranges, formatDomain(ds, cf, formatted), cf.Format); err != nil {
			return err
		}
	}
	return nil
}

func previewColumnSpan(column *runtimev1.WorkbookColumn) int {
	switch column.Type {
	case runtimev1.WorkbookColumnType_WORKBOOK_COLUMN_TYPE_NUMBER, runtimev1.WorkbookColumnType_WORKBOOK_COLUMN_TYPE_BOOL:
		return previewNumberCols
	default:
		return previewTextCols
	}
}

// textPart is a part of a text block with the grid rows it takes: a run of text or the preview of a table.
type textPart struct {
	text string
	// The data sheet of a table.
	dataSheetID string
	// Rows of a table's part that don't show the table's rows:
	// the preview's header and link, and a blank row that sets the part below it apart.
	extraRows int
	// A text needs a row per wrapped line. A table needs to show its first rows and can show all of them.
	minRows, maxRows int
}

// textParts returns the parts of a text block that hold content, for a block that is cols grid columns wide.
func (w *writer) textParts(t *runtimev1.WorkbookTextBlock, cols int) []textPart {
	parts := t.Parts
	if len(parts) == 0 {
		parts = []*runtimev1.WorkbookTextPart{{Part: &runtimev1.WorkbookTextPart_Text{Text: t.Text}}}
	}
	var res []textPart
	for _, p := range parts {
		switch {
		case p.GetDataSheetId() != "":
			ds := w.dataSheets[p.GetDataSheetId()]
			rows := max(1, len(ds.Rows))
			if ds.Totals != nil {
				rows++
			}
			extra := previewExtraRows + 1
			res = append(res, textPart{
				dataSheetID: p.GetDataSheetId(),
				extraRows:   extra,
				minRows:     extra + min(rows, textTableMinRows),
				maxRows:     extra + rows,
			})
		case p.GetText() != "":
			// A cell holds at most excelize.TotalCellChars UTF-16 units and excelize cuts off the rest,
			// so the rows are counted for the text the cell keeps.
			text, units := p.GetText(), 0
			for i, r := range text {
				if units += utf16.RuneLen(r); units > excelize.TotalCellChars {
					text = text[:i]
					break
				}
			}
			lines := 0
			for _, line := range strings.Split(text, "\n") {
				lines += max(1, (utf8.RuneCountInString(line)+cols*textCharsPerCol-1)/(cols*textCharsPerCol))
			}
			res = append(res, textPart{text: text, minRows: lines, maxRows: lines})
		}
	}
	// Nothing follows the last part, so a table there needs no blank row below it.
	if n := len(res); n > 0 && res[n-1].dataSheetID != "" {
		res[n-1].extraRows--
		res[n-1].minRows--
		res[n-1].maxRows--
	}
	return res
}

// writeText writes a text block over its area, top to bottom in reading order:
// each run of text as one merged, wrapping cell and each table as a preview that links to its data sheet.
// A part takes the rows it can use, as long as the parts below it keep the rows they need (see blockGeometries),
// and a text that comes last takes what is left, so a block of only text fills its whole area.
// A block without text or tables writes nothing: an empty merged cell would only hide the grid.
func (w *writer) writeText(sheet string, b *runtimev1.WorkbookBlock, t *runtimev1.WorkbookTextBlock, col, row int, g geometry) error {
	cols := max(1, g.width/cellWidthPx)
	used, err := w.writeHeading(sheet, b, col, row, cols)
	if err != nil {
		return err
	}
	row += used
	rows := max(1, g.height/cellHeightPx-used)

	parts := w.textParts(t, cols)
	// Rows the parts further down need at least.
	reserved := 0
	for _, p := range parts {
		reserved += p.minRows
	}
	for i, p := range parts {
		reserved -= p.minRows
		n := max(p.minRows, min(p.maxRows, rows-reserved))
		if p.dataSheetID != "" {
			if err := w.writePreviewTable(sheet, p.dataSheetID, col, row, cols, n-p.extraRows); err != nil {
				return err
			}
		} else {
			if i == len(parts)-1 {
				n = max(n, rows)
			}
			from, err := excelize.CoordinatesToCellName(col, row)
			if err != nil {
				return err
			}
			to, err := excelize.CoordinatesToCellName(col+cols-1, row+n-1)
			if err != nil {
				return err
			}
			if from != to {
				if err := w.f.MergeCell(sheet, from, to); err != nil {
					return err
				}
			}
			if err := w.f.SetCellStr(sheet, from, p.text); err != nil {
				return err
			}
			if err := w.f.SetCellStyle(sheet, from, from, w.styles.wrap); err != nil {
				return err
			}
		}
		row += n
		rows -= n
	}
	return nil
}

// writeNote writes the block's heading across its width and the note below it.
func (w *writer) writeNote(sheet string, b *runtimev1.WorkbookBlock, text string, col, row int, g geometry) error {
	used, err := w.writeHeading(sheet, b, col, row, max(previewTextCols, g.width/cellWidthPx))
	if err != nil {
		return err
	}
	return w.setCell(sheet, col, row+used, text, w.styles.note)
}

// writeHeading writes a block's title and subtitle (its description), each merged across the block's
// width so they show in full, and returns the number of rows used.
func (w *writer) writeHeading(sheet string, b *runtimev1.WorkbookBlock, col, row, cols int) (int, error) {
	used := 0
	if b.Title != "" {
		if err := w.setMergedCell(sheet, col, row, cols, b.Title, w.styles.label); err != nil {
			return 0, err
		}
		used++
	}
	if b.Subtitle != "" {
		if err := w.setMergedCell(sheet, col, row+used, cols, b.Subtitle, w.styles.muted); err != nil {
			return 0, err
		}
		used++
	}
	return used, nil
}

// setCell writes a value and, when styleID is not 0, its style.
func (w *writer) setCell(sheet string, col, row int, value any, styleID int) error {
	cell, err := excelize.CoordinatesToCellName(col, row)
	if err != nil {
		return err
	}
	if err := w.f.SetCellValue(sheet, cell, value); err != nil {
		return err
	}
	if styleID == 0 {
		return nil
	}
	return w.f.SetCellStyle(sheet, cell, cell, styleID)
}

// setMergedCell writes a value into a cell merged across `cols` columns, so text and numbers wider
// than one grid column show in full.
func (w *writer) setMergedCell(sheet string, col, row, cols int, value any, styleID int) error {
	if cols > 1 {
		from, err := excelize.CoordinatesToCellName(col, row)
		if err != nil {
			return err
		}
		to, err := excelize.CoordinatesToCellName(col+cols-1, row)
		if err != nil {
			return err
		}
		if err := w.f.MergeCell(sheet, from, to); err != nil {
			return err
		}
	}
	return w.setCell(sheet, col, row, value, styleID)
}

// setNumber writes a numeric JSON value, or the placeholder text when there is none.
func (w *writer) setNumber(sheet string, col, row int, v *structpb.Value, styleID int, placeholder string) error {
	return w.setMergedNumber(sheet, col, row, 1, v, styleID, placeholder)
}

func (w *writer) setMergedNumber(sheet string, col, row, cols int, v *structpb.Value, styleID int, placeholder string) error {
	if n, ok := v.GetKind().(*structpb.Value_NumberValue); ok {
		return w.setMergedCell(sheet, col, row, cols, n.NumberValue, styleID)
	}
	if placeholder == "" {
		return nil
	}
	return w.setMergedCell(sheet, col, row, cols, placeholder, w.styles.muted)
}

func isNumber(v *structpb.Value) bool {
	_, ok := v.GetKind().(*structpb.Value_NumberValue)
	return ok
}

func (w *writer) setLink(sheet string, col, row int, text, target, linkType string) error {
	if err := w.setCell(sheet, col, row, text, w.styles.link); err != nil {
		return err
	}
	cell, err := excelize.CoordinatesToCellName(col, row)
	if err != nil {
		return err
	}
	return w.f.SetCellHyperLink(sheet, cell, target, linkType)
}
