package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

func (s *Server) APIForName(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
	// todo :: check for any permissions

	ctx := context.Background()
	if pathParams["name"] == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		s.logger.Info("failed to read APIForName request", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	reqParams := make(map[string]interface{})
	if len(body) > 0 { // post request
		if err := json.Unmarshal(body, &reqParams); err != nil {
			s.logger.Info("failed to parse APIForName request body", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	queryParams := req.URL.Query()
	for k, v := range queryParams {
		reqParams[k] = v
	}

	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", pathParams["instance_id"]),
		attribute.String("args.name", pathParams["name"]),
	)

	s.addInstanceRequestAttributes(ctx, pathParams["instance_id"])

	res, err := s.runtime.APIForName(ctx, pathParams["instance_id"], pathParams["name"], reqParams)
	if err != nil {
		// todo :: set correct error
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(res)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to write response data: %s", err), http.StatusInternalServerError)
		return
	}
}
