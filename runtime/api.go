package runtime

import (
	"context"
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

var ErrAPINotFound = fmt.Errorf("API not found")

var BuiltinAPIs = map[string]*runtimev1.API{}

func (r *Runtime) APIForName(ctx context.Context, instanceID, name string, reqParams map[string]any) (*runtimev1.API, error) {
	if api, ok := BuiltinAPIs[name]; ok {
		return api, nil
	}
	ctrl, err := r.Controller(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	resource, err := ctrl.Get(ctx, &runtimev1.ResourceName{Name: name, Kind: ResourceKindAPI}, false)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return nil, ErrAPINotFound
		}
		return nil, err
	}

	api := resource.GetApi()
	if api == nil {
		return nil, ErrAPINotFound
	}
	return api, nil
}
