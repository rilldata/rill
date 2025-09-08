package river

import (
	"context"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

type DeleteExpiredAuthCodesArgs struct{}

func (DeleteExpiredAuthCodesArgs) Kind() string { return "delete_expired_auth_codes" }

type DeleteExpiredAuthCodesWorker struct {
	river.WorkerDefaults[DeleteExpiredAuthCodesArgs]
	admin  *admin.Service
	logger *zap.Logger
}

func (w *DeleteExpiredAuthCodesWorker) Work(ctx context.Context, job *river.Job[DeleteExpiredAuthCodesArgs]) error {
	// Delete auth codes that have been expired for more than 24 hours.
	// By delaying deletion past the expiration time, we can provide a nicer error message for expired codes.
	// (The user will see "code has expired" instead of "code not found".)
	return w.admin.DB.DeleteExpiredAuthorizationCodes(ctx, 24*time.Hour)
}
