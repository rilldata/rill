package reconcilers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/rilldata/rill/runtime/queries"
	"go.opentelemetry.io/otel/attribute"
)

// exportAttachmentSizeLimit is the maximum size of an export attached to emails.
// Exports larger than this are not attached.
const exportAttachmentSizeLimit = 10 * 1024 * 1024

var errAttachmentTooLarge = errors.New("export exceeds the attachment size limit")

// exportAttachmentOptions are the options for exportAttachment.
type exportAttachmentOptions struct {
	ResolverProperties map[string]any    // Properties of the "legacy_metrics" resolver to export
	Annotations        map[string]string // Annotations of the report or alert, used to find the explore it originates from
	ExecutionTime      time.Time
	Claims             *runtime.SecurityClaims
	Format             runtimev1.ExportFormat
	IncludeHeader      bool
	Limit              uint64 // Zero means no limit
	Priority           int
}

// exportAttachment exports the results of a legacy metrics query for attaching them to emails.
// It exports the same data as a download link for the query (see ExportReport in runtime/server).
// It returns errAttachmentTooLarge if the export is larger than exportAttachmentSizeLimit.
func exportAttachment(ctx context.Context, c *runtime.Controller, opts *exportAttachmentOptions) (*email.Attachment, error) {
	var ext, contentType string
	switch opts.Format {
	case runtimev1.ExportFormat_EXPORT_FORMAT_CSV:
		ext, contentType = "csv", "text/csv"
	case runtimev1.ExportFormat_EXPORT_FORMAT_XLSX:
		ext, contentType = "xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET:
		ext, contentType = "parquet", "application/octet-stream"
	default:
		return nil, fmt.Errorf("export format %q cannot be attached", opts.Format.String())
	}

	queryName, _ := opts.ResolverProperties["query_name"].(string)
	queryArgsJSON, _ := opts.ResolverProperties["query_args_json"].(string)
	qpb, err := queries.ProtoFromJSON(queryName, queryArgsJSON, &opts.ExecutionTime)
	if err != nil {
		return nil, err
	}
	q, err := queries.ProtoToExportQuery(qpb, opts.Claims, int64(opts.Limit), &opts.ExecutionTime)
	if err != nil {
		return nil, err
	}

	// Exports execute the query directly via q.Export (bypassing runtime.Query/Resolve), so emit the billable query metric here.
	if c.Activity != nil {
		queryAttrs := append(c.Runtime.GetInstanceAttributes(ctx, c.InstanceID), attribute.String("source", string(runtime.RequestSourceFromContext(ctx))))
		c.Activity.RecordMetric(ctx, "query", 1, queryAttrs...)
	}

	var originDashboard *runtimev1.ResourceName
	if explore := exploreNameFromAnnotations(opts.Annotations, ""); explore != "" {
		originDashboard = &runtimev1.ResourceName{Kind: runtime.ResourceKindExplore, Name: explore}
	}

	var filename string
	w := &limitedBuffer{limit: exportAttachmentSizeLimit}
	err = q.Export(ctx, c.Runtime, c.InstanceID, w, &runtime.ExportOptions{
		Format:   opts.Format,
		Priority: opts.Priority,
		PreWriteHook: func(name string) error {
			filename = name
			return nil
		},
		IncludeHeader:   opts.IncludeHeader,
		OriginDashboard: originDashboard,
	})
	if w.exceeded {
		return nil, errAttachmentTooLarge
	}
	if err != nil {
		return nil, err
	}

	return &email.Attachment{
		Filename:    fmt.Sprintf("%s_%s.%s", filename, opts.ExecutionTime.UTC().Format("20060102150405"), ext),
		ContentType: contentType,
		Data:        w.buf.Bytes(),
	}, nil
}

// limitedBuffer is an io.Writer that buffers data up to a size limit.
// It records whether a write exceeded the limit, since writers wrapping it may not propagate the error it returns.
// NOTE: It deliberately does not embed bytes.Buffer, which would let io.Copy bypass the limit through its io.ReaderFrom implementation.
type limitedBuffer struct {
	buf      bytes.Buffer
	limit    int
	exceeded bool
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if b.buf.Len()+len(p) > b.limit {
		b.exceeded = true
		return 0, errAttachmentTooLarge
	}
	return b.buf.Write(p)
}
