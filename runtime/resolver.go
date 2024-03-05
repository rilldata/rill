package runtime

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

type Result struct {
	Data  []byte // marshalled array of Jsons
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
		return nil, fmt.Errorf("no resolver found of type %q", opts.API.Spec.Resolver)
	}
	resolver, err := resolverInitializer(ctx, opts)
	if err != nil {
		return nil, err
	}
	defer resolver.Close()

	// Get dependency cache keys
	ctrl, err := opts.Runtime.Controller(ctx, opts.InstanceID)
	if err != nil {
		return nil, err
	}
	depKeys := make([]string, 0, len(resolver.Deps()))
	for _, dep := range resolver.Deps() {
		res, err := ctrl.Get(ctx, dep, false)
		if err != nil {
			// Deps are approximate, not exact (see docstring for Deps()), so they may not all exist
			continue
		}
		// Using StateUpdatedOn instead of StateVersion because the state version is reset when the resource is deleted and recreated.
		key := fmt.Sprintf("%s:%s:%d:%d", res.Meta.Name.Kind, res.Meta.Name.Name, res.Meta.StateUpdatedOn.Seconds, res.Meta.StateUpdatedOn.Nanos/int32(time.Millisecond))
		depKeys = append(depKeys, key)
	}

	depKey := strings.Join(depKeys, ";")

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
			opts.Runtime.queryCache.cache.Set(key, res.Data, int64(len(res.Data)))
		}
		return res.Data, nil
	})
	if err != nil {
		return nil, err
	}
	return val.([]byte), nil
}

// TODO: Add a function for exporting the result to a file
