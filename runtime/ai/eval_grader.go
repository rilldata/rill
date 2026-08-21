package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// gradeEvalCase grades a completed case run in tiers, cheapest first with short-circuit:
// structural assertions are deterministic and produce precise reasons,
// so the LLM judge only runs when they pass.
// The verdict and reasons are written onto res.
func (r *Runner) gradeEvalCase(ctx context.Context, s *Session, opts *EvalCaseRunOptions, answer string, res *EvalCaseRunResult) {
	// Tier 1: structural assertions on the agent's queries.
	reasons := gradeEvalStructural(s, opts.Case)
	if len(reasons) > 0 {
		res.Verdict = runtimev1.AIEvalVerdict_AI_EVAL_VERDICT_FAIL
		res.FailureReasons = reasons
		return
	}

	// Tier 2: LLM judge on the final answer.
	if opts.Case.ExpectAnswer != "" {
		verdict, err := judgeEvalAnswer(ctx, s, opts, answer)
		if err != nil {
			res.Verdict = runtimev1.AIEvalVerdict_AI_EVAL_VERDICT_ERROR
			res.FailureReasons = []string{fmt.Sprintf("judge failed: %s", err.Error())}
			return
		}
		res.JudgeReasoning = verdict.Reasoning
		if !verdict.Pass {
			res.Verdict = runtimev1.AIEvalVerdict_AI_EVAL_VERDICT_FAIL
			res.FailureReasons = []string{fmt.Sprintf("answer did not match expectation: %s", verdict.Reasoning)}
			return
		}
	}

	res.Verdict = runtimev1.AIEvalVerdict_AI_EVAL_VERDICT_PASS
}

// gradeEvalStructural checks which metrics view, measures, and dimensions the agent's queries used
// against the case's structural expectations.
// It uses union semantics across all of the agent's successful query_metrics_view calls,
// so exploratory or summary queries don't cause false failures; queries that errored don't count as queried.
func gradeEvalStructural(s *Session, c *runtimev1.AIEvalCase) []string {
	if c.ExpectMetricsView == "" && len(c.ExpectMeasures) == 0 && len(c.ExpectDimensions) == 0 {
		return nil
	}

	// Map call IDs to their results so failed queries can be excluded.
	resultsByCall := map[string]*Message{}
	for _, m := range s.Messages(FilterByType(MessageTypeResult), FilterByTool(QueryMetricsViewName)) {
		resultsByCall[m.ParentID] = m
	}

	views := map[string]bool{}
	measures := map[string]bool{}
	dimensions := map[string]bool{}
	for _, msg := range s.Messages(FilterByType(MessageTypeCall), FilterByTool(QueryMetricsViewName)) {
		result, ok := resultsByCall[msg.ID]
		if !ok || result.ContentType == MessageContentTypeError {
			continue
		}
		var args map[string]any
		if err := json.Unmarshal([]byte(msg.Content), &args); err != nil {
			continue
		}
		if mv, ok := args["metrics_view"].(string); ok && mv != "" {
			views[strings.ToLower(mv)] = true
		}
		collectEvalFieldNames(args["measures"], measures)
		collectEvalFieldNames(args["dimensions"], dimensions)
	}

	var reasons []string
	if c.ExpectMetricsView != "" && !views[strings.ToLower(c.ExpectMetricsView)] {
		reasons = append(reasons, fmt.Sprintf("expected metrics view %q was never queried%s", c.ExpectMetricsView, formatEvalQueriedSet("metrics views", views)))
	}
	if missing := missingEvalFields(c.ExpectMeasures, measures); len(missing) > 0 {
		reasons = append(reasons, fmt.Sprintf("expected measures were never queried: %s%s", strings.Join(missing, ", "), formatEvalQueriedSet("measures", measures)))
	}
	if missing := missingEvalFields(c.ExpectDimensions, dimensions); len(missing) > 0 {
		reasons = append(reasons, fmt.Sprintf("expected dimensions were never queried: %s%s", strings.Join(missing, ", "), formatEvalQueriedSet("dimensions", dimensions)))
	}
	return reasons
}

