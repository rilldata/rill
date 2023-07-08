package usage

import (
	"context"
	"time"
)

type Client struct {
	sinkPeriod time.Duration
	snk        Sink
	queue      chan Event
	stop       chan bool
}

type Conf struct {
	Sink       Sink
	SinkPeriod time.Duration
	QueueSize  int
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

func NewClient(conf Conf) *Client {
	client := &Client{
		sinkPeriod: conf.SinkPeriod,
		snk:        conf.Sink,
		queue:      make(chan Event, conf.QueueSize),
		stop:       make(chan bool),
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
	c.queue <- Event{Time: time.Now(), Name: name, Value: value, Dims: dims}
}

func (c *Client) Stop() {
	c.stop <- true
}

func (c *Client) start() {
	ticker := time.NewTicker(c.sinkPeriod)

	for {
		select {
		case <-ticker.C:
			c.sink()
		case <-c.stop:
			ticker.Stop()
			close(c.queue)
			return
		}
	}
}

func (c *Client) sink() {
	// Process all events in the queue
	var events []Event
	for len(c.queue) > 0 {
		event := <-c.queue
		events = append(events, event)
	}

	// If there are events, use snk to process them
	if len(events) > 0 {
		c.snk.Sink(events)
	}
}
