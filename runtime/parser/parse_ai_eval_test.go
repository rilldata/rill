package parser

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
)

func TestAIEval(t *testing.T) {
	files := map[string]string{
		`rill.yaml`: ``,
		// Kind inferred from the /evals directory
		`evals/revenue.yaml`: `
display_name: Revenue questions
notes: Revenue always means net_revenue.
defaults:
  agent: analyst_agent
  explore: sales_explore
concurrency: 3
timeout: 20m
case_timeout: 2m
cases:
  - name: monthly_revenue
    question: What was our total revenue per month in 2024?
    notes: Look for 12 rows.
    expect:
      answer: Monthly totals of net_revenue for all 12 months of 2024.
      metrics_view: sales
      measures: [net_revenue]
      dimensions: [month]
  - name: top_countries
    question: Which countries have the highest revenue?
    explore: countries_explore
    expect:
      metrics_view: sales
`,
		// Explicit type with the ai_eval alias, outside the evals directory
		`checks.yaml`: `
type: ai_eval
cases:
  - name: basic
    question: How many orders were there yesterday?
    expect:
      answer: A single count of orders for yesterday.
`,
	}

	resources := []*Resource{
		{
			Name:  ResourceName{Kind: ResourceKindAIEval, Name: "revenue"},
			Paths: []string{"/evals/revenue.yaml"},
			Refs: []ResourceName{
				{Kind: ResourceKindMetricsView, Name: "sales"},
				{Kind: ResourceKindExplore, Name: "sales_explore"},
				{Kind: ResourceKindExplore, Name: "countries_explore"},
			},
			AIEvalSpec: &runtimev1.AIEvalSpec{
				DisplayName:        "Revenue questions",
				Agent:              "analyst_agent",
				Notes:              "Revenue always means net_revenue.",
				Explore:            "sales_explore",
				TimeoutSeconds:     1200,
				CaseTimeoutSeconds: 120,
				Concurrency:        3,
				Cases: []*runtimev1.AIEvalCase{
					{
						Name:              "monthly_revenue",
						Question:          "What was our total revenue per month in 2024?",
						Notes:             "Look for 12 rows.",
						ExpectAnswer:      "Monthly totals of net_revenue for all 12 months of 2024.",
						ExpectMetricsView: "sales",
						ExpectMeasures:    []string{"net_revenue"},
						ExpectDimensions:  []string{"month"},
					},
					{
						Name:              "top_countries",
						Question:          "Which countries have the highest revenue?",
						Explore:           "countries_explore",
						ExpectMetricsView: "sales",
					},
				},
			},
		},
		{
			Name:  ResourceName{Kind: ResourceKindAIEval, Name: "checks"},
			Paths: []string{"/checks.yaml"},
			AIEvalSpec: &runtimev1.AIEvalSpec{
				DisplayName: "Checks",
				Agent:       "analyst_agent",
				Cases: []*runtimev1.AIEvalCase{
					{
						Name:         "basic",
						Question:     "How many orders were there yesterday?",
						ExpectAnswer: "A single count of orders for yesterday.",
					},
				},
			},
		},
	}

	ctx := context.Background()
	repo := makeRepo(t, files)
	p, err := Parse(ctx, repo, "", "", "duckdb", true)
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, nil)
}

func TestAIEvalErrors(t *testing.T) {
	cases := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name:    "missing cases",
			yaml:    "type: eval\n",
			wantErr: `missing required property "cases"`,
		},
		{
			name: "duplicate case names",
			yaml: `
type: eval
cases:
  - name: a
    question: q1
    expect: {answer: x}
  - name: A
    question: q2
    expect: {answer: y}
`,
			wantErr: `duplicate case name "A"`,
		},
		{
			name: "missing question",
			yaml: `
type: eval
cases:
  - name: a
    expect: {answer: x}
`,
			wantErr: `missing required property "question"`,
		},
		{
			name: "empty expect",
			yaml: `
type: eval
cases:
  - name: a
    question: q
    expect: {}
`,
			wantErr: `"expect" must set at least one of`,
		},
		{
			name: "invalid agent",
			yaml: `
type: eval
defaults:
  agent: developer_agent
cases:
  - name: a
    question: q
    expect: {answer: x}
`,
			wantErr: `invalid agent "developer_agent"`,
		},
		{
			name: "invalid case name",
			yaml: `
type: eval
cases:
  - name: "has spaces"
    question: q
    expect: {answer: x}
`,
			wantErr: "name may only contain letters, numbers, underscores and hyphens",
		},
		{
			name: "negative timeout",
			yaml: `
type: eval
timeout: -5m
cases:
  - name: a
    question: q
    expect: {answer: x}
`,
			wantErr: `invalid "timeout": must be at least 1s`,
		},
		{
			name: "sub-second case timeout",
			yaml: `
type: eval
case_timeout: 500ms
cases:
  - name: a
    question: q
    expect: {answer: x}
`,
			wantErr: `invalid "case_timeout": must be at least 1s`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			files := map[string]string{
				`rill.yaml`:      ``,
				`evals/bad.yaml`: tc.yaml,
			}
			errors := []*runtimev1.ParseError{
				{
					Message:  tc.wantErr,
					FilePath: "/evals/bad.yaml",
				},
			}
			ctx := context.Background()
			repo := makeRepo(t, files)
			p, err := Parse(ctx, repo, "", "", "duckdb", true)
			require.NoError(t, err)
			requireResourcesAndErrors(t, p, nil, errors)
		})
	}
}
