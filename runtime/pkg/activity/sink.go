package activity

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Sink is used by a bufferedClient to flush collected Event-s.
type Sink interface {
	Sink(events []Event) error
	Close() error
}

type noopSink struct{}

func NewNoopSink() *noopSink {
	return &noopSink{}
}

func (n *noopSink) Sink(_ []Event) error {
	return nil
}

func (n *noopSink) Close() error {
	return nil
}

type consoleSink struct {
	logger *zap.Logger
	level  zapcore.Level
}

// NewConsoleSink might be used for a local run
func NewConsoleSink(logger *zap.Logger, level zapcore.Level) *consoleSink {
	return &consoleSink{logger: logger, level: level}
}

func (s *consoleSink) Sink(events []Event) error {
	if s.logger.Core().Enabled(s.level) {
		for _, e := range events {
			jsonEvent, err := e.Marshal()
			if err != nil {
				return err
			}
			s.logger.Log(s.level, string(jsonEvent))
		}
	}
	return nil
}

func (s *consoleSink) Close() error {
	return nil
}
