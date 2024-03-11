# Events telemetry

## Types of telemetry in Rill

In Rill, we have two kinds of telemetry:

- OpenTelemetry: Used for emitting and collecting traditional logging, tracing, and metrics in a hosted setting. Encapsulated in the `runtime/pkg/observability` package.
- Events telemetry: Used for event telemetry, such as user behavior, browser errors, system metrics, billing events, etc. Encapsulated in `runtime/pkg/activity` (this package).

Why do we have a separate implementation for events telemetry? Two reasons:
1. The event telemetry is collected from all combinations of frontend/CLI/backend services and local/cloud-hosted environments. At the time of writing, OpenTelemetry is not well-suited for capturing data across such disparate environments.
2. At the time of writing, OpenTelemetry does not have mature abstractions for capturing event data.

## Sources of events telemetry and how it propagates

Event telemetry propagate in various ways depending on the source service and environment:
1. The cloud admin server sends events directly to Kafka.
2. The cloud runtime sends events directly to Kafka.
3. The cloud admin UI sends events to the cloud admin server, which proxies them to Kafka.
4. The local CLI sends events to Rill's intake API.
5. The local UI sends events to the local server hosted by the CLI, which proxies them to Rill's intake API.
6. The local runtime sends events to Rill's intake API (using the same client as the CLI).

## Required event format

We do not enforce strict schemas for different event types, but we do require all events to have the following format:

```json
{
    "event_id": "<generated unique ID for deduplication>",
    "event_time": "<timestamp in ISO8601 format>",
    "event_type": "<high-level event type, such as 'behavioral', 'metric', ...>",
    "event_name": "<unique identifer of the event within the event type, such as 'login_start' or 'duckdb_estimated_size'>"
    // Other event-specific attributes
}
```

Events that do not contain these fields cannot be generated or proxied through the client.

## Common event types, names and attributes

The `event.go` file in this package contains constants for common event types/names and common attributes.

