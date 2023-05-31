package downloads

import (
	"context"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/queries"
)

const (
	Limit            string = "limit"
	Compression      string = "compression"
	Format           string = "format"
	Request          string = "request"
	downloadPriority int    = 0
)

type DownloadHandler struct {
	Runtime *runtime.Runtime
}

func (h *DownloadHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/csv")
	w.WriteHeader(http.StatusOK)
	limit, err := strconv.Atoi(req.URL.Query().Get(Limit))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse request: %s", err), http.StatusBadRequest)
		return
	}

	var dst []byte
	_, err = base64.StdEncoding.Decode(
		dst,
		[]byte(
			req.URL.Query().Get(Request),
		),
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse request: %s", err), http.StatusBadRequest)
		return
	}

	var request any
	err = json.Unmarshal(dst, &request)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse request: %s", err), http.StatusBadRequest)
		return
	}

	switch v := request.(type) {
	case *runtimev1.MetricsViewToplistRequest:
		v.Limit = int64(limit)
		err = h.executeToplist(req.Context(), w, v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, fmt.Sprintf("Unsupported request type: %s", err), http.StatusBadRequest)
		return
	}
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
