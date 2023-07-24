package activity

import (
	"go.uber.org/zap"
)

// Sink is used by a bufferedClient to flush collected Event-s.
type Sink interface {
	Sink(events []Event) error
	Close() error
}

type NoopSink struct{}

func NewNoopSink() *NoopSink {
	return &NoopSink{}
}

func (n *NoopSink) Sink(_ []Event) error {
	return nil
}

func (n *NoopSink) Close() error {
	return nil
}

type ConsoleSink struct {
	logger *zap.Logger
}

// NewConsoleSink might be used for a local run
func NewConsoleSink(logger *zap.Logger) *ConsoleSink {
	return &ConsoleSink{logger: logger}
}

func (s *ConsoleSink) Sink(events []Event) error {
	for _, e := range events {
		jsonEvent, err := e.Marshal()
		if err != nil {
			return err
		}
		s.logger.Info(string(jsonEvent))
	}
	return nil
}

func (s *ConsoleSink) Close() error {
	return nil
}
