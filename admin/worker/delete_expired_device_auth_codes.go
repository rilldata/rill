package worker

import (
	"context"
	"time"
)

func (w *Worker) deleteExpiredDeviceAuthCodes(ctx context.Context) error {
	// Delete device auth codes that have been expired for more than 24 hours.
	// By delaying deletion past the expiration time, we can provide a nicer error message for expired codes.
	// (The user will see "code has expired" instead of "code not found".)
	return w.admin.DB.DeleteExpiredDeviceAuthCodes(ctx, 24*time.Hour)
}
