package resolvers

import (
	"slices"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
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
