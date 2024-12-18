package resolvers

import (
	"context"
	"errors"
	"slices"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
)

// normalizeRefs sorts and deduplicates the given refs.
// It modifies the input slice in place and returns it.
func normalizeRefs(refs []*runtimev1.ResourceName) []*runtimev1.ResourceName {
	// Sort
	slices.SortFunc(refs, func(a, b *runtimev1.ResourceName) int {
		if a.Kind != b.Kind {
			return strings.Compare(a.Kind, b.Kind)
		}
		return strings.Compare(a.Name, b.Name)
	})

	// Compact
	return slices.CompactFunc(refs, func(a, b *runtimev1.ResourceName) bool {
		return a.Kind == b.Kind && a.Name == b.Name
	})
}

func cacheKeyForMetricsView(ctx context.Context, r *runtime.Runtime, instanceID, name string, priority int) ([]byte, bool, error) {
	cacheKeyResolver, err := r.Resolve(ctx, &runtime.ResolveOptions{
		InstanceID:         instanceID,
		Resolver:           "metrics_cache_key",
		ResolverProperties: map[string]any{"metrics_view": name},
		Args:               map[string]any{"priority": priority},
		Claims:             &runtime.SecurityClaims{SkipChecks: true},
	})
	if err != nil {
		if errors.Is(err, errCachingDisabled) {
			return nil, false, nil
		}
		return nil, false, err
	}
	cacheKey, err := cacheKeyResolver.MarshalJSON()
	if err != nil {
		return nil, false, err
	}
	return cacheKey, true, nil
}
