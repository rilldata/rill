package sqlite

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// TTL for AI sessions; sessions with no messages newer than this are deleted by deleteExpiredAISessionsLoop.
var aiSessionTTL = 90 * 24 * time.Hour // 3 months

// aiSessionDeleteBatchSize bounds how many sessions are deleted per transaction.
// Each batch holds the sqlite write lock briefly, so other registry operations can interleave.
const aiSessionDeleteBatchSize = 200

// deleteExpiredAISessionsLoop runs deleteExpiredAISessions once at startup and then daily.
// It runs in the background so it doesn't block Migrate (and therefore runtime startup).
func (c *connection) deleteExpiredAISessionsLoop() {
	for {
		if err := c.deleteExpiredAISessions(c.ctx); err != nil && c.ctx.Err() == nil {
			c.logger.Error("sqlite: failed to delete expired AI sessions", zap.Error(err))
		}
		select {
		case <-c.ctx.Done():
			return
		case <-time.After(24 * time.Hour):
		}
	}
}

// deleteExpiredAISessions deletes AI sessions that have no messages newer than aiSessionTTL.
// It snapshots expired session IDs first, then deletes messages and sessions in batched transactions to bound lock-hold time.
func (c *connection) deleteExpiredAISessions(ctx context.Context) error {
	cutoff := time.Now().UTC().Add(-aiSessionTTL)

	ids, err := c.expiredAISessionIDs(ctx, cutoff)
	if err != nil {
		return err
	}

	// Delete in batches; each batch atomically deletes a session's messages and the session itself.
	for start := 0; start < len(ids); start += aiSessionDeleteBatchSize {
		end := start + aiSessionDeleteBatchSize
		if end > len(ids) {
			end = len(ids)
		}
		if err := c.deleteAISessionBatch(ctx, ids[start:end]); err != nil {
			return err
		}
	}

	return nil
}

// expiredAISessionIDs returns the IDs of AI sessions older than cutoff with no messages newer than cutoff.
// It does a single full scan of ai_messages (filtered by created_on >= cutoff) and a single scan of ai_sessions.
func (c *connection) expiredAISessionIDs(ctx context.Context, cutoff time.Time) ([]string, error) {
	rows, err := c.db.QueryContext(ctx, `
		SELECT id FROM ai_sessions
		WHERE created_on < ?
		  AND id NOT IN (SELECT session_id FROM ai_messages WHERE created_on >= ?)
	`, cutoff, cutoff)
	if err != nil {
		return nil, fmt.Errorf("failed to query expired AI sessions: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan expired AI session id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read expired AI sessions: %w", err)
	}
	return ids, nil
}

func (c *connection) deleteAISessionBatch(ctx context.Context, ids []string) error {
	placeholders := strings.Repeat("?,", len(ids)-1) + "?"
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin AI session cleanup transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`DELETE FROM ai_messages WHERE session_id IN (%s)`, placeholders), args...); err != nil {
		return fmt.Errorf("failed to delete expired AI messages: %w", err)
	}
	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`DELETE FROM ai_sessions WHERE id IN (%s)`, placeholders), args...); err != nil {
		return fmt.Errorf("failed to delete expired AI sessions: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit AI session cleanup: %w", err)
	}
	return nil
}
