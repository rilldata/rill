package driverutil_test

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/driverutil"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"

	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
)

const sheet = "Sheet1"

func TestWriteXLSX(t *testing.T) {
	rt, instanceID := testruntime.NewInstance(t)
	olap, release, err := rt.OLAP(t.Context(), instanceID, "")
	require.NoError(t, err)
	defer release()

	// writeXLSX writes the result of sql in an XLSX file and returns the file and its raw cell values.
	writeXLSX := func(t *testing.T, sql string, opts driverutil.FileOptions) (*excelize.File, [][]string) {
		res, err := olap.Query(t.Context(), &drivers.Statement{Query: sql})
		require.NoError(t, err)
		defer res.Close()
		var buf bytes.Buffer
		require.NoError(t, driverutil.ResultToFile(res, &buf, drivers.FileFormatXLSX, opts))
		f, err := excelize.OpenReader(&buf)
		require.NoError(t, err)
		t.Cleanup(func() { _ = f.Close() })
		rows, err := f.GetRows(sheet, excelize.Options{RawCellValue: true})
		require.NoError(t, err)
		return f, rows
	}
	// numFmt returns the number format of a cell: a built-in format ID or a custom format.
	numFmt := func(t *testing.T, f *excelize.File, cell string) (int, string) {
		id, err := f.GetCellStyle(sheet, cell)
		require.NoError(t, err)
		style, err := f.GetStyle(id)
		require.NoError(t, err)
		if style.CustomNumFmt != nil {
			return style.NumFmt, *style.CustomNumFmt
		}
		return style.NumFmt, ""
	}

	t.Run("formats every row, not only the first ones", func(t *testing.T) {
		// The last row, past the rows read ahead to size the columns, holds a small share, a long label and a time of day.
		f, rows := writeXLSX(t, `
SELECT
	i AS n,
	IF(i < 149, 1.0, 0.0004)::DOUBLE AS share,
	IF(i < 149, 1.0, 0.0004)::DOUBLE AS amount,
	IF(i < 149, 'short', repeat('long label ', 10)) AS label,
	IF(i < 149, TIMESTAMP '2024-01-01' + INTERVAL (i) DAY, TIMESTAMP '2024-06-01 13:45:00') AS ts
FROM range(150) t(i)
ORDER BY i`, driverutil.FileOptions{NumberFormats: []driverutil.XLSXNumberFormat{
			{Column: 0},
			{Column: 1, Format: driverutil.MeasureFormat{Preset: "percentage"}},
			{Column: 2},
		}})

		// Every row, in order.
		require.Len(t, rows, 151)
		for i, row := range rows[1:] {
			require.Equal(t, strconv.Itoa(i), row[0])
		}

		// Integers or decimals by the column's type, not by the first values.
		_, format := numFmt(t, f, "A151")
		require.Equal(t, "#,##0", format)
		_, format = numFmt(t, f, "C151")
		require.Equal(t, "#,##0.00", format)
		// Each percentage gets the decimals it needs.
		_, format = numFmt(t, f, "B2")
		require.Equal(t, "0.0%", format)
		_, format = numFmt(t, f, "B151")
		require.Equal(t, "0.000%", format)
		// A timestamp at midnight UTC is a date, and one with a time of day a date and time.
		id, _ := numFmt(t, f, "E2")
		require.Equal(t, 14, id)
		id, _ = numFmt(t, f, "E151")
		require.Equal(t, 22, id)

		// The columns are sized to the first rows: the long label isn't among them.
		width, err := f.GetColWidth(sheet, "D")
		require.NoError(t, err)
		require.Equal(t, 10.0, width)
		// Dates with times get the width of their format.
		width, err = f.GetColWidth(sheet, "E")
		require.NoError(t, err)
		require.Equal(t, 18.0, width)
	})

	t.Run("reads ahead a bounded amount of data", func(t *testing.T) {
		// Each row holds 10 KB, so the rows read ahead stop well before the 40th, whose label is long.
		f, rows := writeXLSX(t, `
SELECT repeat('x', 10000) AS blob, IF(i < 40, 'short', repeat('long label ', 4)) AS label
FROM range(60) t(i)
ORDER BY i`, driverutil.FileOptions{})
		require.Len(t, rows, 61)
		width, err := f.GetColWidth(sheet, "B")
		require.NoError(t, err)
		require.Equal(t, 10.0, width)
		width, err = f.GetColWidth(sheet, "A")
		require.NoError(t, err)
		require.Equal(t, 60.0, width)
	})

	t.Run("dates", func(t *testing.T) {
		f, rows := writeXLSX(t, `
SELECT * FROM (VALUES
	(DATE '1850-05-01', TIMESTAMP '1850-05-01 00:00:00', TIMESTAMP '1850-05-01 06:00:00'),
	(DATE '2024-02-10', TIMESTAMP '2024-02-10 10:30:00', TIMESTAMP '2024-02-10 05:00:00'),
	(NULL, NULL, NULL)
) t(d, ts, lima_day)`, driverutil.FileOptions{TimeColumns: []driverutil.XLSXTimeColumn{
			{Column: 2, Date: true, TimeZone: "America/Lima"},
		}})

		// Excel can't hold dates before 1900: they keep their text, without a format.
		require.Equal(t, []string{"1850-05-01", "1850-05-01 00:00:00", "1850-05-01"}, rows[1])
		for _, cell := range []string{"A2", "B2", "C2"} {
			id, format := numFmt(t, f, cell)
			require.Zero(t, id, cell)
			require.Empty(t, format, cell)
		}

		// A date column, a timestamp with a time of day, and a day in Lima, which starts at 05:00 UTC.
		require.Equal(t, []string{"45332", "45332.4375", "45332"}, rows[2])
		for cell, want := range map[string]int{"A3": 14, "B3": 22, "C3": 14} {
			id, _ := numFmt(t, f, cell)
			require.Equal(t, want, id, cell)
		}

		// Nulls are empty cells.
		require.Len(t, rows, 3)
	})

	t.Run("empty result", func(t *testing.T) {
		f, rows := writeXLSX(t, "SELECT 1 AS a, 'x' AS b WHERE false", driverutil.FileOptions{Headers: []string{"Title"}})
		require.Equal(t, [][]string{{"Title"}, {"a", "b"}}, rows)
		// An Excel table needs at least one data row.
		tables, err := f.GetTables(sheet)
		require.NoError(t, err)
		require.Empty(t, tables)
	})

	t.Run("header lines", func(t *testing.T) {
		f, rows := writeXLSX(t, "SELECT 1 AS a", driverutil.FileOptions{Headers: []string{"Title", "Filters: none", ""}})
		require.Equal(t, [][]string{{"Title"}, {"Filters: none"}, nil, {"a"}, {"1"}}, rows)
		for i, want := range []string{"bold", "muted", ""} {
			cell := fmt.Sprintf("A%d", i+1)
			id, err := f.GetCellStyle(sheet, cell)
			require.NoError(t, err)
			style, err := f.GetStyle(id)
			require.NoError(t, err)
			switch want {
			case "bold":
				require.True(t, style.Font.Bold, cell)
			case "muted":
				require.Equal(t, "6B7280", style.Font.Color, cell)
			default:
				require.Zero(t, id, cell)
			}
		}
	})
}
