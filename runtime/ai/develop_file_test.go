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

// TestDevelopFileSeededContext verifies that the tool calls develop_file pre-invokes (list_files, project_status,
// read_file and the skill preload) are actually passed to the sub-agent's first completion request as paired
// tool call/result messages.
// It reproduces a bug where the messages were filtered by the session ID instead of the current call ID,
// so the sub-agent ran without the file listing, project status, current file contents or the project's skills.
func TestDevelopFileSeededContext(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"models/orders.sql": `SELECT 1 AS revenue, 'web' AS channel`,
			// Always-apply developer skill: must be preloaded into the sub-agent's context
			"skills/conventions/SKILL.md": `---
description: Modeling conventions.
agents: [developer]
always_apply: true
---

Always materialize models.`,
			// Analyst-only skill: must not be preloaded for the developer sub-agent
			"skills/glossary/SKILL.md": `---
description: Business glossary.
agents: [analyst]
always_apply: true
---

ARPU excludes trial users.`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	s := newSession(t, rt, instanceID)
	llm := &capturingAIService{}
	s.SetLLM(func(ctx context.Context) (drivers.AIService, func(), error) {
		return llm, func() {}, nil
	})

	var res *ai.DevelopFileResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.DevelopFileName, &res, &ai.DevelopFileArgs{
		Path:   "/models/orders.sql",
		Type:   "model",
		Prompt: "Add a created_at timestamp column to the model.",
	})
	require.NoError(t, err)
	require.Len(t, llm.calls, 1)

	// Index the tool calls and results in the first completion request.
	// Every seeded tool call must be immediately followed by its result so the provider accepts the message sequence.
	messages := llm.calls[0].Messages
	callIDs := make(map[string]bool)
	results := make(map[string]string) // Tool name to result content
	for i, msg := range messages {
		for _, block := range msg.Content {
			call := block.GetToolCall()
			if call == nil {
				continue
			}
			require.False(t, callIDs[call.Id], "duplicate tool call %q", call.Id)
			callIDs[call.Id] = true
			require.Less(t, i+1, len(messages), "tool call %q has no result", call.Name)
			require.Len(t, messages[i+1].Content, 1)
			result := messages[i+1].Content[0].GetToolResult()
			require.NotNil(t, result, "tool call %q is not followed by a tool result", call.Name)
			require.Equal(t, call.Id, result.Id, "tool result does not match the preceding tool call %q", call.Name)
			results[call.Name] = result.Content
		}
	}

	// The file listing, project status and current file contents are seeded.
	require.Contains(t, results, ai.ListFilesName)
	require.Contains(t, results[ai.ListFilesName], "models/orders.sql")
	require.Contains(t, results, ai.ProjectStatusName)
	require.Contains(t, results, ai.ReadFileName)
	require.Contains(t, results[ai.ReadFileName], "SELECT 1 AS revenue, 'web' AS channel")

	// The project's skills are seeded, but only the developer's always-apply skills are loaded.
	require.Contains(t, results, ai.ListSkillsName)
	require.Contains(t, results[ai.ListSkillsName], "conventions")
	require.Contains(t, results, ai.LoadSkillName)
	require.Contains(t, results[ai.LoadSkillName], "Always materialize models.")
	require.NotContains(t, results[ai.LoadSkillName], "ARPU excludes trial users.")
	require.Len(t, s.Messages(ai.FilterByType(ai.MessageTypeCall), ai.FilterByTool(ai.LoadSkillName)), 1)

	// The seeded calls come after the system and user prompts.
	require.Equal(t, string(ai.RoleSystem), messages[0].Role)
	var firstCallIdx, userPromptIdx int
	for i, msg := range messages {
		if msg.Role == string(ai.RoleUser) && userPromptIdx == 0 {
			userPromptIdx = i
		}
		if msg.Content[0].GetToolCall() != nil && firstCallIdx == 0 {
			firstCallIdx = i
		}
	}
	require.Greater(t, firstCallIdx, userPromptIdx)
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
