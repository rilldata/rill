package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-openapi/spec"
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

	// TODO: Should it resolve security and check access here?

	// Resolve the API to JSON data
	res, err := s.runtime.Resolve(ctx, &runtime.ResolveOptions{
		InstanceID:         instanceID,
		Resolver:           api.Spec.Resolver,
		ResolverProperties: api.Spec.ResolverProperties.AsMap(),
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
	if !auth.GetClaims(ctx).CanInstance(instanceID, auth.ReadAPI) {
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

func (s *Server) generateOpenAPISpec(ctx context.Context, instanceID string, apis map[string]*runtimev1.API) (*spec.Swagger, error) {
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
		title = fmt.Sprintf("Rill project API of %s/%s", organization, project)
	} else {
		title = "Rill project API"
	}
	baseSpec := &spec.Swagger{
		SwaggerProps: spec.SwaggerProps{
			Swagger:  "2.0",
			Produces: []string{"application/json"},
			Info: &spec.Info{
				InfoProps: spec.InfoProps{
					Title:   title,
					Version: "1.0.0",
				},
			},
			BasePath: fmt.Sprintf("/v1/instances/%s/api", instanceID),
			Paths: &spec.Paths{
				Paths: make(map[string]spec.PathItem),
			},
		},
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

	baseSpec.SwaggerProps.Host = runtimeHost

	for name, api := range apis {
		pathItem, err := s.generatePathItemSpec(name, api)
		if err != nil {
			return nil, err
		}

		baseSpec.Paths.Paths[fmt.Sprintf("/%s", name)] = *pathItem
	}

	return baseSpec, nil
}

func (s *Server) generatePathItemSpec(name string, api *runtimev1.API) (*spec.PathItem, error) {
	var err error
	reqSummary := ""
	if api.Spec.OpenapiSummary != "" {
		reqSummary = api.Spec.OpenapiSummary
	}
	if reqSummary == "" {
		reqSummary = fmt.Sprintf("Query %s API", name)
	}

	var params []spec.Parameter
	if api.Spec.OpenapiParameters != nil {
		maps := make([]map[string]any, len(api.Spec.OpenapiParameters))
		for i, param := range api.Spec.OpenapiParameters {
			maps[i] = param.AsMap()
		}
		params, err = openapiutil.MapToParameters(maps)
		if err != nil {
			return nil, err
		}
	}

	var schema *spec.Schema
	if api.Spec.OpenapiResponseSchema != nil {
		schema, err = openapiutil.MapToSchema(api.Spec.OpenapiResponseSchema.AsMap())
		if err != nil {
			return nil, err
		}
	} else {
		schema = &spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{"object"},
			},
		}
	}

	pathItem := spec.PathItem{
		PathItemProps: spec.PathItemProps{
			Get: &spec.Operation{
				OperationProps: spec.OperationProps{
					Summary:    reqSummary,
					Parameters: params,
					Responses: &spec.Responses{
						ResponsesProps: spec.ResponsesProps{
							StatusCodeResponses: map[int]spec.Response{
								200: {
									ResponseProps: spec.ResponseProps{
										Description: fmt.Sprintf("Successful response of the %s API", name),
										Schema: &spec.Schema{
											SchemaProps: spec.SchemaProps{
												Type: []string{"array"},
												Items: &spec.SchemaOrArray{
													Schema: schema,
												},
											},
										},
									},
								},
								400: {
									ResponseProps: spec.ResponseProps{
										Description: "Bad request",
										Schema: &spec.Schema{
											SchemaProps: spec.SchemaProps{
												Type: []string{"object"},
												Properties: map[string]spec.Schema{
													"error": {
														SchemaProps: spec.SchemaProps{
															Type: []string{"string"},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return &pathItem, nil
}
