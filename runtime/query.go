package runtime

import (
	"context"
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
	// TODO: Add caching here
	return query.Resolve(ctx, r, instanceID, priority)
}
