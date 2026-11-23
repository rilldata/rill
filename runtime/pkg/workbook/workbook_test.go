package workbook

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"strings"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
	"google.golang.org/protobuf/types/known/structpb"
)

func testPNG(t *testing.T, width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for i := range img.Pix {
		img.Pix[i] = 0xff
	}
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))
	return buf.Bytes()
}

func row(t *testing.T, values ...any) *structpb.ListValue {
	lv, err := structpb.NewList(values)
	require.NoError(t, err)
	return lv
}

func num(v float64) *structpb.Value {
	return structpb.NewNumberValue(v)
}

func geo(x, y, width, height uint32) *runtimev1.WorkbookGeometry {
	return &runtimev1.WorkbookGeometry{XPx: x, YPx: y, WidthPx: width, HeightPx: height}
}

// testRequest covers every block type and column type on one dashboard sheet, with a second sheet for a tab.
func testRequest(t *testing.T) *runtimev1.CanvasWorkbookExportRequest {
	return &runtimev1.CanvasWorkbookExportRequest{
		Canvas: "campaign",
		Meta: &runtimev1.WorkbookMeta{
			Title:          "Campaign report",
			GeneratedAt:    "2026-10-21 18:40 UTC",
			DashboardUrl:   "https://ui.rilldata.com/org/proj/canvas/campaign",
			TimeRangeText:  "Jan 1 – Mar 31, 2026",
			ComparisonText: "Oct 1 – Dec 31, 2025",
			TimeZone:       "UTC",
			FilterLines:    []string{"Brand: Northwind"},
		},
		DataSheets: []*runtimev1.WorkbookDataSheet{
			{
				Id:    "trend",
				Title: "Spend and CPA by week",
				Columns: []*runtimev1.WorkbookColumn{
					{Name: "week", DisplayName: "Week ending", Type: runtimev1.WorkbookColumnType_WORKBOOK_COLUMN_TYPE_DATE},
					{Name: "spend", DisplayName: "Client Spend", Type: runtimev1.WorkbookColumnType_WORKBOOK_COLUMN_TYPE_NUMBER, NumberFormat: `"$"#,##0.00`},
					{Name: "cpa", DisplayName: "CPA", Type: runtimev1.WorkbookColumnType_WORKBOOK_COLUMN_TYPE_NUMBER, NumberFormat: `"$"#,##0.00`},
				},
				Rows: []*structpb.ListValue{
					row(t, "2026-01-04", 71.28, 17.82),
					row(t, "2026-01-11", 17239.29, 4.91),
					row(t, "2026-01-18", nil, 3.02),
				},
			},
			{
				Id:    "lines",
				Title: "Line items",
				Columns: []*runtimev1.WorkbookColumn{
					{Name: "line_item", DisplayName: "Line Item", Type: runtimev1.WorkbookColumnType_WORKBOOK_COLUMN_TYPE_STRING},
					{Name: "impressions", DisplayName: "Impressions", Type: runtimev1.WorkbookColumnType_WORKBOOK_COLUMN_TYPE_NUMBER, NumberFormat: "#,##0"},
					{Name: "active", DisplayName: "Active", Type: runtimev1.WorkbookColumnType_WORKBOOK_COLUMN_TYPE_BOOL},
					{Name: "updated", DisplayName: "Updated", Type: runtimev1.WorkbookColumnType_WORKBOOK_COLUMN_TYPE_TIMESTAMP},
				},
				Rows: []*structpb.ListValue{
					row(t, "Display", 43196497, true, "2026-03-31T13:31:01Z"),
					row(t, "Video", "1641148", "false", "2026-03-30T00:00:00"),
				},
				ConditionalFormats: []*runtimev1.WorkbookConditionalFormat{{
					Columns: []string{"impressions"},
					Format:  &runtimev1.ExportConditionalFormat{Format: &runtimev1.ExportConditionalFormat_DataBar{DataBar: &runtimev1.ExportDataBar{Color: "#638ec6"}}},
				}},
			},
			{
				Id:      "empty",
				Title:   "Empty",
				Columns: []*runtimev1.WorkbookColumn{{Name: "a", Type: runtimev1.WorkbookColumnType_WORKBOOK_COLUMN_TYPE_STRING}},
			},
			{
				Id:    "kpis-trend",
				Title: "Overall trend",
				Columns: []*runtimev1.WorkbookColumn{
					{Name: "time", DisplayName: "Time", Type: runtimev1.WorkbookColumnType_WORKBOOK_COLUMN_TYPE_DATE},
					{Name: "kpi_0", DisplayName: "Impressions", Type: runtimev1.WorkbookColumnType_WORKBOOK_COLUMN_TYPE_NUMBER, NumberFormat: "#,##0"},
				},
				Rows: []*structpb.ListValue{
					row(t, "2026-01-01", 10),
					row(t, "2026-01-02", 12),
					row(t, "2026-01-03", 9),
				},
			},
		},
		DashboardSheets: []*runtimev1.WorkbookDashboardSheet{
			{
				Title: "Campaign report",
				Notes: []string{"Tab group: Overview: see the sheet of the same name"},
				Blocks: []*runtimev1.WorkbookBlock{
					{
						Component: "combo", Title: "Spend and CPA by Week", Geometry: geo(0, 0, 600, 300),
						Content: &runtimev1.WorkbookBlock_Chart{Chart: &runtimev1.WorkbookChartBlock{
							DataSheetId:    "trend",
							Type:           runtimev1.WorkbookChartType_WORKBOOK_CHART_TYPE_COMBO,
							CategoryColumn: "week",
							Series: []*runtimev1.WorkbookChartSeries{
								{Column: "spend", Mark: runtimev1.WorkbookSeriesMark_WORKBOOK_SERIES_MARK_BAR},
								{Column: "cpa", Mark: runtimev1.WorkbookSeriesMark_WORKBOOK_SERIES_MARK_LINE, SecondaryAxis: true},
							},
							LegendPosition: "top",
							XAxisTitle:     "Week Ending",
							YAxisTitle:     "Client Spend",
							NumberFormat:   `"$"#,##0`,
						}},
					},
					{
						Component: "pie", Title: "Impressions by line item", Geometry: geo(620, 0, 400, 300),
						Content: &runtimev1.WorkbookBlock_Chart{Chart: &runtimev1.WorkbookChartBlock{
							DataSheetId:    "lines",
							Type:           runtimev1.WorkbookChartType_WORKBOOK_CHART_TYPE_PIE,
							CategoryColumn: "line_item",
							Series:         []*runtimev1.WorkbookChartSeries{{Column: "impressions"}},
						}},
					},
					{
						Component: "snapshot", Title: "Heatmap", Geometry: geo(0, 320, 400, 200),
						Content: &runtimev1.WorkbookBlock_Image{Image: &runtimev1.WorkbookImageBlock{Png: testPNG(t, 800, 400)}},
					},
					{
						Component: "kpis", Title: "Overall", Geometry: geo(420, 320, 600, 200),
						Content: &runtimev1.WorkbookBlock_Kpi{Kpi: &runtimev1.WorkbookKpiBlock{
							TrendDataSheetId: "kpis-trend",
							Values: []*runtimev1.WorkbookKpiValue{
								{Label: "Impressions", Value: num(44837645), NumberFormat: "#,##0", ComparisonLabel: "Previous quarter", ComparisonValue: num(41200000), Delta: num(3637645), DeltaPercent: num(0.088), TrendColumn: "kpi_0"},
								{Label: "CTR", Value: structpb.NewNullValue(), NumberFormat: "0.00%"},
							},
						}},
					},
					{
						Component: "table", Title: "Line item performance", Geometry: geo(0, 540, 1000, 160),
						Content: &runtimev1.WorkbookBlock_Table{Table: &runtimev1.WorkbookTableBlock{DataSheetIds: []string{"lines"}}},
					},
					{
						Component: "md", Geometry: geo(0, 680, 400, 60),
						Content: &runtimev1.WorkbookBlock_Text{Text: &runtimev1.WorkbookTextBlock{Text: "Prepared by the analytics team"}},
					},
					{
						Component: "broken", Title: "Funnel", Geometry: geo(420, 680, 400, 60),
						Content: &runtimev1.WorkbookBlock_Note{Note: &runtimev1.WorkbookNoteBlock{Text: "No data"}},
					},
				},
			},
			{Title: "Overview"},
		},
	}
}

