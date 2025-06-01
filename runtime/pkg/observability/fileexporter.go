package observability

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"
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
	filepath, err := dotrill.New("").ResolveFilename("otel_traces.log", true)
	if err != nil {
		return nil, err
	}
	// Setup lumberjack for log rotation
	logWriter := &lumberjack.Logger{
		Filename:   filepath, // File path
		MaxSize:    100,      // Max file size in MB
		MaxBackups: 1,        // Keep 1 backup
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
		spanData := traceSpan{
			TraceID:   span.SpanContext().TraceID().String(),
			ID:        span.SpanContext().SpanID().String(),
			ParentID:  span.Parent().SpanID().String(),
			Name:      span.Name(),
			Kind:      span.SpanKind().String(),
			Timestamp: span.StartTime().UnixMicro(),
			Duration:  span.EndTime().UnixMicro() - span.StartTime().UnixMicro(),
		}
		attributes := span.Attributes()
		m := make(map[string]string, len(attributes))
		spanData.Tags = m
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

func SearchTracesFile(ctx context.Context, traceID, resourceName string) ([]byte, error) {
	db, err := sqlx.Open("duckdb", "")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	fp, err := dotrill.New("").ResolveFilename("otel_traces*.log", true)
	if err != nil {
		return nil, err
	}

	var query string
	var args []any
	if resourceName != "" {
		query = fmt.Sprintf(`
			SELECT 
				replace(traceID::VARCHAR, '-', '') AS traceID, 
				tags::JSON::BLOB AS tags, 
				* EXCLUDE (traceID, tags)
			FROM read_json_auto(%s)
			WHERE traceID = (
				SELECT traceID
				FROM read_json_auto(%s)
				WHERE name ILIKE '%%Reconcile%%' AND json_extract_string(tags::JSON, '$.name') ILIKE ?
				ORDER BY timestamp DESC
				LIMIT 1
			)
		`, escapeStringValue(fp), escapeStringValue(fp))
		args = []any{resourceName}
	} else {
		query = fmt.Sprintf(`
			SELECT 
				replace(traceID::VARCHAR, '-', '') AS traceID, 
				tags::JSON::BLOB AS tags, 
				* EXCLUDE (traceID, tags) 
			FROM 
				read_json_auto(%s) 
			WHERE 
				traceID = ?`, escapeStringValue(fp))
		args = []any{traceID}
	}
	rows, err := db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var spans []traceSpan

	for rows.Next() {
		var span traceSpan
		if err := rows.StructScan(&span); err != nil {
			return nil, err
		}

		// Manually decode tags string to map
		if len(span.TagsRaw) > 0 {
			if err := json.Unmarshal(span.TagsRaw, &span.Tags); err != nil {
				return nil, err
			}
		}

		spans = append(spans, span)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(spans) == 0 {
		return nil, sql.ErrNoRows
	}

	data, err := json.Marshal(spans)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func escapeStringValue(s string) string {
	return fmt.Sprintf("'%s'", strings.ReplaceAll(s, "'", "''"))
}

type traceSpan struct {
	TraceID   string            `db:"traceID" json:"traceID"`
	Duration  int64             `db:"duration" json:"duration"`
	ID        string            `db:"id" json:"id"`
	Kind      string            `db:"kind" json:"kind"`
	Name      string            `db:"name" json:"name"`
	ParentID  string            `db:"parentId" json:"parentId"`
	TagsRaw   []byte            `db:"tags" json:"-"`
	Tags      map[string]string `json:"tags,omitempty"`
	Timestamp int64             `db:"timestamp" json:"timestamp"`
}
