package runtime

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/mitchellh/hashstructure/v2"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/jsonval"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var ErrMetricsViewCachingDisabled = errors.New("metrics_cache_key: caching is disabled")

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
	// CacheKey returns a key that can be used for caching. It can be a large string since the value will be hashed.
	// The key should include all the properties and args that affect the output.
	// It does not need to include the instance ID or resolver name, as those are added separately to the cache key.
	//
	// If the resolver result is not cacheable, ok is set to false.
	CacheKey(ctx context.Context) (key []byte, ok bool, err error)
	// Refs access by the resolver. The output may be approximate, i.e. some of the refs may not exist.
	// The output should avoid duplicates and be stable between invocations.
	Refs() []*runtimev1.ResourceName
	// Validate the properties and args without running any expensive operations.
	Validate(ctx context.Context) error
	// ResolveInteractive resolves data for interactive use (e.g. API requests or alerts).
	ResolveInteractive(ctx context.Context) (ResolverResult, error)
	// ResolveExport resolve data for export (e.g. downloads or reports).
	ResolveExport(ctx context.Context, w io.Writer, opts *ResolverExportOptions) error
	// InferRequiredSecurityRules attempts to infer the security rules that are required to be able to execute the resolver for the currently configured properties.
	InferRequiredSecurityRules() ([]*runtimev1.SecurityRule, error)
}

// ResolverResult is the result of a resolver's execution.
type ResolverResult interface {
	// Close should be called to release resources
	Close() error
	// Meta can contain arbitrary metadata about the result.
	// For example, the metrics resolver will return information about the result fields, like display names and formatting rules.
	Meta() map[string]any
	// Schema is the schema for the Data
	Schema() *runtimev1.StructType
	// Next returns the next row of data. It returns io.EOF when there are no more rows.
	Next() (map[string]any, error)
	// MarshalJSON is a convenience method to serialize the result to JSON.
	MarshalJSON() ([]byte, error)
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
	Runtime    *Runtime
	InstanceID string
	Properties map[string]any
	Args       map[string]any
	Claims     *SecurityClaims
	ForExport  bool
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
	Claims             *SecurityClaims
}

