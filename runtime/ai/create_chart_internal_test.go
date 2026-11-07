package ai

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestChartMarkdownLongValues checks that the markdown fallback caps long cell values,
// so that a dimension with long values does not make the text of a result outgrow the chart data.
func TestChartMarkdownLongValues(t *testing.T) {
	long := strings.Repeat("x", 5000)
	text := chartMarkdown(&CreateChartResult{
		ChartType: "Bar Chart",
		Title:     "Bids by URL",
		ChartDataResult: &ChartDataResult{
			Data: []map[string]any{{"url": long, "bids": float64(3)}},
			Fields: map[string]*ChartField{
				"url":  {Name: "url", Kind: "dimension", DisplayName: "URL"},
				"bids": {Name: "bids", Kind: "measure", DisplayName: "Bids"},
			},
		},
	})
	require.NotContains(t, text, long)
	require.Contains(t, text, "| "+strings.Repeat("x", chartMarkdownCellRunes-1)+"… | 3 |")
	require.Less(t, len(text), 500)
}
