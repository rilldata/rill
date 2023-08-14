package server

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/bufbuild/connect-go"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/server/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func (s *Server) Export(ctx context.Context, req *connect.Request[runtimev1.ExportRequest]) (*connect.Response[runtimev1.ExportResponse], error) {
	if !auth.GetClaims(ctx).CanInstance(req.Msg.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	if s.opts.DownloadRowLimit != nil {
		if req.Msg.Limit == nil {
			req.Msg.Limit = s.opts.DownloadRowLimit
		}
		if *req.Msg.Limit > *s.opts.DownloadRowLimit {
			return nil, status.Errorf(codes.InvalidArgument, "limit must be less than or equal to %d", *s.opts.DownloadRowLimit)
		}
	}

	r, err := proto.Marshal(req.Msg)
	if err != nil {
		return nil, err
	}

	out := fmt.Sprintf("/v1/download?%s=%s", "request", base64.URLEncoding.EncodeToString(r))

	return connect.NewResponse(&runtimev1.ExportResponse{
		DownloadUrlPath: out,
	}), nil
}

func (s *Server) downloadHandler(w http.ResponseWriter, req *http.Request) {
	marshalled, err := base64.URLEncoding.DecodeString(req.URL.Query().Get("request"))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse request: %s", err), http.StatusBadRequest)
		return
	}

	request := &runtimev1.ExportRequest{}
	err = proto.Unmarshal(marshalled, request)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse request: %s", err), http.StatusBadRequest)
		return
	}

	if !auth.GetClaims(req.Context()).CanInstance(request.InstanceId, auth.ReadMetrics) {
		http.Error(w, "action not allowed", http.StatusUnauthorized)
		return
	}

	if s.opts.DownloadRowLimit != nil && (request.Limit == nil || *request.Limit > *s.opts.DownloadRowLimit) {
		http.Error(w, fmt.Sprintf("limit must be less than or equal to %d", *s.opts.DownloadRowLimit), http.StatusBadRequest)
		return
	}

	var q runtime.Query
	switch v := request.Request.(type) {
	case *runtimev1.ExportRequest_MetricsViewToplistRequest:
		r := v.MetricsViewToplistRequest
		err := validateInlineMeasures(r.InlineMeasures)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		q = &queries.MetricsViewToplist{
			MetricsViewName: r.MetricsViewName,
			DimensionName:   r.DimensionName,
			MeasureNames:    r.MeasureNames,
			InlineMeasures:  r.InlineMeasures,
			TimeStart:       r.TimeStart,
			TimeEnd:         r.TimeEnd,
			Sort:            r.Sort,
			Filter:          r.Filter,
			Limit:           request.Limit,
		}
	case *runtimev1.ExportRequest_MetricsViewRowsRequest:
		r := v.MetricsViewRowsRequest
		q = &queries.MetricsViewRows{
			MetricsViewName: r.MetricsViewName,
			TimeStart:       r.TimeStart,
			TimeEnd:         r.TimeEnd,
			Filter:          r.Filter,
			Sort:            r.Sort,
			Limit:           request.Limit,
			TimeZone:        r.TimeZone,
		}
	default:
		http.Error(w, fmt.Sprintf("unsupported request type: %s", reflect.TypeOf(v).Name()), http.StatusBadRequest)
		return
	}

	err = q.Export(req.Context(), s.runtime, request.InstanceId, w, &runtime.ExportOptions{
		Format: request.Format,
		PreWriteHook: func(filename string) error {
			// Add timestamp to filename
			filename += "_" + time.Now().Format("20060102150405")

			// Write HTTP headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			switch request.Format {
			case runtimev1.ExportFormat_EXPORT_FORMAT_CSV:
				w.Header().Set("Content-Type", "text/csv")
				w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.csv\"", filename))
			case runtimev1.ExportFormat_EXPORT_FORMAT_XLSX:
				w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
				w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.xlsx\"", filename))
			default:
				return fmt.Errorf("unsupported format %q", request.Format.String())
			}
			return nil
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