// Resolve resolves a query using the given options.
// The caller must call Close on the result when done consuming it.
func (r *Runtime) Resolve(ctx context.Context, opts *ResolveOptions) (res ResolverResult, resErr error) {
	// Since claims don't really make sense for some resolver use cases, it's easy to forget to set them.
	// Adding an early panic to catch this.
	if opts.Claims == nil {
		panic("received nil claims")
	}

	ctx, span := tracer.Start(ctx, "runtime.Resolve", trace.WithAttributes(attribute.String("resolver", opts.Resolver)))
	var cacheHit bool
	defer func() {
		span.SetAttributes(attribute.Bool("cache_hit", cacheHit))
		if resErr != nil {
			span.SetAttributes(attribute.String("err", resErr.Error()))
		}
		span.End()
	}()

	// Initialize the resolver
	initializer, ok := ResolverInitializers[opts.Resolver]
	if !ok {
		return nil, fmt.Errorf("no resolver found for name %q", opts.Resolver)
	}
	resolver, err := initializer(ctx, &ResolverOptions{
		Runtime:    r,
		InstanceID: opts.InstanceID,
		Properties: opts.ResolverProperties,
		Args:       opts.Args,
		Claims:     opts.Claims,
		ForExport:  false,
	})
	if err != nil {
		return nil, err
	}
	defer resolver.Close()

	// Get the cache key
	cacheKey, ok, err := resolver.CacheKey(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		cacheHit = false
		// If not cacheable, just resolve and return
		return resolver.ResolveInteractive(ctx)
	}

	// Build cache key based on the resolver's key and refs
	ctrl, err := r.Controller(ctx, opts.InstanceID)
	if err != nil {
		return nil, err
	}
	hash := md5.New()
	if _, err := hash.Write(cacheKey); err != nil {
		return nil, err
	}
	if opts.Claims.UserAttributes != nil {
		h, err := hashstructure.Hash(opts.Claims.UserAttributes, hashstructure.FormatV2, nil)
		if err != nil {
			return nil, err
		}
		if _, err = hash.Write([]byte(strconv.FormatUint(h, 16))); err != nil {
			return nil, err
		}
	}
	for _, ref := range resolver.Refs() {
		res, err := ctrl.Get(ctx, ref, false)
		if err != nil {
			// Refs are approximate, not exact (see docstring for Refs()), so they may not all exist
			continue
		}

		if _, err := hash.Write([]byte(res.Meta.Name.Kind)); err != nil {
			return nil, err
		}
		if _, err := hash.Write([]byte(res.Meta.Name.Name)); err != nil {
			return nil, err
		}
		if err := binary.Write(hash, binary.BigEndian, res.Meta.StateUpdatedOn.Seconds); err != nil {
			return nil, err
		}
		if err := binary.Write(hash, binary.BigEndian, res.Meta.StateUpdatedOn.Nanos); err != nil {
			return nil, err
		}
	}
	sum := hex.EncodeToString(hash.Sum(nil))
	key := fmt.Sprintf("inst:%s:resolver:%s:hash:%s", opts.InstanceID, opts.Resolver, sum)

	// Try to get from cache
	if val, ok := r.queryCache.cache.Get(key); ok {
		cacheHit = true
		return val.(*cachedResolverResult).copy(), nil
	}
	// Load with singleflight
	val, err := r.queryCache.singleflight.Do(ctx, key, func(ctx context.Context) (any, error) {
		// Try cache again
		if val, ok := r.queryCache.cache.Get(key); ok {
			cacheHit = true
			return val.(*cachedResolverResult), nil
		}

		// Resolve
		// NOTE: We can under no circumstances return the res directly since we're in a singleflight and results can have iterator state.
		cacheHit = false
		res, err := resolver.ResolveInteractive(ctx)
		if err != nil {
			return nil, err
		}
		defer res.Close()

		// Cache the result
		cRes, err := newCachedResolverResult(res)
		if err != nil {
			return nil, err
		}
		r.queryCache.cache.Set(key, cRes, int64(len(cRes.data)))
		return cRes, nil
	})
	if err != nil {
		return nil, err
	}

	// Need to call copy() even on the first result to prevent the parsed rows staying in memory for a long time.
	return val.(*cachedResolverResult).copy(), nil
}

// NewDriverResolverResult creates a ResolverResult from a drivers.Result.
func NewDriverResolverResult(result *drivers.Result, meta map[string]any) ResolverResult {
	return &driverResolverResult{
		rows: result,
		meta: meta,
	}
}

type driverResolverResult struct {
	rows     *drivers.Result
	meta     map[string]any
	closeErr error
}

var _ ResolverResult = &driverResolverResult{}

// Close implements ResolverResult.
func (r *driverResolverResult) Close() error {
	if r.closeErr != nil {
		return r.closeErr
	}
	// it is okay to call Close multiple times
	// so we don't need to track if it was already called
	return r.rows.Close()
}

// Meta implements ResolverResult.
func (r *driverResolverResult) Meta() map[string]any {
	return r.meta
}

// Schema implements ResolverResult.
func (r *driverResolverResult) Schema() *runtimev1.StructType {
	return r.rows.Schema
}

// Next implements ResolverResult.
func (r *driverResolverResult) Next() (map[string]any, error) {
	if !r.rows.Next() {
		r.closeErr = r.rows.Close()
		return nil, io.EOF
	}
	row := make(map[string]any)
	err := r.rows.MapScan(row)
	if err != nil {
		return nil, err
	}
	return row, nil
}

