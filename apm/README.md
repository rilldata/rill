This Docker Compose configuration brings a metrics/tracing backend in local environment, these components are started:
* Open Telemetry collector
* Zipkin
* Prometheus

Run Runtime with this environment variable defined `RILL_OTEL_EXPORTER_ENDPOINT=localhost:4317`