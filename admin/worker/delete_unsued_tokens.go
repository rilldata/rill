package worker

import (
	"context"
	"time"
)

// The retention period is defined as 180 days (6 months).
const tokenRetentionPeriod = time.Hour * 24 * 180

// deleteUnusedTokens revokes both user and service tokens that haven't been used for more than the retention period.
func (w *Worker) deleteUnusedTokens(ctx context.Context) error {
	err := w.admin.DB.DeleteInactiveUserAuthTokens(ctx, tokenRetentionPeriod)
	if err != nil {
		return err
	}
	err = w.admin.DB.DeleteInactiveServiceAuthTokens(ctx, tokenRetentionPeriod)
	if err != nil {
		return err
	}

	return nil
}
