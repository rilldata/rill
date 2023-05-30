package worker

import (
	"context"
)

func (w *Worker) deleteExpiredDeviceAuthCodes(ctx context.Context) error {
	return w.admin.DB.DeleteExpiredDeviceAuthCodes(ctx)
}
