package river

import (
	"context"

	"github.com/rilldata/rill/admin"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

type SyncAllUserFilesToGitArgs struct{}

func (SyncAllUserFilesToGitArgs) Kind() string { return "sync_all_user_files_to_git" }

// SyncAllUserFilesToGitWorker is a periodic sweep that enqueues a user-files sync for every project with
// auto-sync enabled that is due for one and has staged files. Auto-sync is opt-in per project: admins enable
// it and choose the interval (projects.user_files_sync_interval_seconds), so the sweep runs frequently but
// only touches projects that asked for it.
type SyncAllUserFilesToGitWorker struct {
	river.WorkerDefaults[SyncAllUserFilesToGitArgs]
	admin  *admin.Service
	logger *zap.Logger
}

func (w *SyncAllUserFilesToGitWorker) Work(ctx context.Context, job *river.Job[SyncAllUserFilesToGitArgs]) error {
	ids, err := w.admin.DB.FindProjectIDsForUserFilesAutoSync(ctx)
	if err != nil {
		return err
	}
	for _, id := range ids {
		// Uniqueness on the sync job dedupes this with manual triggers for the same project.
		if _, err := w.admin.Jobs.SyncUserFilesToGit(ctx, id); err != nil {
			w.logger.Error("failed to enqueue user files sync", zap.String("project_id", id), zap.Error(err))
		}
	}
	return nil
}
