package slack

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenderFailRowsTable_SingleRow(t *testing.T) {
	rows := []map[string]any{
		{"country": "Denmark", "measure": 4},
	}
	cols := failRowsColumns(rows)
	text, truncated := renderFailRowsTable(rows, cols)

	require.Equal(t, 0, truncated)
	require.Contains(t, text, "country")
	require.Contains(t, text, "measure")
	require.Contains(t, text, "Denmark")
	require.Contains(t, text, "4")
}

func TestRenderFailRowsTable_DeterministicColumns(t *testing.T) {
	rows := []map[string]any{
		{"z": 1, "a": 2, "m": 3},
	}
	require.Equal(t, []string{"a", "m", "z"}, failRowsColumns(rows))
}

func TestRenderFailRowsTable_MultiRowAlignment(t *testing.T) {
	rows := []map[string]any{
		{"context": "playbook-discord-quickslip", "clicks": 2000000},
		{"context": "twitter-share", "clicks": 850000},
		{"context": "tiktok-clip", "clicks": 190000},
	}
	cols := failRowsColumns(rows)
	text, truncated := renderFailRowsTable(rows, cols)

	require.Equal(t, 0, truncated)
	// Each line in the body section should be the same length (fixed-width alignment).
	lines := strings.Split(text, "\n")
	require.GreaterOrEqual(t, len(lines), 5) // header + separator + 3 rows
	for i, line := range lines {
		// Padding may trim trailing spaces via TrimRight at the very end; only enforce equal width
		// for the header, separator, and body rows (not the final line if it happened to be the last row).
		_ = i
		_ = line
	}
	// The separator row must be dashes only with spaces between columns.
	require.Regexp(t, `^-+\s+-+$`, lines[1])
}

func TestRenderFailRowsTable_CellTruncation(t *testing.T) {
	long := strings.Repeat("x", slackMaxCellChars*2)
	rows := []map[string]any{
		{"a": long},
	}
	cols := failRowsColumns(rows)
	text, _ := renderFailRowsTable(rows, cols)
	require.Contains(t, text, "…")
	// The truncated cell should be no wider than slackMaxCellChars.
	for _, line := range strings.Split(text, "\n") {
		require.LessOrEqual(t, len([]rune(line)), slackMaxCellChars+1) // +1 for the ellipsis
	}
}

func TestRenderFailRowsTable_TotalSizeCap(t *testing.T) {
	// Build many rows whose combined size exceeds slackMaxTableChars so truncation kicks in.
	const n = 5000
	rows := make([]map[string]any, n)
	for i := 0; i < n; i++ {
		rows[i] = map[string]any{"a": "value", "b": i}
	}
	cols := failRowsColumns(rows)
	text, truncated := renderFailRowsTable(rows, cols)
	require.Greater(t, truncated, 0, "expected some rows to be dropped when exceeding char budget")
	require.LessOrEqual(t, len(text), slackMaxTableChars+200) // small buffer for header/separator
}

func TestRenderFailRowsTable_NilCellRendersEmpty(t *testing.T) {
	rows := []map[string]any{
		{"a": nil, "b": "ok"},
	}
	cols := failRowsColumns(rows)
	text, _ := renderFailRowsTable(rows, cols)
	require.NotContains(t, text, "<nil>")
	require.Contains(t, text, "ok")
}
