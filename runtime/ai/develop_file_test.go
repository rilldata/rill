package ai_test

import (
	"context"
	"strings"
	"testing"

	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

// TestDevelopFileMetricsViewFieldNamesInPrompt verifies that the develop_file prompt includes the
// exact dimension and measure names of the project's metrics views.
// It reproduces AI canvas generation for a model with a hyphen in its name (e.g. `bids-data`), where
// the metrics view's measure names differ from their display names: without the field names in the
// prompt, the sub-agent guessed measure names from display names and produced a broken dashboard.
func TestDevelopFileMetricsViewFieldNamesInPrompt(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"models/bids-data.yaml": `
type: model
sql: |
  SELECT TIMESTAMP '2024-01-01 00:00:00' AS __time, 'Above the Fold' AS ad_position, 1 AS bid_cnt, 2.5 AS media_spend_usd
`,
			"metrics/bids-data_metrics.yaml": `
type: metrics_view
model: bids-data
timeseries: __time
dimensions:
  - column: ad_position
measures:
  - name: total_bids_measure
    display_name: Total Bids
    expression: SUM(bid_cnt)
  - name: total_media_spend_usd_measure
    display_name: Total Media Spend (USD)
    expression: SUM(media_spend_usd)
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	// Initialize a test session with a mocked LLM that captures the prompts passed to it.
	s := newSession(t, rt, instanceID)
	llm := &capturingAIService{}
	s.SetLLM(func(ctx context.Context) (drivers.AIService, func(), error) {
		return llm, func() {}, nil
	})

	// Ask the developer to create a canvas dashboard using display names only (as the parent agent does).
	var res *ai.DevelopFileResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.DevelopFileName, &res, &ai.DevelopFileArgs{
		Path:   "/dashboards/bids-data_metrics_canvas.yaml",
		Type:   "canvas",
		Prompt: `Create a canvas dashboard based on the "bids-data_metrics" metrics view with KPIs for Total Bids and Total Media Spend (USD).`,
	})
	require.NoError(t, err)

	// The prompt should contain the exact field names of the metrics view, including the mapping
	// from display names to measure names.
	prompt := llm.promptText()
	require.Contains(t, prompt, "The project's metrics views and their exact field names")
	require.Contains(t, prompt, "bids-data_metrics")
	require.Contains(t, prompt, "ad_position")
	require.Contains(t, prompt, `total_bids_measure (display name "Total Bids")`)
	require.Contains(t, prompt, `total_media_spend_usd_measure (display name "Total Media Spend (USD)")`)
}

// TestDevelopFileExistingExploresInPrompt verifies that the develop_file prompt tells the sub-agent
// which explore dashboards each metrics view already has, and whether they are inline.
// This prevents the sub-agent from writing a stand-alone explore file for a metrics view that already
// emits an inline explore, which would show up as a duplicate dashboard.
func TestDevelopFileExistingExploresInPrompt(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"models/orders.yaml": `
type: model
sql: SELECT TIMESTAMP '2024-01-01 00:00:00' AS created_at, 'web' AS channel, 10.0 AS revenue
`,
			// Inline explore
			"metrics/orders_inline.yaml": `
version: 1
type: metrics_view
model: orders
timeseries: created_at
dimensions:
  - column: channel
measures:
  - name: total_revenue
    expression: SUM(revenue)
explore:
  display_name: Orders Dashboard
`,
			// Stand-alone explore
			"metrics/orders_standalone.yaml": `
version: 1
type: metrics_view
model: orders
timeseries: created_at
dimensions:
  - column: channel
measures:
  - name: total_revenue
    expression: SUM(revenue)
explore:
  skip: true
`,
			"dashboards/orders_standalone_explore.yaml": `
type: explore
metrics_view: orders_standalone
`,
			// No explore
			"metrics/orders_none.yaml": `
version: 1
type: metrics_view
model: orders
timeseries: created_at
dimensions:
  - column: channel
measures:
  - name: total_revenue
    expression: SUM(revenue)
explore:
  skip: true
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 7, 0, 0)

	s := newSession(t, rt, instanceID)
	llm := &capturingAIService{}
	s.SetLLM(func(ctx context.Context) (drivers.AIService, func(), error) {
		return llm, func() {}, nil
	})

	var res *ai.DevelopFileResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.DevelopFileName, &res, &ai.DevelopFileArgs{
		Path:   "/dashboards/orders.yaml",
		Type:   "explore",
		Prompt: `Create an explore dashboard for the "orders_inline" metrics view.`,
	})
	require.NoError(t, err)

	prompt := llm.promptText()
	require.Contains(t, prompt, "orders_inline: dimensions: [created_at, channel]; measures: [total_revenue (display name \"Total Revenue\")]; existing explores: orders_inline (inline in the metrics view file)")
	require.Contains(t, prompt, "orders_standalone: dimensions: [created_at, channel]; measures: [total_revenue (display name \"Total Revenue\")]; existing explores: orders_standalone_explore (stand-alone file /dashboards/orders_standalone_explore.yaml)")
	require.Contains(t, prompt, "orders_none: dimensions: [created_at, channel]; measures: [total_revenue (display name \"Total Revenue\")]; existing explores: none")
	require.Contains(t, prompt, "Do NOT create a stand-alone explore file for a metrics view that already has an inline explore")
}

// capturingAIService is a drivers.AIService that captures the messages passed to it and always
// returns a plain text response (ending completion loops after one iteration).
type capturingAIService struct {
	calls []*drivers.CompleteOptions
}

func (a *capturingAIService) Complete(ctx context.Context, opts *drivers.CompleteOptions) (*drivers.CompleteResult, error) {
	a.calls = append(a.calls, opts)
	return &drivers.CompleteResult{
		Message: ai.NewTextCompletionMessage(ai.RoleAssistant, "Done."),
	}, nil
}

// promptText returns the text content of all messages passed to the service.
func (a *capturingAIService) promptText() string {
	var sb strings.Builder
	for _, call := range a.calls {
		for _, msg := range call.Messages {
			for _, block := range msg.Content {
				if text := block.GetText(); text != "" {
					sb.WriteString(text)
					sb.WriteString("\n")
				}
			}
		}
	}
	return sb.String()
}
