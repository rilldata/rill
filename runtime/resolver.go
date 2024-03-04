package runtime

import (
	"context"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"io"
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
var APIResolverInitializers map[string]APIResolverInitializer = make(map[string]APIResolverInitializer)

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
}
