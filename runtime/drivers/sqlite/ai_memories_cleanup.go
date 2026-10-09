package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
)

// aiMemoryTombstoneTTL is how long a deleted AI memory is kept as a tombstone (so the deletion can be undone) before it is purged.
var aiMemoryTombstoneTTL = 7 * 24 * time.Hour

// deleteExpiredAIMemoryTombstonesLoop purges expired AI memory tombstones once at startup and then daily.
// Active memories are never touched: unlike AI sessions, memories have no TTL.
func (c *connection) deleteExpiredAIMemoryTombstonesLoop() {
	for {
		if err := c.deleteExpiredAIMemoryTombstones(c.ctx); err != nil && c.ctx.Err() == nil {
			c.logger.Error("sqlite: failed to delete expired AI memory tombstones", zap.Error(err))
		}
		select {
		case <-c.ctx.Done():
			return
		case <-time.After(24 * time.Hour):
		}
	}
}

func (c *connection) deleteExpiredAIMemoryTombstones(ctx context.Context) error {
	cutoff := time.Now().UTC().Add(-aiMemoryTombstoneTTL)
	_, err := c.db.ExecContext(ctx, `DELETE FROM ai_memories WHERE status = ? AND updated_on < ?`, drivers.AIMemoryStatusDeleted, cutoff)
	if err != nil {
		return fmt.Errorf("failed to delete expired AI memory tombstones: %w", err)
	}
	return nil
}