func TestWrite(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, Write(&buf, testRequest(t)))

	f, err := excelize.OpenReader(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)
	defer f.Close()

	const dashboard = "Campaign report"
	require.Equal(t, []string{dashboard, "Overview", "Data - Spend and CPA by week", "Data - Line items", "Data - Empty", "Data - Overall trend"}, f.GetSheetList())

	cell := func(sheet, ref string) string {
		v, err := f.GetCellValue(sheet, ref)
		require.NoError(t, err)
		return v
	}

	t.Run("header", func(t *testing.T) {
		require.Equal(t, "Campaign report", cell(dashboard, "A1"))
		require.Equal(t, "Jan 1 – Mar 31, 2026 vs Oct 1 – Dec 31, 2025 · UTC", cell(dashboard, "A2"))
		require.Equal(t, "Brand: Northwind", cell(dashboard, "A3"))
		require.Equal(t, "Generated 2026-10-21 18:40 UTC", cell(dashboard, "A4"))
		require.Equal(t, "Open in Rill", cell(dashboard, "A5"))
		ok, link, err := f.GetCellHyperLink(dashboard, "A5")
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, "https://ui.rilldata.com/org/proj/canvas/campaign", link)
		require.Equal(t, "Tab group: Overview: see the sheet of the same name", cell(dashboard, "A6"))
		// The tab's sheet repeats the workbook header with its own title.
		require.Equal(t, "Overview", cell("Overview", "A1"))
		require.Equal(t, "Jan 1 – Mar 31, 2026 vs Oct 1 – Dec 31, 2025 · UTC", cell("Overview", "A2"))
	})

	t.Run("data sheets", func(t *testing.T) {
		const trend = "Data - Spend and CPA by week"
		require.Equal(t, "Week ending", cell(trend, "A1"))
		require.Equal(t, "Client Spend", cell(trend, "B1"))
		require.Equal(t, "2026-01-04", cell(trend, "A2"))
		require.Equal(t, "$71.28", cell(trend, "B2"))
		require.Equal(t, "", cell(trend, "B4"))
		require.Equal(t, "$3.02", cell(trend, "C4"))
		tables, err := f.GetTables(trend)
		require.NoError(t, err)
		require.Len(t, tables, 1)
		require.Equal(t, "Table_Spend_and_CPA_by_week", tables[0].Name)
		require.Equal(t, "A1:C4", tables[0].Range)

		const lines = "Data - Line items"
		require.Equal(t, "43,196,497", cell(lines, "B2"))
		require.Equal(t, "TRUE", cell(lines, "C2"))
		require.Equal(t, "2026-03-31 13:31:01", cell(lines, "D2"))
		// Numbers and booleans sent as strings are still typed.
		require.Equal(t, "1,641,148", cell(lines, "B3"))
		require.Equal(t, "FALSE", cell(lines, "C3"))
		require.Equal(t, "2026-03-30 00:00:00", cell(lines, "D3"))

		// The conditional format covers the column's data rows.
		formats, err := f.GetConditionalFormats(lines)
		require.NoError(t, err)
		require.Len(t, formats["B2:B3"], 1)
		require.Equal(t, "data_bar", formats["B2:B3"][0].Type)
		require.Equal(t, "638EC6", formats["B2:B3"][0].BarColor)
		// The bars run from zero to the largest value in the column, written as numbers.
		require.Equal(t, "0", formats["B2:B3"][0].MinValue)
		require.Equal(t, "4.3196497E+07", formats["B2:B3"][0].MaxValue)

		// A sheet without rows keeps its header but gets no table.
		require.Equal(t, "a", cell("Data - Empty", "A1"))
		tables, err = f.GetTables("Data - Empty")
		require.NoError(t, err)
		require.Empty(t, tables)
	})

	mergedRanges := func() []string {
		merged, err := f.GetMergeCells(dashboard)
		require.NoError(t, err)
		ranges := make([]string, len(merged))
		for i, m := range merged {
			ranges[i] = m.GetStartAxis() + ":" + m.GetEndAxis()
		}
		return ranges
	}

	// The body starts at row 8, after the six header lines and a blank row; 64 px per column and
	// 20 px per row, with every block inset by 10 px.
	t.Run("kpi block", func(t *testing.T) {
		// The 600 px block spans nine columns, so each of its two KPIs gets four.
		require.Equal(t, "Overall", cell(dashboard, "G24"))
		require.Contains(t, mergedRanges(), "G24:O24")
		require.Equal(t, "Impressions", cell(dashboard, "G25"))
		require.Equal(t, "44.8M", cell(dashboard, "G26"))
		require.Contains(t, mergedRanges(), "G26:J26")
		require.Equal(t, "Previous quarter", cell(dashboard, "G27"))
		require.Equal(t, "41.2M", cell(dashboard, "J27"))
		require.Equal(t, "Change", cell(dashboard, "G28"))
		require.Equal(t, "3.6M", cell(dashboard, "J28"))
		require.Equal(t, "% change", cell(dashboard, "G29"))
		require.Equal(t, "8.8%", cell(dashboard, "J29"))
		// The second KPI has no value and no comparison rows.
		require.Equal(t, "CTR", cell(dashboard, "K25"))
		require.Equal(t, "No data", cell(dashboard, "K26"))
		require.Equal(t, "", cell(dashboard, "K27"))
		// The first KPI's sparkline fills the rest of its band below the comparison rows, bound to the trend sheet.
		require.Contains(t, mergedRanges(), "G30:J32")
		sheet := sheetXML(t, buf.Bytes(), "sheet1.xml")
		require.Contains(t, sheet, "<x14:sparkline>")
		require.Contains(t, sheet, "Data - Overall trend&#39;!$B$2:$B$4")
		require.Contains(t, sheet, "<xm:sqref>G30</xm:sqref>")
	})

	t.Run("table preview", func(t *testing.T) {
		// Title, a blank row, then the header: text columns span three grid columns, numbers two.
		require.Equal(t, "Line item performance", cell(dashboard, "A35"))
		require.Equal(t, "", cell(dashboard, "A36"))
		require.Equal(t, "Line Item", cell(dashboard, "A37"))
		require.Equal(t, "Impressions", cell(dashboard, "D37"))
		require.Equal(t, "Updated", cell(dashboard, "H37"))
		require.Equal(t, "Display", cell(dashboard, "A38"))
		require.Equal(t, "43.2M", cell(dashboard, "D38"))
		require.Equal(t, "Video", cell(dashboard, "A39"))
		// The header is filled like a table header and the rows are banded.
		headerStyle, err := f.GetCellStyle(dashboard, "A37")
		require.NoError(t, err)
		header, err := f.GetStyle(headerStyle)
		require.NoError(t, err)
		require.Equal(t, []string{"4472C4"}, header.Fill.Color)
		require.True(t, header.Font.Bold)
		first, err := f.GetCellStyle(dashboard, "A38")
		require.NoError(t, err)
		second, err := f.GetCellStyle(dashboard, "A39")
		require.NoError(t, err)
		require.NotEqual(t, first, second)
		banded, err := f.GetStyle(first)
		require.NoError(t, err)
		require.Equal(t, []string{"D9E1F2"}, banded.Fill.Color)
		// The preview's numbers get the data sheet's conditional format, scaled to the full sheet.
		formats, err := f.GetConditionalFormats(dashboard)
		require.NoError(t, err)
		require.Len(t, formats["D38:D39"], 1)
		require.Equal(t, "data_bar", formats["D38:D39"][0].Type)
		require.Equal(t, "4.3196497E+07", formats["D38:D39"][0].MaxValue)
		require.Equal(t, "Full data: Data - Line items (2 rows)", cell(dashboard, "A40"))
		ok, link, err := f.GetCellHyperLink(dashboard, "A40")
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, "'Data - Line items'!A1", link)
	})

	t.Run("text and note blocks", func(t *testing.T) {
		require.Equal(t, "Prepared by the analytics team", cell(dashboard, "A42"))
		require.Contains(t, mergedRanges(), "A42:E43")
		require.Equal(t, "Funnel", cell(dashboard, "G42"))
		require.Equal(t, "No data", cell(dashboard, "G43"))
	})

	t.Run("image", func(t *testing.T) {
		pictures, err := f.GetPictures(dashboard, "A24")
		require.NoError(t, err)
		require.Len(t, pictures, 1)
		require.Equal(t, ".png", pictures[0].Extension)
	})

	t.Run("charts", func(t *testing.T) {
		charts := chartParts(t, buf.Bytes())
		require.Len(t, charts, 2)
		// The chart XML uses the default namespace, and the sheet quotes are XML-escaped.
		var combo, pie string
		for _, c := range charts {
			if strings.Contains(c, "<pieChart>") {
				pie = c
			} else {
				combo = c
			}
		}
		require.Contains(t, combo, "<barChart>")
		require.Contains(t, combo, "<lineChart>")
		require.Contains(t, combo, "Data - Spend and CPA by week&#39;!$A$2:$A$4")
		require.Contains(t, combo, "Data - Spend and CPA by week&#39;!$B$2:$B$4")
		require.Contains(t, combo, "Data - Spend and CPA by week&#39;!$C$2:$C$4")
		require.Contains(t, combo, "Data - Spend and CPA by week&#39;!$B$1")
		require.Contains(t, combo, "Spend and CPA by Week")
		require.Contains(t, combo, "Week Ending")
		require.Contains(t, combo, `<legendPos val="t">`)
		// The value axis abbreviates thousands and millions; the category axis follows the date cells.
		require.Contains(t, combo, `formatCode="[&gt;=1000000]&#34;$&#34;#,##0.0,,&#34;M&#34;;[&gt;=1000]&#34;$&#34;#,##0.0,&#34;k&#34;;&#34;$&#34;#,##0"`)
		require.Contains(t, combo, `formatCode="yyyy-mm-dd" sourceLinked="true"`)
		// Two value axes: the line series sits on the secondary axis.
		require.Equal(t, 2, strings.Count(combo, "<valAx>"))
		// The secondary axis shows its own measure in that measure's format (CPA keeps its cents).
		secondaryAxis := combo[strings.LastIndex(combo, "<valAx>"):]
		require.Contains(t, secondaryAxis, `formatCode="[&gt;=1000000]&#34;$&#34;#,##0.0,,&#34;M&#34;;[&gt;=1000]&#34;$&#34;#,##0.0,&#34;k&#34;;&#34;$&#34;#,##0.00"`)

		require.Contains(t, pie, "Data - Line items&#39;!$A$2:$A$3")
		require.Contains(t, pie, "Data - Line items&#39;!$B$2:$B$3")
	})
}

