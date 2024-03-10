package activity

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Sink is a destination for sending telemetry events.
type Sink interface {
	Emit(event Event) error
	Close()
}

type noopSink struct{}

// NewNoopSink returns a sink that drops all events.
func NewNoopSink() Sink {
	return &noopSink{}
}

func (n *noopSink) Emit(_ Event) error {
	return nil
}

func (n *noopSink) Close() {}

type loggerSink struct {
	logger *zap.Logger
	level  zapcore.Level
}

// NewLoggerSink returns a sink that logs events to the given logger.
func NewLoggerSink(logger *zap.Logger, level zapcore.Level) Sink {
	if logger.Core().Enabled(level) {
		return &loggerSink{logger: logger, level: level}
	}
	return NewNoopSink()
}

func (s *loggerSink) Emit(event Event) error {
	jsonEvent, err := event.MarshalJSON()
	if err != nil {
		return err
	}
	s.logger.Log(s.level, string(jsonEvent))
	return nil
}

func (s *loggerSink) Close() {}

type filterSink struct {
	sink Sink
	fn   func(Event) bool
}

// NewFilterSink returns a sink that filters events based on the provided filter function.
// Only events for which the filter function returns true will be emitted to the wrapped sink.
func NewFilterSink(sink Sink, fn func(Event) bool) Sink {
	return &filterSink{sink: sink, fn: fn}
}

func (s *filterSink) Emit(event Event) error {
	if s.fn(event) {
		return s.sink.Emit(event)
	}
	return nil
}

func (s *filterSink) Close() {
	s.sink.Close()
}
