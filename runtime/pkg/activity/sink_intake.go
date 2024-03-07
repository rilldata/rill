package activity

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

type intakeSink struct {
	opts     IntakeSinkOptions
	logger   *zap.Logger
	buffer   []Event
	bufferMu sync.Mutex
	stop     chan struct{}
	wg       sync.WaitGroup
}

// IntakeSinkOptions provides options for NewIntakeSink.
type IntakeSinkOptions struct {
	IntakeURL      string
	IntakeUser     string
	IntakePassword string
	BufferSize     int
	SinkInterval   time.Duration
}

// NewIntakeSink creates a new sink that sends events to the Rill intake API.
func NewIntakeSink(logger *zap.Logger, opts IntakeSinkOptions) Sink {
	sink := &intakeSink{
		opts:   opts,
		logger: logger,
		buffer: make([]Event, 0, opts.BufferSize),
		stop:   make(chan struct{}),
	}

	go sink.runBackground()

	return sink
}

func (s *intakeSink) Emit(event Event) error {
	s.bufferMu.Lock()
	defer s.bufferMu.Unlock()

	s.buffer = append(s.buffer, event)
	if len(s.buffer) >= s.opts.BufferSize {
		s.flush()
	}

	return nil
}

func (s *intakeSink) Close() {
	// Prevent new flushes from being triggered
	close(s.stop)

	// Flush any remaining events in the buffer
	s.bufferMu.Lock()
	s.flush()
	s.bufferMu.Unlock()

	// Wait for all flushes to complete
	s.wg.Wait()
}

// runBackground periodically flushes the buffer.
func (s *intakeSink) runBackground() {
	ticker := time.NewTicker(s.opts.SinkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.bufferMu.Lock()
			s.flush()
			s.bufferMu.Unlock()
		case <-s.stop:
			return
		}
	}
}

// flush sends the buffered events to the intake server and resets the buffer.
// It is not a blocking operation. Use s.wg to wait for flushes to complete.
// The caller must hold s.mu when calling flush.
func (s *intakeSink) flush() {
	if len(s.buffer) == 0 {
		return
	}

	events := s.buffer
	s.buffer = make([]Event, 0, s.opts.BufferSize)

	s.wg.Add(1)
	defer s.wg.Done()

	go func() {
		err := s.send(events)
		if err != nil {
			s.logger.Error("could not flush activity events", zap.Error(err))
		}
	}()
}

// send sends the given events to the intake server.
func (s *intakeSink) send(events []Event) error {
	body := make([]byte, 0)
	for _, event := range events {
		data, err := event.Marshal()
		if err != nil {
			return fmt.Errorf("could not marshal event: %w", err)
		}
		body = append(body, data...)
		body = append(body, '\n')
	}

	req, err := http.NewRequest(http.MethodPost, s.opts.IntakeURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("could not create intake request: %w", err)
	}
	req.SetBasicAuth(s.opts.IntakeUser, s.opts.IntakePassword)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not send telemetry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("send telemetry failed with status code %d", resp.StatusCode)
	}

	return nil
}