// MarshalJSON implements ResolverResult.
func (r *driverResolverResult) MarshalJSON() ([]byte, error) {
	defer func() {
		r.closeErr = r.rows.Close()
	}()
	var out []map[string]any
	for r.rows.Next() {
		row := make(map[string]any)
		err := r.rows.MapScan(row)
		if err != nil {
			return nil, err
		}

		ret, err := jsonval.ToValue(row, &runtimev1.Type{StructType: r.rows.Schema})
		if err != nil {
			return nil, err
		}
		row, ok := ret.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("Resolver.MarshalJSON unexpected type %T", ret)
		}
		out = append(out, row)
	}
	if r.rows.Err() != nil {
		return nil, r.rows.Err()
	}
	if out == nil { // fixes 'null' output when there are no rows
		out = []map[string]any{}
	}
	return json.Marshal(out)
}

// NewMapsResolverResult creates a ResolverResult from a slice of maps.
func NewMapsResolverResult(result []map[string]any, schema *runtimev1.StructType) ResolverResult {
	return &mapsResolverResult{
		rows:   result,
		schema: schema,
	}
}

type mapsResolverResult struct {
	rows   []map[string]any
	meta   map[string]any
	schema *runtimev1.StructType
	idx    int
}

var _ ResolverResult = &mapsResolverResult{}

// Close implements ResolverResult.
func (r *mapsResolverResult) Close() error {
	return nil
}

// Meta implements ResolverResult.
func (r *mapsResolverResult) Meta() map[string]any {
	return r.meta
}

// Schema implements ResolverResult.
func (r *mapsResolverResult) Schema() *runtimev1.StructType {
	return r.schema
}

// Next implements ResolverResult.
func (r *mapsResolverResult) Next() (map[string]any, error) {
	if r.idx >= len(r.rows) {
		return nil, io.EOF
	}
	row := r.rows[r.idx]
	r.idx++
	return row, nil
}

// MarshalJSON implements ResolverResult.
func (r *mapsResolverResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.rows)
}

// newCachedResolverResult wraps a ResolverResult such that it is cacheable.
// Unlike other ResolverResult implementations, it can be kept in memory for a long time and read multiple times.
// When used multiple times, call .copy() to get a copy with a reset iteration cursor.
func newCachedResolverResult(res ResolverResult) (*cachedResolverResult, error) {
	data, err := res.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return &cachedResolverResult{
		data:   data,
		meta:   res.Meta(),
		schema: res.Schema(),
	}, nil
}

type cachedResolverResult struct {
	data   []byte
	meta   map[string]any
	schema *runtimev1.StructType

	// Iterator fields. Should only be populated on short-lived copies obtained via copy().
	rows []map[string]any
	idx  int
}

var _ ResolverResult = &cachedResolverResult{}

// Close implements ResolverResult.
func (r *cachedResolverResult) Close() error {
	return nil
}

// Meta implements ResolverResult.
func (r *cachedResolverResult) Meta() map[string]any {
	return r.meta
}

// Schema implements ResolverResult.
func (r *cachedResolverResult) Schema() *runtimev1.StructType {
	return r.schema
}

// Next implements ResolverResult.
func (r *cachedResolverResult) Next() (map[string]any, error) {
	if r.rows == nil {
		var rows []map[string]any
		err := json.Unmarshal(r.data, &rows)
		if err != nil {
			return nil, err
		}
		r.rows = rows
	}
	if r.idx >= len(r.rows) {
		return nil, io.EOF
	}
	row := r.rows[r.idx]
	r.idx++
	return row, nil
}

// MarshalJSON implements ResolverResult.
func (r *cachedResolverResult) MarshalJSON() ([]byte, error) {
	return r.data, nil
}

func (r *cachedResolverResult) copy() *cachedResolverResult {
	return &cachedResolverResult{
		data:   r.data,
		schema: r.schema,
		meta:   r.meta,
		rows:   nil,
		idx:    0,
	}
}
