package runtime

import (
	"context"
	"errors"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (r *Runtime) APIForName(ctx context.Context, instanceID, name string, reqParams map[string]any) ([]byte, error) {
	ctrl, err := r.Controller(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	resource, err := ctrl.Get(ctx, &runtimev1.ResourceName{Name: name, Kind: ResourceKindAPI}, false)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return nil, status.Error(codes.NotFound, "resource not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	api := resource.GetApi()
	if api == nil {
		return nil, status.Error(codes.InvalidArgument, "")
	}

	return []byte(api.Spec.Sql), nil
	// var resolverInitializer APIResolverInitializer
	// var ok bool
	// if api.Spec.Sql != "" {
	// 	resolverInitializer, ok = APIResolverInitializers["SQLResolver"]
	// 	if !ok {
	// 		panic("no SQLResolver")
	// 	}
	// } else {
	// 	resolverInitializer, ok = APIResolverInitializers["MetricsSQLResolver"]
	// 	if !ok {
	// 		return nil, status.Error(codes.InvalidArgument, "MetricsSQLResolver not found")
	// 	}
	// }

	// resolver, err := resolverInitializer(ctx, &APIResolverOptions{
	// 	Runtime:    r,
	// 	InstanceID: instanceID,
	// 	API:        api,
	// 	Args:       reqParams,
	// 	// UserAttributes: reqParams, // TODO: user attributes
	// })
	// if err != nil {
	// 	return nil, status.Error(codes.InvalidArgument, err.Error())
	// }

	// iter, err := resolver.ResolveInteractive(ctx, 100)
	// if err != nil {
	// 	return nil, err
	// }

	// return iter.Marshal(), nil
}
