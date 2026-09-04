package ai

import (
	"fmt"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
)

func TestGradeEvalStructural(t *testing.T) {
	// query emits a call message plus its result; only calls with non-error results count as queried.
	callID := 0
	query := func(content string, errored bool) []*Message {
		callID++
		id := fmt.Sprintf("call-%d", callID)
		contentType := MessageContentTypeJSON
		if errored {
			contentType = MessageContentTypeError
		}
		return []*Message{
			{ID: id, Type: MessageTypeCall, Tool: QueryMetricsViewName, Content: content},
			{ParentID: id, Type: MessageTypeResult, Tool: QueryMetricsViewName, ContentType: contentType, Content: "{}"},
		}
	}
	ok := func(content string) []*Message { return query(content, false) }
	errored := func(content string) []*Message { return query(content, true) }
	concat := func(groups ...[]*Message) []*Message {
		var res []*Message
		for _, g := range groups {
			res = append(res, g...)
		}
		return res
	}

	cases := []struct {
		name        string
		messages    []*Message
		evalCase    *runtimev1.AIEvalCase
		wantReasons int
		wantContain string
	}{
		{
			name: "no expectations passes without queries",
			evalCase: &runtimev1.AIEvalCase{
				ExpectAnswer: "anything",
			},
			wantReasons: 0,
		},
		{
			name:     "expected metrics view queried",
			messages: ok(`{"metrics_view": "Sales", "measures": [{"name": "net_revenue"}], "dimensions": [{"name": "month"}]}`),
			evalCase: &runtimev1.AIEvalCase{
				ExpectMetricsView: "sales",
				ExpectMeasures:    []string{"NET_REVENUE"},
				ExpectDimensions:  []string{"month"},
			},
			wantReasons: 0,
		},
		{
			name: "union across multiple queries",
			messages: concat(
				ok(`{"metrics_view": "sales", "measures": [{"name": "net_revenue"}]}`),
				ok(`{"metrics_view": "sales", "measures": [{"name": "order_count"}], "dimensions": [{"name": "country"}]}`),
			),
			evalCase: &runtimev1.AIEvalCase{
				ExpectMetricsView: "sales",
				ExpectMeasures:    []string{"net_revenue", "order_count"},
				ExpectDimensions:  []string{"country"},
			},
			wantReasons: 0,
		},
		{
			name:     "computed time_floor dimension matches the underlying dimension",
			messages: ok(`{"metrics_view": "sales", "dimensions": [{"name": "month", "compute": {"time_floor": {"dimension": "event_time", "grain": "month"}}}]}`),
			evalCase: &runtimev1.AIEvalCase{
				ExpectDimensions: []string{"event_time"},
			},
			wantReasons: 0,
		},
		{
			name:     "errored query does not count as queried",
			messages: errored(`{"metrics_view": "sales", "measures": [{"name": "net_revenue"}]}`),
			evalCase: &runtimev1.AIEvalCase{
				ExpectMetricsView: "sales",
			},
			wantReasons: 1,
			wantContain: "no metrics views were queried",
		},
		{
			name:     "wrong metrics view",
			messages: ok(`{"metrics_view": "bookings", "measures": [{"name": "gross"}]}`),
			evalCase: &runtimev1.AIEvalCase{
				ExpectMetricsView: "sales",
			},
			wantReasons: 1,
			wantContain: `expected metrics view "sales" was never queried`,
		},
		{
			name:     "missing measure names the gap and what was queried",
			messages: ok(`{"metrics_view": "sales", "measures": [{"name": "gross_bookings"}]}`),
			evalCase: &runtimev1.AIEvalCase{
				ExpectMeasures: []string{"net_revenue"},
			},
			wantReasons: 1,
			wantContain: "net_revenue",
		},
		{
			name: "no queries at all",
			evalCase: &runtimev1.AIEvalCase{
				ExpectMetricsView: "sales",
			},
			wantReasons: 1,
			wantContain: "no metrics views were queried",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := &Session{BaseSession: &BaseSession{messages: tc.messages}}
			reasons := gradeEvalStructural(s, tc.evalCase)
			require.Len(t, reasons, tc.wantReasons)
			if tc.wantContain != "" {
				require.Contains(t, reasons[0], tc.wantContain)
			}
		})
	}
}

func TestTruncateEvalAnswer(t *testing.T) {
	require.Equal(t, "short", truncateEvalAnswer("short"))
	long := make([]rune, evalAnswerPreviewMaxLen+10)
	for i := range long {
		long[i] = 'x'
	}
	got := truncateEvalAnswer(string(long))
	require.Len(t, []rune(got), evalAnswerPreviewMaxLen+1) // truncated + ellipsis
}
