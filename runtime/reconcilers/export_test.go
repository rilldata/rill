package reconcilers

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// SetModelFirstPartitionHookForTest installs a barrier for sequential cancellation tests.
// Install it before creating the test runtime, so the runtime closes before cleanup restores the hook.
func SetModelFirstPartitionHookForTest(t testing.TB, hook func(context.Context, string, *runtimev1.ResourceName)) {
	previous := modelFirstPartitionExecutedHook.Swap(&hook)
	t.Cleanup(func() { modelFirstPartitionExecutedHook.Store(previous) })
}
