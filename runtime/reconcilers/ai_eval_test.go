package reconcilers_test

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"

	_ "github.com/rilldata/rill/runtime/resolvers"
)

func TestAIEvalRunner(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		AIConnector: "mock_ai",
		Files: map[string]string{
			"rill.yaml": ``,
			"/models/bar.sql": `
SELECT '2024-01-01T00:00:00Z'::TIMESTAMP as __time, 'Denmark' as country, 10 as revenue
`,
			"/metrics/mv1.yaml": `
version: 1
type: metrics_view
model: bar
timeseries: __time
dimensions:
- column: country
measures:
- name: total_revenue
  expression: SUM(revenue)
`,
			"/evals/checks.yaml": `
cases:
  - name: case_structural
    question: What is the total revenue by country?
    expect:
      metrics_view: mv1
      measures: [total_revenue]
  - name: case_judge
    question: How many countries are there?
    expect:
      answer: There is exactly one country.
`,
		},
	})

	// Project parser + model + metrics view + eval
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	catalog, release, err := rt.Catalog(t.Context(), id)
	require.NoError(t, err)
	defer release()

	// The cost invariant: reconciling without a trigger must not create any AI sessions.
	sessions, err := catalog.FindAISessions(t.Context(), "", "")
	require.NoError(t, err)
	require.Len(t, sessions, 0)

	res := testruntime.GetResource(t, rt, id, runtime.ResourceKindAIEval, "checks")
	require.False(t, res.GetAiEval().Spec.Trigger)
	require.Nil(t, res.GetAiEval().State.CurrentExecution)
	require.Len(t, res.GetAiEval().State.ExecutionHistory, 0)

	ctrl, err := rt.Controller(t.Context(), id)
	require.NoError(t, err)

	// Trigger a run of only the structural case
	triggerAIEval := func(name string, cases []string) {
		n := &runtimev1.ResourceName{Kind: runtime.ResourceKindRefreshTrigger, Name: name}
		r := &runtimev1.Resource{Resource: &runtimev1.Resource_RefreshTrigger{RefreshTrigger: &runtimev1.RefreshTrigger{
			Spec: &runtimev1.RefreshTriggerSpec{AiEvals: []*runtimev1.AIEvalTrigger{{AiEval: "checks", Cases: cases}}},
		}}}
		require.NoError(t, ctrl.Create(t.Context(), n, nil, nil, nil, nil, false, r))
		require.NoError(t, ctrl.WaitUntilIdle(t.Context(), false))
	}
	triggerAIEval("trigger_selected", []string{"case_structural"})
	testruntime.RequireReconcileState(t, rt, id, -1, 0, 0)

	res = testruntime.GetResource(t, rt, id, runtime.ResourceKindAIEval, "checks")
	eval := res.GetAiEval()
	require.False(t, eval.Spec.Trigger, "trigger should be cleared after the run")
	require.Empty(t, eval.Spec.TriggerCases)
	require.Nil(t, eval.State.CurrentExecution)
	require.Len(t, eval.State.ExecutionHistory, 1)
	require.Equal(t, uint32(1), eval.State.ExecutionCount)

	execution := eval.State.ExecutionHistory[0]
	require.Len(t, execution.CaseResults, 1, "only the selected case should have run")
	cr := execution.CaseResults[0]
	require.Equal(t, "case_structural", cr.CaseName)
	// The mock AI answers without querying the metrics view, so the structural assertion fails.
	require.Equal(t, runtimev1.AIEvalVerdict_AI_EVAL_VERDICT_FAIL, cr.Verdict)
	require.NotEmpty(t, cr.FailureReasons)
	require.Contains(t, cr.FailureReasons[0], "mv1")
	require.NotEmpty(t, cr.SessionId, "the case result should point to its transcript session")
	require.Equal(t, uint32(0), execution.PassedCount)
	require.Equal(t, uint32(1), execution.FailedCount)

	// The run created exactly one AI session, hidden from user chat lists by its user agent.
	sessions, err = catalog.FindAISessions(t.Context(), "", "")
	require.NoError(t, err)
	require.Len(t, sessions, 1)
	require.Equal(t, "rill/eval", sessions[0].UserAgent)

	// Trigger a run of all cases
	triggerAIEval("trigger_all", nil)
	testruntime.RequireReconcileState(t, rt, id, -1, 0, 0)

	res = testruntime.GetResource(t, rt, id, runtime.ResourceKindAIEval, "checks")
	eval = res.GetAiEval()
	require.Len(t, eval.State.ExecutionHistory, 2)
	require.Equal(t, uint32(2), eval.State.ExecutionCount)

	execution = eval.State.ExecutionHistory[0] // newest first
	require.Len(t, execution.CaseResults, 2)
	for _, cr := range execution.CaseResults {
		require.NotEmpty(t, cr.SessionId)
		switch cr.CaseName {
		case "case_structural":
			require.Equal(t, runtimev1.AIEvalVerdict_AI_EVAL_VERDICT_FAIL, cr.Verdict)
		case "case_judge":
			// The mock AI can't produce structured judge output, so the judge errors rather than failing the answer.
			require.Equal(t, runtimev1.AIEvalVerdict_AI_EVAL_VERDICT_ERROR, cr.Verdict)
			require.Contains(t, cr.FailureReasons[0], "judge failed")
		default:
			t.Fatalf("unexpected case %q", cr.CaseName)
		}
	}
}

