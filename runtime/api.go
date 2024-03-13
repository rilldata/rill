package runtime

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// BuiltinAPIs is a map of built-in APIs (i.e. APIs that are not created dynamically from the project's YAML files.)
var BuiltinAPIs = map[string]*runtimev1.API{}

// RegisterBuiltinAPI adds a built-in API with the given name that invokes the given resolver and resolver properties.
func RegisterBuiltinAPI(name, resolver string, resolverProps map[string]any) {
	props, err := structpb.NewStruct(resolverProps)
	if err != nil {
		panic(err)
	}

	api := &runtimev1.API{
		Spec: &runtimev1.APISpec{
			Resolver:           resolver,
			ResolverProperties: props,
		},
		State: &runtimev1.APIState{},
	}

	BuiltinAPIs[name] = api
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
