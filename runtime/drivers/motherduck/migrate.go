package motherduck

import (
	"context"
	"embed"
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

// Embed migrations directory in the binary
//
//go:embed migrations/*.sql
var migrationsFS embed.FS

// Name of the table that tracks migrations.
var migrationVersionTable = "rill.migration_version"

// Migrate implements drivers.Connection.
// Migrate for DuckDB may not be safe for concurrent use.
func (c *connection) Migrate(ctx context.Context) (err error) {
	conn, release, err := c.acquireMetaConn(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = release() }()

	// Create rill schema if it doesn't exist
	_, err = conn.ExecContext(ctx, "create schema if not exists rill")
	if err != nil {
		return err
	}

	// Create migrationVersionTable if it doesn't exist
	_, err = conn.ExecContext(ctx, fmt.Sprintf("create table if not exists %s(version integer not null)", migrationVersionTable))
	if err != nil {
		return err
	}

	// Set the version to 0 if table is empty
	_, err = conn.ExecContext(ctx, fmt.Sprintf("insert into %s(version) select 0 where 0=(select count(*) from %s)", migrationVersionTable, migrationVersionTable))
	if err != nil {
		return err
	}

	// Get version of latest migration
	var currentVersion int
	err = conn.QueryRowContext(ctx, fmt.Sprintf("select version from %s", migrationVersionTable)).Scan(&currentVersion)
	if err != nil {
		return err
	}

	// Iterate over migrations (sorted by filename)
	files, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return err
	}

	for _, file := range files {
		// Extract version number from filename
		version, err := migrationFilenameToVersion(file.Name())
		if err != nil {
			return fmt.Errorf("unexpected migration filename: %s", file.Name())
		}

		// Skip migrations below current version
		if version <= currentVersion {
			continue
		}

		// Read SQL
		sql, err := migrationsFS.ReadFile(path.Join("migrations", file.Name()))
		if err != nil {
			return err
		}

		err = c.migrateSingle(ctx, conn, file.Name(), sql, version)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *connection) migrateSingle(ctx context.Context, conn *sqlx.Conn, name string, sql []byte, version int) (err error) {
	// Start a transaction
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Run migration
	_, err = tx.ExecContext(ctx, string(sql))
	if err != nil {
		return fmt.Errorf("failed to run migration '%s': %w", name, err)
	}

	// Update migration version
	_, err = tx.ExecContext(ctx, fmt.Sprintf("UPDATE %s SET version=?", migrationVersionTable), version)
	if err != nil {
		return err
	}

	// Commit migration
	err = tx.Commit()
	if err != nil {
		return err
	}

	// Force DuckDB to merge WAL into .db file
	_, err = conn.ExecContext(ctx, "CHECKPOINT;")
	if err != nil {
		return err
	}
	return nil
}

// MigrationStatus implements drivers.Connection.
func (c *connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	conn, release, err := c.acquireMetaConn(ctx)
	if err != nil {
		return 0, 0, err
	}
	defer func() { _ = release() }()

	// Get current version
	err = conn.QueryRowxContext(ctx, fmt.Sprintf("select version from %s", migrationVersionTable)).Scan(&current)
	if err != nil {
		return 0, 0, err
	}

	// Set desired to version number of last migration file
	files, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return 0, 0, err
	}
	if len(files) > 0 {
		file := files[len(files)-1]
		version, err := migrationFilenameToVersion(file.Name())
		if err != nil {
			return 0, 0, fmt.Errorf("unexpected migration filename: %s", file.Name())
		}
		desired = version
	}

	return current, desired, nil
}

func migrationFilenameToVersion(name string) (int, error) {
	return strconv.Atoi(strings.TrimSuffix(name, ".sql"))
}
