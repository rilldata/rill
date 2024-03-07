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
	jsonEvent, err := event.Marshal()
	if err != nil {
		return err
	}
	s.logger.Log(s.level, string(jsonEvent))
	return nil
}

func (s *loggerSink) Close() {}
