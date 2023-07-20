package publisher

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

// Client collects and periodically sinks Event-s.
type Client struct {
	sinkPeriod time.Duration
	sink       Sink
	buffer     []Event
	bufferSize int
	bufferMx   sync.Mutex
	stop       chan struct{}
	sinkWg     sync.WaitGroup
}

type Options struct {
	Sink       Sink
	SinkPeriod time.Duration
	BufferSize int
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

func String(name, value string) *Dim {
	return &Dim{Name: name, Value: value}
}

func New(opts Options) *Client {
	client := &Client{
		sinkPeriod: opts.SinkPeriod,
		sink:       opts.Sink,
		buffer:     make([]Event, 0, opts.BufferSize),
		bufferSize: opts.BufferSize,
		stop:       make(chan struct{}),
	}

	go client.start()

	return client
}

func (c *Client) Emit(ctx context.Context, name string, value float64, dims ...Dim) {
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
		go c.flush()
	}
}

func (c *Client) Stop() {
	close(c.stop)
	c.flush()
	// Wait for all Sink calls to complete
	c.sinkWg.Wait()
}

func (c *Client) start() {
	ticker := time.NewTicker(c.sinkPeriod)

	for {
		select {
		case <-ticker.C:
			c.flush()
		case <-c.stop:
			ticker.Stop()
			return
		}
	}
}

func (c *Client) flush() {
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
		c.sink.Sink(events)
	}
}

func (e *Event) Marshal() ([]byte, error) {
	// Create a map to hold the flattened event structure.
	flattened := make(map[string]interface{})

	// Add the non-dims fields.
	flattened["Time"] = e.Time
	flattened["Name"] = e.Name
	flattened["Value"] = e.Value

	// Iterate over the dims slice and add each dim to the map.
	for _, dim := range e.Dims {
		flattened[dim.Name] = dim.Value
	}

	return json.Marshal(flattened)
}
