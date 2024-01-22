package activity

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Sink is used by a bufferedClient to flush collected Event-s.
type Sink interface {
	Sink(ctx context.Context, events []Event) error
	// SetActivity sets activity client so that a sink can emit activity events
	SetActivity(activity Client)
	Close() error
}

type NoopSink struct{}

func NewNoopSink() *NoopSink {
	return &NoopSink{}
}

func (n *NoopSink) Sink(_ context.Context, _ []Event) error {
	return nil
}

func (n *NoopSink) SetActivity(activity Client) {}

func (n *NoopSink) Close() error {
	return nil
}

type ConsoleSink struct {
	logger *zap.Logger
	level  zapcore.Level
}

// NewConsoleSink might be used for a local run
func NewConsoleSink(logger *zap.Logger, level zapcore.Level) *ConsoleSink {
	return &ConsoleSink{logger: logger, level: level}
}

func (s *ConsoleSink) Sink(_ context.Context, events []Event) error {
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

func (s *ConsoleSink) SetActivity(activity Client) {}

func (s *ConsoleSink) Close() error {
	return nil
}