// sheetXML returns the XML of one worksheet part of the workbook.
func sheetXML(t *testing.T, data []byte, name string) string {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	require.NoError(t, err)
	for _, file := range zr.File {
		if file.Name != "xl/worksheets/"+name {
			continue
		}
		rc, err := file.Open()
		require.NoError(t, err)
		content, err := io.ReadAll(rc)
		require.NoError(t, err)
		require.NoError(t, rc.Close())
		return string(content)
	}
	t.Fatalf("worksheet %q not found", name)
	return ""
}

// chartParts returns the XML of every chart in the workbook.
func chartParts(t *testing.T, data []byte) []string {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	require.NoError(t, err)
	var charts []string
	for _, file := range zr.File {
		if !strings.HasPrefix(file.Name, "xl/charts/chart") {
			continue
		}
		rc, err := file.Open()
		require.NoError(t, err)
		content, err := io.ReadAll(rc)
		require.NoError(t, err)
		require.NoError(t, rc.Close())
		charts = append(charts, string(content))
	}
	return charts
}

func TestWriteSupportsEveryChartType(t *testing.T) {
	for _, typ := range []runtimev1.WorkbookChartType{
		runtimev1.WorkbookChartType_WORKBOOK_CHART_TYPE_COLUMN,
		runtimev1.WorkbookChartType_WORKBOOK_CHART_TYPE_BAR,
		runtimev1.WorkbookChartType_WORKBOOK_CHART_TYPE_LINE,
		runtimev1.WorkbookChartType_WORKBOOK_CHART_TYPE_AREA,
		runtimev1.WorkbookChartType_WORKBOOK_CHART_TYPE_PIE,
		runtimev1.WorkbookChartType_WORKBOOK_CHART_TYPE_DOUGHNUT,
		runtimev1.WorkbookChartType_WORKBOOK_CHART_TYPE_SCATTER,
	} {
		for _, stacked := range []bool{false, true} {
			req := testRequest(t)
			req.DashboardSheets[0].Blocks = []*runtimev1.WorkbookBlock{{
				Component: "chart", Title: typ.String(), Geometry: geo(0, 0, 600, 300),
				Content: &runtimev1.WorkbookBlock_Chart{Chart: &runtimev1.WorkbookChartBlock{
					DataSheetId:    "trend",
					Type:           typ,
					CategoryColumn: "week",
					Series:         []*runtimev1.WorkbookChartSeries{{Column: "spend"}, {Column: "cpa"}},
					Stacked:        stacked,
					PercentStacked: stacked,
				}},
			}}
			var buf bytes.Buffer
			require.NoError(t, Write(&buf, req), "%s stacked=%v", typ, stacked)
			require.Len(t, chartParts(t, buf.Bytes()), 1)
		}
	}
}

