package river

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/runtime/pkg/gitutil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// userFilesEnvironment is the environment user files are staged in. UI-created resources always target prod.
const userFilesEnvironment = "prod"

type SyncUserFilesToGitArgs struct {
	ProjectID string
}

func (SyncUserFilesToGitArgs) Kind() string { return "sync_user_files_to_git" }

type SyncUserFilesToGitWorker struct {
	river.WorkerDefaults[SyncUserFilesToGitArgs]
	admin  *admin.Service
	logger *zap.Logger
}

// NewSyncUserFilesToGitWorker creates a worker for direct invocation in tests.
func NewSyncUserFilesToGitWorker(adm *admin.Service, logger *zap.Logger) *SyncUserFilesToGitWorker {
	return &SyncUserFilesToGitWorker{admin: adm, logger: logger}
}

// Work flushes a project's staged user files (alerts, reports, personal canvases) from the virtual_files
// staging buffer into its Git repo as a single commit pushed directly to the primary branch.
// Rows whose flushed content is confirmed on the branch transition to synced; rows edited mid-flight stay
// dirty for the next run. If the primary branch is protected, the job is cancelled with a descriptive error
// and the files stay staged (and live, via the runtime's staging overlay).
func (w *SyncUserFilesToGitWorker) Work(ctx context.Context, job *river.Job[SyncUserFilesToGitArgs]) (err error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.project_id", job.Args.ProjectID))

	// Record any failure on the project, where the status API surfaces it; a later successful run clears it.
	// Context cancellation is excluded: it signals a shutdown rather than a sync failure, and the job will be retried.
	defer func() {
		if err != nil && !errors.Is(err, context.Canceled) {
			w.recordSyncError(ctx, job.Args.ProjectID, err)
		}
	}()

	proj, err := w.admin.DB.FindProject(ctx, job.Args.ProjectID)
	if err != nil {
		return err
	}

	// Only Git-connected projects can be synced. Archive-backed projects have nowhere to push.
	if proj.GitRemote == nil || proj.GithubInstallationID == nil {
		w.logger.Debug("sync user files: project is not connected to git, skipping", zap.String("project_id", proj.ID))
		return nil
	}

	staged, err := w.admin.DB.FindStagedVirtualFiles(ctx, proj.ID, userFilesEnvironment)
	if err != nil {
		return err
	}

	// Rewrite any staged rows still at legacy paths into the user_files layout before flushing,
	// so legacy root paths (e.g. "alerts/x.yaml") are never committed to Git.
	migrated, err := w.migrateLegacyPaths(ctx, proj, staged)
	if err != nil {
		return err
	}
	if migrated {
		staged, err = w.admin.DB.FindStagedVirtualFiles(ctx, proj.ID, userFilesEnvironment)
		if err != nil {
			return err
		}
	}
	if len(staged) == 0 {
		return w.admin.DB.UpdateProjectUserFilesSyncedOn(ctx, proj.ID, "")
	}

	cfg, err := w.admin.GitConfigForProject(ctx, proj)
	if err != nil {
		return err
	}

	// Reconcile against the served branch: mark synced any staged row whose desired state is already
	// reflected there (e.g. a prior sync crashed between push and mark, or someone committed equivalent
	// content by hand). This keeps re-runs from producing empty commits.
	// It also detects conflicts: rows whose Git copy changed while they were staged. The staged copy
	// still wins (the flush below overwrites it), but the overwrite is reported as a sync warning.
	staged, conflicts, err := w.reconcileAgainstServedBranch(ctx, proj, cfg, staged)
	if err != nil {
		return err
	}
	if len(staged) == 0 {
		return w.admin.DB.UpdateProjectUserFilesSyncedOn(ctx, proj.ID, "")
	}

	// Build the set of changes to flush for the rows that remain. Soft-deleted rows become file deletions.
	changes := make([]admin.FileChange, 0, len(staged))
	for _, vf := range staged {
		if _, _, isLegacy := admin.UserFilesRewriteLegacyPath(vf.Path); isLegacy {
			// Defensive: never commit a legacy root path (migrateLegacyPaths should have handled these).
			w.logger.Warn("sync user files: skipping staged file at unexpected legacy path", zap.String("project_id", proj.ID), zap.String("path", vf.Path))
			continue
		}
		if vf.Deleted {
			changes = append(changes, admin.FileChange{Path: vf.Path, Data: nil})
		} else {
			changes = append(changes, admin.FileChange{Path: vf.Path, Data: vf.Data})
		}
	}
	if len(changes) == 0 {
		return nil
	}

	_, err = w.admin.GitWriter.WriteFiles(ctx, &admin.GitWriteOptions{
		Config:     cfg,
		Branch:     proj.PrimaryBranch,
		BaseBranch: proj.PrimaryBranch,
		Subpath:    proj.Subpath,
		Message:    "Sync Rill-managed user files",
		Changes:    changes,
	})
	if err != nil && !errors.Is(err, gitutil.ErrEmptyCommit) {
		if admin.IsProtectedBranchErr(err) {
			// There is no pull request fallback yet, so retrying won't help. Cancel with a clear error;
			// the files stay staged and keep serving via the runtime's staging overlay.
			return river.JobCancel(fmt.Errorf("branch %q is protected and rejected the push; allow Rill to push to it directly", proj.PrimaryBranch))
		}
		return fmt.Errorf("failed to push user files: %w", err)
	}

	// Mark the flushed rows synced. Rows edited between the flush and here fail the (data, deleted)
	// compare-and-swap and stay dirty for the next run.
	for _, vf := range staged {
		if _, err := w.markSynced(ctx, proj.ID, vf); err != nil {
			return err
		}
	}

	warning := gitConflictWarning(conflicts)
	if warning != "" {
		w.logger.Warn("sync user files: overwrote changes made directly in git", zap.String("project_id", proj.ID), zap.Strings("paths", conflicts))
	}
	return w.admin.DB.UpdateProjectUserFilesSyncedOn(ctx, proj.ID, warning)
}

