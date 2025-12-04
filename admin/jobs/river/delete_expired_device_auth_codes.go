package river

import (
	"context"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

type DeleteExpiredDeviceAuthCodesArgs struct{}

func (DeleteExpiredDeviceAuthCodesArgs) Kind() string { return "delete_expired_device_auth_codes" }

type DeleteExpiredDeviceAuthCodesWorker struct {
	river.WorkerDefaults[DeleteExpiredDeviceAuthCodesArgs]
	admin  *admin.Service
	logger *zap.Logger
}

func (w *DeleteExpiredDeviceAuthCodesWorker) Work(ctx context.Context, job *river.Job[DeleteExpiredDeviceAuthCodesArgs]) error {
	// Delete device auth codes that have been expired for more than 24 hours.
	// By delaying deletion past the expiration time, we can provide a nicer error message for expired codes.
	// (The user will see "code has expired" instead of "code not found".)
	return w.admin.DB.DeleteExpiredDeviceAuthCodes(ctx, 24*time.Hour)
}