func TestWriteChartOptions(t *testing.T) {
	// A horizontal bar chart keeps the canvas order (first category at the top), and a hidden legend stays hidden.
	req := testRequest(t)
	req.DashboardSheets[0].Blocks = []*runtimev1.WorkbookBlock{{
		Component: "bars", Title: "Spend by week", Geometry: geo(0, 0, 600, 300),
		Content: &runtimev1.WorkbookBlock_Chart{Chart: &runtimev1.WorkbookChartBlock{
			DataSheetId:    "trend",
			Type:           runtimev1.WorkbookChartType_WORKBOOK_CHART_TYPE_BAR,
			CategoryColumn: "week",
			Series:         []*runtimev1.WorkbookChartSeries{{Column: "spend"}},
			LegendPosition: "none",
		}},
	}}
	var buf bytes.Buffer
	require.NoError(t, Write(&buf, req))
	charts := chartParts(t, buf.Bytes())
	require.Len(t, charts, 1)
	require.Contains(t, charts[0], `<orientation val="maxMin">`)
	valAx := charts[0][strings.Index(charts[0], "<valAx>"):]
	require.Contains(t, valAx, `<tickLblPos val="high">`)
	require.NotContains(t, charts[0], "<legend>")
}

func TestWriteTablePreviewLayout(t *testing.T) {
	// A leaderboard with two dimensions previews its data sheets side by side, each within its share of the block,
	// and a sheet that holds only the first rows of its source says so.
	req := testRequest(t)
	req.DataSheets[0].MoreRows = true
	req.DashboardSheets[0].Blocks = []*runtimev1.WorkbookBlock{{
		Component: "leaderboard", Title: "Top", Geometry: geo(0, 0, 600, 200),
		Content: &runtimev1.WorkbookBlock_Table{Table: &runtimev1.WorkbookTableBlock{DataSheetIds: []string{"lines", "trend"}}},
	}}
	var buf bytes.Buffer
	require.NoError(t, Write(&buf, req))
	f, err := excelize.OpenReader(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)
	defer f.Close()

	const dashboard = "Campaign report"
	cell := func(ref string) string {
		v, err := f.GetCellValue(dashboard, ref)
		require.NoError(t, err)
		return v
	}
	// The 600 px block spans nine columns, so each sheet gets four:
	// the label column narrows to two so the first number fits beside it.
	require.Equal(t, "Top", cell("A8"))
	require.Equal(t, "Line Item", cell("A10"))
	require.Equal(t, "Impressions", cell("C10"))
	require.Equal(t, "Display", cell("A11"))
	require.Equal(t, "43.2M", cell("C11"))
	require.Equal(t, "Full data: Data - Line items (2 rows), 2 more columns", cell("A13"))
	require.Equal(t, "Week ending", cell("F10"))
	require.Equal(t, "Client Spend", cell("H10"))
	require.Equal(t, "2026-01-04", cell("F11"))
	require.Equal(t, "Data: Data - Spend and CPA by week (first 3 rows), 1 more column", cell("F14"))
	merged, err := f.GetMergeCells(dashboard)
	require.NoError(t, err)
	var ranges []string
	for _, m := range merged {
		ranges = append(ranges, m.GetStartAxis()+":"+m.GetEndAxis())
	}
	require.Contains(t, ranges, "A10:B10")
	require.Contains(t, ranges, "C10:D10")
	require.Contains(t, ranges, "F10:G10")
	require.Contains(t, ranges, "H10:I10")
}

func TestWriteTotalsAndRowFormats(t *testing.T) {
	req := testRequest(t)
	// The line items carry a totals row; a KPIs-like sheet holds a different measure on each row.
	req.DataSheets[1].Totals = row(t, "Total", 44837645, nil, nil)
	// A nested pivot sends the domain of its innermost rows, which the formats use instead of the column's numbers.
	req.DataSheets[1].ConditionalFormats[0].Domain = &runtimev1.WorkbookValueRange{Min: 1000, Max: 50000000}
	req.DataSheets = append(req.DataSheets, &runtimev1.WorkbookDataSheet{
		Id:    "kpis",
		Title: "KPIs",
		Columns: []*runtimev1.WorkbookColumn{
			{Name: "measure", DisplayName: "Measure", Type: runtimev1.WorkbookColumnType_WORKBOOK_COLUMN_TYPE_STRING},
			{Name: "value", DisplayName: "Value", Type: runtimev1.WorkbookColumnType_WORKBOOK_COLUMN_TYPE_NUMBER},
			{Name: "delta_percent", DisplayName: "% change", Type: runtimev1.WorkbookColumnType_WORKBOOK_COLUMN_TYPE_NUMBER, NumberFormat: "0.0%"},
		},
		Rows: []*structpb.ListValue{
			row(t, "Spend", 1234.5, 0.25),
			row(t, "CTR", 0.0123, 0.5),
			row(t, "Bids", 98765, nil),
		},
		RowNumberFormats: []string{`"$"#,##0.00`, "0.00%"},
	})
	req.DashboardSheets[0].Blocks = []*runtimev1.WorkbookBlock{{
		Component: "table", Title: "Line items", Geometry: geo(0, 0, 1000, 200),
		Content: &runtimev1.WorkbookBlock_Table{Table: &runtimev1.WorkbookTableBlock{DataSheetIds: []string{"lines"}}},
	}}
	var buf bytes.Buffer
	require.NoError(t, Write(&buf, req))
	f, err := excelize.OpenReader(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)
	defer f.Close()
	cell := func(sheet, ref string) string {
		v, err := f.GetCellValue(sheet, ref)
		require.NoError(t, err)
		return v
	}
	bold := func(sheet, ref string) bool {
		id, err := f.GetCellStyle(sheet, ref)
		require.NoError(t, err)
		style, err := f.GetStyle(id)
		require.NoError(t, err)
		return style.Font != nil && style.Font.Bold
	}

	// The totals row sits directly below the table, outside its range, in bold.
	const lines = "Data - Line items"
	tables, err := f.GetTables(lines)
	require.NoError(t, err)
	require.Len(t, tables, 1)
	require.Equal(t, "A1:D3", tables[0].Range)
	require.Equal(t, "Total", cell(lines, "A4"))
	require.Equal(t, "44,837,645", cell(lines, "B4"))
	require.True(t, bold(lines, "A4"))
	require.True(t, bold(lines, "B4"))
	require.False(t, bold(lines, "B3"))

	// The preview shows the totals row first, in bold, and counts only the data rows.
	const dashboard = "Campaign report"
	require.Equal(t, "Line Item", cell(dashboard, "A10"))
	require.Equal(t, "Total", cell(dashboard, "A11"))
	require.Equal(t, "44.8M", cell(dashboard, "D11"))
	require.True(t, bold(dashboard, "A11"))
	require.Equal(t, "Display", cell(dashboard, "A12"))
	require.Equal(t, "Full data: Data - Line items (2 rows)", cell(dashboard, "A14"))

	// Conditional formats leave the totals row out, on the data sheet and in the preview, and use the given domain.
	formats, err := f.GetConditionalFormats(lines)
	require.NoError(t, err)
	require.Len(t, formats["B2:B3"], 1)
	require.Equal(t, "5E+07", formats["B2:B3"][0].MaxValue)
	formats, err = f.GetConditionalFormats(dashboard)
	require.NoError(t, err)
	require.Len(t, formats["D12:D13"], 1)
	require.Equal(t, "5E+07", formats["D12:D13"][0].MaxValue)

	// A row's format applies to number columns without a format of their own; the rest keep theirs.
	const kpis = "Data - KPIs"
	require.Equal(t, "$1,234.50", cell(kpis, "B2"))
	require.Equal(t, "25.0%", cell(kpis, "C2"))
	require.Equal(t, "1.23%", cell(kpis, "B3"))
	require.Equal(t, "50.0%", cell(kpis, "C3"))
	require.Equal(t, "98765", cell(kpis, "B4"))
}

