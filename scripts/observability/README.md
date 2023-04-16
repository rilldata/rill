# scripts/observability

This directory contains a `docker-compose.yaml` file that configures an Otel Collector, Zipkin, and Prometheus for collecting and exploring telemetry in development (locally).

Run it with:
```bash
docker-compose -f scripts/observability/docker-compose.yaml up 
```

You can explore the telemetry on:
- `http://localhost:9411` for Zipkin
- `http://localhost:9412` for Prometheus

To explore telemetry from the admin server, set:
```
RILL_ADMIN_METRICS_EXPORTER=otel
RILL_ADMIN_TRACES_EXPORTER=otel
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
```

To explore telemetry from the runtime server, set:
```
RILL_RUNTIME_METRICS_EXPORTER=otel
RILL_RUNTIME_TRACES_EXPORTER=otel
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
```

(For details about all `OTEL_` env vars, see [open-telemetry/opentelemetry-go](https://github.com/open-telemetry/opentelemetry-go/tree/main/exporters/otlp/otlptrace).)
