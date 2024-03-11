package activity

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

// Client constructs telemetry events and sends them to a sink.
type Client struct {
	logger    *zap.Logger
	sink      Sink
	withAttrs []attribute.KeyValue
}

// NewClient creates a base telemetry client that sends events to the provided sink.
// The client will close the sink when Close is called.
func NewClient(sink Sink, logger *zap.Logger) *Client {
	client := &Client{
		logger: logger,
		sink:   sink,
	}

	return client
}

// NewNoopClient creates a client that discards all events.
func NewNoopClient() *Client {
	return NewClient(NewNoopSink(), zap.NewNop())
}

// Close the client. Also closes the sink passed to NewClient.
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

// With returns a copy of the client that will set the provided attributes on all events.
func (c *Client) With(attrs ...attribute.KeyValue) *Client {
	if len(attrs) == 0 {
		return c
	}

	res := make([]attribute.KeyValue, len(c.withAttrs)+len(attrs))
	copy(res, c.withAttrs)
	copy(res[len(c.withAttrs):], attrs)

	return &Client{
		logger:    c.logger,
		sink:      c.sink,
		withAttrs: res,
	}
}

// WithServiceName returns a copy of the client with an attribute set for AttrKeyServiceName.
func (c *Client) WithServiceName(serviceName string) *Client {
	return c.With(attribute.String(AttrKeyServiceName, serviceName))
}

// WithServiceVersion returns a copy of the client with attributes set for AttrKeyServiceVersion and AttrKeyServiceCommit.
func (c *Client) WithServiceVersion(number, commit string) *Client {
	return c.With(attribute.String(AttrKeyServiceVersion, number), attribute.String(AttrKeyServiceCommit, commit))
}

// WithIsDev returns a copy of the client with an attribute set for AttrKeyIsDev.
func (c *Client) WithIsDev() *Client {
	return c.With(attribute.Bool(AttrKeyIsDev, true))
}

// WithInstallID returns a copy of the client with an attribute set for AttrKeyInstallID.
func (c *Client) WithInstallID(installID string) *Client {
	return c.With(attribute.String(AttrKeyInstallID, installID))
}

// WithUserID returns a copy of the client with an attribute set for AttrKeyUserID.
func (c *Client) WithUserID(userID string) *Client {
	return c.With(attribute.String(AttrKeyUserID, userID))
}

// Record sends a generic telemetry event with the provided event type and name.
func (c *Client) Record(ctx context.Context, typ, name string, extraAttrs ...attribute.KeyValue) {
	c.emitRaw(Event{
		EventID:   uuid.New().String(),
		EventTime: time.Now(),
		EventType: typ,
		EventName: name,
		Data:      c.resolveAttrs(ctx, extraAttrs),
	})
}

// RecordMetric sends a telemetry event of type "metric" with the provided name and value.
func (c *Client) RecordMetric(ctx context.Context, name string, value float64, attrs ...attribute.KeyValue) {
	c.emitRaw(Event{
		EventID:   uuid.New().String(),
		EventTime: time.Now(),
		EventType: EventTypeMetric,
		EventName: name,
		Data: c.resolveAttrs(ctx, attrs,
			attribute.Float64("value", value),
			// Backwards compatibility with a previous format (before event_name and event_time)
			attribute.String("name", name),
			attribute.String("time", time.Now().Format(time.RFC3339)),
		),
	})
}