func TestWritePreviewTotalsLastAndTextColumns(t *testing.T) {
	req := testRequest(t)
	// Totals shown after the rows, as a pivot with its totals at the bottom does.
	req.DataSheets[1].Totals = row(t, "Total", 44837645, nil, nil)
	// A nested pivot sends the domain of its innermost rows, which the formats use instead of the column's numbers.
	req.DataSheets[1].ConditionalFormats[0].Domain = &runtimev1.WorkbookValueRange{Min: 1000, Max: 50000000}
	req.DataSheets[1].TotalsLast = true
	// Two text columns before the first number: the label keeps its span rather than narrowing for a text column.
	req.DataSheets = append(req.DataSheets, &runtimev1.WorkbookDataSheet{
		Id:    "places",
		Title: "Places",
		Columns: []*runtimev1.WorkbookColumn{
			{Name: "country", DisplayName: "Country", Type: runtimev1.WorkbookColumnType_WORKBOOK_COLUMN_TYPE_STRING},
			{Name: "city", DisplayName: "City", Type: runtimev1.WorkbookColumnType_WORKBOOK_COLUMN_TYPE_STRING},
			{Name: "spend", DisplayName: "Spend", Type: runtimev1.WorkbookColumnType_WORKBOOK_COLUMN_TYPE_NUMBER},
		},
		Rows: []*structpb.ListValue{row(t, "United States of America", "San Francisco", 10)},
	})
	req.DashboardSheets[0].Blocks = []*runtimev1.WorkbookBlock{
		{
			Component: "lines", Title: "Line items", Geometry: geo(0, 0, 1000, 200),
			Content: &runtimev1.WorkbookBlock_Table{Table: &runtimev1.WorkbookTableBlock{DataSheetIds: []string{"lines"}}},
		},
		{
			// 120 px tall: one row for data after the title, blank row, header and link.
			Component: "short", Title: "Short", Geometry: geo(0, 220, 1000, 120),
			Content: &runtimev1.WorkbookBlock_Table{Table: &runtimev1.WorkbookTableBlock{DataSheetIds: []string{"lines"}}},
		},
		{
			Component: "places", Title: "Places", Geometry: geo(0, 360, 300, 120),
			Content: &runtimev1.WorkbookBlock_Table{Table: &runtimev1.WorkbookTableBlock{DataSheetIds: []string{"places"}}},
		},
	}
	var buf bytes.Buffer
	require.NoError(t, Write(&buf, req))
	f, err := excelize.OpenReader(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)
	defer f.Close()
	const dashboard = "Campaign report"
	cell := func(ref string) string {
		v, err := f.GetCellValue(dashboard, ref)
		require.NoError(t, err)
		return v
	}

	// The data rows come first and the totals row after them; banding starts on the first data row.
	require.Equal(t, "Line Item", cell("A10"))
	require.Equal(t, "Display", cell("A11"))
	require.Equal(t, "Video", cell("A12"))
	require.Equal(t, "Total", cell("A13"))
	require.Equal(t, "Full data: Data - Line items (2 rows)", cell("A14"))
	first, err := f.GetCellStyle(dashboard, "A11")
	require.NoError(t, err)
	style, err := f.GetStyle(first)
	require.NoError(t, err)
	require.Equal(t, []string{"D9E1F2"}, style.Fill.Color)

	// A block with room for one row still shows the totals row.
	require.Equal(t, "Short", cell("A19"))
	require.Equal(t, "Total", cell("A22"))
	require.Equal(t, "Full data: Data - Line items (2 rows)", cell("A23"))

	// The 300 px block spans four columns: the country keeps three, so only it fits.
	require.Equal(t, "Country", cell("A28"))
	require.Equal(t, "United States of America", cell("A29"))
	require.Equal(t, "Full data: Data - Places (1 rows), 2 more columns", cell("A30"))
	merged, err := f.GetMergeCells(dashboard)
	require.NoError(t, err)
	var ranges []string
	for _, m := range merged {
		ranges = append(ranges, m.GetStartAxis()+":"+m.GetEndAxis())
	}
	require.Contains(t, ranges, "A28:C28")
}

func TestWriteScatterMarkers(t *testing.T) {
	// Scatter points get an explicit fill: without one, LibreOffice draws no points.
	req := testRequest(t)
	req.DashboardSheets[0].Blocks = []*runtimev1.WorkbookBlock{{
		Component: "scatter", Title: "Spend vs CPA", Geometry: geo(0, 0, 600, 300),
		Content: &runtimev1.WorkbookBlock_Chart{Chart: &runtimev1.WorkbookChartBlock{
			DataSheetId:    "trend",
			Type:           runtimev1.WorkbookChartType_WORKBOOK_CHART_TYPE_SCATTER,
			CategoryColumn: "spend",
			Series:         []*runtimev1.WorkbookChartSeries{{Column: "cpa"}},
		}},
	}}
	var buf bytes.Buffer
	require.NoError(t, Write(&buf, req))
	charts := chartParts(t, buf.Bytes())
	require.Len(t, charts, 1)
	marker := charts[0][strings.Index(charts[0], "<marker>"):strings.Index(charts[0], "</marker>")]
	require.Contains(t, marker, `<a:srgbClr val="4472C4">`)
}

