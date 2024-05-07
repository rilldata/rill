package worker

import (
	"context"
	"time"
)

func (w *Worker) deleteExpiredAuthCodes(ctx context.Context) error {
	// Delete auth codes that have been expired for more than 24 hours.
	// By delaying deletion past the expiration time, we can provide a nicer error message for expired codes.
	// (The user will see "code has expired" instead of "code not found".)
	return w.admin.DB.DeleteExpiredAuthorizationCodes(ctx, 24*time.Hour)
}
