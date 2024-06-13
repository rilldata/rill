package worker

import (
	"context"
	"time"
)

func (w *Worker) deleteExpiredAuthTokens(ctx context.Context) error {
	// Delete auth tokens that have been expired for more than 24 hours
	retention := 24 * time.Hour
	err := w.admin.DB.DeleteExpiredUserAuthTokens(ctx, retention)
	if err != nil {
		return err
	}
	err = w.admin.DB.DeleteExpiredServiceAuthTokens(ctx, retention)
	if err != nil {
		return err
	}
	err = w.admin.DB.DeleteExpiredDeploymentAuthTokens(ctx, retention)
	if err != nil {
		return err
	}
	err = w.admin.DB.DeleteExpiredMagicAuthTokens(ctx, retention)
	if err != nil {
		return err
	}
	return nil
}