func TestWriteNarrowBlocks(t *testing.T) {
	req := testRequest(t)
	kpi := req.DashboardSheets[0].Blocks[3]
	// Two KPIs in a 300 px block get two columns each, too narrow for "Previous quarter" beside a number.
	kpi.Geometry = geo(0, 0, 300, 200)
	note := req.DashboardSheets[0].Blocks[6]
	note.Title = "Funnel of bids to clicks, by stage"
	note.Geometry = geo(0, 220, 600, 60)
	req.DashboardSheets[0].Blocks = []*runtimev1.WorkbookBlock{kpi, note}
	var buf bytes.Buffer
	require.NoError(t, Write(&buf, req))
	f, err := excelize.OpenReader(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)
	defer f.Close()

	const dashboard = "Campaign report"
	cell := func(ref string) string {
		v, err := f.GetCellValue(dashboard, ref)
		require.NoError(t, err)
		return v
	}
	require.Equal(t, "Impressions", cell("A9"))
	require.Equal(t, "Previous", cell("A11"))
	require.Equal(t, "41.2M", cell("B11"))
	// A note's heading spans the block, not just the first three columns.
	require.Equal(t, "Funnel of bids to clicks, by stage", cell("A19"))
	merged, err := f.GetMergeCells(dashboard)
	require.NoError(t, err)
	var ranges []string
	for _, m := range merged {
		ranges = append(ranges, m.GetStartAxis()+":"+m.GetEndAxis())
	}
	require.Contains(t, ranges, "A19:I19")
}

func TestWriteTextParts(t *testing.T) {
	// A markdown component's text and tables are laid out in reading order within its block,
	// and a table whose rows hold different measures takes each row's number format.
	textPart := func(text string) *runtimev1.WorkbookTextPart {
		return &runtimev1.WorkbookTextPart{Part: &runtimev1.WorkbookTextPart_Text{Text: text}}
	}
	tablePart := func(id string) *runtimev1.WorkbookTextPart {
		return &runtimev1.WorkbookTextPart{Part: &runtimev1.WorkbookTextPart_DataSheetId{DataSheetId: id}}
	}
	req := testRequest(t)
	req.DataSheets = append(req.DataSheets, &runtimev1.WorkbookDataSheet{
		Id:    "scorecard",
		Title: "Scorecard",
		Columns: []*runtimev1.WorkbookColumn{
			{Name: "col_1", DisplayName: "Metric", Type: runtimev1.WorkbookColumnType_WORKBOOK_COLUMN_TYPE_STRING},
			{Name: "col_2", DisplayName: "Actual", Type: runtimev1.WorkbookColumnType_WORKBOOK_COLUMN_TYPE_NUMBER},
			{Name: "col_3", DisplayName: "Progress", Type: runtimev1.WorkbookColumnType_WORKBOOK_COLUMN_TYPE_NUMBER, NumberFormat: "0.0%"},
		},
		Rows: []*structpb.ListValue{
			row(t, "Spend", 1234567.891, 0.833),
			row(t, "Deals", 12, 0.8),
			row(t, "Win rate", 0.25, "n/a"),
		},
		RowNumberFormats: []string{`"$"#,##0.00`, "#,##0", "0.0%"},
	})
	req.DashboardSheets[0].Blocks = []*runtimev1.WorkbookBlock{
		{
			Component: "scorecard", Geometry: geo(0, 0, 600, 300),
			Content: &runtimev1.WorkbookBlock_Text{Text: &runtimev1.WorkbookTextBlock{Parts: []*runtimev1.WorkbookTextPart{
				textPart("Scorecard\nFiscal year 2026"),
				tablePart("scorecard"),
				textPart("Targets are set per quarter."),
			}}},
		},
		{
			// Shorter than its table's preview.
			Component: "short", Geometry: geo(620, 0, 300, 80),
			Content: &runtimev1.WorkbookBlock_Text{Text: &runtimev1.WorkbookTextBlock{Parts: []*runtimev1.WorkbookTextPart{tablePart("lines")}}},
		},
		{
			// Cards that sit side by side on the dashboard read as more lines than the component is tall.
			Component: "cards", Geometry: geo(0, 320, 400, 60),
			Content: &runtimev1.WorkbookBlock_Text{Text: &runtimev1.WorkbookTextBlock{Text: "Revenue\n$1.1M\nDeals\n140\nWin rate\n30%"}},
		},
		{
			// Beside it, more cards that need more rows still.
			Component: "more-cards", Geometry: geo(840, 320, 400, 60),
			Content: &runtimev1.WorkbookBlock_Text{Text: &runtimev1.WorkbookTextBlock{Text: "Spend\n$40K\nClicks\n9.1K\nCPC\n$4.40\nCTR\n1.2%"}},
		},
		{
			Component: "styles-only", Geometry: geo(420, 320, 400, 100),
			Content: &runtimev1.WorkbookBlock_Text{Text: &runtimev1.WorkbookTextBlock{}},
		},
		{
			Component: "below", Title: "Below", Geometry: geo(0, 400, 400, 60),
			Content: &runtimev1.WorkbookBlock_Note{Note: &runtimev1.WorkbookNoteBlock{Text: "No data"}},
		},
		{
			// More lines than the 500 rows the cap on a block's side would give.
			Component: "long", Geometry: geo(0, 480, 400, 60),
			Content: &runtimev1.WorkbookBlock_Text{Text: &runtimev1.WorkbookTextBlock{Text: strings.Repeat("Line\n", 519) + "Last line"}},
		},
		{
			Component: "after", Title: "After", Geometry: geo(0, 560, 400, 60),
			Content: &runtimev1.WorkbookBlock_Note{Note: &runtimev1.WorkbookNoteBlock{Text: "No data"}},
		},
		{
			// More text than a cell holds.
			Component: "huge", Geometry: geo(0, 640, 400, 60),
			Content: &runtimev1.WorkbookBlock_Text{Text: &runtimev1.WorkbookTextBlock{Text: strings.Repeat("Line\n", 7000)}},
		},
		{
			Component: "end", Title: "End", Geometry: geo(0, 720, 400, 60),
			Content: &runtimev1.WorkbookBlock_Note{Note: &runtimev1.WorkbookNoteBlock{Text: "No data"}},
		},
	}
	var buf bytes.Buffer
	require.NoError(t, Write(&buf, req))
	f, err := excelize.OpenReader(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)
	defer f.Close()

	const dashboard = "Campaign report"
	cell := func(sheet, ref string) string {
		v, err := f.GetCellValue(sheet, ref)
		require.NoError(t, err)
		return v
	}
	merged, err := f.GetMergeCells(dashboard)
	require.NoError(t, err)
	var ranges []string
	for _, m := range merged {
		ranges = append(ranges, m.GetStartAxis()+":"+m.GetEndAxis())
	}

	// The 600 by 300 px block spans nine columns and fourteen rows from row 8.
	// The text above the table takes a row per line.
	require.Equal(t, "Scorecard\nFiscal year 2026", cell(dashboard, "A8"))
	require.Contains(t, ranges, "A8:I9")
	// The table takes its header, its three rows and the link, with compact numbers in each row's format.
	require.Equal(t, "Metric", cell(dashboard, "A10"))
	require.Equal(t, "Actual", cell(dashboard, "D10"))
	require.Equal(t, "Progress", cell(dashboard, "F10"))
	require.Equal(t, "Spend", cell(dashboard, "A11"))
	require.Equal(t, "$1.2M", cell(dashboard, "D11"))
	require.Equal(t, "83.3%", cell(dashboard, "F11"))
	require.Equal(t, "25.0%", cell(dashboard, "D13"))
	// Text in a number column stays text.
	require.Equal(t, "n/a", cell(dashboard, "F13"))
	require.Equal(t, "Full data: Data - Scorecard (3 rows)", cell(dashboard, "A14"))
	ok, link, err := f.GetCellHyperLink(dashboard, "A14")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "'Data - Scorecard'!A1", link)
	// A blank row, then the text below the table fills the rest of the block.
	require.Equal(t, "", cell(dashboard, "A15"))
	require.Equal(t, "Targets are set per quarter.", cell(dashboard, "A16"))
	require.Contains(t, ranges, "A16:I21")

	// A text block grows to fit its content instead of cutting it off, and pushes the blocks below it down.
	// The 80 px block has three rows where its table's preview needs four, so the blocks below start a row later.
	require.Equal(t, "Line Item", cell(dashboard, "J8"))
	require.Equal(t, "Display", cell(dashboard, "J9"))
	require.Equal(t, "43.2M", cell(dashboard, "L9"))
	require.Equal(t, "Video", cell(dashboard, "J10"))
	require.Equal(t, "Full data: Data - Line items (2 rows), 2 more columns", cell(dashboard, "J11"))
	// The 60 px blocks have two rows each, for six and eight lines. They grow four and six rows,
	// and the block 80 px below their tops moves by the larger growth, so it starts six more rows later.
	require.Equal(t, "Revenue\n$1.1M\nDeals\n140\nWin rate\n30%", cell(dashboard, "A25"))
	require.Contains(t, ranges, "A25:E30")
	require.Equal(t, "Spend\n$40K\nClicks\n9.1K\nCPC\n$4.40\nCTR\n1.2%", cell(dashboard, "N25"))
	require.Contains(t, ranges, "N25:R32")
	require.Equal(t, "Below", cell(dashboard, "A35"))
	// A long text keeps every one of its 520 rows, and the block below it starts after the last of them.
	// It has more line feeds than a cell shows, so it fills cells of 254 lines, one below the other.
	require.Contains(t, ranges, "A39:E292")
	require.Contains(t, ranges, "A293:E546")
	require.Equal(t, strings.Repeat("Line\n", 11)+"Last line", cell(dashboard, "A547"))
	require.Contains(t, ranges, "A547:E558")
	require.Equal(t, "After", cell(dashboard, "A561"))
	// A text longer still fills more such cells and moves the blocks below it as far.
	require.Equal(t, strings.Repeat("Line\n", 253)+"Line", cell(dashboard, "A565"))
	require.Contains(t, ranges, "A565:E818")
	require.Equal(t, strings.Repeat("Line\n", 253)+"Line", cell(dashboard, "A819"))
	require.Equal(t, strings.Repeat("Line\n", 142), cell(dashboard, "A7423"))
	require.Contains(t, ranges, "A7423:E7565")
	require.Equal(t, "End", cell(dashboard, "A7568"))

	// A block without text or tables leaves its cells alone instead of merging them into one empty cell.
	require.Equal(t, "", cell(dashboard, "G25"))
	for _, r := range ranges {
		require.False(t, strings.HasPrefix(r, "G25:"), "unexpected merge %s", r)
	}

	const data = "Data - Scorecard"
	require.Equal(t, "$1,234,567.89", cell(data, "B2"))
	require.Equal(t, "12", cell(data, "B3"))
	require.Equal(t, "25.0%", cell(data, "B4"))
	require.Equal(t, "83.3%", cell(data, "C2"))
	require.Equal(t, "n/a", cell(data, "C4"))
}

