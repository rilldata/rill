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
	"google.golang.org/protobuf/proto"
)

type DownloadHandler struct {
	Runtime *runtime.Runtime
}

func (h *DownloadHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	marshalled, err := base64.StdEncoding.DecodeString(req.URL.Query().Get("request"))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse request: %s", err), http.StatusBadRequest)
		return
	}

	request := &runtimev1.DownloadLinkRequest{}
	err = proto.Unmarshal(marshalled, request)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse request: %s", err), http.StatusBadRequest)
		return
	}

	switch v := request.Request.(type) {
	case *runtimev1.DownloadLinkRequest_MetricsViewToplistRequest:
		v.MetricsViewToplistRequest.Limit = int64(request.Limit)
		err = h.executeToplistQuery(req.Context(), w, v.MetricsViewToplistRequest, request.Format)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case *runtimev1.DownloadLinkRequest_MetricsViewRowsRequest:
		v.MetricsViewRowsRequest.Limit = int32(request.Limit)
		err = h.executeRowsQuery(req.Context(), w, v.MetricsViewRowsRequest, request.Format)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	default:
		http.Error(w, fmt.Sprintf("Unsupported request type: %s", reflect.TypeOf(v).Name()), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *DownloadHandler) executeToplistQuery(ctx context.Context, writer http.ResponseWriter, req *runtimev1.MetricsViewToplistRequest, format runtimev1.DownloadFormat) error {
	err := validateInlineMeasures(req.InlineMeasures)
	if err != nil {
		return err
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

	return q.Export(ctx, h.Runtime, req.InstanceId, 0, format, writer)
}

func (h *DownloadHandler) executeRowsQuery(ctx context.Context, writer http.ResponseWriter, req *runtimev1.MetricsViewRowsRequest, format runtimev1.DownloadFormat) error {
	q := &queries.MetricsViewRows{
		MetricsViewName: req.MetricsViewName,
		TimeStart:       req.TimeStart,
		TimeEnd:         req.TimeEnd,
		Filter:          req.Filter,
		Sort:            req.Sort,
		Limit:           req.Limit,
		Offset:          req.Offset,
	}

	return q.Export(ctx, h.Runtime, req.InstanceId, 0, format, writer)
}
