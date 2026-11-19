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

// TestAnalystDayOverDayEval reproduces APP-960: asked for the day over day difference in a forecast,
// the analyst queried a daily time series of forecast snapshots and did the arithmetic on its rows itself.
// A case passes when the analyst's final answer has the correct numbers,
// it requested a comparison measure, and it does not add up values itself.
// The log also reports the same for the first final answer, which is written before the number check (CheckResponse) can step in,
// so it measures the instructions alone.
// Set EVAL_AI_CONNECTOR to "claude" to use Claude instead of OpenAI, and RILL_RUNTIME_CLAUDE_TEST_MODEL to pick the model.
func TestAnalystDayOverDayEval(t *testing.T) {
	accounts := []string{"Northwind Media", "Acme Retail", "Globex Auto", "Initech Finance", "Umbrella Health", "Stark Travel", "Wayne Foods", "Hooli Tech"}
	start := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	forecast := map[string]map[string]int64{} // snapshot date -> account -> month-end forecast
	var rows []string
	for i := range 34 { // October 1 to November 3, 2026
		day := start.AddDate(0, 0, i).Format("2006-01-02")
		forecast[day] = map[string]int64{}
		for a, account := range accounts {
			amount := int64(40_000_000+a*6_000_000) + int64((i+5)*(a+3)*6151%99991)*23 + int64((i*97+a*53)%8761)
			forecast[day][account] = amount
			rows = append(rows, fmt.Sprintf("('%sT00:00:00Z'::TIMESTAMP, '%s', %d::DOUBLE)", day, account, amount))
		}
	}

	connector := os.Getenv("EVAL_AI_CONNECTOR")
	if connector == "" {
		connector = "openai"
	}
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		AIConnector: connector,
		Files: map[string]string{
			"models/revenue_forecast.yaml": "type: model\nmaterialize: true\nsql: |\n  SELECT * FROM (VALUES\n    " + strings.Join(rows, ",\n    ") + "\n  ) AS t(snapshot_date, account, month_end_forecast)\n",
			"metrics/revenue_forecast.yaml": `
type: metrics_view
display_name: Revenue Forecast
model: revenue_forecast
timeseries: snapshot_date
dimensions:
- column: account
measures:
- name: month_end_forecast
  display_name: Month-end Forecast
  expression: SUM(month_end_forecast)
`,
			"dashboards/revenue_forecast_canvas.yaml": `
type: canvas
display_name: Revenue Forecast
rows:
  - items:
      - line_chart:
          metrics_view: revenue_forecast
          x:
            field: snapshot_date
            type: temporal
          y:
            field: month_end_forecast
            type: quantitative
        width: 12
  - items:
      - leaderboard:
          metrics_view: revenue_forecast
          dimensions:
            - account
          measures:
            - month_end_forecast
          num_rows: 8
        width: 12
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 7, 0, 0)

	// The data ends on November 3; "day over day" can mean November 3 vs 2 or, for the last complete day, November 2 vs 1.
	total := func(day string) int64 {
		var res int64
		for _, v := range forecast[day] {
			res += v
		}
		return res
	}
	topMover := func(cur, prev string) (string, int64) {
		var name string
		var best int64
		for _, a := range accounts {
			if d := forecast[cur][a] - forecast[prev][a]; abs(d) > abs(best) {
				name, best = a, d
			}
		}
		return name, best
	}
	totalDelta := append(amountRenderings(total("2026-11-03")-total("2026-11-02")), amountRenderings(total("2026-11-02")-total("2026-11-01"))...)
	mover1, moverDelta1 := topMover("2026-11-03", "2026-11-02")
	mover2, moverDelta2 := topMover("2026-11-02", "2026-11-01")
	t.Logf("total deltas: %d (Nov 3 vs 2), %d (Nov 2 vs 1); top movers: %s %d, %s %d",
		total("2026-11-03")-total("2026-11-02"), total("2026-11-02")-total("2026-11-01"), mover1, moverDelta1, mover2, moverDelta2)

	// Sums of the accounts' gains and of their declines: no tool returns these,
	// so an answer that states one ("combined, these accounts added $7.0M") added them up itself.
	// Whole millions ("4M") are left out, since they match unrelated rounded amounts.
	var handSums, exactSums []string
	for _, days := range [][2]string{{"2026-11-03", "2026-11-02"}, {"2026-11-02", "2026-11-01"}} {
		var gains, declines int64
		for _, a := range accounts {
			if d := forecast[days[0]][a] - forecast[days[1]][a]; d > 0 {
				gains += d
			} else {
				declines -= d
			}
		}
		for _, n := range []int64{gains, declines} {
			m := float64(n) / 1e6
			handSums = append(handSums, fmt.Sprintf("%d", n), fmt.Sprintf("%.2fM", m), fmt.Sprintf("%.1fM", m))
			exactSums = append(exactSums, fmt.Sprintf("%d", n))
		}
	}
	t.Logf("hand sums: %v", handSums)

	question := "what's the day over day difference in the month-end forecast?"
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
			prompt:   question + " which accounts moved the most?",
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
				args.Canvas = "revenue_forecast_canvas"
				args.TimeStart = time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
				args.TimeEnd = time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC)
			}
			var res *ai.AnalystAgentResult
			_, err = s.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, &res, args)
			require.NoError(t, err)

			// Find the first final answer (a response with text and no tool calls),
			// whether a comparison measure was requested, and whether the number check rejected an answer.
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
							} else {
								t.Logf("query: %s", m.Content)
							}
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

			finalCorrect := correct(res.Response)
			finalSummed := handSummed(res.Response, finalResults)
			t.Logf("RESULT %s/%s/%s: pass=%v final_answer_correct=%v used_comparison=%v final_answer_hand_summed=%v check_fired=%v first_answer_correct=%v first_answer_hand_summed=%v",
				connector, os.Getenv("RILL_RUNTIME_CLAUDE_TEST_MODEL"), tc.name, finalCorrect && usedComparison && !finalSummed,
				finalCorrect, usedComparison, finalSummed, checkFired, correct(firstAnswer), handSummed(firstAnswer, firstResults))
			t.Logf("first answer:\n%s", firstAnswer)
			if checkFired {
				t.Logf("final answer:\n%s", res.Response)
			}
			if !finalCorrect || !usedComparison || finalSummed {
				t.Errorf("the final answer is wrong, was not computed by a comparison query, or adds up values itself")
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

var (
	// linkRegexp matches link targets and URLs, as the number check strips them: their digits are IDs, not values.
	linkRegexp = regexp.MustCompile(`\]\([^)\s]*\)|https?://[^\s)\]>]+`)
	// digitGroupRegexp matches a digit group separator between digits.
	digitGroupRegexp = regexp.MustCompile(`(\d)[,\x{202f}\x{a0}](\d{3})`)
)

// mentionsNumber reports whether the text states any of the candidates as a whole number outside links,
// ignoring digit group separators and a space before a unit: "7.0M" matches "$7.0 million" but not "$17.0M".
func mentionsNumber(text string, candidates []string) bool {
	text = linkRegexp.ReplaceAllString(text, " ")
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