func TestCellTexts(t *testing.T) {
	long := strings.Repeat("x", excelize.TotalCellChars)
	tests := []struct {
		name string
		text string
		want []string
	}{
		{"fits", "a\nb", []string{"a\nb"}},
		{"line feeds", strings.Repeat("a\n", 300), []string{strings.Repeat("a\n", 253) + "a", strings.Repeat("a\n", 46)}},
		{"characters", long[:20_000] + "\n" + long[:20_000], []string{long[:20_000], long[:20_000]}},
		{"long line", "a\n" + long + "yz\nb", []string{"a", long, "yz\nb"}},
		// A character outside the Basic Multilingual Plane takes two UTF-16 units and is never split.
		{"surrogate pair", long[1:] + "😀", []string{long[1:], "😀"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, cellTexts(tt.text))
		})
	}
}

func TestWriteLongCanvas(t *testing.T) {
	// Blocks far down a long canvas keep their own position instead of piling up at a cap.
	req := testRequest(t)
	var blocks []*runtimev1.WorkbookBlock
	for i := range 30 {
		blocks = append(blocks, &runtimev1.WorkbookBlock{
			Component: fmt.Sprintf("markdown-%d", i),
			Geometry:  geo(0, uint32(i*450), 600, 440),
			Content:   &runtimev1.WorkbookBlock_Text{Text: &runtimev1.WorkbookTextBlock{Text: fmt.Sprintf("Row %d", i+1)}},
		})
	}
	req.DashboardSheets[0].Blocks = blocks
	var buf bytes.Buffer
	require.NoError(t, Write(&buf, req))
	f, err := excelize.OpenReader(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)
	defer f.Close()

	rows, err := f.GetRows("Campaign report")
	require.NoError(t, err)
	found := map[string]int{}
	for i, r := range rows {
		for _, v := range r {
			if strings.HasPrefix(v, "Row ") {
				found[v] = i
			}
		}
	}
	require.Len(t, found, 30)
	// 450 px between blocks is 22.5 grid rows of 20 px.
	require.InDelta(t, 29*450/20, found["Row 30"]-found["Row 1"], 1)
}

