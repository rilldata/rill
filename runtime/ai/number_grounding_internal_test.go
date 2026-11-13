package ai

import (
	"testing"

	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestUngroundedNumbers(t *testing.T) {
	// A day-over-day question: revenue by day for Mar 10 and Mar 11.
	timeSeries := toolResultMessage(`{"schema":[{"name":"order_date","type":"TIMESTAMP"},{"name":"revenue","type":"DOUBLE"}],"data":[["2025-03-11T00:00:00Z",3184527906],["2025-03-10T00:00:00Z",3152968451]],"open_url":"https://ui.rilldata.com/acme/sales/-/ai/0b6f3c52/message/7a8b9c0d"}`)
	// The same revenue with comparison measures.
	comparison := toolResultMessage(`{"schema":[],"data":[[3184527906,31559455,0.010009442051343888,3152968451]]}`)

	tests := []struct {
		name       string
		response   string
		messages   []*aiv1.CompletionMessage
		ungrounded []string
	}{
		{
			name:       "manual subtraction gone wrong",
			response:   "Revenue rose by **+35,559,455** (+1.13%) day over day, from 3,152,968,451 on Mar 10 to 3,184,527,906 on Mar 11.",
			messages:   []*aiv1.CompletionMessage{timeSeries},
			ungrounded: []string{"+35,559,455", "+1.13%"},
		},
		{
			name:       "manual subtraction gone right",
			response:   "Revenue rose by 31,559,455 day over day.",
			messages:   []*aiv1.CompletionMessage{timeSeries},
			ungrounded: []string{"31,559,455"},
		},
		{
			name:     "comparison query",
			response: "Revenue rose by **+31,559,455** (1.00%) to $3.18B on Mar 11, up from 3,152,968,451. That's about 31.6M, or 1.0%.",
			messages: []*aiv1.CompletionMessage{timeSeries, comparison},
		},
		{
			name:       "rounded wrong delta",
			response:   "Revenue rose by about 35.6M day over day.",
			messages:   []*aiv1.CompletionMessage{timeSeries},
			ungrounded: []string{"35.6M"},
		},
		{
			name:     "scale words",
			response: "Revenue is 3.185 billion, up 31.6 million.",
			messages: []*aiv1.CompletionMessage{comparison},
		},
		{
			name:     "other locales",
			response: "Der Umsatz stieg um 31.559.455 (1,00 %) auf 3.184.527.906, also rund 31,6 Mio. Le chiffre atteint 3,18 milliards.",
			messages: []*aiv1.CompletionMessage{comparison},
		},
		{
			name:     "dates, times, years, counts and ordinals",
			response: "On Mar 11, 2025 (2025-03-11T00:00Z, 10:18 AM) the top 5 of 12 channels led for the 3rd day in Q1; see the 2nd chart and top_10.",
			messages: []*aiv1.CompletionMessage{timeSeries},
		},
		{
			name:     "links",
			response: "Revenue rose ([DoD revenue](https://ui.rilldata.com/acme/sales/-/ai/0b6f3c52-1d2e-4f8a-9c7b-5e4d3a2b1c0f/message/7a8b9c0d-1e2f-4a5b-8c6d-9e0f1a2b3c4d)), see https://ui.rilldata.com/acme/12345678.",
			messages: []*aiv1.CompletionMessage{timeSeries},
		},
		{
			name:     "numbers from the user",
			response: "I got +35,559,455 by subtracting 3,152,968,451 from 3,184,527,906.",
			messages: []*aiv1.CompletionMessage{timeSeries, textMessage(RoleUser, "How did you get +35,559,455?")},
		},
		{
			name:       "numbers from the assistant",
			response:   "The delta is 35,559,455.",
			messages:   []*aiv1.CompletionMessage{timeSeries, textMessage(RoleAssistant, "Subtracting gives 35,559,455.")},
			ungrounded: []string{"35,559,455"},
		},
		{
			name:     "numbers from tool calls",
			response: "I limited the query to 2,500 rows.",
			messages: []*aiv1.CompletionMessage{toolCallMessage(map[string]any{"limit": 2500})},
		},
		{
			name:       "repeated numbers are reported once",
			response:   "The delta is 35,559,455. Again: 35,559,455.",
			messages:   []*aiv1.CompletionMessage{timeSeries},
			ungrounded: []string{"35,559,455"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.ungrounded, ungroundedNumbers(tt.response, tt.messages))
		})
	}
}

func TestCheckNumbersGrounded(t *testing.T) {
	messages := []*aiv1.CompletionMessage{toolResultMessage(`{"data":[[3184527906],[3152968451]]}`)}
	require.Empty(t, checkNumbersGrounded("Revenue was 3,184,527,906 on Mar 11.", messages))

	feedback := checkNumbersGrounded("Revenue rose by 35,559,455.", messages)
	require.Contains(t, feedback, "35,559,455")
	require.Contains(t, feedback, `"comparison_time_range"`)
	require.Contains(t, feedback, `"comparison_delta"`)
}

func TestParseDigits(t *testing.T) {
	values := func(raw string) []float64 {
		var res []float64
		for _, r := range parseDigits(raw) {
			res = append(res, r.value)
		}
		return res
	}
	require.Equal(t, []float64{3184527906}, values("3,184,527,906"))
	require.Equal(t, []float64{3184527906}, values("3.184.527.906"))
	require.Equal(t, []float64{3184527906}, values("3'184'527'906"))
	require.Equal(t, []float64{3184527906}, values("3 184 527 906"))
	require.Equal(t, []float64{1234.5}, values("1,234.5"))
	require.Equal(t, []float64{1234.5}, values("1.234,5"))
	require.Equal(t, []float64{1.66}, values("1,66"))
	require.Equal(t, []float64{1234, 1.234}, values("1,234"))
	require.Nil(t, values("1.2.3"))
	require.Nil(t, values("10,20,30"))
}

func toolResultMessage(content string) *aiv1.CompletionMessage {
	return &aiv1.CompletionMessage{
		Role:    string(RoleTool),
		Content: []*aiv1.ContentBlock{{BlockType: &aiv1.ContentBlock_ToolResult{ToolResult: &aiv1.ToolResult{Id: "result", Content: content}}}},
	}
}

func toolCallMessage(input map[string]any) *aiv1.CompletionMessage {
	in, err := structpb.NewStruct(input)
	if err != nil {
		panic(err)
	}
	return &aiv1.CompletionMessage{
		Role:    string(RoleAssistant),
		Content: []*aiv1.ContentBlock{{BlockType: &aiv1.ContentBlock_ToolCall{ToolCall: &aiv1.ToolCall{Id: "call", Name: QueryMetricsViewName, Input: in}}}},
	}
}

func textMessage(role Role, text string) *aiv1.CompletionMessage {
	return NewTextCompletionMessage(role, text)
}
