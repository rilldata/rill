package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
	"gocloud.dev/blob"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	// Register database drivers
	_ "github.com/duckdb/duckdb-go/v2"
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
		// Table `instances`.
		// It excludes the JSON columns due to a type mess up: the columns are TEXT, but we've been saving BLOB values to them.
		// SQLite weirdly allows this, but DuckDB chokes on it.
		"instances": "SELECT * EXCLUDE (variables, project_variables, feature_flags, annotations, connectors, project_connectors, public_paths) FROM instances",
		// Table `instance_health`
		"instance_health": "SELECT * FROM instance_health",
		// Table `catalogv2` (NOTE: the `data` column has already been converted to JSON in rewriteSnapshotForAnalytics below).
		"catalog": "SELECT * FROM catalogv2",
		// Table `model_partitions`
		"model_partitions": "SELECT * FROM model_partitions",
		// Table `ai_sessions`
		"ai_sessions": "SELECT * FROM ai_sessions",
		// Table `ai_messages`
		"ai_messages": "SELECT * FROM ai_messages",
	}
)

// startBackups starts a background goroutine that performs periodic backups of the SQLite file to object storage.
//
// It is a no-op unless the following pre-requisites are in place:
// 1. An external bucket is configured on the storage client.
// 2. A backup ID is provided in the connection config (through the "id" config parameter, currently propagates from RILL_RUNTIME_METASTORE_ID).
// 3. The SQLite database is file-based and doesn't exceed backupMaxSizeBytes in size.
//
// It is a best-effort backup used for analytics. There are currently no guarantees on backups and no restore functionality.
// Backups are performed at midnight UTC every day if the runtime is running at that time.
//
// Backups are stored in the external bucket under the path "shared/metastore/{backupID}/" (the "shared/metastore" prefix is not applied here, but where the connection is opened).
// The directory will contain a snapshot.db SQLite file and Parquet files for each of the tables defined in parquetBackupQueries.
func (c *connection) startBackups() {
	// It's a no-op if no backup ID is provided.
	if c.backupID == "" {
		return
	}

	// Open bucket scoped to the backup directory.
	// Return early (no-op) if a bucket isn't available.
	bucket, ok, err := c.storage.OpenBucket(c.ctx, c.backupID)
	if err != nil {
		c.logger.Error("sqlite: could not open backup bucket", zap.Error(err), zap.String("backup_id", c.backupID))
		return
	}
	if !ok {
		return
	}

	// Exit early if the database is in-memory.
	dbPath, err := c.dbFilePath(c.ctx)
	if err != nil {
		c.logger.Error("sqlite: could not find database file path", zap.Error(err), zap.String("backup_id", c.backupID))
		return
	}
	if dbPath == "" {
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
		err := c.backup(c.ctx, bucket)
		if err != nil {
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

	// Check if the database file is too large.
	dbPath, err := c.dbFilePath(c.ctx)
	if err != nil {
		return fmt.Errorf("could not get database file path: %w", err)
	}
	info, err := os.Stat(dbPath)
	if err != nil {
		return fmt.Errorf("failed to stat SQLite snapshot: %w", err)
	}
	if info.Size() > backupMaxSizeBytes {
		return fmt.Errorf("SQLite snapshot size is too big: %d bytes", info.Size())
	}

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

	// Rewrite the snapshot in preparation for Parquet exports.
	// NOTE: We do this after uploading the snapshot.db to the bucket to ensure the backup file has the original data.
	err = c.rewriteSnapshotForAnalytics(ctx, snapshotPath)
	if err != nil {
		return fmt.Errorf("failed to rewrite snapshot for analytics: %w", err)
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
		// Use a lambda so we can use defer for cleanup inside the loop.
		err := func() error {
			// Export name to Parquet
			fname := fmt.Sprintf("%s.parquet", name)
			exportPath := filepath.Join(tmpDir, fname)
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
			err = bucket.Upload(ctx, fname, f, &blob.WriterOptions{
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

// rewriteSnapshotForAnalytics rewrites the snapshot database to prepare it for analytics exports to Parquet.
// This is done after the snapshot.db file has been backed up, so it does not affect the backup file itself.
// Specifically, we do this to convert catalog resources from protobuf to JSON format to make them easy to query in downstream analytics.
func (c *connection) rewriteSnapshotForAnalytics(ctx context.Context, snapshotPath string) error {
	// Open the snapshot database.
	snapshotDB, err := sqlx.Open("sqlite", snapshotPath)
	if err != nil {
		return fmt.Errorf("failed to open snapshot database: %w", err)
	}
	defer snapshotDB.Close()

	// Convert the `data` field in each row in the `catalogv2` table from protobuf to JSON.
	for offset := range 50_000 { // Failsafe to avoid infinite loops; we don't anticipate this many resources.
		// Read one resource.
		var r drivers.Resource
		err := snapshotDB.QueryRowContext(ctx, "SELECT kind, name, data FROM catalogv2 ORDER BY kind, name LIMIT 1 OFFSET ?", offset).Scan(&r.Kind, &r.Name, &r.Data)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				break
			}
			return fmt.Errorf("failed to query catalog resource: %w", err)
		}

		// Convert data from protobuf message rill.runtime.v1.Resource to JSON.
		pb := &runtimev1.Resource{}
		err = proto.Unmarshal(r.Data, pb)
		if err != nil {
			return fmt.Errorf("failed to unmarshal catalog resource protobuf: %w", err)
		}
		dataJSON, err := protojson.Marshal(pb)
		if err != nil {
			return fmt.Errorf("failed to marshal catalog resource to JSON: %w", err)
		}

		// Write the update back to the snapshot database.
		_, err = snapshotDB.ExecContext(ctx, "UPDATE catalogv2 SET data = ? WHERE kind = ? AND name = ?", dataJSON, r.Kind, r.Name)
		if err != nil {
			return fmt.Errorf("failed to update catalog resource with JSON data: %w", err)
		}
	}

	return nil
}

// dbFilePath gets the file path of the SQLite database.
// It returns the empty string if the database is in-memory.
func (c *connection) dbFilePath(ctx context.Context) (string, error) {
	var file string
	err := c.db.QueryRowContext(ctx, `SELECT file FROM pragma_database_list WHERE name = 'main';`).Scan(&file)
	if err != nil {
		return "", err
	}
	if file == "" || file == ":memory:" || file == "file::memory:" {
		return "", nil
	}
	return file, nil
}
