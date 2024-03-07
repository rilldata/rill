package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/httputil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
)

func (s *Server) apiForName(w http.ResponseWriter, req *http.Request) error {
	if !auth.GetClaims(req.Context()).CanInstance(req.PathValue("instance_id"), auth.ReadAPI) {
		return httputil.Errorf(http.StatusForbidden, "does not have access to read APIs")
	}

	ctx := req.Context()
	if req.PathValue("name") == "" {
		return httputil.Errorf(http.StatusBadRequest, "invalid path")
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return httputil.Errorf(http.StatusBadRequest, "failed to read request body: %w", err)
	}

	reqParams := make(map[string]interface{})
	if len(body) > 0 { // post
		if err := json.Unmarshal(body, &reqParams); err != nil {
			return httputil.Errorf(http.StatusBadRequest, "failed to unmarshal request body: %w", err)
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
			return httputil.Errorf(http.StatusNotFound, "api with name %q does not exist", req.PathValue("name"))
		}
		return httputil.Error(http.StatusInternalServerError, err)
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
		return httputil.Error(http.StatusInternalServerError, err)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(res)
	if err != nil {
		return httputil.Error(http.StatusInternalServerError, err)
	}
	return nil
}
