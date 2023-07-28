package activity

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
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
	time.Sleep(time.Millisecond * 100) // wait for the event to be processed

	if len(sink.Events) == 0 {
		t.Fatalf("Expected at least one event, but got none")
	}

	event := sink.Events[0]
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
	Events []Event
}

func NewTestSink() *TestSink {
	return &TestSink{}
}

func (n *TestSink) Sink(ctx context.Context, events []Event) error {
	n.Events = append(n.Events, events...)
	return nil
}

func (n *TestSink) Close() error {
	return nil
}
