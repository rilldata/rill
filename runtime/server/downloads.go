package server

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"reflect"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/server/auth"
	"google.golang.org/protobuf/proto"
)

func (s *Server) Export(ctx context.Context, req *runtimev1.ExportRequest) (*runtimev1.ExportResponse, error) {
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	r, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	out := fmt.Sprintf("/v1/download?%s=%s", "request", base64.StdEncoding.EncodeToString(r))

	return &runtimev1.ExportResponse{
		DownloadUrlPath: out,
	}, nil
}

func (s *Server) downloadHandler(w http.ResponseWriter, req *http.Request) {
	marshalled, err := base64.StdEncoding.DecodeString(req.URL.Query().Get("request"))
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

	var q runtime.Query
	switch v := request.Request.(type) {
	case *runtimev1.ExportRequest_MetricsViewToplistRequest:
		v.MetricsViewToplistRequest.Limit = int64(request.Limit)
		q, err = createToplistQuery(req.Context(), w, v.MetricsViewToplistRequest, request.Format)
	case *runtimev1.ExportRequest_MetricsViewRowsRequest:
		v.MetricsViewRowsRequest.Limit = request.Limit
		q, err = createRowsQuery(req.Context(), w, v.MetricsViewRowsRequest, request.Format)
	default:
		http.Error(w, fmt.Sprintf("Unsupported request type: %s", reflect.TypeOf(v).Name()), http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !auth.GetClaims(req.Context()).CanInstance(request.InstanceId, auth.ReadMetrics) {
		http.Error(w, "action not allowed", http.StatusUnauthorized)
		return
	}

	err = q.Export(req.Context(), s.runtime, request.InstanceId, 0, request.Format, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if request.Format == runtimev1.ExportFormat_EXPORT_FORMAT_CSV {
		w.Header().Set("Content-Type", "text/csv")
	}

	w.WriteHeader(http.StatusOK)
}

func createToplistQuery(ctx context.Context, writer http.ResponseWriter, req *runtimev1.MetricsViewToplistRequest, format runtimev1.ExportFormat) (runtime.Query, error) {
	err := validateInlineMeasures(req.InlineMeasures)
	if err != nil {
		return nil, err
	}

	q := &queries.MetricsViewToplist{
		MetricsViewName: req.MetricsViewName,
		DimensionName:   req.DimensionName,
		MeasureNames:    req.MeasureNames,
		InlineMeasures:  req.InlineMeasures,
		TimeStart:       req.TimeStart,
		TimeEnd:         req.TimeEnd,
		Limit:           req.Limit,
		Offset:          req.Offset,
		Sort:            req.Sort,
		Filter:          req.Filter,
	}

	return q, nil
}

func createRowsQuery(ctx context.Context, writer http.ResponseWriter, req *runtimev1.MetricsViewRowsRequest, format runtimev1.ExportFormat) (runtime.Query, error) {
	q := &queries.MetricsViewRows{
		MetricsViewName: req.MetricsViewName,
		TimeStart:       req.TimeStart,
		TimeEnd:         req.TimeEnd,
		Filter:          req.Filter,
		Sort:            req.Sort,
		Limit:           req.Limit,
		Offset:          req.Offset,
	}

	return q, nil
}
