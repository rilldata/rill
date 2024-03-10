package activity

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

func TestClientEmitMetric(t *testing.T) {
	sink := newMockSink()
	client := NewClient(sink, zap.NewNop())

	client.RecordMetric(context.Background(), "test_event", 1.0, attribute.String("test_dim", "test_val"))

	require.Equal(t, 1, len(sink.Events()))

	e := sink.Events()[0]
	require.Len(t, e.EventID, 36) // Length of a UUIDv4 string
	require.False(t, e.EventTime.IsZero())
	require.Equal(t, "metric", e.EventType)
	require.Equal(t, "test_event", e.EventName)
}

func TestEventMarshal(t *testing.T) {
	sink := newMockSink()
	client := NewClient(sink, zap.NewNop())

	client.RecordMetric(context.Background(), "test_event", 1.0,
		attribute.Bool("bool", true),
		attribute.String("string", "value"),
		attribute.Int("int", 0),
		attribute.Int64("int64", 0),
		attribute.Float64("float64", 0.0),
		attribute.BoolSlice("bool_slice", []bool{false, true}),
		attribute.StringSlice("string_slice", []string{"value1", "value2"}),
		attribute.IntSlice("int_slice", []int{-1, 0, 1}),
		attribute.Int64Slice("int64_slice", []int64{-1, 0, 1}),
		attribute.Float64Slice("float64_slice", []float64{-1.0, 0.0, 1.0}),
	)

	require.Equal(t, 1, len(sink.Events()))

	e := sink.Events()[0]
	e.EventID = "8cb858a4-2d5a-4a80-ae9b-fb5b905f18a2"
	e.EventTime = time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	data, err := e.MarshalJSON()
	require.NoError(t, err)

	expected := `{"bool":true,"bool_slice":[false,true],"event_id":"8cb858a4-2d5a-4a80-ae9b-fb5b905f18a2","event_name":"test_event","event_time":"2023-01-01T00:00:00Z","event_type":"metric","float64":0,"float64_slice":[-1,0,1],"int":0,"int64":0,"int64_slice":[-1,0,1],"int_slice":[-1,0,1],"string":"value","string_slice":["value1","value2"],"value":1}`
	require.Equal(t, expected, string(data))
}

type mockSink struct {
	events   []Event
	eventsMu sync.Mutex
}

var _ Sink = (*mockSink)(nil)

func newMockSink() *mockSink {
	return &mockSink{}
}

func (n *mockSink) Emit(e Event) error {
	n.eventsMu.Lock()
	defer n.eventsMu.Unlock()
	n.events = append(n.events, e)
	return nil
}

func (n *mockSink) Events() []Event {
	n.eventsMu.Lock()
	defer n.eventsMu.Unlock()
	return n.events
}

func (n *mockSink) Close() {}
