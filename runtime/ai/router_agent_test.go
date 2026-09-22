package ai_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestRouterAgent(t *testing.T) {
	// Setup empty project
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		AIConnector: "openai",
	})

	// NOTE: We use a single eval session, so each subsequent prompt will have the context of the previous ones.
	s := newEval(t, rt, instanceID)
	cases := []struct {
		prompt string
		agent  string
	}{
		{
			prompt: "What country has the highest revenue?",
			agent:  ai.AnalystAgentName,
		},
		{
			prompt: "Repeat the answer you gave to my last question",
			agent:  ai.AnalystAgentName,
		},
		{
			prompt: "Create a model called 'sales_data' that selects all columns from the 'orders' table.",
			agent:  ai.DeveloperAgentName,
		},
		{
			prompt: "Do another one for the 'customers' table.",
			agent:  ai.DeveloperAgentName,
		},
		{
			prompt: "What is 2 + 2?",
			agent:  ai.AnalystAgentName,
		},
		{
			prompt: "developer: Add the metric code_churn",
			agent:  ai.DeveloperAgentName,
		},
		{
			prompt: "developer_agent: Add a new metric code_churn",
			agent:  ai.DeveloperAgentName,
		},
		{
			prompt: "Add a new metric code_churn. Use the developer agent.",
			agent:  ai.DeveloperAgentName,
		},
	}
	for _, c := range cases {
		var res *ai.RouterAgentResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.RouterAgentName, &res, ai.RouterAgentArgs{
			Prompt:      c.prompt,
			SkipHandoff: true,
		})
		require.NoError(t, err)
		require.Equal(t, c.agent, res.Agent, "prompt: %q", c.prompt)
	}
}

// TestRouterAgentTitle verifies that the conversation title shows the chat references in the first prompt as readable text instead of their markup.
func TestRouterAgentTitle(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml":         ``,
			"models/orders.sql": `SELECT 1 AS revenue, 'US' AS country`,
			"metrics/orders.yaml": `
type: metrics_view
display_name: Sales
model: orders
dimensions:
- name: country
  display_name: Customer Country
  column: country
measures:
- name: revenue
  display_name: Net Revenue
  expression: SUM(revenue)
`,
			"canvases/overview.yaml": `
type: canvas
display_name: Sales Overview
rows:
- items:
  - kpi_grid:
      metrics_view: orders
      measures: [revenue]
- items:
  - bar_chart:
      metrics_view: orders
      title: Revenue by country
      x:
        field: country
        type: nominal
      y:
        field: revenue
        type: quantitative
`,
			"canvases/restricted.yaml": `
type: canvas
display_name: Restricted
security:
  access: false
rows:
- items:
  - bar_chart:
      metrics_view: orders
      title: Revenue by customer
      x:
        field: country
        type: nominal
      y:
        field: revenue
        type: quantitative
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 9, 0, 0)

	cases := []struct {
		prompt string
		title  string
	}{
		{
			prompt: `<chat-reference>type="metricsView" metricsView="orders"</chat-reference> by month`,
			title:  "Sales by month",
		},
		{
			prompt: `Explain what drives <chat-reference>type="measure" metricsView="orders" measure="revenue"</chat-reference>?`,
			title:  "Explain what drives Net Revenue?",
		},
		{
			prompt: `Revenue by <chat-reference>type="dimension" metricsView="orders" dimension="country"</chat-reference>`,
			title:  "Revenue by Customer Country",
		},
		{
			prompt: `Summarize <chat-reference>type="canvas" canvas="overview"</chat-reference>`,
			title:  "Summarize Sales Overview",
		},
		{
			// A canvas component is shown by its title.
			prompt: `Explain <chat-reference>type="canvasComponent" canvas="overview" canvasComponent="overview--component-1-0"</chat-reference>`,
			title:  "Explain Revenue by country",
		},
		{
			// A canvas component without a title is shown by the canvas it belongs to.
			prompt: `Explain <chat-reference>type="canvasComponent" canvas="overview" canvasComponent="overview--component-0-0"</chat-reference>`,
			title:  "Explain Sales Overview",
		},
		{
			// A component that doesn't belong to the referenced canvas isn't shown by its title.
			prompt: `Explain <chat-reference>type="canvasComponent" canvas="overview" canvasComponent="restricted--component-0-0"</chat-reference>`,
			title:  "Explain Sales Overview",
		},
		{
			// The time range is shown as dates in the reference's time zone.
			prompt: `<chat-reference>type="timeRange" timeZone="Europe/Madrid" granularity="day" timeRange="2026-07-31T22:00:00.000Z to 2026-08-30T22:00:00.000Z"</chat-reference> vs last year`,
			title:  "2026-08-01 – 2026-08-31 vs last year",
		},
		{
			prompt: `<chat-reference>type="skill" skill="monthly-close"</chat-reference> for March`,
			title:  "/monthly-close for March",
		},
		{
			prompt: `Nulls in <chat-reference>type="column" model="orders" column="revenue" columnType="INTEGER"</chat-reference>`,
			title:  "Nulls in revenue",
		},
		{
			// A reference to a resource that doesn't exist shows its name.
			prompt: `<chat-reference>type="metricsView" metricsView="missing"</chat-reference> trend`,
			title:  "missing trend",
		},
		{
			// A reference of an unknown type is dropped.
			prompt: `<chat-reference>type="unknown" unknown="x"</chat-reference> hello`,
			title:  "hello",
		},
		{
			// The title is truncated after the references are replaced.
			prompt: `<chat-reference>type="metricsView" metricsView="orders"</chat-reference> how much did we sell to each customer last month?`,
			title:  "Sales how much did we sell to each customer las...",
		},
		{
			// The truncation doesn't split a multi-byte rune: the 47-byte cut lands inside the first "ú".
			prompt: strings.Repeat("a", 46) + "úúúú",
			title:  strings.Repeat("a", 46) + "...",
		},
	}
	for _, c := range cases {
		s := newSession(t, rt, instanceID)
		var res *ai.RouterAgentResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.RouterAgentName, &res, ai.RouterAgentArgs{
			Prompt:      c.prompt,
			Agent:       ai.AnalystAgentName,
			SkipHandoff: true,
		})
		require.NoError(t, err)
		require.Equal(t, c.title, s.Title(), "prompt: %q", c.prompt)
	}

	// With security checks, a component of a canvas that the session can't access isn't shown by its title.
	s, err := ai.NewRunner(rt, activity.NewNoopClient()).Session(t.Context(), &ai.SessionOptions{
		InstanceID: instanceID,
		Claims:     &runtime.SecurityClaims{UserID: uuid.NewString(), Permissions: []runtime.Permission{runtime.UseAI, runtime.ReadObjects}},
		UserAgent:  "rill-evals",
	})
	require.NoError(t, err)
	var res *ai.RouterAgentResult
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.RouterAgentName, &res, ai.RouterAgentArgs{
		Prompt:      `Explain <chat-reference>type="canvasComponent" canvas="restricted" canvasComponent="restricted--component-0-0"</chat-reference>`,
		Agent:       ai.AnalystAgentName,
		SkipHandoff: true,
	})
	require.NoError(t, err)
	require.Equal(t, "Explain restricted--component-0-0", s.Title())
}
