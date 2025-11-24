package observability

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-logr/zapr"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	"go.uber.org/zap"
)

// Exporter lists available telemetry exporters
type Exporter string

const (
	NoopExporter       Exporter = ""
	OtelExporter       Exporter = "otel"
	FileBasedExporter  Exporter = "file"
	PrometheusExporter Exporter = "prometheus"
)

// Options for configuring telemetry setup
type Options struct {
	MetricsExporter Exporter
	TracesExporter  Exporter
	ServiceName     string
	ServiceVersion  string
}

// ShutdownFunc stops global telemetry
type ShutdownFunc func(context.Context) error

// Start configures traces and metrics globally.
// Use "go.opentelemetry.io/otel.Tracer" to access global tracers.
// Use "go.opentelemetry.io/otel.Meter" to access global meters.
// If using OtelExporter (otel collector), make sure to set the OTEL_EXPORTER_OTLP_ENDPOINT env var.
// For a full list of Otel env vars, see: https://github.com/open-telemetry/opentelemetry-go/tree/main/exporters/otlp/otlptrace.
func Start(ctx context.Context, logger *zap.Logger, opts *Options) (ShutdownFunc, error) {
	// Log otel info and errors using Zap
	otel.SetLogger(zapr.NewLogger(logger))
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		logger.Error("otel error", zap.Error(err))
	}))

	// Create resource representing the currently running service
	res, err := resource.New(ctx,
		resource.WithContainer(),
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithAttributes(
			semconv.ServiceName(opts.ServiceName),
			semconv.ServiceVersion(opts.ServiceVersion),
		),
		resource.WithFromEnv(),
	)
	if err != nil {
		if !errors.Is(err, resource.ErrPartialResource) && !errors.Is(err, resource.ErrSchemaURLConflict) {
			return nil, err
		}
		logger.Info("otel resource only partially detected", zap.Error(err))
	}

	// Create global metrics exporter
	var meterProvider *metric.MeterProvider
	switch opts.MetricsExporter {
	case PrometheusExporter:
		exp, err := prometheus.New()
		if err != nil {
			return nil, err
		}
		meterProvider = metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(exp))
	case OtelExporter:
		exp, err := otlpmetricgrpc.New(ctx)
		if err != nil {
			return nil, err
		}
		r := metric.NewPeriodicReader(exp, metric.WithInterval(15*time.Second))
		meterProvider = metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(r))
	case NoopExporter:
		// Nothing to do
	default:
		panic(fmt.Errorf("unexpected metrics exporter %q", opts.MetricsExporter))
	}

	// Set global meter provider
	if meterProvider != nil {
		otel.SetMeterProvider(meterProvider)
	}

	// Create global traces exporter
	var tracerProvider *trace.TracerProvider
	switch opts.TracesExporter {
	case OtelExporter:
		client := otlptracegrpc.NewClient()
		exp, err := otlptrace.New(ctx, client)
		if err != nil {
			return nil, err
		}
		bsp := trace.NewBatchSpanProcessor(exp)
		tracerProvider = trace.NewTracerProvider(
			trace.WithSampler(trace.AlwaysSample()),
			trace.WithResource(res),
			trace.WithSpanProcessor(bsp),
		)
	case FileBasedExporter:
		exp, err := NewFileExporter()
		if err != nil {
			return nil, err
		}
		bsp := trace.NewBatchSpanProcessor(exp)
		tracerProvider = trace.NewTracerProvider(
			trace.WithSampler(trace.AlwaysSample()),
			trace.WithResource(res),
			trace.WithSpanProcessor(bsp),
		)
	case NoopExporter:
		// Nothing to do
	default:
		panic(fmt.Errorf("unexpected traces exporter %q", opts.TracesExporter))
	}

	// Set global tracer provider
	if tracerProvider != nil {
		otel.SetTracerProvider(tracerProvider)
	}

	// Collect metrics from the Go runtime (polls every 15s by default)
	if meterProvider != nil {
		err = runtime.Start()
		if err != nil {
			return nil, err
		}
	}

	// Create callback to shut down globals
	shutdown := func(ctx context.Context) error {
		var errs []error
		if meterProvider != nil {
			if err := meterProvider.Shutdown(ctx); err != nil {
				errs = append(errs, err)
			}
		}
		if tracerProvider != nil {
			if err := tracerProvider.ForceFlush(ctx); err != nil {
				errs = append(errs, fmt.Errorf("failed to flush spans: %w", err))
			}
			if err := tracerProvider.Shutdown(ctx); err != nil {
				errs = append(errs, err)
			}
		}
		if len(errs) == 1 {
			return fmt.Errorf("error while shutting down: %w", errs[0])
		}
		if len(errs) > 1 {
			return fmt.Errorf("multiple errors while shutting down: %v", errs)
		}
		return nil
	}
	return shutdown, nil
}
