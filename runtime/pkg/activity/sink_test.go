package activity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	zap_observer "go.uber.org/zap/zaptest/observer"
)

func TestNoopSink_Sink(t *testing.T) {
	sink := NewNoopSink()

	err := sink.Emit(&MetricEvent{
		Time:  time.Now(),
		Name:  "TestEvent",
		Value: 1.23,
		Attrs: []attribute.KeyValue{attribute.String("testDim", "value")},
	})

	require.NoError(t, err, "NoopSink.Sink should not return an error")
}

func TestNoopSink_Close(t *testing.T) {
	sink := NewNoopSink()
	sink.Close()
}

func TestLoggerSink_Sink(t *testing.T) {
	// Set up a zap logger that records all logs for assertions.
	observer, logs := zap_observer.New(zapcore.DebugLevel)
	logger := zap.New(observer)

	sink := NewLoggerSink(logger, zap.DebugLevel)

	event := &MetricEvent{
		Time:  time.Now(),
		Name:  "TestEvent",
		Value: 1.23,
		Attrs: []attribute.KeyValue{attribute.String("testDim", "value")},
	}

	err := sink.Emit(event)

	require.NoError(t, err, "LoggerSink.Sink should not return an error")

	// Assert that the logger has recorded the correct number of logs.
	require.Len(t, logs.All(), 1, "LoggerSink.Sink should log all events")

	// Assert that the logger has recorded the correct log.
	// Convert event to JSON for comparison
	jsonEvent, err := event.Marshal()
	require.NoError(t, err)

	// Check the logged message
	require.Equal(t, string(jsonEvent), logs.All()[0].Message, "LoggerSink.Sink should log correct event")
}

func TestLoggerSink_Sink_LogLevel(t *testing.T) {
	// Set up a zap logger that records all logs for assertions.
	observer, logs := zap_observer.New(zapcore.InfoLevel)
	logger := zap.New(observer)

	// Set the LoggerSink to ErrorLevel
	sink := NewLoggerSink(logger, zap.DebugLevel)

	event := &MetricEvent{
		Time:  time.Now(),
		Name:  "TestEvent",
		Value: 1.23,
		Attrs: []attribute.KeyValue{attribute.String("testDim", "value")},
	}

	err := sink.Emit(event)

	require.NoError(t, err, "LoggerSink.Sink should not return an error")

	// Assert that the logger has not recorded any logs
	require.Len(t, logs.All(), 0, "LoggerSink.Sink should not log anything because log level is higher than debug level")
}

func TestLoggerSink_Close(t *testing.T) {
	sink := NewLoggerSink(zap.NewNop(), zap.DebugLevel)
	sink.Close()
}
