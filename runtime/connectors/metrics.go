package connectors

import (
	"context"
	"time"

	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
)

var (
	meter                 = global.Meter("runtime/connectors")
	downloadTimeHistogram = observability.Must(meter.Float64Histogram("download.time", instrument.WithUnit("s")))
	downloadSizeCounter   = observability.Must(meter.Int64UpDownCounter("download.size", instrument.WithUnit("bytes")))
	downloadSpeedCounter  = observability.Must(meter.Float64UpDownCounter("download.speed", instrument.WithUnit("bytes/s")))
)

type DownloadMetrics struct {
	Connector string
	Ext       string
	Partial   bool
	Duration  time.Duration
	Size      int64
}

func RecordDownloadMetrics(ctx context.Context, m *DownloadMetrics) {
	attrs := attribute.NewSet(
		attribute.String("connector", m.Connector),
		attribute.String("ext", m.Ext),
		attribute.Bool("partial", m.Partial),
	)

	downloadTimeHistogram.Record(ctx, m.Duration.Seconds(), metric.WithAttributeSet(attrs))
	downloadSizeCounter.Add(ctx, m.Size, metric.WithAttributeSet(attrs))

	secs := m.Duration.Seconds()
	if secs != 0 {
		downloadSpeedCounter.Add(ctx, float64(m.Size)/secs, metric.WithAttributeSet(attrs))
	}
}
