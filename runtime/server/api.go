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

func (s *Server) apiHandler(w http.ResponseWriter, req *http.Request) error {
	// Parse path parameters
	ctx := req.Context()
	instanceID := req.PathValue("instance_id")
	apiName := req.PathValue("name")

	// Add observability attributes
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", instanceID),
		attribute.String("args.name", apiName),
	)
	s.addInstanceRequestAttributes(ctx, instanceID)

	// Check if user has access to query for API data
	if !auth.GetClaims(ctx).CanInstance(instanceID, auth.ReadAPI) {
		return httputil.Errorf(http.StatusForbidden, "does not have access to custom APIs")
	}

	// Parse args from the request body and URL query
	args := make(map[string]any)
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return httputil.Errorf(http.StatusBadRequest, "failed to read request body: %w", err)
	}
	if len(body) > 0 { // For POST requests
		if err := json.Unmarshal(body, &args); err != nil {
			return httputil.Errorf(http.StatusBadRequest, "failed to unmarshal request body: %w", err)
		}
	}
	for k, v := range req.URL.Query() {
		// Set only the first value so that client does need to put array accessors in templates.
		args[k] = v[0]
	}

	// Find the API resource
	api, err := s.runtime.APIForName(ctx, instanceID, apiName)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return httputil.Errorf(http.StatusNotFound, "api with name %q not found", apiName)
		}
		return httputil.Error(http.StatusInternalServerError, err)
	}

	// Resolve the API to JSON data
	res, err := s.runtime.Resolve(ctx, &runtime.ResolveOptions{
		InstanceID:         instanceID,
		Resolver:           api.Spec.Resolver,
		ResolverProperties: api.Spec.ResolverProperties.AsMap(),
		Args:               args,
		UserAttributes:     auth.GetClaims(ctx).Attributes(),
	})
	if err != nil {
		return httputil.Error(http.StatusBadRequest, err)
	}

	// Write the response
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(res.Data)
	if err != nil {
		return httputil.Error(http.StatusInternalServerError, err)
	}

	return nil
}
