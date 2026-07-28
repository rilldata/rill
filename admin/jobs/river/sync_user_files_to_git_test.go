package river

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-github/v71/github"
	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	_ "github.com/rilldata/rill/admin/database/postgres"
	"github.com/rilldata/rill/admin/pkg/pgtestcontainer"
	"github.com/rilldata/rill/runtime/pkg/gitutil"
	"github.com/riverqueue/river"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestSyncUserFilesToGitWorker(t *testing.T) {
	ctx := context.Background()
	const env = "prod"

	// Ephemeral Postgres with a minimal admin.Service around it. The GitHub client is a no-op mock whose
	// empty installation token makes GitConfigForProject produce credential-less configs, so a local
	// file:// bare repo can stand in for GitHub.
	pg := pgtestcontainer.New(t)
	t.Cleanup(func() { pg.Terminate(t) })

	encKeyRing, err := database.NewRandomKeyring()
	require.NoError(t, err)
	conf, err := json.Marshal(encKeyRing)
	require.NoError(t, err)
	db, err := database.Open("postgres", pg.DatabaseURL, string(conf))
	require.NoError(t, err)
	require.NoError(t, db.Migrate(ctx))
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	adm := &admin.Service{
		DB:        db,
		Github:    noopGithub{},
		GitWriter: admin.NewGitWriter(),
		Logger:    zap.NewNop(),
	}
	worker := NewSyncUserFilesToGitWorker(adm, zap.NewNop())

	// Fixtures: an org, a project connected to a file:// bare repo, and the owning user.
	org, err := db.InsertOrganization(ctx, &database.InsertOrganizationOptions{Name: "test"})
	require.NoError(t, err)
	proj, err := db.InsertProject(ctx, &database.InsertProjectOptions{
		OrganizationID: org.ID,
		Name:           "proj",
		PrimaryBranch:  "main",
	})
	require.NoError(t, err)
	user, err := db.InsertUser(ctx, &database.InsertUserOptions{Email: "jane.doe@acme.com", DisplayName: "Jane Doe"})
	require.NoError(t, err)

	// InsertProject validates GitRemote as an HTTP URL, so the file:// remote is set with raw SQL.
	remote := setupBareRepoWithCommit(t)
	rawDB, err := sql.Open("pgx", pg.DatabaseURL)
	require.NoError(t, err)
	t.Cleanup(func() { _ = rawDB.Close() })
	_, err = rawDB.ExecContext(ctx,
		"UPDATE projects SET git_remote=$1, github_installation_id=123, github_repo_id=456 WHERE id=$2",
		remote, proj.ID)
	require.NoError(t, err)

	upsert := func(path string, data []byte, ownerID *string) {
		t.Helper()
		require.NoError(t, db.UpsertVirtualFile(ctx, &database.InsertVirtualFileOptions{
			ProjectID:   proj.ID,
			Environment: env,
			Path:        path,
			Data:        data,
			OwnerID:     ownerID,
		}))
	}
	runJob := func() error {
		return worker.Work(ctx, &river.Job[SyncUserFilesToGitArgs]{Args: SyncUserFilesToGitArgs{ProjectID: proj.ID}})
	}

	// Stage: a legacy alert whose owner is recoverable from its YAML annotation, a legacy personal-file
	// tombstone, and a file already in the new layout.
	legacyAlertYAML := []byte("type: alert\nannotations:\n  admin_owner_user_id: " + user.ID + "\n")
	upsert("alerts/my-alert.yaml", legacyAlertYAML, nil)
	upsert("personal/old-canvas.yaml", []byte("type: canvas\n"), &user.ID)
	require.NoError(t, db.UpdateVirtualFileDeleted(ctx, proj.ID, env, "personal/old-canvas.yaml"))
	newCanvasPath := "user_files/jane-doe-" + user.ID + "/canvas/my-canvas.yaml"
	upsert(newCanvasPath, []byte("type: canvas\nv: 1\n"), &user.ID)

	require.NoError(t, runJob())

	// The legacy alert was migrated: committed under user_files/<segment derived from the annotation>,
	// never at its legacy root path.
	migratedAlertPath := "user_files/jane-doe-" + user.ID + "/alerts/my-alert.yaml"
	files := lsRemote(t, remote)
	require.Contains(t, files, migratedAlertPath)
	require.Contains(t, files, newCanvasPath)
	require.NotContains(t, files, "alerts/my-alert.yaml")
	require.NotContains(t, files, "personal/old-canvas.yaml")

	// All rows settled: the flushed rows are synced, the legacy rows are synced tombstones.
	count, err := db.CountStagedVirtualFiles(ctx, proj.ID, env)
	require.NoError(t, err)
	require.Zero(t, count)
	vf, err := db.FindVirtualFile(ctx, proj.ID, env, migratedAlertPath)
	require.NoError(t, err)
	require.Equal(t, database.VirtualFileSyncStateSynced, vf.SyncState)
	vf, err = db.FindVirtualFile(ctx, proj.ID, env, "alerts/my-alert.yaml")
	require.NoError(t, err)
	require.True(t, vf.Deleted)
	require.Equal(t, database.VirtualFileSyncStateSynced, vf.SyncState)

	proj, err = db.FindProject(ctx, proj.ID)
	require.NoError(t, err)
	require.NotNil(t, proj.UserFilesSyncedOn)
	require.Empty(t, proj.UserFilesSyncError)

	// A second run with nothing staged is a no-op.
	headBefore := remoteHead(t, remote)
	require.NoError(t, runJob())
	require.Equal(t, headBefore, remoteHead(t, remote))

	// An edit re-stages the row; the next sync commits the new content.
	upsert(newCanvasPath, []byte("type: canvas\nv: 2\n"), &user.ID)
	require.NoError(t, runJob())
	got, err := adm.GitWriter.ReadFiles(ctx, &gitutil.Config{Remote: remote}, "main", "", []string{newCanvasPath})
	require.NoError(t, err)
	require.Equal(t, []byte("type: canvas\nv: 2\n"), got[newCanvasPath])

	// Content that already matches the branch reconciles to synced without a new commit.
	upsert(newCanvasPath, []byte("type: canvas\nv: 2\n"), &user.ID)
	headBefore = remoteHead(t, remote)
	require.NoError(t, runJob())
	require.Equal(t, headBefore, remoteHead(t, remote))
	vf, err = db.FindVirtualFile(ctx, proj.ID, env, newCanvasPath)
	require.NoError(t, err)
	require.Equal(t, database.VirtualFileSyncStateSynced, vf.SyncState)

	// A deletion is flushed as a file removal.
	require.NoError(t, db.UpdateVirtualFileDeleted(ctx, proj.ID, env, newCanvasPath))
	require.NoError(t, runJob())
	require.NotContains(t, lsRemote(t, remote), newCanvasPath)
	vf, err = db.FindVirtualFile(ctx, proj.ID, env, newCanvasPath)
	require.NoError(t, err)
	require.True(t, vf.Deleted)
	require.Equal(t, database.VirtualFileSyncStateSynced, vf.SyncState)

	// A protected branch cancels the job with a descriptive error and leaves the rows staged.
	upsert(migratedAlertPath, []byte("type: alert\nv: 2\n"), nil)
	hook := filepath.Join(remote[len("file://"):], "hooks", "pre-receive")
	require.NoError(t, os.WriteFile(hook, []byte("#!/bin/sh\necho \"GH006: Protected branch update failed\" >&2\nexit 1\n"), 0o755))
	err = runJob()
	require.Error(t, err)
	require.ErrorContains(t, err, "protected")
	vf, err = db.FindVirtualFile(ctx, proj.ID, env, migratedAlertPath)
	require.NoError(t, err)
	require.Equal(t, database.VirtualFileSyncStateDirty, vf.SyncState)

	// The failure is recorded on the project so the status API can surface it.
	proj, err = db.FindProject(ctx, proj.ID)
	require.NoError(t, err)
	require.Contains(t, proj.UserFilesSyncError, "protected")

	// Removing the protection lets the next run flush the pending row and clear the error.
	require.NoError(t, os.Remove(hook))
	require.NoError(t, runJob())
	proj, err = db.FindProject(ctx, proj.ID)
	require.NoError(t, err)
	require.Empty(t, proj.UserFilesSyncError)

	// A file edited directly in Git while its row is staged: the staged copy wins (sync must converge),
	// but the overwrite is reported as a sync warning.
	upsert(migratedAlertPath, []byte("type: alert\nv: 3\n"), nil)
	commitFileToRemote(t, remote, migratedAlertPath, []byte("type: alert\nhand-edited: true\n"))
	require.NoError(t, runJob())
	got, err = adm.GitWriter.ReadFiles(ctx, &gitutil.Config{Remote: remote}, "main", "", []string{migratedAlertPath})
	require.NoError(t, err)
	require.Equal(t, []byte("type: alert\nv: 3\n"), got[migratedAlertPath])
	proj, err = db.FindProject(ctx, proj.ID)
	require.NoError(t, err)
	require.Contains(t, proj.UserFilesSyncWarning, migratedAlertPath)
	vf, err = db.FindVirtualFile(ctx, proj.ID, env, migratedAlertPath)
	require.NoError(t, err)
	require.Equal(t, database.VirtualFileSyncStateSynced, vf.SyncState)

	// A normal edit (the Git copy still matches the version the edit was based on) syncs cleanly
	// and clears the warning.
	upsert(migratedAlertPath, []byte("type: alert\nv: 4\n"), nil)
	require.NoError(t, runJob())
	proj, err = db.FindProject(ctx, proj.ID)
	require.NoError(t, err)
	require.Empty(t, proj.UserFilesSyncWarning)
}

