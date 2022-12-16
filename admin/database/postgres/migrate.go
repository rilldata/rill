package postgres

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strconv"
	"strings"
)

// Embed migrations directory in the binary
//
//go:embed migrations/*.sql
var migrationsFS embed.FS

// Fixed advisory lock number to prevent concurrent migrations.
var migrationLockNumber = int64(5103805673824918) // random number

// Name of the table that tracks migrations.
var migrationVersionTable = "admin_migration_version"

// Migrate runs migrations. It's safe for concurrent invocations.
// Adapted from: https://github.com/jackc/tern
func (c *connection) Migrate(ctx context.Context) (err error) {
	// Acquire advisory lock
	_, err = c.db.ExecContext(ctx, "select pg_advisory_lock($1)", migrationLockNumber)
	if err != nil {
		return err
	}
	defer func() {
		// Release advisory lock when this function returns
		_, unlockErr := c.db.ExecContext(ctx, "select pg_advisory_unlock($1)", migrationLockNumber)
		if err == nil && unlockErr != nil {
			err = unlockErr
		}
	}()

	// Check if migrationVersionTable exists
	var exists int
	err = c.db.QueryRowContext(ctx, "select count(*) from pg_catalog.pg_class where relname=$1 and relkind='r' and pg_table_is_visible(oid)", migrationVersionTable).Scan(&exists)
	if err != nil {
		return err
	}

	// Create migrationVersionTable if it doesn't exist
	if exists == 0 {
		_, err = c.db.ExecContext(ctx, fmt.Sprintf("create table if not exists %s(version int4 not null)", migrationVersionTable))
		if err != nil {
			return err
		}

		// Set the version to 0 if table is empty (note: defensive coding, table should always be empty)
		_, err = c.db.ExecContext(ctx, fmt.Sprintf("insert into %s(version) select 0 where 0=(select count(*) from %s)", migrationVersionTable, migrationVersionTable))
		if err != nil {
			return err
		}
	}

	// Get version of latest migration
	var currentVersion int
	err = c.db.QueryRowContext(ctx, fmt.Sprintf("select version from %s", migrationVersionTable)).Scan(&currentVersion)
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
		version, err := strconv.Atoi(strings.TrimSuffix(file.Name(), ".sql"))
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

		err = migrateSingle(ctx, c, file, sql, version)
		if err != nil {
			return err
		}
	}

	return nil
}

func migrateSingle(ctx context.Context, c *connection, file fs.DirEntry, sql []byte, version int) (err error) {
	// Start a transaction
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Run migration
	_, err = tx.ExecContext(ctx, string(sql))
	if err != nil {
		return fmt.Errorf("failed to run migration '%s': %w", file.Name(), err)
	}

	// Update migration version
	_, err = tx.ExecContext(ctx, fmt.Sprintf("UPDATE %s SET version=$1", migrationVersionTable), version)
	if err != nil {
		return err
	}

	// Commit migration
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
