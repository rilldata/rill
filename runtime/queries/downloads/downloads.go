package downloads

import (
	"context"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/queries"
	"google.golang.org/protobuf/proto"
)

const (
	Limit            string = "limit"
	Format           string = "format"
	Request          string = "request"
	downloadPriority int    = 0
)

type DownloadHandler struct {
	Runtime *runtime.Runtime
}

func (h *DownloadHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/csv")

	limitString := req.URL.Query().Get(Limit)
	var limit int
	var err error
	if limitString != "" {
		limit, err = strconv.Atoi(limitString)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to parse request: %s", err), http.StatusBadRequest)
			return
		}
	}

	marshalled, err := base64.StdEncoding.DecodeString(req.URL.Query().Get(Request))
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
		v.MetricsViewToplistRequest.Limit = int64(limit)
		err = h.executeToplist(req.Context(), w, v.MetricsViewToplistRequest)
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

func (h *DownloadHandler) executeToplist(ctx context.Context, writer http.ResponseWriter, req *runtimev1.MetricsViewToplistRequest) error {
	// err := server.ValidateInlineMeasures(req.InlineMeasures)
	// if err != nil {
	// return err
	// }

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
	err := q.Resolve(ctx, h.Runtime, req.InstanceId, downloadPriority)
	if err != nil {
		return err
	}

	w := csv.NewWriter(writer)

	record := make([]string, 0, len(q.Result.Meta))
	for _, structs := range q.Result.Data {
		for _, field := range structs.Fields {
			record = append(record, field.GetStringValue())
		}
		if err := w.Write(record); err != nil {
			return err
		}
		record = record[:0]
	}

	return nil
}
