package ai_test

import (
	"context"
	"fmt"
	"math"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

// TestAnalystDayOverDayEval reproduces APP-960: asked for a day over day difference,
// the analyst queried a daily time series and did the arithmetic on its rows itself.
// A case passes when the analyst's first final answer has the correct numbers,
// it requested a comparison measure before writing it, and it does not add up values itself.
// The first final answer is written before the number check (CheckResponse) can step in, so it measures the instructions alone;
// the log also reports whether the check fired and whether the final answer is correct.
// Set EVAL_AI_CONNECTOR to "claude" to use Claude instead of OpenAI, and RILL_RUNTIME_CLAUDE_TEST_MODEL to pick the model.
func TestAnalystDayOverDayEval(t *testing.T) {
	customers := []string{"Northwind Media", "Acme Retail", "Globex Auto", "Initech Finance", "Umbrella Health", "Stark Travel", "Wayne Foods", "Hooli Tech"}
	start := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	revenue := map[string]map[string]int64{} // day -> customer -> revenue
	var rows []string
	for i := range 34 { // October 1 to November 3, 2026
		day := start.AddDate(0, 0, i).Format("2006-01-02")
		revenue[day] = map[string]int64{}
		for c, customer := range customers {
			amount := int64(90_000_000+c*11_000_000) + int64((i+3)*(c+5)*7919%104729)*37 + int64((i*131+c*71)%9973)
			revenue[day][customer] = amount
			rows = append(rows, fmt.Sprintf("('%sT00:00:00Z'::TIMESTAMP, '%s', %d::DOUBLE)", day, customer, amount))
		}
	}

	connector := os.Getenv("EVAL_AI_CONNECTOR")
	if connector == "" {
		connector = "openai"
	}
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		AIConnector: connector,
		Files: map[string]string{
			"models/customer_revenue.yaml": "type: model\nmaterialize: true\nsql: |\n  SELECT * FROM (VALUES\n    " + strings.Join(rows, ",\n    ") + "\n  ) AS t(order_date, customer, revenue)\n",
			"metrics/customer_revenue.yaml": `
type: metrics_view
display_name: Customer Revenue
model: customer_revenue
timeseries: order_date
dimensions:
- column: customer
measures:
- name: revenue
  display_name: Revenue
  expression: SUM(revenue)
`,
			"dashboards/customer_revenue_canvas.yaml": `
type: canvas
display_name: Customer Revenue
rows:
  - items:
      - line_chart:
          metrics_view: customer_revenue
          x:
            field: order_date
            type: temporal
          y:
            field: revenue
            type: quantitative
        width: 12
  - items:
      - leaderboard:
          metrics_view: customer_revenue
          dimensions:
            - customer
          measures:
            - revenue
          num_rows: 8
        width: 12
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 7, 0, 0)

	// The data ends on November 3; "day over day" can mean November 3 vs 2 or, for the last complete day, November 2 vs 1.
	total := func(day string) int64 {
		var res int64
		for _, v := range revenue[day] {
			res += v
		}
		return res
	}
	topMover := func(cur, prev string) (string, int64) {
		var name string
		var best int64
		for _, c := range customers {
			if d := revenue[cur][c] - revenue[prev][c]; abs(d) > abs(best) {
				name, best = c, d
			}
		}
		return name, best
	}
	totalDelta := append(amountRenderings(total("2026-11-03")-total("2026-11-02")), amountRenderings(total("2026-11-02")-total("2026-11-01"))...)
	mover1, moverDelta1 := topMover("2026-11-03", "2026-11-02")
	mover2, moverDelta2 := topMover("2026-11-02", "2026-11-01")
	t.Logf("total deltas: %d (Nov 3 vs 2), %d (Nov 2 vs 1); top movers: %s %d, %s %d",
		total("2026-11-03")-total("2026-11-02"), total("2026-11-02")-total("2026-11-01"), mover1, moverDelta1, mover2, moverDelta2)

	// Sums of the customers' gains and of their declines: no tool returns these,
	// so an answer that states one ("combined, these customers added $7.0M") added them up itself.
	var handSums, exactSums []string
	for _, days := range [][2]string{{"2026-11-03", "2026-11-02"}, {"2026-11-02", "2026-11-01"}} {
		var gains, declines int64
		for _, c := range customers {
			if d := revenue[days[0]][c] - revenue[days[1]][c]; d > 0 {
				gains += d
			} else {
				declines -= d
			}
		}
		for _, n := range []int64{gains, declines} {
			m := float64(n) / 1e6
			handSums = append(handSums, fmt.Sprintf("%d", n), fmt.Sprintf("%.2fM", m), fmt.Sprintf("%.1fM", m), fmt.Sprintf("%.0fM", m))
			exactSums = append(exactSums, fmt.Sprintf("%d", n))
		}
	}
	t.Logf("hand sums: %v", handSums)

	question := "what's the day over day difference in revenue?"
	cases := []struct {
		name     string
		canvas   bool
		prompt   string
		expected [][]string // Each group must have at least one rendering in the answer.
	}{
		{name: "NoContext", prompt: question, expected: [][]string{totalDelta}},
		{name: "Canvas", canvas: true, prompt: question, expected: [][]string{totalDelta}},
		{
			name:     "CanvasWithMovers",
			canvas:   true,
			prompt:   question + " which customers moved the most?",
			expected: [][]string{totalDelta, append(amountRenderings(moverDelta1), amountRenderings(moverDelta2)...)},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := newSession(t, rt, instanceID)
			llm, release, err := rt.AI(t.Context(), instanceID)
			require.NoError(t, err)
			t.Cleanup(release)
			rec := &recordingAIService{ai: llm}
			s.SetLLM(func(context.Context) (drivers.AIService, func(), error) { return rec, func() {}, nil })

			args := &ai.AnalystAgentArgs{Prompt: tc.prompt}
			if tc.canvas {
				args.Canvas = "customer_revenue_canvas"
				args.TimeStart = time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
				args.TimeEnd = time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC)
			}
			var res *ai.AnalystAgentResult
			_, err = s.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, &res, args)
			require.NoError(t, err)

			// Find the first final answer (a response with text and no tool calls),
			// whether a comparison measure was requested before it, and whether the number check rejected it.
			// Also keep the tool results the LLM had seen when it wrote each answer,
			// so that a sum only counts as hand-computed if no tool result contains it.
			var firstAnswer, firstResults, finalResults string
			var usedComparison, checkFired bool
			for _, call := range rec.calls {
				var results strings.Builder
				for _, m := range call.Input {
					if m.Role == string(ai.RoleUser) && strings.Contains(m.Content, "do not appear in any tool result") {
						checkFired = true
					}
					if m.ContentType == "tool_response" {
						results.WriteString(m.Content)
					}
				}
				finalResults = results.String()
				var text strings.Builder
				hasToolCall := false
				for _, m := range call.Response {
					switch m.ContentType {
					case "tool_call":
						hasToolCall = true
						if m.ToolName == ai.QueryMetricsViewName {
							if firstAnswer != "" {
								t.Logf("query after check: %s", m.Content)
								continue
							}
							t.Logf("query: %s", m.Content)
							if strings.Contains(m.Content, "comparison_time_range") {
								usedComparison = true
							}
						}
					case "text":
						text.WriteString(m.Content)
					}
				}
				if firstAnswer == "" && !hasToolCall && text.Len() > 0 {
					firstAnswer, firstResults = text.String(), results.String()
				}
			}
			handSummed := func(answer, results string) bool {
				return mentionsNumber(answer, handSums) && !slices.ContainsFunc(exactSums, func(n string) bool { return strings.Contains(results, n) })
			}
			correct := func(answer string) bool {
				return !slices.ContainsFunc(tc.expected, func(group []string) bool { return !mentionsNumber(answer, group) })
			}

			firstCorrect := correct(firstAnswer)
			firstSummed := handSummed(firstAnswer, firstResults)
			t.Logf("RESULT %s/%s/%s: pass=%v first_answer_correct=%v used_comparison=%v first_answer_hand_summed=%v check_fired=%v final_answer_correct=%v final_answer_hand_summed=%v",
				connector, os.Getenv("RILL_RUNTIME_CLAUDE_TEST_MODEL"), tc.name, firstCorrect && usedComparison && !firstSummed,
				firstCorrect, usedComparison, firstSummed, checkFired, correct(res.Response), handSummed(res.Response, finalResults))
			t.Logf("first answer:\n%s", firstAnswer)
			if checkFired {
				t.Logf("final answer:\n%s", res.Response)
			}
			if !firstCorrect || !usedComparison || firstSummed {
				t.Errorf("the first answer is wrong, was not computed by a comparison query, or adds up values itself")
			}
		})
	}
}

// amountRenderings returns common ways of writing the absolute value of n.
func amountRenderings(n int64) []string {
	f := math.Abs(float64(n))
	return []string{
		fmt.Sprintf("%d", int64(f)),
		fmt.Sprintf("%.2fM", f/1e6), fmt.Sprintf("%.1fM", f/1e6),
		fmt.Sprintf("%.0fK", f/1e3), fmt.Sprintf("%.1fK", f/1e3),
	}
}

// digitGroupRegexp matches a digit group separator between digits.
var digitGroupRegexp = regexp.MustCompile(`(\d)[,\x{202f}\x{a0}](\d{3})`)

// mentionsNumber reports whether the text states any of the candidates as a whole number,
// ignoring digit group separators and a space before a unit: "7.0M" matches "$7.0 million" but not "$17.0M".
func mentionsNumber(text string, candidates []string) bool {
	for range 4 { // Repeat to remove every separator in long numbers.
		text = digitGroupRegexp.ReplaceAllString(text, "$1$2")
	}
	text = strings.NewReplacer(" million", "M", " M", "M", " K", "K").Replace(text)
	for _, c := range candidates {
		if regexp.MustCompile(`(^|[^\d.])` + regexp.QuoteMeta(c) + `($|[^\d])`).MatchString(text) {
			return true
		}
	}
	return false
}

func abs(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}
