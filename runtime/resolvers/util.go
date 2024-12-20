package resolvers

import (
	"context"
	"errors"
	"fmt"
	"io"
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
		if errors.Is(err, runtime.ErrMetricsViewCachingDisabled) {
			return nil, false, nil
		}
		return nil, false, err
	}
	defer cacheKeyResolver.Close()

	row, err := cacheKeyResolver.Next()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, false, fmt.Errorf("`metrics_cache_key` resolver returned no rows")
		}
		return nil, false, err
	}
	res, ok := row["key"].(string)
	if !ok {
		// should never happen but just in case
		return nil, false, errors.New("`metrics_cache_key`: expected a column key of type string in result")
	}
	return []byte(res), true, nil
}
