package runtime

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

var BuiltinAPIs = map[string]*runtimev1.API{}

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