func TestValidate(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(req *runtimev1.CanvasWorkbookExportRequest)
		want   string
	}{
		{
			name:   "no dashboard sheet",
			mutate: func(req *runtimev1.CanvasWorkbookExportRequest) { req.DashboardSheets = nil },
			want:   "at least one dashboard sheet",
		},
		{
			name: "block beyond the sheet",
			mutate: func(req *runtimev1.CanvasWorkbookExportRequest) {
				req.DashboardSheets[0].Blocks[0].Geometry.YPx = maxBlockYPx + 1
			},
			want: "beyond the edge of the sheet",
		},
		{
			name: "conditional format of an unknown column",
			mutate: func(req *runtimev1.CanvasWorkbookExportRequest) {
				req.DataSheets[1].ConditionalFormats[0].Columns = []string{"clicks"}
			},
			want: `references unknown column "clicks"`,
		},
		{
			name: "conditional format with a bad color",
			mutate: func(req *runtimev1.CanvasWorkbookExportRequest) {
				req.DataSheets[1].ConditionalFormats[0].Format.GetDataBar().Color = "blue"
			},
			want: `color "blue" is not a hex RGB color`,
		},
		{
			name: "conditional format with a domain out of order",
			mutate: func(req *runtimev1.CanvasWorkbookExportRequest) {
				req.DataSheets[1].ConditionalFormats[0].Domain = &runtimev1.WorkbookValueRange{Min: 2, Max: 1}
			},
			want: "has an invalid domain",
		},
		{
			name:   "duplicate data sheet id",
			mutate: func(req *runtimev1.CanvasWorkbookExportRequest) { req.DataSheets[1].Id = "trend" },
			want:   `data sheet id "trend" is used more than once`,
		},
		{
			name: "too many columns",
			mutate: func(req *runtimev1.CanvasWorkbookExportRequest) {
				ds := req.DataSheets[2]
				for i := 0; i <= MaxColumnsPerSheet; i++ {
					ds.Columns = append(ds.Columns, &runtimev1.WorkbookColumn{Name: strings.Repeat("c", i+2)})
				}
			},
			want: "exceeds the limit of 200",
		},
		{
			name: "too many cells",
			mutate: func(req *runtimev1.CanvasWorkbookExportRequest) {
				ds := req.DataSheets[2]
				for i := 0; i < 10; i++ {
					ds.Columns = append(ds.Columns, &runtimev1.WorkbookColumn{Name: strings.Repeat("c", i+2)})
				}
				ds.Rows = make([]*structpb.ListValue, MaxCells/len(ds.Columns)+1)
			},
			want: "more than 1000000 cells",
		},
		{
			name: "row wider than the sheet",
			mutate: func(req *runtimev1.CanvasWorkbookExportRequest) {
				req.DataSheets[2].Rows = []*structpb.ListValue{row(t, "a", "b")}
			},
			want: "has 2 values but the sheet has 1 columns",
		},
		{
			name: "totals row wider than the sheet",
			mutate: func(req *runtimev1.CanvasWorkbookExportRequest) {
				req.DataSheets[2].Totals = row(t, "a", "b")
			},
			want: "the totals row of data sheet \"Empty\" has 2 values",
		},
		{
			name: "more row formats than rows",
			mutate: func(req *runtimev1.CanvasWorkbookExportRequest) {
				req.DataSheets[2].RowNumberFormats = []string{"0.0%"}
			},
			want: "has 1 row number formats but 0 rows",
		},
		{
			name: "unknown category column",
			mutate: func(req *runtimev1.CanvasWorkbookExportRequest) {
				req.DashboardSheets[0].Blocks[0].GetChart().CategoryColumn = "nope"
			},
			want: `unknown category column "nope"`,
		},
		{
			name: "combo series without mark",
			mutate: func(req *runtimev1.CanvasWorkbookExportRequest) {
				req.DashboardSheets[0].Blocks[0].GetChart().Series[0].Mark = runtimev1.WorkbookSeriesMark_WORKBOOK_SERIES_MARK_UNSPECIFIED
			},
			want: "series without a mark",
		},
		{
			name: "image is not a png",
			mutate: func(req *runtimev1.CanvasWorkbookExportRequest) {
				req.DashboardSheets[0].Blocks[2].GetImage().Png = []byte("GIF89a")
			},
			want: `image "snapshot" is not a PNG`,
		},
		{
			name: "table references unknown sheet",
			mutate: func(req *runtimev1.CanvasWorkbookExportRequest) {
				req.DashboardSheets[0].Blocks[4].GetTable().DataSheetIds = []string{"nope"}
			},
			want: `unknown data sheet "nope"`,
		},
		{
			name: "text references unknown sheet",
			mutate: func(req *runtimev1.CanvasWorkbookExportRequest) {
				req.DashboardSheets[0].Blocks[5].GetText().Parts = []*runtimev1.WorkbookTextPart{
					{Part: &runtimev1.WorkbookTextPart_DataSheetId{DataSheetId: "nope"}},
				}
			},
			want: `text "md" references unknown data sheet "nope"`,
		},
		{
			name: "kpi trend column unknown",
			mutate: func(req *runtimev1.CanvasWorkbookExportRequest) {
				req.DashboardSheets[0].Blocks[3].GetKpi().Values[0].TrendColumn = "nope"
			},
			want: `unknown trend column "nope"`,
		},
		{
			name: "block without size",
			mutate: func(req *runtimev1.CanvasWorkbookExportRequest) {
				req.DashboardSheets[0].Blocks[5].Geometry = nil
			},
			want: `block "md" has no size`,
		},
		{
			name: "block without content",
			mutate: func(req *runtimev1.CanvasWorkbookExportRequest) {
				req.DashboardSheets[0].Blocks[5].Content = nil
			},
			want: `block "md" has no content`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := testRequest(t)
			tc.mutate(req)
			err := Write(io.Discard, req)
			require.Error(t, err)
			require.True(t, IsValidationError(err), "expected a validation error, got %v", err)
			require.Contains(t, err.Error(), tc.want)
		})
	}
}

func TestCompactNumberFormat(t *testing.T) {
	require.Equal(t, `[>=1000000]#,##0.0,,"M";[>=1000]#,##0.0,"k";#,##0`, compactNumberFormat("#,##0"))
	require.Equal(t, `[>=1000000]"$"#,##0.0,,"M";[>=1000]"$"#,##0.0,"k";"$"#,##0.00`, compactNumberFormat(`"$"#,##0.00`))
	require.Equal(t, `[>=1000000]#,##0.0,,"M";[>=1000]#,##0.0,"k";General`, compactNumberFormat("General"))
	require.Equal(t, `[>=1000000]#,##0.0,,"M""€";[>=1000]#,##0.0,"k""€";#,##0.00"€"`, compactNumberFormat(`#,##0.00"€"`))
	require.Equal(t, "0.0%", compactNumberFormat("0.0%"))
	require.Equal(t, "yyyy-mm-dd", compactNumberFormat("yyyy-mm-dd"))
	require.Equal(t, "", compactNumberFormat(""))
}
