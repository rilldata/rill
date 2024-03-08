package resolvers

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/server/auth"
)

func init() {
	runtime.RegisterAPIResolverInitializer("API", newAPI)
}

type apiProps struct {
	API  string         `mapstructure:"api"`
	Args map[string]any `mapstructure:"args"`
}

// newAPI creates a resolver that proxies to another API.
func newAPI(ctx context.Context, opts *runtime.APIResolverOptions) (runtime.APIResolver, error) {
	if opts.ResolverProperties == nil {
		return nil, fmt.Errorf("resolver properties not found")
	}
	props := &apiProps{}
	if err := mapstructure.Decode(opts.ResolverProperties.AsMap(), props); err != nil {
		return nil, err
	}

	api, err := opts.Runtime.APIForName(ctx, opts.InstanceID, props.API)
	if err != nil {
		return nil, err
	}

	initializer, ok := runtime.APIResolverInitializers[api.Spec.Resolver]
	if !ok {
		return nil, fmt.Errorf("no resolver found of type %q", api.Spec.Resolver)
	}

	return initializer(ctx, &runtime.APIResolverOptions{
		Runtime:            opts.Runtime,
		InstanceID:         opts.InstanceID,
		Resolver:           api.Spec.Resolver,
		ResolverProperties: api.Spec.ResolverProperties,
		Args:               opts.Args,
		UserAttributes:     auth.GetClaims(ctx).Attributes(),
		Priority:           opts.Priority,
	})
}
