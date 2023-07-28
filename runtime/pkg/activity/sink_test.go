package activity

import (
	"context"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	zap_observer "go.uber.org/zap/zaptest/observer"
	"testing"
	"time"
)

func TestNoopSink_Sink(t *testing.T) {
	sink := NewNoopSink()

	err := sink.Sink(context.Background(), []Event{
		{
			Time:  time.Now(),
			Name:  "TestEvent",
			Value: 1.23,
			Dims:  []attribute.KeyValue{attribute.String("testDim", "value")},
		},
	})

	require.NoError(t, err, "NoopSink.Sink should not return an error")
}

func TestNoopSink_Close(t *testing.T) {
	sink := NewNoopSink()

	err := sink.Close()

	require.NoError(t, err, "NoopSink.Close should not return an error")
}

func TestConsoleSink_Sink(t *testing.T) {
	// Set up a zap logger that records all logs for assertions.
	observer, logs := zap_observer.New(zapcore.DebugLevel)
	logger := zap.New(observer)

	sink := NewConsoleSink(logger, zap.DebugLevel)

	events := []Event{
		{
			Time:  time.Now(),
			Name:  "TestEvent",
			Value: 1.23,
			Dims:  []attribute.KeyValue{attribute.String("testDim", "value")},
		},
	}

	err := sink.Sink(context.Background(), events)

	require.NoError(t, err, "ConsoleSink.Sink should not return an error")

	// Assert that the logger has recorded the correct number of logs.
	require.Len(t, logs.All(), len(events), "ConsoleSink.Sink should log all events")

	// Assert that the logger has recorded the correct log.
	// Convert event to JSON for comparison
	jsonEvent, err := events[0].Marshal()
	require.NoError(t, err)

	// Check the logged message
	require.Equal(t, string(jsonEvent), logs.All()[0].Message, "ConsoleSink.Sink should log correct event")
}

func TestConsoleSink_Sink_LogLevel(t *testing.T) {
	// Set up a zap logger that records all logs for assertions.
	observer, logs := zap_observer.New(zapcore.InfoLevel)
	logger := zap.New(observer)

	// Set the ConsoleSink to ErrorLevel
	sink := NewConsoleSink(logger, zap.DebugLevel)

	events := []Event{
		{
			Time:  time.Now(),
			Name:  "TestEvent",
			Value: 1.23,
			Dims:  []attribute.KeyValue{attribute.String("testDim", "value")},
		},
	}

	err := sink.Sink(context.Background(), events)

	require.NoError(t, err, "ConsoleSink.Sink should not return an error")

	// Assert that the logger has not recorded any logs
	require.Len(t, logs.All(), 0, "ConsoleSink.Sink should not log anything because log level is higher than debug level")
}

func TestConsoleSink_Close(t *testing.T) {
	sink := NewConsoleSink(zap.NewNop(), zap.DebugLevel)

	err := sink.Close()

	require.NoError(t, err, "ConsoleSink.Close should not return an error")
}
