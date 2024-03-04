package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rilldata/rill/runtime"
	"io"
	"net/http"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
)

func (s *Server) APIForName(w http.ResponseWriter, req *http.Request) {
	if !auth.GetClaims(req.Context()).CanInstance(req.PathValue("instance_id"), auth.ReadOLAP) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	ctx := req.Context()
	if req.PathValue("name") == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	reqParams := make(map[string]interface{})
	if len(body) > 0 { // post
		if err := json.Unmarshal(body, &reqParams); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	queryParams := req.URL.Query()
	for k, v := range queryParams {
		// set only the first value so that client does need to put array accessors in templates.
		reqParams[k] = v[0]
	}

	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.PathValue("instance_id")),
		attribute.String("args.name", req.PathValue("name")),
	)

	s.addInstanceRequestAttributes(ctx, req.PathValue("instance_id"))

	api, err := s.runtime.APIForName(ctx, req.PathValue("instance_id"), req.PathValue("name"))
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := runtime.Resolve(ctx, &runtime.APIResolverOptions{
		Runtime:        s.runtime,
		InstanceID:     req.PathValue("instance_id"),
		API:            api,
		Args:           reqParams,
		UserAttributes: auth.GetClaims(ctx).Attributes(),
		Priority:       0,
	})
	if err != nil {
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
