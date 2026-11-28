package workbook

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSheetName(t *testing.T) {
	n := newNamer()
	require.Equal(t, "Data - SpendCPA by week", n.sheetName(`Data - Spend/CPA: by [week]?*`))
	require.Equal(t, "Sheet", n.sheetName("   "))
	require.Equal(t, "History sheet", n.sheetName("history"))
	require.Equal(t, "Quoted", n.sheetName("'Quoted'"))
	// Cutting a name to length can expose an apostrophe at its end, which Excel rejects.
	require.Equal(t, "Data - Revenue by the customer", n.sheetName("Data - Revenue by the customer's region"))

	long := strings.Repeat("abcdefghij", 4)
	require.Equal(t, long[:31], n.sheetName(long))
	// Duplicates get a numbered suffix that still fits the 31-character limit.
	require.Equal(t, long[:27]+" (2)", n.sheetName(long))

	require.Equal(t, "Overview", n.sheetName("Overview"))
	require.Equal(t, "overview (2)", n.sheetName("overview"))
	require.Equal(t, "Overview (3)", n.sheetName("Overview"))
}

func TestTableName(t *testing.T) {
	n := newNamer()
	require.Equal(t, "Table_Spend_and_CPA_by_week", n.tableName("Spend and CPA by week"))
	require.Equal(t, "Table_100_stacked_test", n.tableName("100% stacked (test) "))
	require.Equal(t, "Table", n.tableName("—"))
	require.Equal(t, "Table_2", n.tableName(""))
	require.Equal(t, "Table_spend_and_cpa_by_week_2", n.tableName("spend and cpa by week"))
}

func TestQuoteSheet(t *testing.T) {
	require.Equal(t, "'Data - Mike''s view'", quoteSheet("Data - Mike's view"))
}
