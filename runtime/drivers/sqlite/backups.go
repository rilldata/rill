package sqlite

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"gocloud.dev/blob"

	// Register database drivers
	_ "github.com/marcboeker/go-duckdb/v2"
	_ "modernc.org/sqlite"
)

var (
	// Maximum size of the SQLite snapshot for backup.
	backupMaxSizeBytes int64 = 1024 * 1024 * 1024 // 1 GB

	// Max time a backup may run for.
	backupMaxDuration = 10 * time.Minute

	// Queries to backup as Parquet.
	// Each will be exported to {output_dir}/{key}.parquet.
	parquetBackupQueries = map[string]string{
		"instances":        "SELECT * EXCLUDE (variables, project_variables, feature_flags, annotations, connectors, project_connectors, public_paths) FROM instances",
		"instance_health":  "SELECT * FROM instance_health",
		"catalog":          "SELECT * EXCLUDE (data) FROM catalogv2",
		"model_partitions": "SELECT * FROM model_partitions",
		"ai_sessions":      "SELECT * FROM ai_sessions",
		"ai_messages":      "SELECT * FROM ai_messages",
	}
)

// startBackups starts a background goroutine that performs periodic backups of the SQLite file to object storage.
//
// It is also a no-op unless the following pre-requisites are in place:
// 1. An external bucket is configured on the storage client.
// 2. A backup ID is provided in the connection config (through the "id" config parameter, currently propagates from RILL_RUNTIME_METASTORE_ID).
//
// It is a best-effort backup used for analytics. There are currently no guarantees on backups and no restore functionality.
// Backups are performed at midnight UTC every day if the runtime is running at that time.
//
// Backups are stored in the external bucket under the path "shared/metastore/{backupID}/" (the "shared/metastore" prefix is applied by the code that opens the connection).
// The directory will contain a snapshot.db SQLite file and Parquet files for each of the tables defined in parquetBackupQueries.
func (c *connection) startBackups() {
	// No-op if no backup ID is provided.
	if c.backupID == "" {
		return
	}

	// No-op if no external bucket is configured.
	bucket, ok, err := c.storage.OpenBucket(c.ctx, c.backupID)
	if err != nil {
		c.logger.Error("sqlite: could not open backup bucket", zap.Error(err))
		return
	}
	if !ok {
		return
	}

	// No-op if the SQLite file is in-memory.
	ok, err = c.isInMemory(c.ctx)
	if err != nil {
		c.logger.Error("sqlite: could not check database type for backups", zap.Error(err))
		return
	}
	if ok {
		return
	}

	// Run a backup every day at midnight UTC.
	for {
		// Calculate duration until next midnight UTC.
		now := time.Now().UTC()
		midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
		delay := midnight.Sub(now)
		if delay <= 0 { // Just to be safe
			delay = 24 * time.Hour
		}

		// Wait until next midnight UTC or context cancellation.
		select {
		case <-c.ctx.Done():
			// Handle closed, exit.
			return
		case <-time.After(delay):
			// Time to perform backup.
		}

		// Perform backup.
		c.logger.Info("sqlite: backup started", zap.String("backup_id", c.backupID))
		if err := c.backup(c.ctx, bucket); err != nil {
			c.logger.Error("sqlite: backup failed", zap.String("backup_id", c.backupID), zap.Error(err))
		} else {
			c.logger.Info("sqlite: backup completed successfully", zap.String("backup_id", c.backupID))
		}
	}
}

// backup performs a backup of the SQLite file to the provided storage bucket.
// It assumes the bucket is already scoped to the correct backup directory for the current backup ID.
// See startBackups() for details.
func (c *connection) backup(ctx context.Context, bucket *blob.Bucket) error {
	// Set a timeout for the entire backup operation.
	ctx, cancel := context.WithTimeout(ctx, backupMaxDuration)
	defer cancel()

	// Setup a temporary directory for intermediate files
	tmpDir, err := os.MkdirTemp("", "sqlite-backup-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Capture a snapshot of the SQLite database
	snapshotPath := filepath.Join(tmpDir, "snapshot.db")
	_, err = c.db.ExecContext(ctx, fmt.Sprintf("VACUUM INTO '%s'", snapshotPath))
	if err != nil {
		return fmt.Errorf("failed to create SQLite snapshot: %w", err)
	}

	// Check snapshot has a reasonable size
	info, err := os.Stat(snapshotPath)
	if err != nil {
		return fmt.Errorf("failed to stat SQLite snapshot: %w", err)
	}
	if info.Size() >= backupMaxSizeBytes {
		return fmt.Errorf("SQLite snapshot size is too small: %d bytes", info.Size())
	}

	// Upload the snapshot file itself for safekeeping.
	f, err := os.Open(snapshotPath)
	if err != nil {
		return fmt.Errorf("failed to open SQLite snapshot for upload: %w", err)
	}
	defer f.Close()
	err = bucket.Upload(ctx, "snapshot.db", f, &blob.WriterOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return fmt.Errorf("failed to upload SQLite file: %w", err)
	}

	// Open an in-memory DuckDB handle with 1 CPU and 256MB memory limit.
	// We'll use this to create Parquet files for the tables in the SQLite database.
	duckdb, err := sqlx.Open("duckdb", "?threads=1&memory_limit=256MB")
	if err != nil {
		return fmt.Errorf("failed to open DuckDB: %w", err)
	}
	duckdb.SetMaxOpenConns(1)
	defer duckdb.Close()

	// Attach the SQLite database to DuckDB and export tables to Parquet.
	_, err = duckdb.ExecContext(ctx, fmt.Sprintf("ATTACH '%s' AS sqlite_db (TYPE SQLITE); USE sqlite_db;", snapshotPath))
	if err != nil {
		return fmt.Errorf("failed to attach SQLite database to DuckDB: %w", err)
	}
	for name, query := range parquetBackupQueries {
		// Use a lamda so we can use defer for cleanup inside the loop.
		err := func() error {
			// Export name to Parquet
			exportPath := filepath.Join(tmpDir, fmt.Sprintf("%s.parquet", name))
			_, err = duckdb.ExecContext(ctx, fmt.Sprintf("COPY (%s) TO '%s' (FORMAT PARQUET)", query, exportPath))
			if err != nil {
				return fmt.Errorf("failed to export query %q: %w", name, err)
			}
			defer os.Remove(exportPath)

			// Open the Parquet file for reading.
			f, err := os.Open(exportPath)
			if err != nil {
				return fmt.Errorf("failed to open Parquet file for name %s: %w", name, err)
			}
			defer f.Close()

			// Upload the Parquet files to the storage bucket.
			blobKey := fmt.Sprintf("%s.parquet", name)
			err = bucket.Upload(ctx, blobKey, f, &blob.WriterOptions{
				ContentType: "application/octet-stream",
			})
			if err != nil {
				return fmt.Errorf("failed to upload Parquet file for name %s: %w", name, err)
			}
			return nil
		}()
		if err != nil {
			return err
		}
	}

	return nil
}

// isInMemory checks if the SQLite database is in-memory.
func (c *connection) isInMemory(ctx context.Context) (bool, error) {
	var seq int
	var name, file string
	row := c.db.QueryRowContext(ctx, "PRAGMA database_list;")
	err := row.Scan(&seq, &name, &file)
	if err != nil {
		return false, err
	}
	if file == "" || file == ":memory:" || file == "file::memory:" {
		return true, nil
	}
	return false, nil
}
