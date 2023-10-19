package worker

import (
	"context"
	"time"
)

func (w *Worker) deleteExpiredVirtualFiles(ctx context.Context) error {
	// Delete virtual files that have been soft deleted for more than 24 hours
	retention := 24 * time.Hour
	err := w.admin.DB.DeleteExpiredVirtualFiles(ctx, retention)
	if err != nil {
		return err
	}
	return nil
}
