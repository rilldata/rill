package runtime

import (
	"context"
	"fmt"
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
	Resolve(ctx context.Context, rt *Runtime, instanceID string, priority int) error
}

func (r *Runtime) Query(ctx context.Context, instanceID string, query Query, priority int) error {
	// take all deps and their last updated time and add them to the query key, prefix query key with instanceID
	cacheKey := instanceID + query.Key()
	deps := query.Deps()
	service, err := r.catalogCache.get(ctx, r, instanceID)
	if err != nil {
		return err
	}
	for _, dep := range deps {
		entry, found := service.FindEntry(ctx, dep)
		if !found {
			r.logger.Error(fmt.Sprintf("dependency %s not found, ignoring it!", dep))
			continue
		}
		cacheKey += entry.Name + entry.UpdatedOn.String()
	}
	val, ok := r.queryCache.get(cacheKey)
	if ok {
		return query.UnmarshalResult(val)
	}
	err = query.Resolve(ctx, r, instanceID, priority)
	if err != nil {
		return err
	}
	r.queryCache.add(cacheKey, query.MarshalResult())
	return nil
}
