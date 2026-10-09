package ai

import (
	"context"

	"github.com/rilldata/rill/runtime/drivers"
)

// This file exposes internals to the external ai_test package. It is only compiled during tests.

// PreloadSkills exposes preloadSkills to the external test package.
var PreloadSkills = preloadSkills

// ApplyMemoryOpsForTest applies a batch of memory operations, as background extraction does.
func ApplyMemoryOpsForTest(ctx context.Context, s *Session, ops []MemoryOp) ([]MemoryOp, error) {
	return applyMemoryOps(ctx, s, ops, drivers.AIMemorySourceExtracted, "")
}
