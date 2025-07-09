package runtime

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/openapiutil"
	"google.golang.org/protobuf/types/known/structpb"
)

// BuiltinAPIs is a map of built-in APIs (i.e. predefined APIs that are not created dynamically from a project's YAML files.)
var BuiltinAPIs = map[string]*runtimev1.API{}

// BuiltinAPIOptions hold options for registering built-in APIs.
type BuiltinAPIOptions struct {
	Name                  string
	Resolver              string
	ResolverProperties    map[string]any
	OpenAPISummary        string
	OpenAPIRequestSchema  string
	OpenAPIResponseSchema string
}

// RegisterBuiltinAPI adds a built-in API with the given options.
func RegisterBuiltinAPI(opts *BuiltinAPIOptions) {
	props, err := structpb.NewStruct(opts.ResolverProperties)
	if err != nil {
		panic(err)
	}

	if opts.OpenAPIRequestSchema != "" {
		_, _, err = openapiutil.ParseJSONSchema(opts.Name, opts.OpenAPIRequestSchema)
		if err != nil {
			panic(err)
		}
	}

	if opts.OpenAPIResponseSchema != "" {
		_, _, err = openapiutil.ParseJSONSchema(opts.Name, opts.OpenAPIResponseSchema)
		if err != nil {
			panic(err)
		}
	}

	api := &runtimev1.API{
		Spec: &runtimev1.APISpec{
			Resolver:                  opts.Resolver,
			ResolverProperties:        props,
			OpenapiSummary:            opts.OpenAPISummary,
			OpenapiRequestSchemaJson:  opts.OpenAPIRequestSchema,
			OpenapiResponseSchemaJson: opts.OpenAPIResponseSchema,
			OpenapiDefsPrefix:         "", // Not adding definitions prefix for built-in APIs
		},
		State: &runtimev1.APIState{},
	}

	BuiltinAPIs[opts.Name] = api
}

// APIForName returns the API with the given name for the given instance.
// It gives precedence to built-in APIs over project-specific dynamically created APIs.
func (r *Runtime) APIForName(ctx context.Context, instanceID, name string) (*runtimev1.API, error) {
	if api, ok := BuiltinAPIs[name]; ok {
		return api, nil
	}

	ctrl, err := r.Controller(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	resource, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: ResourceKindAPI, Name: name}, false)
	if err != nil {
		return nil, err
	}

	return resource.GetApi(), nil
}
