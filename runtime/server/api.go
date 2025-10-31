package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/getkin/kin-openapi/openapi3"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/httputil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/pkg/openapiutil"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"gopkg.in/yaml.v3"
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
	claims := auth.GetClaims(ctx, instanceID)
	if !claims.Can(runtime.ReadAPI) {
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

	// Rewrite the claims before passing them to the resolver
	if api.Spec.SkipNestedSecurity {
		claims.SkipChecks = true
	}

	// Resolve the API to JSON data
	res, err := s.runtime.Resolve(ctx, &runtime.ResolveOptions{
		InstanceID:         instanceID,
		Resolver:           api.Spec.Resolver,
		ResolverProperties: api.Spec.ResolverProperties.AsMap(),
		Args:               args,
		Claims:             claims,
	})
	if err != nil {
		return httputil.Error(http.StatusBadRequest, err)
	}
	defer res.Close()

	// Write the response
	data, err := res.MarshalJSON()
	if err != nil {
		return httputil.Error(http.StatusInternalServerError, err)
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		return httputil.Error(http.StatusInternalServerError, err)
	}

	return nil
}

func (s *Server) combinedOpenAPISpec(w http.ResponseWriter, req *http.Request) error {
	// Parse path parameters
	ctx := req.Context()
	instanceID := req.PathValue("instance_id")

	// Add observability attributes
	observability.AddRequestAttributes(ctx)
	s.addInstanceRequestAttributes(ctx, instanceID)

	// Only GET request is allowed
	if req.Method != http.MethodGet {
		return httputil.Error(http.StatusMethodNotAllowed, fmt.Errorf("GET only"))
	}

	// Check if user has access to query for API data
	if !auth.GetClaims(ctx, instanceID).Can(runtime.ReadAPI) {
		return httputil.Errorf(http.StatusForbidden, "does not have access to custom APIs")
	}

	apis := make(map[string]*runtimev1.API)

	// Get all built-in APIs
	for name, api := range runtime.BuiltinAPIs {
		apis[name] = api
	}

	// Get all custom APIs
	ctrl, err := s.runtime.Controller(ctx, instanceID)
	if err != nil {
		return httputil.Error(http.StatusBadRequest, err)
	}

	list, err := ctrl.List(ctx, runtime.ResourceKindAPI, "", false)
	if err != nil {
		return err
	}

	for _, res := range list {
		apis[res.Meta.Name.Name] = res.GetApi()
	}

	// Generate the OpenAPI spec
	combinedAPISpec, err := s.generateOpenAPISpec(ctx, instanceID, apis)
	if err != nil {
		return httputil.Error(http.StatusInternalServerError, err)
	}

	// Check accepted format of the OpenAPI spec: JSON or YAML
	accept := req.Header.Get("Accept")

	// YAML
	if accept == "application/yaml" || accept == "application/x-yaml" {
		data, err := yaml.Marshal(combinedAPISpec)
		if err != nil {
			return httputil.Error(http.StatusInternalServerError, err)
		}

		w.Header().Set("Content-Type", "application/yaml")
		_, err = w.Write(data)
		if err != nil {
			return httputil.Error(http.StatusInternalServerError, err)
		}

		return nil
	}

	// JSON
	data, err := combinedAPISpec.MarshalJSON()
	if err != nil {
		return httputil.Error(http.StatusInternalServerError, err)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		return httputil.Error(http.StatusInternalServerError, err)
	}

	return nil
}

