# scripts/observability

This directory contains a `docker-compose.yaml` file that configures an Otel Collector, Zipkin, and Prometheus for collecting and exploring telemetry in development (locally).

You can explore the telemetry on:
- `localhost:9411` for Zipkin
- `localhost:9412` for Prometheus

To collect telemetry, run the admin/runtime server with `OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317`.
