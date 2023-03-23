package runtime

import (
	"context"
	"strings"
)

type Query interface {
	// Key should return a cache key that uniquely identifies the query
	Key() string
	// Deps should return the source and model names that the query targets.
	// It's used to invalidate cached queries when the underlying data changes.
	Deps() []string
	// MarshalResult should return the query result for caching.
	// TODO: Also return estimated cost in bytes.
	MarshalResult() any
	// UnmarshalResult should populate a query with a cached result
	UnmarshalResult(v any) error
	// Resolve should execute the query against the instance's infra.
	// Error can be nil along with a nil result in general, i.e. when a model contains no rows aggregation results can be nil.
	Resolve(ctx context.Context, rt *Runtime, instanceID string, priority int) error
}

type queryCacheKey struct {
	instanceID    string
	queryKey      string
	dependencyKey string
}

func (r *Runtime) Query(ctx context.Context, instanceID string, query Query, priority int) error {
	// if key is empty, skip caching
	if query.Key() == "" {
		return query.Resolve(ctx, r, instanceID, priority)
	}
	deps := query.Deps()
	depKeys := make([]string, len(deps))
	for i, dep := range deps {
		entry, err := r.GetCatalogEntry(ctx, instanceID, dep)
		if err != nil {
			// return fmt.Errorf("query dependency %q not found", dep)
			continue
		}
		depKeys[i] = entry.Name + ":" + entry.RefreshedOn.String()
	}
	depKey := strings.Join(depKeys, ";")
	key := queryCacheKey{
		instanceID:    instanceID,
		queryKey:      query.Key(),
		dependencyKey: depKey,
	}
	val, ok := r.queryCache.get(key)
	if ok {
		return query.UnmarshalResult(val)
	}
	err := query.Resolve(ctx, r, instanceID, priority)
	if err != nil {
		return err
	}
	r.queryCache.add(key, query.MarshalResult())
	return nil
}
