package observability

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/rilldata/rill/cli/pkg/dotrill"
	"go.opentelemetry.io/otel/sdk/trace"
	"gopkg.in/natefinch/lumberjack.v2"
)

// FileExporter writes OpenTelemetry traces to a rotating log file
type FileExporter struct {
	logger *log.Logger
	mu     sync.Mutex
}

var _ trace.SpanExporter = (*FileExporter)(nil)

// NewFileExporter initializes a file exporter with log rotation
func NewFileExporter() (*FileExporter, error) {
	filepath, err := dotrill.ResolveFilename("otel_traces.log", true)
	if err != nil {
		return nil, err
	}
	// Setup lumberjack for log rotation
	logWriter := &lumberjack.Logger{
		Filename:   filepath, // File path
		MaxSize:    10,       // Max file size in MB
		MaxBackups: 3,        // Keep up to 3 old log files
		MaxAge:     7,        // Keep logs for 7 days
		Compress:   false,    // Do not Compress old log files
	}

	// Create a logger
	logger := log.New(logWriter, "", 0)

	return &FileExporter{
		logger: logger,
	}, nil
}

// ExportSpans writes spans to the log file in JSON format
func (fe *FileExporter) ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error {
	fe.mu.Lock() // Prevent race conditions
	defer fe.mu.Unlock()

	for _, span := range spans {
		spanData := map[string]interface{}{
			"traceID":   span.SpanContext().TraceID().String(),
			"id":        span.SpanContext().SpanID().String(),
			"parentId":  span.Parent().SpanID().String(),
			"name":      span.Name(),
			"kind":      span.SpanKind().String(),
			"timestamp": span.StartTime().UnixMicro(),
			"duration":  span.EndTime().UnixMicro() - span.StartTime().UnixMicro(),
		}
		attributes := span.Attributes()
		m := make(map[string]string, len(attributes))
		spanData["tags"] = m
		for _, attr := range attributes {
			m[string(attr.Key)] = attr.Value.Emit()
		}

		jsonData, err := json.Marshal(spanData)
		if err != nil {
			fe.logger.Println("Error marshalling span:", err)
			continue
		}
		fe.logger.Println(string(jsonData)) // Write to file
	}

	return nil
}

// Shutdown closes the file (if needed)
func (fe *FileExporter) Shutdown(ctx context.Context) error {
	return nil // No explicit shutdown needed for lumberjack
}