// EmitBehavioral sends a telemetry event of type "behavioral" with the provided name and attributes.
// The event additionally has all the attributes associated with out legacy behavioral events.
// It will panic if all of WithServiceName, WithServiceVersion, WithInstallID and WithUserID have not been called on the client.
func (c *Client) RecordBehavioralLegacy(name string, extraAttrs ...attribute.KeyValue) {
	// For compatibility with the legacy behavioral events, we need to ensure the output has at least these properties:
	//     app_name       string
	//     install_id     string
	//     build_id       string
	//     version        string
	//     user_id        string
	//     is_dev         bool
	//     mode           string
	//     action         string
	//     medium         string
	//     space          string
	//     screen_name    string
	//     event_datetime int64
	//     event_type     string
	//     payload        map[string]any

	data := c.resolveAttrs(context.Background(), extraAttrs)

	if _, ok := data["install_id"]; !ok {
		panic("install_id is required for a legacy behavioral event")
	}

	if _, ok := data["user_id"]; !ok {
		panic("user_id is required for a legacy behavioral event")
	}

	val, ok := data[AttrKeyServiceCommit]
	if !ok {
		panic("service_commit is required for a legacy behavioral event")
	}
	data["build_id"] = val

	val, ok = data[AttrKeyServiceVersion]
	if !ok {
		panic("service_version is required for a legacy behavioral event")
	}
	data["version"] = val

	if val, ok := data["olap_connector"]; ok {
		payload := make(map[string]any)
		payload["olap_connector"] = val
		if conns, ok := data["connectors"]; ok {
			payload["connectors"] = conns
		}
		data["payload"] = payload
	}

	data["app_name"] = "rill-developer"
	data["mode"] = "edit"
	data["action"] = name
	data["medium"] = "cli"
	data["space"] = "terminal"
	data["screen_name"] = "terminal"

	t := time.Now()
	data["event_datetime"] = t.Unix() * 1000

	c.emitRaw(Event{
		EventID:   uuid.New().String(),
		EventTime: t,
		EventType: EventTypeBehavioral,
		EventName: name,
		Data:      data,
	})
}

// RecordRaw proxies a raw event represented as a map to the client's sink.
// It does not enrich the provided event with any of the client's contextual attributes.
// It returns an error if the event does not contain the required fields (see the Event type for required fields).
func (c *Client) RecordRaw(data map[string]any) error {
	// Ensure the event is not nil
	if data == nil {
		return fmt.Errorf("empty event")
	}

	// Pop event_id
	id, ok := data["event_id"].(string)
	if !ok {
		return fmt.Errorf("missing event_id")
	}
	delete(data, "event_id")

	// Pop event_time
	tStr, ok := data["event_time"].(string)
	if !ok {
		return fmt.Errorf("missing event_time")
	}
	t, err := time.Parse(time.RFC3339Nano, tStr)
	if err != nil {
		return fmt.Errorf("failed to parse event_time: %w", err)
	}
	delete(data, "event_time")

	// Pop event_type
	typ, ok := data["event_type"].(string)
	if !ok {
		return fmt.Errorf("missing event_type")
	}
	delete(data, "event_type")

	// Pop event_name
	name, ok := data["event_name"].(string)
	if !ok {
		return fmt.Errorf("missing event_name")
	}
	delete(data, "event_name")

	// Emit the event
	c.emitRaw(Event{
		EventID:   id,
		EventTime: t,
		EventType: typ,
		EventName: name,
		Data:      data,
	})
	return nil
}

// emitRaw sends an event to the sink.
func (c *Client) emitRaw(e Event) {
	err := c.sink.Emit(e)
	if err != nil {
		c.logger.Error("Failed to emit event", zap.Error(err))
	}
}

// resolveAttrs combines the attributes from the client, context, and args into a map.
func (c *Client) resolveAttrs(ctx context.Context, extraAttrs []attribute.KeyValue, extraExtraAttrs ...attribute.KeyValue) map[string]any {
	n := len(c.withAttrs) + len(extraAttrs) + len(extraExtraAttrs)
	attrsFromCtx := attrsFromContext(ctx)
	if attrsFromCtx != nil {
		n += len(*attrsFromCtx)
	}

	if n == 0 {
		return nil
	}

	data := make(map[string]any, n+4) // +4 to leave room for the common fields without reallocation.

	for _, a := range c.withAttrs {
		data[string(a.Key)] = a.Value.AsInterface()
	}

	if attrsFromCtx != nil {
		for _, a := range *attrsFromCtx {
			data[string(a.Key)] = a.Value.AsInterface()
		}
	}

	for _, a := range extraAttrs {
		data[string(a.Key)] = a.Value.AsInterface()
	}

	for _, a := range extraExtraAttrs {
		data[string(a.Key)] = a.Value.AsInterface()
	}

	return data
}
