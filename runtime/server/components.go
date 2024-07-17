package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/httputil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
)

func (s *Server) parseTemplate(w http.ResponseWriter, req *http.Request) error {
	// Parse path parameters
	ctx := req.Context()
	instanceID := req.PathValue("instance_id")
	component := req.PathValue("name")

	// Add observability attributes
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", instanceID),
		attribute.String("args.name", component),
	)
	s.addInstanceRequestAttributes(ctx, instanceID)

	// Check if user has access to query for component data (we use the ReadAPI permission for this for now)
	if !auth.GetClaims(req.Context()).CanInstance(instanceID, auth.ReadAPI) {
		return httputil.Errorf(http.StatusForbidden, "does not have access to component data")
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
		args[k] = v[0]
	}

	parsedComponent := s.getParsedComponent(ctx, instanceID, component, args)
	if parsedComponent == nil {
		return httputil.Error(http.StatusInternalServerError, err)
	}

	response := map[string]string{"content": *parsedComponent}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(response)
}

func (s *Server) componentDataHandler(w http.ResponseWriter, req *http.Request) error {
	// Parse path parameters
	ctx := req.Context()
	instanceID := req.PathValue("instance_id")
	component := req.PathValue("name")
	// Add observability attributes
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", instanceID),
		attribute.String("args.name", component),
	)
	s.addInstanceRequestAttributes(ctx, instanceID)

	// Check if user has access to query for component data (we use the ReadAPI permission for this for now)
	if !auth.GetClaims(req.Context()).CanInstance(instanceID, auth.ReadAPI) {
		return httputil.Errorf(http.StatusForbidden, "does not have access to component data")
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

	// Find the component resource
	componentSpec, err := s.getComponent(ctx, instanceID, component)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return httputil.Errorf(http.StatusNotFound, "component with name %q not found", component)
		}
		return httputil.Error(http.StatusInternalServerError, err)
	}

	// Call the component's data resolver
	res, err := s.runtime.Resolve(ctx, &runtime.ResolveOptions{
		InstanceID:         instanceID,
		Resolver:           componentSpec.Resolver,
		ResolverProperties: componentSpec.ResolverProperties.AsMap(),
		Args:               args,
		Claims:             auth.GetClaims(ctx).SecurityClaims(),
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

func (s *Server) getComponent(ctx context.Context, instanceID, name string) (*runtimev1.ComponentSpec, error) {
	ctrl, err := s.runtime.Controller(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	res, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindComponent, Name: name}, false)
	if err != nil {
		return nil, err
	}

	ch := res.GetComponent()
	spec := ch.Spec
	if spec == nil {
		return nil, fmt.Errorf("component %q is invalid", name)
	}

	return spec, nil
}

func (s *Server) getParsedComponent(ctx context.Context, instanceID, name string, args map[string]any) *string {
	ctrl, err := s.runtime.Controller(ctx, instanceID)
	if err != nil {
		return nil
	}
	fmt.Println(name, args)

	res, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindComponent, Name: name}, false)
	if err != nil {
		return nil
	}

	blob, _, err := s.runtime.GetFile(ctx, instanceID, res.Meta.GetFilePaths()[0])
	if err != nil {
		return nil
	}

	fmt.Println(blob)

	resolved, err := rillv1.ResolveTemplate(blob, rillv1.TemplateData{
		ExtraProps: map[string]any{
			"args": args,
		},
	})
	if err != nil {
		fmt.Println(err)
		return nil
	}

	fmt.Println(resolved)

	return &resolved
}
