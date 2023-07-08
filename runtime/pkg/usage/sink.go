package usage

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
)

// Sink is used by a usage Client to sink accumulated events.
type Sink interface {
	Sink(events []Event)
	Close() error
}

type NoopSink struct{}

func NewNoopSink() *NoopSink {
	return &NoopSink{}
}

func (n *NoopSink) Sink(events []Event) {}

func (n *NoopSink) Close() error {
	return nil
}

type ConsoleSink struct {
	logger *zap.Logger
}

func NewConsoleSink(logger *zap.Logger) *ConsoleSink {
	return &ConsoleSink{logger: logger}
}

func (s *ConsoleSink) Sink(events []Event) {
	for _, e := range events {
		s.logger.Info(fmt.Sprintf("%v", e))
	}
}

func (s *ConsoleSink) Close() error {
	return nil
}

type FileSink struct {
	mu     sync.Mutex
	file   *os.File
	logger *zap.Logger
}

func NewFileSink(filename string, logger *zap.Logger) (*FileSink, error) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}

	return &FileSink{file: f, logger: logger}, nil
}

func (s *FileSink) Sink(events []Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, event := range events {
		data, err := convertEventToBytes(event)
		if err != nil {
			s.logger.Debug(fmt.Sprintf("could not serialize event: %v", event), zap.Error(err))
		}
		_, err = s.file.Write(data)
		if err != nil {
			s.logger.Debug(fmt.Sprintf("could not append event to a file: %v", event), zap.Error(err))
			continue
		}
		_, err = s.file.WriteString("\n")
		if err != nil {
			s.logger.Debug(fmt.Sprintf("could not append a separator to a file after event: %v", event), zap.Error(err))
		}
	}
}

func (s *FileSink) Close() error {
	return s.file.Close()
}

func convertEventToBytes(event Event) ([]byte, error) {
	// Create a map to hold the flattened event structure.
	flattened := make(map[string]interface{})

	// Add the non-dims fields.
	flattened["Time"] = event.Time
	flattened["Name"] = event.Name
	flattened["Value"] = event.Value

	// Iterate over the dims slice and add each dim to the map.
	for _, dim := range event.Dims {
		flattened[dim.Name] = dim.Value
	}

	return json.Marshal(flattened)
}