func TestAIEvalRefusesToRunWithoutHostAccess(t *testing.T) {
	// Simulates a shared production (cloud) runtime, where eval runs must be refused:
	// they execute with security checks skipped, which is only safe in development contexts.
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		AIConnector:       "mock_ai",
		DisableHostAccess: true,
		Environment:       "prod",
		Files: map[string]string{
			"rill.yaml": ``,
			"/evals/checks.yaml": `
cases:
  - name: a
    question: What is the answer?
    expect:
      answer: The answer.
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, id, 2, 0, 0)

	ctrl, err := rt.Controller(t.Context(), id)
	require.NoError(t, err)
	n := &runtimev1.ResourceName{Kind: runtime.ResourceKindRefreshTrigger, Name: "trigger_cloud"}
	r := &runtimev1.Resource{Resource: &runtimev1.Resource_RefreshTrigger{RefreshTrigger: &runtimev1.RefreshTrigger{
		Spec: &runtimev1.RefreshTriggerSpec{AiEvals: []*runtimev1.AIEvalTrigger{{AiEval: "checks"}}},
	}}}
	require.NoError(t, ctrl.Create(t.Context(), n, nil, nil, nil, nil, false, r))
	require.NoError(t, ctrl.WaitUntilIdle(t.Context(), false))

	res := testruntime.GetResource(t, rt, id, runtime.ResourceKindAIEval, "checks")
	require.Contains(t, res.Meta.ReconcileError, "can only run in development environments")
	require.False(t, res.GetAiEval().Spec.Trigger, "the refused trigger should be cleared")
	require.Nil(t, res.GetAiEval().State.CurrentExecution)
	require.Len(t, res.GetAiEval().State.ExecutionHistory, 0, "no run should have executed")

	// The refusal must happen before any LLM session is created
	catalog, release, err := rt.Catalog(t.Context(), id)
	require.NoError(t, err)
	defer release()
	sessions, err := catalog.FindAISessions(t.Context(), "", "")
	require.NoError(t, err)
	require.Len(t, sessions, 0)
}

func TestAIEvalRunsInCloudDevEnvironment(t *testing.T) {
	// Simulates a cloud dev deployment (edit session): no host access, but environment "dev".
	// Eval runs are allowed there; only production deployments refuse.
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		AIConnector:       "mock_ai",
		DisableHostAccess: true,
		Environment:       "dev",
		Files: map[string]string{
			"rill.yaml": ``,
			"/evals/checks.yaml": `
cases:
  - name: a
    question: What is the answer?
    expect:
      answer: The answer.
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, id, 2, 0, 0)

	ctrl, err := rt.Controller(t.Context(), id)
	require.NoError(t, err)
	n := &runtimev1.ResourceName{Kind: runtime.ResourceKindRefreshTrigger, Name: "trigger_dev"}
	r := &runtimev1.Resource{Resource: &runtimev1.Resource_RefreshTrigger{RefreshTrigger: &runtimev1.RefreshTrigger{
		Spec: &runtimev1.RefreshTriggerSpec{AiEvals: []*runtimev1.AIEvalTrigger{{AiEval: "checks"}}},
	}}}
	require.NoError(t, ctrl.Create(t.Context(), n, nil, nil, nil, nil, false, r))
	require.NoError(t, ctrl.WaitUntilIdle(t.Context(), false))

	res := testruntime.GetResource(t, rt, id, runtime.ResourceKindAIEval, "checks")
	require.Empty(t, res.Meta.ReconcileError)
	require.False(t, res.GetAiEval().Spec.Trigger, "the trigger should be cleared after the run")
	require.Len(t, res.GetAiEval().State.ExecutionHistory, 1, "the run should have executed")
	require.Len(t, res.GetAiEval().State.ExecutionHistory[0].CaseResults, 1)
}
