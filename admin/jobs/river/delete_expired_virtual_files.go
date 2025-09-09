package river

import (
	"context"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/riverqueue/river"
)

type DeleteExpiredVirtualFilesArgs struct{}

func (DeleteExpiredVirtualFilesArgs) Kind() string { return "delete_expired_virtual_files" }

type DeleteExpiredVirtualFilesWorker struct {
	river.WorkerDefaults[DeleteExpiredVirtualFilesArgs]
	admin *admin.Service
}

func (w *DeleteExpiredVirtualFilesWorker) Work(ctx context.Context, job *river.Job[DeleteExpiredVirtualFilesArgs]) error {
	// Delete virtual files that have been soft deleted for more than 24 hours
	retention := 24 * time.Hour
	err := w.admin.DB.DeleteExpiredVirtualFiles(ctx, retention)
	if err != nil {
		return err
	}
	return nil
}
