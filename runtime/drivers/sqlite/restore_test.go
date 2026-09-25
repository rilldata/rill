package sqlite

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gocloud.dev/blob"
	"gocloud.dev/blob/fileblob"
)

func TestBackupAndRestore(t *testing.T) {
	// File paths for the test
	tmpdir := t.TempDir()
	dbPath := filepath.Join(tmpdir, "data.sqlite")
	storageDir := filepath.Join(tmpdir, "storage")
	bucketDir := filepath.Join(storageDir, "bucket")

	// Create local bucket
	bucket, err := fileblob.OpenBucket(bucketDir, &fileblob.Options{CreateDir: true})
	require.NoError(t, err)
	// Production scopes the bucket to the backup directory; see connection_cache.go and startBackups().
	bucket = blob.PrefixedBucket(bucket, "shared/metastore/test-restore/")

	cfg := map[string]any{
		"dsn":            dbPath,
		"id":             "test-restore",
		"backups_enable": true,
	}

	// Create a database with an instance in it, then back it up.
	h, err := driver{}.Open(context.Background(), "", "", cfg, storage.MustNew(storageDir, nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	require.NoError(t, h.Migrate(t.Context()))
	registry, ok := h.AsRegistry()
	require.True(t, ok)
	require.NoError(t, registry.CreateInstance(t.Context(), &drivers.Instance{ID: "a"}))
	require.NoError(t, h.(*connection).backup(t.Context(), bucket))
	require.NoError(t, h.Close())

	// Simulate the loss of the local database file, then restore it.
	require.NoError(t, os.Remove(dbPath))
	require.NoError(t, restoreBackup(t.Context(), bucket, dbPath, zap.NewNop()))

	// Reopen the database and check the instance is back.
	h, err = driver{}.Open(context.Background(), "", "", cfg, storage.MustNew(storageDir, nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	defer h.Close()
	require.NoError(t, h.Migrate(t.Context()))
	registry, ok = h.AsRegistry()
	require.True(t, ok)
	instances, err := registry.FindInstances(t.Context())
	require.NoError(t, err)
	require.Len(t, instances, 1)
	require.Equal(t, "a", instances[0].ID)
}

func TestRestoreCorruptSnapshot(t *testing.T) {
	tmpdir := t.TempDir()
	dbPath := filepath.Join(tmpdir, "data.sqlite")

	bucket, err := fileblob.OpenBucket(filepath.Join(tmpdir, "bucket"), &fileblob.Options{CreateDir: true})
	require.NoError(t, err)
	bucket = blob.PrefixedBucket(bucket, "shared/metastore/test-corrupt/")
	require.NoError(t, bucket.WriteAll(t.Context(), backupSnapshotName, []byte("not a database"), nil))

	// The restore must fail rather than leave a broken database behind for the next backup to overwrite.
	require.Error(t, restoreBackup(t.Context(), bucket, dbPath, zap.NewNop()))
	require.NoFileExists(t, dbPath)
	require.NoFileExists(t, dbPath+".restore")
}

func TestRestoreMissingSnapshot(t *testing.T) {
	tmpdir := t.TempDir()
	bucket, err := fileblob.OpenBucket(filepath.Join(tmpdir, "bucket"), &fileblob.Options{CreateDir: true})
	require.NoError(t, err)
	bucket = blob.PrefixedBucket(bucket, "shared/metastore/test-missing/")

	// An empty backup directory is the normal case for a new deployment, so it must not be an error.
	dbPath := filepath.Join(tmpdir, "data.sqlite")
	require.NoError(t, restoreBackup(t.Context(), bucket, dbPath, zap.NewNop()))
	require.NoFileExists(t, dbPath)

	// A key that merely shares a prefix with the snapshot is not a snapshot.
	require.NoError(t, bucket.WriteAll(t.Context(), backupSnapshotName+".old", []byte("decoy"), nil))
	require.NoError(t, restoreBackup(t.Context(), bucket, dbPath, zap.NewNop()))
	require.NoFileExists(t, dbPath)
}

func TestShouldRestoreBackup(t *testing.T) {
	// EvalSymlinks because SQLite reports the resolved path, which differs from t.TempDir() on macOS.
	tmpdir, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)

	// In-memory databases can't be restored into.
	for _, dsn := range []string{":memory:", "file:rill?mode=memory&cache=shared", "file::memory:?cache=shared"} {
		dbPath, ok, err := shouldRestoreBackup(t.Context(), dsn)
		require.NoError(t, err)
		require.False(t, ok, "dsn %q", dsn)
		require.Empty(t, dbPath)
	}

	// A database file that doesn't exist yet has no migrations applied.
	dsn := filepath.Join(tmpdir, "data.sqlite")
	dbPath, ok, err := shouldRestoreBackup(t.Context(), dsn)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, dsn, dbPath)

	// Nor does one that exists but was never migrated (the call above created it).
	require.FileExists(t, dsn)
	dbPath, ok, err = shouldRestoreBackup(t.Context(), dsn)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, dsn, dbPath)

	// Nor does one where Migrate() created the version table and then crashed before inserting the version row.
	db, err := sqlx.Open("sqlite", dsn)
	require.NoError(t, err)
	_, err = db.ExecContext(t.Context(), fmt.Sprintf("CREATE TABLE %s(version integer not null)", migrationVersionTable))
	require.NoError(t, err)
	require.NoError(t, db.Close())

	dbPath, ok, err = shouldRestoreBackup(t.Context(), dsn)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, dsn, dbPath)

	// Once migrated, the database has data and must not be overwritten by a restore.
	h, err := driver{}.Open(context.Background(), "", "", map[string]any{"dsn": dsn}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	require.NoError(t, h.Migrate(t.Context()))
	require.NoError(t, h.Close())

	dbPath, ok, err = shouldRestoreBackup(t.Context(), dsn)
	require.NoError(t, err)
	require.False(t, ok)
	require.Empty(t, dbPath)
}
