package runtime

import (
	"context"
	"fmt"
	"io"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

type Result struct {
	Rows  []byte
	Cache bool
}

type APIResolver interface {
	// Key that can be used for caching
	Key() string
	// Deps referenced by the query
	Deps() []*runtimev1.ResourceName
	// Validate the query without running any "expensive" operations
	Validate(ctx context.Context) error
	// ResolveInteractive Resolve for interactive use (e.g. API requests or alerts)
	ResolveInteractive(ctx context.Context, priority int) (Result, error)
	// ResolveExport Resolve for export to a file (e.g. downloads or reports)
	ResolveExport(ctx context.Context, w io.Writer, opts *ExportOptions) error
	// Close any resources that needs to be released
	Close() error
}

// APIResolverInitializers Resolvers should register themselves in this map from their package's init() function
var APIResolverInitializers = make(map[string]APIResolverInitializer)

func RegisterAPIResolverInitializer(name string, resolverInitializer APIResolverInitializer) {
	APIResolverInitializers[name] = resolverInitializer
}

type APIResolverInitializer func(ctx context.Context, opts *APIResolverOptions) (APIResolver, error)

type APIResolverOptions struct {
	Runtime        *Runtime
	InstanceID     string
	API            *runtimev1.API
	Args           map[string]any
	UserAttributes map[string]any
	Priority       int
}

func Resolve(ctx context.Context, opts *APIResolverOptions) ([]byte, error) {
	resolverInitializer, ok := APIResolverInitializers[opts.API.Spec.Resolver]
	if !ok {
		return nil, fmt.Errorf("no resolver found of type %s", opts.API.Spec.Resolver)
	}
	resolver, err := resolverInitializer(ctx, opts)
	if err != nil {
		return nil, err
	}
	defer resolver.Close()
	if err := resolver.Validate(ctx); err != nil {
		return nil, err
	}
	depKey := ""
	for _, dep := range resolver.Deps() {
		depKey += dep.Kind + ":" + dep.Name
	}
	key := queryCacheKey{
		instanceID:    opts.InstanceID,
		queryKey:      resolver.Key(),
		dependencyKey: depKey,
	}.String()

	// Try to get from cache
	if val, ok := opts.Runtime.queryCache.cache.Get(key); ok {
		return val.([]byte), nil
	}

	// Load with singleflight
	val, err := opts.Runtime.queryCache.singleflight.Do(ctx, key, func(ctx context.Context) (any, error) {
		// Try cache again
		if val, ok := opts.Runtime.queryCache.cache.Get(key); ok {
			return val, nil
		}

		res, err := resolver.ResolveInteractive(ctx, opts.Priority)
		if err != nil {
			return nil, err
		}

		if res.Cache {
			opts.Runtime.queryCache.cache.Set(key, res.Rows, int64(len(res.Rows)))
		}
		return res.Rows, nil
	})
	if err != nil {
		return nil, err
	}
	return val.([]byte), nil
}

// TODO: Add a function for exporting the result to a file