// noopGithub is a minimal admin.Github mock. Its empty installation token yields credential-less
// git configs, which work against local file:// remotes.
type noopGithub struct{}

func (noopGithub) AppClient() *github.Client { return nil }
func (noopGithub) InstallationClient(installationID int64, repoID *int64) *github.Client {
	return nil
}

func (noopGithub) InstallationToken(ctx context.Context, installationID, repoID int64) (string, time.Time, error) {
	return "", time.Time{}, nil
}

func (noopGithub) InstallationTokenForOrg(ctx context.Context, org string) (string, time.Time, error) {
	return "", time.Time{}, nil
}

func (noopGithub) DeleteBranch(ctx context.Context, installationID, repoID int64, remote, branch string) error {
	return nil
}

func (noopGithub) CreateManagedRepo(ctx context.Context, repoPrefix string, autoInit bool) (*github.Repository, error) {
	return nil, fmt.Errorf("not implemented")
}

func (noopGithub) DeleteManagedRepo(ctx context.Context, repo string) error {
	return fmt.Errorf("not implemented")
}

func (noopGithub) ManagedOrgInstallationID() (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

// setupBareRepoWithCommit creates a bare Git repo with a single commit on main and returns its file:// URL.
func setupBareRepoWithCommit(t *testing.T) string {
	remoteDir := t.TempDir()
	runGitCmd(t, "", "init", "--bare", remoteDir)
	runGitCmd(t, remoteDir, "symbolic-ref", "HEAD", "refs/heads/main")

	workDir := t.TempDir()
	runGitCmd(t, "", "clone", remoteDir, workDir)
	runGitCmd(t, workDir, "config", "user.name", "Test")
	runGitCmd(t, workDir, "config", "user.email", "test@example.com")
	require.NoError(t, os.WriteFile(filepath.Join(workDir, "rill.yaml"), []byte("compiler: rillv1\n"), 0o644))
	runGitCmd(t, workDir, "add", ".")
	runGitCmd(t, workDir, "commit", "-m", "Initial commit")
	runGitCmd(t, workDir, "branch", "-M", "main")
	runGitCmd(t, workDir, "push", "origin", "main")

	return "file://" + remoteDir
}

// commitFileToRemote commits content at path on the remote's main branch, simulating a direct Git edit.
func commitFileToRemote(t *testing.T, remote, path string, data []byte) {
	dir := t.TempDir()
	runGitCmd(t, "", "clone", "--branch", "main", remote, dir)
	runGitCmd(t, dir, "config", "user.name", "Test")
	runGitCmd(t, dir, "config", "user.email", "test@example.com")
	full := filepath.Join(dir, filepath.FromSlash(path))
	require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
	require.NoError(t, os.WriteFile(full, data, 0o644))
	runGitCmd(t, dir, "add", ".")
	runGitCmd(t, dir, "commit", "-m", "Direct edit")
	runGitCmd(t, dir, "push", "origin", "main")
}

// lsRemote returns all file paths on the main branch of the remote.
func lsRemote(t *testing.T, remote string) []string {
	dir := t.TempDir()
	runGitCmd(t, "", "clone", "--branch", "main", "--depth", "1", remote, dir)
	var files []string
	err := filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	require.NoError(t, err)
	return files
}

// remoteHead returns the commit hash of main on the remote.
func remoteHead(t *testing.T, remote string) string {
	out, err := exec.Command("git", "ls-remote", remote, "refs/heads/main").Output()
	require.NoError(t, err)
	return string(out)
}

func runGitCmd(t *testing.T, dir string, args ...string) {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v failed: %s", args, out)
}