// recordSyncError persists the outcome of a failed sync attempt on the project, where the status API
// surfaces it. Recording is best-effort: it must not mask the original error.
func (w *SyncUserFilesToGitWorker) recordSyncError(ctx context.Context, projectID string, syncErr error) {
	// Unwrap river's cancel wrapper so its "JobCancelError:" prefix doesn't leak into the stored message.
	var cancelErr *rivertype.JobCancelError
	if errors.As(syncErr, &cancelErr) {
		syncErr = cancelErr.Unwrap()
	}
	msg := syncErr.Error()
	// Git errors can embed full command output; keep the stored message readable.
	if len(msg) > 500 {
		msg = msg[:500] + "…"
	}
	if err := w.admin.DB.UpdateProjectUserFilesSyncError(ctx, projectID, msg); err != nil {
		w.logger.Warn("sync user files: failed to record sync error", zap.String("project_id", projectID), zap.Error(err))
	}
}

// migrateLegacyPaths rewrites staged rows at legacy paths ("alerts/", "reports/", "personal/" at the repo
// root) into the user_files layout: the content is re-staged at its new path and the legacy row becomes a
// synced tombstone (legacy paths were never committed to Git, so there is nothing to flush for them; the
// tombstone tells runtimes to drop the staged copy at the old path). It reports whether any row was migrated.
func (w *SyncUserFilesToGitWorker) migrateLegacyPaths(ctx context.Context, proj *database.Project, staged []*database.VirtualFile) (bool, error) {
	migrated := false
	for _, vf := range staged {
		if _, _, isLegacy := admin.UserFilesRewriteLegacyPath(vf.Path); !isLegacy {
			continue
		}
		migrated = true

		if !vf.Deleted {
			newPath, ok, err := w.userFilesPathForLegacyFile(ctx, vf)
			if err != nil {
				return false, err
			}
			if !ok {
				continue
			}
			err = w.admin.DB.UpsertVirtualFile(ctx, &database.InsertVirtualFileOptions{
				ProjectID:   proj.ID,
				Environment: userFilesEnvironment,
				Path:        newPath,
				Data:        vf.Data,
				OwnerID:     vf.OwnerID,
			})
			if err != nil {
				return false, err
			}
			if err := w.admin.DB.UpdateVirtualFileDeleted(ctx, proj.ID, userFilesEnvironment, vf.Path); err != nil {
				return false, err
			}
			w.logger.Info("sync user files: migrated legacy path", zap.String("project_id", proj.ID), zap.String("from", vf.Path), zap.String("to", newPath))
		}

		// The legacy row is now (or already was) a tombstone. Mark it synced right away: its path was never
		// in Git, so there is no deletion to flush, and synced tombstones are eventually garbage-collected.
		if _, err := w.admin.DB.MarkVirtualFileSynced(ctx, proj.ID, userFilesEnvironment, vf.Path, []byte{}, true); err != nil {
			return false, err
		}
	}
	return migrated, nil
}

