package runtime

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// Resolver represents logic, such as a SQL query, that produces output data.
// Resolvers are used to evaluate API requests, alerts, reports, etc.
//
// A resolver has two levels of configuration: static properties and dynamic arguments.
// For example, a SQL resolver has a static property for the SQL query and dynamic arguments for the query parameters.
// The static properties are usually declared in advance, such as in the YAML for a custom API, whereas the dynamic arguments are provided just prior to execution, such as in an API request.
type Resolver interface {
	// Close is called when done with the resolver.
	// Note that the Resolve method may not have been called when Close is called (in case of cache hits or validation failures).
	Close() error
	// Key that can be used for caching. It can be a large string since the value will be hashed.
	// The key should include all the properties and args that affect the output.
	// It does not need to include the instance ID or resolver name, as those are added separately to the cache key.
	Key() string
	// Refs access by the resolver. The output may be approximate, i.e. some of the refs may not exist.
	// The output should avoid duplicates and be stable between invocations.
	Refs() []*runtimev1.ResourceName
	// Validate the properties and args without running any expensive operations.
	Validate(ctx context.Context) error
	// ResolveInteractive resolves data for interactive use (e.g. API requests or alerts).
	ResolveInteractive(ctx context.Context) (*ResolverResult, error)
	// ResolveExport resolve data for export (e.g. downloads or reports).
	ResolveExport(ctx context.Context, w io.Writer, opts *ResolverExportOptions) error
}

// ResolverResult is the result of a resolver's execution.
type ResolverResult struct {
	// Data is a JSON encoded array of objects.
	Data []byte
	// Schema is the schema for the Data
	Schema *runtimev1.StructType
	// Cache indicates whether the result can be cached.
	Cache bool
}

// ResolverExportOptions are the options passed to a resolver's ResolveExport method.
type ResolverExportOptions struct {
	// Format is the format to export the result in.
	Format runtimev1.ExportFormat
	// PreWriteHook is a function that is called after the export has been prepared, but before the first bytes are output to the io.Writer.
	PreWriteHook func(filename string) error
}

// ResolverOptions are the options passed to a resolver initializer.
type ResolverOptions struct {
	Runtime        *Runtime
	InstanceID     string
	Properties     map[string]any
	Args           map[string]any
	UserAttributes map[string]any
	ForExport      bool
	// internal use only
	SQLArgs []any
}

// ResolverInitializer is a function that initializes a resolver.
type ResolverInitializer func(ctx context.Context, opts *ResolverOptions) (Resolver, error)

// ResolverInitializers tracks resolver initializers by name.
var ResolverInitializers = make(map[string]ResolverInitializer)

// RegisterResolverInitializer registers a resolver initializer by name.
func RegisterResolverInitializer(name string, initializer ResolverInitializer) {
	if ResolverInitializers[name] != nil {
		panic(fmt.Errorf("resolver already registered for name %q", name))
	}
	ResolverInitializers[name] = initializer
}

// ResolveOptions are the options passed to the runtime's Resolve method.
type ResolveOptions struct {
	InstanceID         string
	Resolver           string
	ResolverProperties map[string]any
	Args               map[string]any
	UserAttributes     map[string]any
}

// ResolveResult is subset of ResolverResult that is cached
type ResolveResult struct {
	Data   []byte
	Schema *runtimev1.StructType
}

// Resolve resolves a query using the given options.
func (r *Runtime) Resolve(ctx context.Context, opts *ResolveOptions) (ResolveResult, error) {
	// Initialize the resolver
	initializer, ok := ResolverInitializers[opts.Resolver]
	if !ok {
		return ResolveResult{}, fmt.Errorf("no resolver found for name %q", opts.Resolver)
	}
	resolver, err := initializer(ctx, &ResolverOptions{
		Runtime:        r,
		InstanceID:     opts.InstanceID,
		Properties:     opts.ResolverProperties,
		Args:           opts.Args,
		UserAttributes: opts.UserAttributes,
		ForExport:      false,
	})
	if err != nil {
		return ResolveResult{}, err
	}
	defer resolver.Close()

	// Build cache key based on the resolver's key and refs
	ctrl, err := r.Controller(ctx, opts.InstanceID)
	if err != nil {
		return ResolveResult{}, err
	}
	hash := md5.New()
	if _, err := hash.Write([]byte(resolver.Key())); err != nil {
		return ResolveResult{}, err
	}
	for _, ref := range resolver.Refs() {
		res, err := ctrl.Get(ctx, ref, false)
		if err != nil {
			// Refs are approximate, not exact (see docstring for Refs()), so they may not all exist
			continue
		}

		if _, err := hash.Write([]byte(res.Meta.Name.Kind)); err != nil {
			return ResolveResult{}, err
		}
		if _, err := hash.Write([]byte(res.Meta.Name.Name)); err != nil {
			return ResolveResult{}, err
		}
		if err := binary.Write(hash, binary.BigEndian, res.Meta.StateUpdatedOn.Seconds); err != nil {
			return ResolveResult{}, err
		}
		if err := binary.Write(hash, binary.BigEndian, res.Meta.StateUpdatedOn.Nanos); err != nil {
			return ResolveResult{}, err
		}
	}
	sum := hex.EncodeToString(hash.Sum(nil))
	key := fmt.Sprintf("inst:%s:resolver:%s:hash:%s", opts.InstanceID, opts.Resolver, sum)

	// Try to get from cache
	if val, ok := r.queryCache.cache.Get(key); ok {
		return val.(ResolveResult), nil
	}

	// Load with singleflight
	val, err := r.queryCache.singleflight.Do(ctx, key, func(ctx context.Context) (any, error) {
		// Try cache again
		if val, ok := r.queryCache.cache.Get(key); ok {
			return val, nil
		}

		res, err := resolver.ResolveInteractive(ctx)
		if err != nil {
			return ResolveResult{}, err
		}

		cRes := ResolveResult{
			Data:   res.Data,
			Schema: res.Schema,
		}
		if res.Cache {
			r.queryCache.cache.Set(key, cRes, int64(len(res.Data)))
		}
		return cRes, nil
	})
	if err != nil {
		return ResolveResult{}, err
	}
	return val.(ResolveResult), nil
}