func (s *Server) generateOpenAPISpec(ctx context.Context, instanceID string, apis map[string]*runtimev1.API) (*openapi3.T, error) {
	attributes := s.runtime.GetInstanceAttributes(ctx, instanceID)
	var organization, project string
	for _, attr := range attributes {
		if attr.Key == "organization" {
			organization = attr.Value.AsString()
		} else if attr.Key == "project" {
			project = attr.Value.AsString()
		}
	}
	var title string
	if organization != "" && project != "" {
		title = fmt.Sprintf("Rill %s/%s project API", organization, project)
	} else {
		title = "Rill project API"
	}

	spec := &openapi3.T{
		OpenAPI: "3.0.3",
		Info: &openapi3.Info{
			Title:   title,
			Version: "1.0.0",
		},
		Paths: &openapi3.Paths{},
	}

	var runtimeHost string
	if s.opts.AuthAudienceURL != "" {
		runtimeURL, err := url.Parse(s.opts.AuthAudienceURL)
		if err != nil {
			return nil, err
		}
		runtimeHost = runtimeURL.Host
	} else {
		runtimeHost = fmt.Sprintf("localhost:%d", s.opts.HTTPPort)
	}

	spec.Servers = openapi3.Servers{
		&openapi3.Server{
			URL: fmt.Sprintf("http://%s/v1/instances/%s/api", runtimeHost, instanceID),
		},
	}

	for name, api := range apis {
		pathItem, componentsForPath, err := s.generatePathItemSpec(name, api)
		if err != nil {
			return nil, err
		}

		spec.Paths.Set(fmt.Sprintf("/%s", name), pathItem)

		for k, v := range componentsForPath {
			if spec.Components == nil {
				spec.Components = &openapi3.Components{}
			}
			if spec.Components.Schemas == nil {
				spec.Components.Schemas = make(map[string]*openapi3.SchemaRef)
			}
			spec.Components.Schemas[k] = v
		}
	}

	return spec, nil
}

func (s *Server) generatePathItemSpec(name string, api *runtimev1.API) (*openapi3.PathItem, map[string]*openapi3.SchemaRef, error) {
	summary := ""
	if api.Spec.OpenapiSummary != "" {
		summary = api.Spec.OpenapiSummary
	}
	if summary == "" {
		summary = fmt.Sprintf("Query %s resolver", name)
	}

	var parameters openapi3.Parameters
	if api.Spec.OpenapiParametersJson != "" {
		var err error
		parameters, err = openapiutil.ParseJSONParameters(api.Spec.OpenapiParametersJson)
		if err != nil {
			return nil, nil, err
		}
	}

	components := make(map[string]*openapi3.SchemaRef)

	var requestBody *openapi3.RequestBodyRef
	if api.Spec.OpenapiRequestSchemaJson != "" {
		s, cs, err := openapiutil.ParseJSONSchema(api.Spec.OpenapiDefsPrefix, api.Spec.OpenapiRequestSchemaJson)
		if err != nil {
			return nil, nil, err
		}

		for k, v := range cs {
			components[k] = v
		}

		requestBody = &openapi3.RequestBodyRef{
			Value: &openapi3.RequestBody{
				Content: openapi3.NewContentWithJSONSchema(s),
			},
		}
	}

	var responseSchema *openapi3.Schema
	if api.Spec.OpenapiResponseSchemaJson != "" {
		s, cs, err := openapiutil.ParseJSONSchema(api.Spec.OpenapiDefsPrefix, api.Spec.OpenapiResponseSchemaJson)
		if err != nil {
			return nil, nil, err
		}

		for k, v := range cs {
			components[k] = v
		}

		responseSchema = s
	} else {
		responseSchema = &openapi3.Schema{
			Type: &openapi3.Types{"object"},
		}
	}

	op := &openapi3.Operation{
		Summary:     summary,
		Parameters:  parameters,
		RequestBody: requestBody,
		Responses: openapi3.NewResponses(
			openapi3.WithStatus(200, &openapi3.ResponseRef{
				Value: openapi3.NewResponse().WithDescription(
					fmt.Sprintf("Successful response of %s resolver", name),
				).WithContent(
					openapi3.NewContentWithJSONSchema(&openapi3.Schema{
						Type: &openapi3.Types{"array"},
						Items: &openapi3.SchemaRef{
							Value: responseSchema,
						},
					}),
				),
			}),
			openapi3.WithStatus(400, &openapi3.ResponseRef{
				Value: openapi3.NewResponse().WithDescription(
					"Bad request",
				).WithContent(
					openapi3.NewContentWithJSONSchema(&openapi3.Schema{
						Type: &openapi3.Types{"object"},
						Properties: map[string]*openapi3.SchemaRef{
							"error": {
								Value: &openapi3.Schema{
									Type: &openapi3.Types{"string"},
								},
							},
						},
					}),
				),
			}),
		),
	}

	// If the API has a request body schema, use POST method. Otherwise, use GET method.
	var pathItem *openapi3.PathItem
	if requestBody != nil {
		pathItem = &openapi3.PathItem{Post: op}
	} else {
		pathItem = &openapi3.PathItem{Get: op}
	}

	return pathItem, components, nil
}
