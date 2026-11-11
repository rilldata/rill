package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/httputil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/pkg/workbook"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/encoding/protojson"
)

// maxWorkbookRequestBytes caps the body of a workbook export: the data sheets and PNG snapshots of a large canvas.
const maxWorkbookRequestBytes = 64 << 20

// exportWorkbookHandler renders a canvas dashboard as an Excel workbook.
// The body is a protojson-encoded CanvasWorkbookExportRequest (see proto/rill/runtime/v1/workbook.proto)
// that the browser builds from the data it already displays, so the handler only formats it;
// it requires the caller to be able to open the canvas, which keeps the endpoint from serving as a generic formatter.
func (s *Server) exportWorkbookHandler(w http.ResponseWriter, req *http.Request) error {
	ctx := req.Context()
	instanceID := req.PathValue("instance_id")
	observability.AddRequestAttributes(ctx, attribute.String("args.instance_id", instanceID))
	s.addInstanceRequestAttributes(ctx, instanceID)

	claims := auth.GetClaims(ctx, instanceID)
	if !claims.Can(runtime.ReadMetrics) {
		return httputil.Errorf(http.StatusUnauthorized, "action not allowed")
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, req.Body, maxWorkbookRequestBytes))
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return httputil.Errorf(http.StatusRequestEntityTooLarge, "the export is larger than the limit of %d bytes", maxWorkbookRequestBytes)
		}
		return httputil.Errorf(http.StatusBadRequest, "failed to read request: %s", err)
	}
	wb := &runtimev1.CanvasWorkbookExportRequest{}
	// Unknown fields are ignored like on the gRPC-gateway routes, so a UI shipped ahead of the runtime can still export.
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(body, wb); err != nil {
		return httputil.Errorf(http.StatusBadRequest, "failed to parse request: %s", err)
	}
	if wb.Canvas == "" {
		return httputil.Errorf(http.StatusBadRequest, "canvas is required")
	}

	_, err = s.runtime.ResolveCanvas(ctx, instanceID, wb.Canvas, claims, &runtime.ResolveCanvasOptions{})
	if err != nil {
		if errors.Is(err, runtime.ErrForbidden) {
			return httputil.Errorf(http.StatusForbidden, "action not allowed")
		}
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return httputil.Errorf(http.StatusNotFound, "canvas %q not found", wb.Canvas)
		}
		return httputil.Error(http.StatusInternalServerError, err)
	}

	// Render into memory before writing the headers, so a failure returns an error status instead of a truncated file.
	var buf bytes.Buffer
	if err := workbook.Write(&buf, wb); err != nil {
		if workbook.IsValidationError(err) {
			return httputil.Error(http.StatusBadRequest, err)
		}
		return httputil.Error(http.StatusInternalServerError, err)
	}

	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.xlsx\"", workbookFilename(wb)))
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	_, err = w.Write(buf.Bytes())
	return err
}

// workbookFilename derives a header-safe file name from the workbook title (or the canvas name) plus a timestamp,
// like the query export downloads.
func workbookFilename(wb *runtimev1.CanvasWorkbookExportRequest) string {
	name := strings.TrimSpace(wb.Meta.GetTitle())
	if name == "" {
		name = wb.Canvas
	}
	name = strings.Map(func(r rune) rune {
		if r < utf8.RuneSelf && (unicode.IsLetter(r) || unicode.IsDigit(r)) {
			return r
		}
		return '_'
	}, name)
	return name + "_" + time.Now().Format("20060102150405")
}
