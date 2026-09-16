package sqlite

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/runtime/storage"
	"go.uber.org/zap"
	"gocloud.dev/blob"
	"gocloud.dev/gcerrors"

	// Register the SQLite driver
	_ "modernc.org/sqlite"
)

// Max time a restore may run for.
// It blocks runtime startup, so we bound it instead of letting a slow download hang the process.
var restoreMaxDuration = 10 * time.Minute

// restoreBackupIfEmpty restores the SQLite database at the given DSN from the latest backup in object storage,
// but only if the database has no data yet. It is called from driver.Open() before the connection handle is created,
// since it replaces the database file. See startBackups() for how backups are produced and where they are stored.
//
// It is a no-op (returning nil) if the database is in-memory, if it already has data,
// if no bucket is configured on the storage client, or if the backup directory doesn't contain a snapshot.
// Any other failure is returned as an error, which fails runtime startup.
// Starting with an empty database would be worse: the next backup would overwrite the snapshot we failed to restore.
func restoreBackupIfEmpty(ctx context.Context, st *storage.Client, backupID, dsn string, logger *zap.Logger) error {
	ctx, cancel := context.WithTimeout(ctx, restoreMaxDuration)
	defer cancel()

	// Check if the database is a restore candidate before touching object storage.
	// Opening a bucket resolves cloud credentials, which we shouldn't require on a startup that doesn't need it.
	dbPath, ok, err := shouldRestoreBackup(ctx, dsn)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	// Open bucket scoped to the backup directory.
	// Return early (no-op) if a bucket isn't available.
	bucket, ok, err := st.OpenBucket(ctx, backupID)
	if err != nil {
		return fmt.Errorf("could not open backup bucket: %w", err)
	}
	if !ok {
		return nil
	}
	defer bucket.Close()

	return restoreBackup(ctx, bucket, dbPath, logger.With(zap.String("backup_id", backupID)))
}

// shouldRestoreBackup reports whether the SQLite database at the given DSN is empty and should be restored from a backup,
// along with the path of the database file to restore into.
//
// "Empty" means migrations have never run on the database.
// That is a wider condition than the database file not existing, and deliberately so:
// it also covers a zero-length file and a previous startup that created the file and then crashed before migrating.
// Without those cases, a single failed startup would permanently disable restores,
// and the next backup would overwrite the good snapshot with an empty database.
//
// It closes its connection before returning, which is what makes it safe for the caller to replace the database file:
// an open SQLite connection keeps referencing the file it originally opened.
func shouldRestoreBackup(ctx context.Context, dsn string) (dbPath string, ok bool, err error) {
	db, err := sqlx.Open("sqlite", dsn)
	if err != nil {
		return "", false, fmt.Errorf("failed to open database: %w", err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()

	// Ask SQLite for the file path instead of parsing the DSN, which can take several forms.
	// An empty path means the database is in-memory and can't be restored into.
	err = db.QueryRowContext(ctx, `SELECT file FROM pragma_database_list WHERE name = 'main';`).Scan(&dbPath)
	if err != nil {
		return "", false, fmt.Errorf("failed to find database file path: %w", err)
	}
	if dbPath == "" || dbPath == ":memory:" || dbPath == "file::memory:" {
		return "", false, nil
	}

	// The migration version table doesn't exist on a database that has never been migrated.
	// We check sqlite_master instead of matching on the "no such table" error string.
	var tables int
	err = db.QueryRowContext(ctx, `SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = ?`, migrationVersionTable).Scan(&tables)
	if err != nil {
		return "", false, fmt.Errorf("failed to check for migration version table: %w", err)
	}
	if tables != 0 {
		// The table is created with version 0 before the first migration is applied, so it may exist on an empty database.
		var version int
		err = db.QueryRowContext(ctx, fmt.Sprintf("SELECT version FROM %s", migrationVersionTable)).Scan(&version)
		if err != nil {
			return "", false, fmt.Errorf("failed to read migration version: %w", err)
		}
		if version > 0 {
			return "", false, nil
		}
	}

	return dbPath, true, nil
}

// restoreBackup downloads the snapshot from the backup bucket and moves it into place at dbPath.
// It assumes the bucket is already scoped to the correct backup directory and that dbPath is not currently open.
// It is a no-op (returning nil) if the backup directory doesn't contain a snapshot,
// which is the normal case for a new deployment.
func restoreBackup(ctx context.Context, bucket *blob.Bucket, dbPath string, logger *zap.Logger) error {
	attrs, err := bucket.Attributes(ctx, backupSnapshotName)
	if err != nil {
		if gcerrors.Code(err) == gcerrors.NotFound {
			logger.Info("sqlite: no backup found, starting with an empty database")
			return nil
		}
		return fmt.Errorf("failed to check for backup snapshot: %w", err)
	}

	err = os.MkdirAll(filepath.Dir(dbPath), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Download the snapshot to a temporary file in the same directory as the database, so the rename below is atomic.
	// The name is deterministic and the file is truncated on create, so a crashed restore doesn't leak a file per attempt.
	tmpPath := dbPath + ".restore"
	defer os.Remove(tmpPath)
	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temporary file for snapshot: %w", err)
	}
	defer f.Close() // Best-effort; we also close it explicitly below.
	err = bucket.Download(ctx, backupSnapshotName, f, nil)
	if err != nil {
		return fmt.Errorf("failed to download snapshot: %w", err)
	}
	err = f.Close()
	if err != nil {
		return fmt.Errorf("failed to write snapshot: %w", err)
	}

	// Verify the download is a well-formed SQLite database containing a metastore schema.
	// This catches a truncated or corrupted download before we overwrite anything.
	// We deliberately don't run PRAGMA integrity_check or quick_check: snapshots can be several GB,
	// and the underlying object storage already verifies transfer integrity.
	snapshotDB, err := sqlx.Open("sqlite", tmpPath)
	if err != nil {
		return fmt.Errorf("failed to open downloaded snapshot: %w", err)
	}
	snapshotDB.SetMaxOpenConns(1)
	defer snapshotDB.Close() // Idempotent; we also close it explicitly below.
	var snapshotVersion int
	err = snapshotDB.QueryRowContext(ctx, fmt.Sprintf("SELECT version FROM %s", migrationVersionTable)).Scan(&snapshotVersion)
	if err != nil {
		return fmt.Errorf("downloaded snapshot is not a valid metastore database: %w", err)
	}
	err = snapshotDB.Close()
	if err != nil {
		return fmt.Errorf("failed to close downloaded snapshot: %w", err)
	}

	// Remove journal files left behind by the database we're replacing.
	// They describe a different database and would corrupt the restored one.
	for _, suffix := range []string{"-wal", "-shm", "-journal"} {
		err := os.Remove(dbPath + suffix)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("failed to remove stale journal file: %w", err)
		}
	}

	err = os.Rename(tmpPath, dbPath)
	if err != nil {
		return fmt.Errorf("failed to move snapshot into place: %w", err)
	}

	// Log at warn level: this only happens after data loss, and the snapshot may be up to a day old.
	logger.Warn("sqlite: restored database from backup",
		zap.Time("snapshot_time", attrs.ModTime),
		zap.Int64("snapshot_size_bytes", attrs.Size),
		zap.Int("snapshot_migration_version", snapshotVersion),
	)
	return nil
}
