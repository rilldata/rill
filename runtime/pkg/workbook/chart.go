package workbook

import (
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/xuri/excelize/v2"
)

// scatterMarkerColor is Excel's first theme accent, the colour a scatter series gets by default.
const scatterMarkerColor = "#4472C4"

// addChart adds a native chart bound to the cells of the block's data sheet.
func (w *writer) addChart(sheet, title string, c *runtimev1.WorkbookChartBlock, cell string, g geometry, offsetX, offsetY int) error {
	ds := w.dataSheets[c.DataSheetId]
	dsName := w.dataSheetNames[c.DataSheetId]
	// With no rows, the chart references the empty first data row rather than nothing at all.
	rows := max(1, len(ds.Rows))
	categoryIndex := columnIndex(ds, c.CategoryColumn)
	categories, err := columnRange(dsName, categoryIndex, rows)
	if err != nil {
		return err
	}

	// Combo charts split their series by mark and axis; every other type draws all series in one chart.
	type group struct {
		typ       excelize.ChartType
		secondary bool
		series    []excelize.ChartSeries
	}
	var groups []*group
	addSeries := func(typ excelize.ChartType, secondary bool, s excelize.ChartSeries) {
		for _, grp := range groups {
			if grp.typ == typ && grp.secondary == secondary {
				grp.series = append(grp.series, s)
				return
			}
		}
		groups = append(groups, &group{typ: typ, secondary: secondary, series: []excelize.ChartSeries{s}})
	}
	for _, s := range c.Series {
		idx := columnIndex(ds, s.Column)
		values, err := columnRange(dsName, idx, rows)
		if err != nil {
			return err
		}
		headerCell, err := excelize.CoordinatesToCellName(idx, 1, true)
		if err != nil {
			return err
		}
		cs := excelize.ChartSeries{
			Name:       quoteSheet(dsName) + "!" + headerCell,
			Categories: categories,
			Values:     values,
		}
		typ, err := chartType(c, s.Mark)
		if err != nil {
			return err
		}
		switch typ {
		case excelize.Line:
			cs.Marker = excelize.ChartMarker{Symbol: "none"}
		case excelize.Scatter:
			// Points only: Excel's default scatter style joins the points with lines.
			// The marker needs an explicit fill: excelize drops the theme colour it would get otherwise,
			// and LibreOffice then draws no points at all.
			cs.Line = excelize.LineOptions{Type: excelize.LineNone}
			cs.Marker = excelize.ChartMarker{
				Symbol: "circle",
				Size:   5,
				Fill:   excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{scatterMarkerColor}},
			}
		}
		secondary := c.Type == runtimev1.WorkbookChartType_WORKBOOK_CHART_TYPE_COMBO && s.SecondaryAxis
		addSeries(typ, secondary, cs)
	}

	varyColors := c.Type == runtimev1.WorkbookChartType_WORKBOOK_CHART_TYPE_PIE || c.Type == runtimev1.WorkbookChartType_WORKBOOK_CHART_TYPE_DOUGHNUT
	charts := make([]*excelize.Chart, 0, len(groups))
	for i, grp := range groups {
		chart := &excelize.Chart{
			Type:       grp.typ,
			Series:     grp.series,
			VaryColors: &varyColors,
		}
		chart.YAxis.Secondary = grp.secondary
		if i == 0 {
			// The first chart carries the frame: size, position, title, legend and axes.
			chart.Dimension = excelize.ChartDimension{Width: uint(g.width), Height: uint(g.height)}
			chart.Format = excelize.GraphicOptions{OffsetX: offsetX, OffsetY: offsetY}
			// A small legend font keeps long legends (one entry per split value) inside the chart.
			chart.Legend = excelize.ChartLegend{Position: legendPosition(c.LegendPosition), Font: &excelize.Font{Size: 8}}
			if title != "" {
				chart.Title = excelize.ChartTitle{Paragraph: []excelize.RichTextRun{{Text: title}}}
			}
			if c.XAxisTitle != "" {
				chart.XAxis.Title = excelize.ChartTitle{Paragraph: []excelize.RichTextRun{{Text: c.XAxisTitle}}}
			}
			// Excel draws the first category of a bar chart at the bottom; reversing the axis keeps the canvas order, top to bottom.
			if c.Type == runtimev1.WorkbookChartType_WORKBOOK_CHART_TYPE_BAR {
				chart.XAxis.ReverseOrder = true
			}
			// Category labels follow the data cells' format, so dates don't show as serial numbers.
			if format := columnFormat(ds.Columns[categoryIndex-1]); format != "" {
				chart.XAxis.NumFmt = excelize.ChartNumFmt{CustomNumFmt: format, SourceLinked: true}
			}
			if c.YAxisTitle != "" {
				chart.YAxis.Title = excelize.ChartTitle{Paragraph: []excelize.RichTextRun{{Text: c.YAxisTitle}}}
			}
			if c.NumberFormat != "" {
				chart.YAxis.NumFmt = excelize.ChartNumFmt{CustomNumFmt: compactNumberFormat(c.NumberFormat)}
			}
		}
		charts = append(charts, chart)
	}
	return w.f.AddChart(sheet, cell, charts[0], charts[1:]...)
}

