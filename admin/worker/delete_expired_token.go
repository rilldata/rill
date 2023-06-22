package worker

import (
	"context"
	"time"
)

func (w *Worker) deleteExpiredUserAuthTokens(ctx context.Context) error {
	// Delete user auth tokens that have been expired for more than 24 hours.
	return w.admin.DB.DeleteExpiredUserAuthTokens(ctx, 24*time.Hour)
}