// userFilesPathForLegacyFile resolves the user_files path for a legacy row. Legacy alert/report rows don't
// store an owner in the database, so the owner is recovered from the "admin_owner_user_id" annotation their
// YAML carries; rows whose owner can't be resolved go to the shared segment.
func (w *SyncUserFilesToGitWorker) userFilesPathForLegacyFile(ctx context.Context, vf *database.VirtualFile) (string, bool, error) {
	if vf.OwnerID == nil || *vf.OwnerID == "" {
		if ownerID := ownerFromUserFileYAML(vf.Data); ownerID != "" {
			subdir, name, ok := admin.UserFilesRewriteLegacyPath(vf.Path)
			if !ok {
				return "", false, nil
			}
			newPath, err := w.admin.UserFilesPathForOwner(ctx, ownerID, subdir, name)
			if err != nil {
				return "", false, err
			}
			return newPath, true, nil
		}
	}
	return w.admin.UserFilesPathForVirtualFile(ctx, vf)
}

// ownerFromUserFileYAML extracts the owning user ID from a user file's "admin_owner_user_id" annotation.
// It returns "" if the data is not valid YAML or carries no owner annotation.
func ownerFromUserFileYAML(data []byte) string {
	var doc struct {
		Annotations struct {
			AdminOwnerUserID string `yaml:"admin_owner_user_id"`
		} `yaml:"annotations"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return ""
	}
	return doc.Annotations.AdminOwnerUserID
}

// reconcileAgainstServedBranch marks synced any staged row whose desired content already matches the served
// branch, and returns the rows that still need to be flushed. A non-deleted row is satisfied when the branch
// holds identical bytes; a deleted row (tombstone) is satisfied when the file is absent from the branch.
// It also returns the paths of unsatisfied rows in conflict with the branch: the branch copy differs from
// the version the staged change was based on (the row's base_data snapshot), meaning the file was edited
// directly in Git while the row was staged and the flush will overwrite that edit. The detection is
// best-effort: an edit pushed between this read and the flush is not caught.
func (w *SyncUserFilesToGitWorker) reconcileAgainstServedBranch(ctx context.Context, proj *database.Project, cfg *gitutil.Config, staged []*database.VirtualFile) ([]*database.VirtualFile, []string, error) {
	paths := make([]string, len(staged))
	for i, vf := range staged {
		paths[i] = vf.Path
	}

	present, err := w.admin.GitWriter.ReadFiles(ctx, cfg, proj.PrimaryBranch, proj.Subpath, paths)
	if err != nil {
		// If we can't read the branch (e.g. it doesn't exist yet), just proceed to flush everything.
		w.logger.Debug("sync user files: could not read served branch for reconcile", zap.String("project_id", proj.ID), zap.Error(err))
		return staged, nil, nil
	}

	remaining := staged[:0]
	var conflicts []string
	for _, vf := range staged {
		got, ok := present[vf.Path]
		satisfied := (vf.Deleted && !ok) || (!vf.Deleted && ok && bytes.Equal(got, vf.Data))
		if !satisfied {
			remaining = append(remaining, vf)
			// A branch copy that matches the row's base is the normal case (unchanged since the staged
			// change was based on it). Anything else present on the branch was put there directly in Git:
			// either edited while the row was staged, or created at a path Rill has never synced.
			if ok && (vf.BaseData == nil || !bytes.Equal(got, vf.BaseData)) {
				conflicts = append(conflicts, vf.Path)
			}
			continue
		}
		if _, err := w.markSynced(ctx, proj.ID, vf); err != nil {
			return nil, nil, err
		}
	}
	return remaining, conflicts, nil
}

// gitConflictWarning renders the sync warning for files whose direct-Git edits were overwritten by a flush.
// It returns "" for a clean sync.
func gitConflictWarning(conflicts []string) string {
	if len(conflicts) == 0 {
		return ""
	}
	const maxListed = 5
	listed := conflicts
	var more string
	if len(conflicts) > maxListed {
		listed = conflicts[:maxListed]
		more = fmt.Sprintf(" and %d more", len(conflicts)-maxListed)
	}
	return fmt.Sprintf("the sync overwrote changes made directly in Git to %s%s; the previous versions remain in the repository's Git history", strings.Join(listed, ", "), more)
}

// markSynced transitions a staged row to synced, gated on the (data, deleted) state that was flushed.
func (w *SyncUserFilesToGitWorker) markSynced(ctx context.Context, projectID string, vf *database.VirtualFile) (bool, error) {
	expected := vf.Data
	if vf.Deleted {
		expected = []byte{}
	}
	return w.admin.DB.MarkVirtualFileSynced(ctx, projectID, userFilesEnvironment, vf.Path, expected, vf.Deleted)
}
