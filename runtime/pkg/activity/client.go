package activity

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Client interface {
	Emit(ctx context.Context, name string, value float64, dims ...Dim)
	Close() error
}

// bufferedClient collects and periodically sinks Event-s.
type bufferedClient struct {
	sinkPeriod time.Duration
	sink       Sink
	buffer     []Event
	bufferSize int
	bufferMx   sync.Mutex
	stop       chan struct{}
	sinkWg     sync.WaitGroup
	logger     *zap.Logger
}

type BufferedClientOptions struct {
	Sink       Sink
	SinkPeriod time.Duration
	BufferSize int
	Logger     *zap.Logger
}

func NewBufferedClient(opts BufferedClientOptions) Client {
	client := &bufferedClient{
		sinkPeriod: opts.SinkPeriod,
		sink:       opts.Sink,
		buffer:     make([]Event, 0, opts.BufferSize),
		bufferSize: opts.BufferSize,
		stop:       make(chan struct{}),
	}

	go client.init()

	return client
}

func (c *bufferedClient) Emit(ctx context.Context, name string, value float64, dims ...Dim) {
	dimsFromCtx := GetDimsFromContext(ctx)
	if dimsFromCtx == nil {
		dimsFromCtx = &[]Dim{}
	}

	if dims == nil {
		dims = []Dim{}
	}
	dims = append(*dimsFromCtx, dims...)

	event := Event{Time: time.Now(), Name: name, Value: value, Dims: dims}

	c.bufferMx.Lock()
	defer c.bufferMx.Unlock()

	c.buffer = append(c.buffer, event)

	if len(c.buffer) >= c.bufferSize {
		go func() {
			err := c.flush()
			if err != nil {
				c.logger.Error("could not flush activity events", zap.Error(err))
			}
		}()
	}
}

func (c *bufferedClient) Close() error {
	close(c.stop)
	err := c.flush() // Do not return the error immediately so concurrent flush calls can complete
	// Wait for all Sink calls to complete
	c.sinkWg.Wait()
	return err
}

func (c *bufferedClient) init() {
	ticker := time.NewTicker(c.sinkPeriod)

	for {
		select {
		case <-ticker.C:
			err := c.flush()
			if err != nil {
				c.logger.Error("could not flush activity events", zap.Error(err))
			}
		case <-c.stop:
			ticker.Stop()
			return
		}
	}
}

func (c *bufferedClient) flush() error {
	c.sinkWg.Add(1)
	defer c.sinkWg.Done()

	var events []Event
	// Sink call may take some time to process, so it's better to unlock the mutex early
	func() {
		c.bufferMx.Lock()
		defer c.bufferMx.Unlock()

		events = c.buffer
		c.buffer = make([]Event, 0, c.bufferSize)
	}()

	// If there are events, use a sink to process them
	if len(events) > 0 {
		err := c.sink.Sink(events)
		if err != nil {
			return err
		}
	}

	return nil
}

type noopClient struct{}

func NewNoopClient() Client {
	return &noopClient{}
}

func (n *noopClient) Emit(ctx context.Context, name string, value float64, dims ...Dim) {
}

func (n *noopClient) Close() error {
	return nil
}

type Event struct {
	Time  time.Time
	Name  string
	Value float64
	Dims  []Dim
}

type Dim struct {
	Name  string
	Value string
}

func String(name, value string) Dim {
	return Dim{Name: name, Value: value}
}

func (e *Event) Marshal() ([]byte, error) {
	// Create a map to hold the flattened event structure.
	flattened := make(map[string]interface{})

	// Add the non-dims fields.
	flattened["time"] = e.Time
	flattened["name"] = e.Name
	flattened["value"] = e.Value

	// Iterate over the dims slice and add each dim to the map.
	for _, dim := range e.Dims {
		flattened[dim.Name] = dim.Value
	}

	return json.Marshal(flattened)
}
