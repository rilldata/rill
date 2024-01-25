package activity

import (
	"context"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"sync"
	"testing"
	"time"
)

func TestBufferedClientEmit(t *testing.T) {
	sink := NewTestSink()
	client := NewBufferedClient(BufferedClientOptions{
		Sink:       sink,
		SinkPeriod: time.Millisecond * 10,
		BufferSize: 1,
	})

	client.Emit(context.Background(), "test_event", 1.0, attribute.String("test_dim", "test_val"))

	require.Eventually(t, func() bool { return len(sink.GetEvents()) == 1 }, time.Second*2, time.Millisecond*10)

	event := sink.GetEvents()[0]
	if event.Name != "test_event" {
		t.Errorf("Expected 'test_event', but got '%s'", event.Name)
	}
}

func TestNoopClientEmit(t *testing.T) {
	client := NewNoopClient()

	// This should not cause any errors
	client.Emit(context.Background(), "test_event", 1.0, attribute.String("test_dim", "test_val"))
}

func TestEventMarshal(t *testing.T) {
	event := &Event{
		Time:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		Name:  "test_event",
		Value: 1.0,
		Dims: []attribute.KeyValue{
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
	if err != nil {
		t.Fatalf("Expected no error, but got '%s'", err)
	}

	expected := `{"bool":true,"bool_slice":[false,true],"float64":0,"float64_slice":[-1,0,1],"int":0,"int64":0,"int64_slice":[-1,0,1],"int_slice":[-1,0,1],"name":"test_event","string":"value","string_slice":["value1","value2"],"time":"2023-01-01T00:00:00Z","value":1}`
	if string(data) != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, string(data))
	}
}

type TestSink struct {
	events   []Event
	eventsMu sync.Mutex
}

func NewTestSink() *TestSink {
	return &TestSink{}
}

func (n *TestSink) Sink(_ context.Context, events []Event) error {
	n.eventsMu.Lock()
	defer n.eventsMu.Unlock()
	n.events = append(n.events, events...)
	return nil
}

func (n *TestSink) GetEvents() []Event {
	n.eventsMu.Lock()
	defer n.eventsMu.Unlock()
	return n.events
}

func (n *TestSink) Close() error {
	return nil
}
