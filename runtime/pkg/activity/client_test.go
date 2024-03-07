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

func TestBufferedClientEmit(t *testing.T) {
	sink := newMockSink()
	client := NewClient(sink, zap.NewNop(), &ClientOptions{})

	client.EmitMetric(context.Background(), "test_event", 1.0, attribute.String("test_dim", "test_val"))

	require.Equal(t, 1, len(sink.Events()))

	e := sink.Events()[0]
	me, ok := e.(*MetricEvent)
	require.True(t, ok)
	require.Equal(t, "test_event", me.Name)
}

func TestEventMarshal(t *testing.T) {
	event := &MetricEvent{
		Time:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		Name:  "test_event",
		Value: 1.0,
		Attrs: []attribute.KeyValue{
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
		},
	}
	data, err := event.Marshal()
	require.NoError(t, err)

	expected := `{"bool":true,"bool_slice":[false,true],"float64":0,"float64_slice":[-1,0,1],"int":0,"int64":0,"int64_slice":[-1,0,1],"int_slice":[-1,0,1],"name":"test_event","string":"value","string_slice":["value1","value2"],"time":"2023-01-01T00:00:00Z","value":1}`
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