// collectEvalFieldNames extracts the field names referenced by a query's measures/dimensions list.
// Computed entries (e.g. a time_floor dimension aliased as "month") also record the underlying
// dimension/measure from the compute spec, so expectations can target either name.
func collectEvalFieldNames(v any, into map[string]bool) {
	items, ok := v.([]any)
	if !ok {
		return
	}
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if name, ok := m["name"].(string); ok && name != "" {
			into[strings.ToLower(name)] = true
		}
		compute, ok := m["compute"].(map[string]any)
		if !ok {
			continue
		}
		for _, spec := range compute {
			cm, ok := spec.(map[string]any)
			if !ok {
				continue
			}
			if d, ok := cm["dimension"].(string); ok && d != "" {
				into[strings.ToLower(d)] = true
			}
			if mm, ok := cm["measure"].(string); ok && mm != "" {
				into[strings.ToLower(mm)] = true
			}
		}
	}
}

func missingEvalFields(expected []string, queried map[string]bool) []string {
	var missing []string
	for _, e := range expected {
		if !queried[strings.ToLower(e)] {
			missing = append(missing, e)
		}
	}
	return missing
}

func formatEvalQueriedSet(label string, set map[string]bool) string {
	if len(set) == 0 {
		return fmt.Sprintf("; no %s were queried", label)
	}
	names := make([]string, 0, len(set))
	for name := range set {
		names = append(names, name)
	}
	return fmt.Sprintf("; the agent queried %s: %s", label, strings.Join(names, ", "))
}

// evalJudgeVerdict is the structured output type for the LLM judge.
type evalJudgeVerdict struct {
	Pass      bool   `json:"pass" jsonschema:"Whether the actual answer satisfies the expected answer."`
	Reasoning string `json:"reasoning" jsonschema:"1-3 sentences explaining the verdict, citing specifics."`
}

// judgeEvalAnswer grades the agent's final answer against the case's expected answer using a structured LLM completion.
func judgeEvalAnswer(ctx context.Context, s *Session, opts *EvalCaseRunOptions, answer string) (*evalJudgeVerdict, error) {
	notes := strings.TrimSpace(strings.Join([]string{opts.SuiteNotes, opts.Case.Notes}, "\n"))

	var verdict evalJudgeVerdict
	err := s.Complete(ctx, "Eval judge", &verdict, &CompleteOptions{
		Messages: []*aiv1.CompletionMessage{
			NewTextCompletionMessage(RoleSystem, evalJudgeSystemPrompt()),
			NewTextCompletionMessage(RoleUser, buildEvalJudgePrompt(opts.Case.Question, notes, opts.Case.ExpectAnswer, answer)),
		},
		// UnwrapCall keeps the judge as an internal implementation detail; its reasoning is stored on the case result.
		UnwrapCall: true,
	})
	if err != nil {
		return nil, err
	}
	return &verdict, nil
}

func evalJudgeSystemPrompt() string {
	return mustExecuteTemplate(`
<role>
You are grading an AI analyst's answer against an expected answer for a business-metrics question.
</role>

<rubric>
Grade semantic equivalence of the substantive claims: numbers, entities, time periods, and the direction of trends.
Do not grade wording, formatting, or level of detail beyond what the expectation requires.
Domain guidance provided by the eval author takes precedence over your own intuitions.
Minor numeric rounding (within about 1%) is acceptable unless the guidance says otherwise.
Fail the answer if facts required by the expectation are missing, wrong, or contradicted.
</rubric>
`, nil)
}

func buildEvalJudgePrompt(question, notes, expected, actual string) string {
	return mustExecuteTemplate(`
Question asked: {{ .question }}

{{ if .notes }}
Guidance from the eval author:
{{ .notes }}
{{ end }}
Expected answer:
{{ .expected }}

Actual answer:
{{ .actual }}

Grade whether the actual answer satisfies the expected answer.
`, map[string]any{
		"question": question,
		"notes":    notes,
		"expected": expected,
		"actual":   actual,
	})
}
