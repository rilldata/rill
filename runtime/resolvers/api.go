package resolvers

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/server/auth"
)

func init() {
	runtime.RegisterResolverInitializer("API", newAPI)
}

type apiProps struct {
	API  string         `mapstructure:"api"`
	Args map[string]any `mapstructure:"args"`
}

// newAPI creates a resolver that proxies to the resolver of an API.
func newAPI(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	// Parse props
	props := &apiProps{}
	if err := mapstructure.Decode(opts.Properties, props); err != nil {
		return nil, err
	}

	// Find the API
	api, err := opts.Runtime.APIForName(ctx, opts.InstanceID, props.API)
	if err != nil {
		return nil, err
	}

	// Merge the user-provided args with the static args defined in the API proxy props.
	// Note: The static args take precedence over the user-provided args (basically, a user can't override them).
	if opts.Args == nil {
		opts.Args = make(map[string]any)
	}
	for k, v := range props.Args {
		opts.Args[k] = v
	}

	// We need to protect against infinite recursion where API A proxies to another API that proxies back to API A.
	// For convenience, we overload the args with a special key that we use to track the APIs we've visited.
	key := "__internal__apis_visited"
	visited, ok := opts.Args[key].([]string)
	if !ok {
		visited = []string{}
	}
	for _, v := range visited {
		if v == props.API {
			return nil, fmt.Errorf("infinite recursion detected: the API %q proxies to itself", v)
		}
	}
	visited = append(visited, props.API)
	opts.Args[key] = visited

	// Initialize the resolver of the API to proxy to
	initializer, ok := runtime.ResolverInitializers[api.Spec.Resolver]
	if !ok {
		return nil, fmt.Errorf("no resolver found of type %q", api.Spec.Resolver)
	}
	return initializer(ctx, &runtime.ResolverOptions{
		Runtime:        opts.Runtime,
		InstanceID:     opts.InstanceID,
		Properties:     api.Spec.ResolverProperties.AsMap(),
		Args:           opts.Args,
		UserAttributes: auth.GetClaims(ctx).Attributes(),
	})
}
