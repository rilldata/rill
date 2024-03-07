package activity

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

// Client constructs telemetry events and sends them to a sink.
type Client struct {
	logger     *zap.Logger
	sink       Sink
	withOpts   *ClientOptions
	withAttrs  []attribute.KeyValue
	withAnonID string
	withUserID string
}

// ClientOptions provides options for creating a client.
type ClientOptions struct {
	ServiceName   string
	Version       string
	VersionCommit string
	VersionDev    bool
}

// NewClient creates a base telemetry client that sends events to the provided sink.
// The client will close the sink when Close is called.
func NewClient(sink Sink, logger *zap.Logger, opts *ClientOptions) *Client {
	client := &Client{
		logger:   logger,
		sink:     sink,
		withOpts: opts,
	}

	var attrs []attribute.KeyValue
	if opts.ServiceName != "" {
		attrs = append(attrs, attribute.String("service_name", opts.ServiceName))
	}
	if opts.Version != "" {
		attrs = append(attrs, attribute.String("service_version", opts.Version))
	}

	return client.With(attrs...)
}

// NewNoopClient creates a client that discards all events.
func NewNoopClient() *Client {
	return NewClient(NewNoopSink(), zap.NewNop(), &ClientOptions{})
}

func (c *Client) Close(ctx context.Context) error {
	// Close the sink in the background.
	done := make(chan struct{})
	go func() {
		defer close(done)
		c.sink.Close()
	}()

	// Wait for the sink to close or the context to be done.
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Client) With(attrs ...attribute.KeyValue) *Client {
	if len(attrs) == 0 {
		return c
	}

	return &Client{
		logger:     c.logger,
		sink:       c.sink,
		withOpts:   c.withOpts,
		withAttrs:  append(attrs, c.withAttrs...),
		withAnonID: c.withAnonID,
		withUserID: c.withUserID,
	}
}

func (c *Client) WithIdentity(anonymousID, userID string) *Client {
	return &Client{
		logger:     c.logger,
		sink:       c.sink,
		withOpts:   c.withOpts,
		withAttrs:  c.withAttrs,
		withAnonID: anonymousID,
		withUserID: userID,
	}
}

func (c *Client) EmitMetric(ctx context.Context, name string, value float64, attrs ...attribute.KeyValue) {
	attrsFromCtx := attrsFromContext(ctx)
	if attrsFromCtx == nil {
		attrsFromCtx = &[]attribute.KeyValue{}
	}

	if attrs == nil {
		attrs = []attribute.KeyValue{}
	}
	attrs = append(*attrsFromCtx, attrs...)

	c.emitRaw(&MetricEvent{
		Time:  time.Now(),
		Name:  name,
		Value: value,
		Attrs: attrs,
	})
}

func (c *Client) EmitUserAction(action string, attrs ...attribute.KeyValue) {
	// Note: Not adding attrs to c.withAttrs because these user events are so broken.
	var payload map[string]any
	if len(attrs) != 0 {
		payload = make(map[string]any, len(attrs))
		for _, attr := range attrs {
			payload[string(attr.Key)] = attr.Value.AsInterface()
		}
	}

	e := &UserEvent{
		AppName:       c.withOpts.ServiceName,
		InstallID:     c.withAnonID,
		BuildID:       c.withOpts.VersionCommit,
		Version:       c.withOpts.Version,
		UserID:        c.withUserID,
		IsDev:         c.withOpts.VersionDev,
		Mode:          "edit",
		Action:        action,
		Medium:        "cli",
		Space:         "terminal",
		ScreenName:    "terminal",
		EventDatetime: time.Now().Unix() * 1000,
		EventType:     "behavioral",
		Payload:       payload,
	}

	c.emitRaw(e)
}

func (c *Client) EmitUserActionRaw(jsonData []byte) error {
	var e *UserEvent
	err := json.Unmarshal(jsonData, &e)
	if err != nil {
		return fmt.Errorf("failed to unmarshal user event: %w", err)
	}

	if e == nil {
		return fmt.Errorf("empty user event")
	}

	// TODO: More validation?

	c.emitRaw(e)
	return nil
}

func (c *Client) emitRaw(e Event) {
	err := c.sink.Emit(e)
	if err != nil {
		c.logger.Error("Failed to emit event", zap.Error(err))
	}
}