// chartType maps a block's chart type and layout flags to the excelize chart type.
// For combo charts the series mark decides between columns and a line.
func chartType(c *runtimev1.WorkbookChartBlock, mark runtimev1.WorkbookSeriesMark) (excelize.ChartType, error) {
	stacked := func(plain, stacked, percent excelize.ChartType) excelize.ChartType {
		if c.PercentStacked {
			return percent
		}
		if c.Stacked {
			return stacked
		}
		return plain
	}
	switch c.Type {
	case runtimev1.WorkbookChartType_WORKBOOK_CHART_TYPE_COLUMN:
		return stacked(excelize.Col, excelize.ColStacked, excelize.ColPercentStacked), nil
	case runtimev1.WorkbookChartType_WORKBOOK_CHART_TYPE_BAR:
		return stacked(excelize.Bar, excelize.BarStacked, excelize.BarPercentStacked), nil
	case runtimev1.WorkbookChartType_WORKBOOK_CHART_TYPE_LINE:
		return excelize.Line, nil
	case runtimev1.WorkbookChartType_WORKBOOK_CHART_TYPE_AREA:
		return stacked(excelize.Area, excelize.AreaStacked, excelize.AreaPercentStacked), nil
	case runtimev1.WorkbookChartType_WORKBOOK_CHART_TYPE_PIE:
		return excelize.Pie, nil
	case runtimev1.WorkbookChartType_WORKBOOK_CHART_TYPE_DOUGHNUT:
		return excelize.Doughnut, nil
	case runtimev1.WorkbookChartType_WORKBOOK_CHART_TYPE_SCATTER:
		return excelize.Scatter, nil
	case runtimev1.WorkbookChartType_WORKBOOK_CHART_TYPE_COMBO:
		if mark == runtimev1.WorkbookSeriesMark_WORKBOOK_SERIES_MARK_LINE {
			return excelize.Line, nil
		}
		return stacked(excelize.Col, excelize.ColStacked, excelize.ColPercentStacked), nil
	default:
		return 0, fmt.Errorf("unsupported chart type %q", c.Type)
	}
}

// legendPosition maps the block's legend position to one excelize accepts; "none" leaves the legend out.
func legendPosition(position string) string {
	switch position {
	case "none", "top", "bottom", "left", "right":
		return position
	default:
		return "bottom"
	}
}

// columnRange returns the formula reference to a column's data cells, e.g. 'Data - Spend'!$B$2:$B$11.
func columnRange(sheet string, col, rows int) (string, error) {
	first, err := excelize.CoordinatesToCellName(col, 2, true)
	if err != nil {
		return "", err
	}
	last, err := excelize.CoordinatesToCellName(col, rows+1, true)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s!%s:%s", quoteSheet(sheet), first, last), nil
}
