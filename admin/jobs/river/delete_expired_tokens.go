package river

import (
	"context"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/riverqueue/river"
)

type DeleteExpiredTokensArgs struct{}

func (DeleteExpiredTokensArgs) Kind() string { return "delete_expired_tokens" }

type DeleteExpiredTokensWorker struct {
	river.WorkerDefaults[DeleteExpiredTokensArgs]
	admin *admin.Service
}

func (w *DeleteExpiredTokensWorker) Work(ctx context.Context, job *river.Job[DeleteExpiredTokensArgs]) error {
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
