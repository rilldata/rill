package resolvers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/mitchellh/mapstructure"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
)

func init() {
	runtime.RegisterResolverInitializer("union", newUnion)
}

type unionResolver struct {
	resolvers []runtime.Resolver
}

type unionProps struct {
	Resolvers []unionResolverEntry `mapstructure:"resolvers"`
}

type unionResolverEntry struct {
	Name       string         `mapstructure:"name"`
	Properties map[string]any `mapstructure:"properties"`
}

// newUnion creates a resolver that invokes multiple resolvers and returns the union of their results.
func newUnion(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	props := &unionProps{}
	if err := mapstructure.Decode(opts.Properties, props); err != nil {
		return nil, err
	}

	if len(props.Resolvers) == 0 {
		return nil, fmt.Errorf("union resolver requires at least one resolver")
	}

	resolvers := make([]runtime.Resolver, 0, len(props.Resolvers))
	for _, entry := range props.Resolvers {
		initializer, ok := runtime.ResolverInitializers[entry.Name]
		if !ok {
			closeResolvers(resolvers)
			return nil, fmt.Errorf("no resolver found of type %q", entry.Name)
		}
		r, err := initializer(ctx, &runtime.ResolverOptions{
			Runtime:    opts.Runtime,
			InstanceID: opts.InstanceID,
			Properties: entry.Properties,
			Args:       opts.Args,
			Claims:     opts.Claims,
			ForExport:  opts.ForExport,
		})
		if err != nil {
			closeResolvers(resolvers)
			return nil, fmt.Errorf("failed to initialize %q resolver in union: %w", entry.Name, err)
		}
		resolvers = append(resolvers, r)
	}

	return &unionResolver{resolvers: resolvers}, nil
}

func (r *unionResolver) Close() error {
	return closeResolvers(r.resolvers)
}

func (r *unionResolver) CacheKey(ctx context.Context) ([]byte, bool, error) {
	var parts []string
	for _, resolver := range r.resolvers {
		key, ok, err := resolver.CacheKey(ctx)
		if err != nil {
			return nil, false, err
		}
		if !ok {
			return nil, false, nil
		}
		parts = append(parts, string(key))
	}
	return []byte(strings.Join(parts, "\n")), true, nil
}

func (r *unionResolver) Refs() []*runtimev1.ResourceName {
	var refs []*runtimev1.ResourceName
	for _, resolver := range r.resolvers {
		refs = append(refs, resolver.Refs()...)
	}
	return normalizeRefs(refs)
}

func (r *unionResolver) Validate(ctx context.Context) error {
	for _, resolver := range r.resolvers {
		if err := resolver.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (r *unionResolver) ResolveInteractive(ctx context.Context) (runtime.ResolverResult, error) {
	var allRows []map[string]any
	var schema *runtimev1.StructType
	for _, resolver := range r.resolvers {
		res, err := resolver.ResolveInteractive(ctx)
		if err != nil {
			return nil, err
		}

		for {
			row, err := res.Next()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				res.Close()
				return nil, err
			}
			allRows = append(allRows, row)
		}
		res.Close()

		schema = mergeSchemas(schema, res.Schema())
	}

	if schema == nil { // Fallback
		schema = &runtimev1.StructType{}
	}

	return runtime.NewMapsResolverResult(allRows, schema), nil
}

func (r *unionResolver) ResolveExport(ctx context.Context, w io.Writer, opts *runtime.ResolverExportOptions) error {
	return errors.New("union resolver does not support export")
}

func (r *unionResolver) InferRequiredSecurityRules() ([]*runtimev1.SecurityRule, error) {
	var rules []*runtimev1.SecurityRule
	for _, resolver := range r.resolvers {
		rs, err := resolver.InferRequiredSecurityRules()
		if err != nil {
			return nil, err
		}
		rules = append(rules, rs...)
	}
	return rules, nil
}

// mergeSchemas does a best-effort merge of two StructTypes.
// If two fields have the same name, the type of the first is used.
func mergeSchemas(a, b *runtimev1.StructType) *runtimev1.StructType {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}

	existing := make(map[string]bool, len(a.Fields))
	for _, f := range a.Fields {
		existing[f.Name] = true
	}

	merged := &runtimev1.StructType{
		Fields: slices.Clone(a.Fields),
	}

	for _, f := range b.Fields {
		if !existing[f.Name] {
			merged.Fields = append(merged.Fields, f)
			existing[f.Name] = true
		}
	}

	return merged
}

// closeResolvers attempts to close all resolvers and aggregates any errors that occur.
func closeResolvers(resolvers []runtime.Resolver) error {
	var errs []error
	for _, r := range resolvers {
		if err := r.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
