package rduckdb

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/marcboeker/go-duckdb"
	"github.com/mitchellh/mapstructure"
	"go.opentelemetry.io/otel/attribute"
	"gocloud.dev/blob"
)

type DB interface {
	// Close closes the database.
	Close() error

	// AcquireReadConnection returns a connection to the database for reading.
	// Once done the connection should be released by calling the release function.
	// This connection must only be used for select queries or for creating and working with temporary tables.
	AcquireReadConnection(ctx context.Context) (conn Conn, release func() error, err error)

	// AcquireWriteConnection returns a connection to the database for writing.
	// Once done the connection should be released by calling the release function.
	// Any persistent changes to the database should be done by calling CRUD APIs on this connection.
	AcquireWriteConnection(ctx context.Context) (conn Conn, release func() error, err error)

	// Size returns the size of the database in bytes.
	// It is currently implemented as sum of the size of all serving `.db` files.
	Size() int64

	// CRUD APIs

	// CreateTableAsSelect creates a new table by name from the results of the given SQL query.
	CreateTableAsSelect(ctx context.Context, name string, sql string, opts *CreateTableOptions) error

	// InsertTableAsSelect inserts the results of the given SQL query into the table.
	InsertTableAsSelect(ctx context.Context, name string, sql string, opts *InsertTableOptions) error

	// DropTable removes a table from the database.
	DropTable(ctx context.Context, name string) error

	// RenameTable renames a table in the database.
	RenameTable(ctx context.Context, oldName, newName string) error

	// AddTableColumn adds a column to the table.
	AddTableColumn(ctx context.Context, tableName, columnName, typ string) error

	// AlterTableColumn alters the type of a column in the table.
	AlterTableColumn(ctx context.Context, tableName, columnName, newType string) error
}

type DBOptions struct {
	// Clean specifies whether to start with a clean database or download data from cloud storage and start with backed up data.
	Clean bool
	// LocalPath is the path where local db files will be stored. Should be unique for each database.
	LocalPath string

	Backup *blob.Bucket

	// ReadSettings are settings applied the read duckDB handle.
	ReadSettings map[string]string
	// WriteSettings are settings applied the write duckDB handle.
	WriteSettings map[string]string
	// InitQueries are the queries to run when the database is first created.
	InitQueries []string

	Logger         *slog.Logger
	OtelAttributes []attribute.KeyValue
}

// TODO :: revisit this logic
func (d *DBOptions) ValidateSettings() error {
	read := &settings{}
	err := mapstructure.Decode(d.ReadSettings, read)
	if err != nil {
		return fmt.Errorf("read settings: %w", err)
	}

	write := &settings{}
	err = mapstructure.Decode(d.WriteSettings, write)
	if err != nil {
		return fmt.Errorf("write settings: %w", err)
	}

	// no memory limits defined
	// divide memory equally between read and write
	if read.MaxMemory == "" && write.MaxMemory == "" {
		connector, err := duckdb.NewConnector("", nil)
		if err != nil {
			return fmt.Errorf("unable to create duckdb connector: %w", err)
		}
		defer connector.Close()
		db := sql.OpenDB(connector)
		defer db.Close()

		row := db.QueryRow("SELECT value FROM duckdb_settings() WHERE name = 'max_memory'")
		var maxMemory string
		err = row.Scan(&maxMemory)
		if err != nil {
			return fmt.Errorf("unable to get max_memory: %w", err)
		}

		bytes, err := humanReadableSizeToBytes(maxMemory)
		if err != nil {
			return fmt.Errorf("unable to parse max_memory: %w", err)
		}

		read.MaxMemory = fmt.Sprintf("%d bytes", int64(bytes)/2)
		write.MaxMemory = fmt.Sprintf("%d bytes", int64(bytes)/2)
	}

	if read.MaxMemory == "" != (write.MaxMemory == "") {
		// only one is defined
		var mem string
		if read.MaxMemory != "" {
			mem = read.MaxMemory
		} else {
			mem = write.MaxMemory
		}

		bytes, err := humanReadableSizeToBytes(mem)
		if err != nil {
			return fmt.Errorf("unable to parse max_memory: %w", err)
		}

		read.MaxMemory = fmt.Sprintf("%d bytes", int64(bytes)/2)
		write.MaxMemory = fmt.Sprintf("%d bytes", int64(bytes)/2)
	}

	var readThread, writeThread int
	if read.Threads != "" {
		readThread, err = strconv.Atoi(read.Threads)
		if err != nil {
			return fmt.Errorf("unable to parse read threads: %w", err)
		}
	}
	if write.Threads != "" {
		writeThread, err = strconv.Atoi(write.Threads)
		if err != nil {
			return fmt.Errorf("unable to parse write threads: %w", err)
		}
	}

	if readThread == 0 != (writeThread == 0) {
		// only one is defined
		var threads int
		if readThread != 0 {
			threads = readThread
		} else {
			threads = writeThread
		}

		read.Threads = strconv.Itoa(threads)
		write.Threads = strconv.Itoa(threads)
	}

	err = mapstructure.WeakDecode(read, &d.ReadSettings)
	if err != nil {
		return fmt.Errorf("failed to update read settings: %w", err)
	}

	err = mapstructure.WeakDecode(write, &d.WriteSettings)
	if err != nil {
		return fmt.Errorf("failed to update write settings: %w", err)
	}
	return nil
}

type CreateTableOptions struct {
	// View specifies whether the created table is a view.
	View bool
}

type IncrementalStrategy string

const (
	IncrementalStrategyUnspecified IncrementalStrategy = ""
	IncrementalStrategyAppend      IncrementalStrategy = "append"
	IncrementalStrategyMerge       IncrementalStrategy = "merge"
)

type InsertTableOptions struct {
	ByName    bool
	Strategy  IncrementalStrategy
	UniqueKey []string
}

// NewDB creates a new DB instance.
// This can be a slow operation if the backup is large.
// dbIdentifier is a unique identifier for the database reported in metrics.
func NewDB(ctx context.Context, opts *DBOptions) (DB, error) {
	err := opts.ValidateSettings()
	if err != nil {
		return nil, err
	}
	db := &db{
		opts:       opts,
		localPath:  opts.LocalPath,
		writeDirty: true,
		logger:     opts.Logger,
	}
	if opts.Backup != nil {
		db.backup = opts.Backup
	}
	// if clean is true, remove the backup
	if opts.Clean {
		err = db.deleteBackup(ctx, "", "")
		if err != nil {
			return nil, fmt.Errorf("unable to clean backup: %w", err)
		}
	}

	// create local path
	err = os.MkdirAll(db.localPath, fs.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("unable to create read path: %w", err)
	}

	// sync local data
	err = db.syncLocalWithBackup(ctx)
	if err != nil {
		return nil, err
	}

	// create read handle
	db.readHandle, err = db.openDBAndAttach(ctx, true, "")
	if err != nil {
		if strings.Contains(err.Error(), "Symbol not found") {
			fmt.Printf("Your version of macOS is not supported. Please upgrade to the latest major release of macOS. See this link for details: https://support.apple.com/en-in/macos/upgrade")
			os.Exit(1)
		}
		return nil, err
	}

	return db, nil
}

type db struct {
	opts *DBOptions

	localPath    string
	readHandle   *sqlx.DB
	readMu       sync.RWMutex
	writeMu      sync.Mutex
	writeDirty   bool
	inconsistent bool

	backup *blob.Bucket

	logger *slog.Logger
}

var _ DB = &db{}

func (d *db) Close() error {
	d.writeMu.Lock()
	defer d.writeMu.Unlock()

	d.readMu.Lock()
	defer d.readMu.Unlock()

	return d.readHandle.Close()
}

func (d *db) AcquireReadConnection(ctx context.Context) (Conn, func() error, error) {
	d.readMu.RLock()

	c, err := d.readHandle.Connx(ctx)
	if err != nil {
		d.readMu.RUnlock()
		return nil, nil, err
	}

	release := func() error {
		err := c.Close()
		d.readMu.RUnlock()
		return err
	}
	conn := &conn{
		Conn: c,
		db:   d,
	}
	return conn, release, nil
}

func (d *db) AcquireWriteConnection(ctx context.Context) (Conn, func() error, error) {
	d.writeMu.Lock()
	defer d.writeMu.Unlock()
	c, release, err := d.acquireWriteConn(ctx)
	if err != nil {
		return nil, nil, err
	}

	return &conn{
		Conn: c,
		db:   d,
	}, release, nil
}

func (d *db) CreateTableAsSelect(ctx context.Context, name, query string, opts *CreateTableOptions) error {
	d.logger.Debug("create table", slog.String("name", name), slog.Bool("view", opts.View))
	d.writeMu.Lock()
	defer d.writeMu.Unlock()
	conn, release, err := d.acquireWriteConn(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = release()
	}()
	return d.createTableAsSelect(ctx, conn, release, name, query, opts)
}

func (d *db) createTableAsSelect(ctx context.Context, conn *sqlx.Conn, releaseConn func() error, name, query string, opts *CreateTableOptions) error {
	// check if some older version exists
	oldVersion, oldVersionExists, _ := tableVersion(d.localPath, name)
	d.logger.Debug("old version", slog.String("version", oldVersion), slog.Bool("exists", oldVersionExists))

	// create new version directory
	newVersion := newVersion()
	newVersionDir := filepath.Join(d.localPath, name, newVersion)
	err := os.MkdirAll(newVersionDir, fs.ModePerm)
	if err != nil {
		return fmt.Errorf("create: unable to create dir %q: %w", name, err)
	}

	var m *meta
	if opts.View {
		// create view - validates that SQL is correct
		_, err = conn.ExecContext(ctx, fmt.Sprintf("CREATE OR REPLACE VIEW %s AS (%s\n)", safeSQLName(name), query))
		if err != nil {
			return err
		}

		m = &meta{ViewSQL: query}
	} else {
		// create db file
		dbFile := filepath.Join(newVersionDir, "data.db")
		safeDBName := safeSQLName(dbName(name))

		// detach existing db
		_, err = conn.ExecContext(ctx, fmt.Sprintf("DETACH DATABASE IF EXISTS %s", safeDBName), nil)
		if err != nil {
			_ = os.RemoveAll(newVersionDir)
			return fmt.Errorf("create: detach %q db failed: %w", safeDBName, err)
		}

		// attach new db
		_, err = conn.ExecContext(ctx, fmt.Sprintf("ATTACH %s AS %s", safeSQLString(dbFile), safeDBName), nil)
		if err != nil {
			_ = os.RemoveAll(newVersionDir)
			return fmt.Errorf("create: attach %q db failed: %w", dbFile, err)
		}

		// ingest data
		_, err = conn.ExecContext(ctx, fmt.Sprintf("CREATE OR REPLACE TABLE %s.default AS (%s\n)", safeDBName, query), nil)
		if err != nil {
			_ = os.RemoveAll(newVersionDir)
			return fmt.Errorf("create: create %q.default table failed: %w", safeDBName, err)
		}
	}

	// write meta
	err = writeMeta(newVersionDir, m)
	if err != nil {
		_ = os.RemoveAll(newVersionDir)
		return err
	}

	// update version.txt
	d.writeDirty = true
	err = d.setTableVersion(name, newVersion)
	if err != nil {
		_ = os.RemoveAll(newVersionDir)
		return fmt.Errorf("create: write version file failed: %w", err)
	}

	// close write handle before syncing read so that temp files or wal files if any are removed
	err = releaseConn()
	if err != nil {
		return err
	}

	if err := d.syncBackupWithLocal(ctx, name, oldVersion); err != nil {
		return fmt.Errorf("create: replicate failed: %w", err)
	}
	d.logger.Debug("table created", slog.String("name", name))
	// backup and local are now in sync
	d.writeDirty = false
	if oldVersionExists {
		_ = d.deleteLocalTable(name, oldVersion)
	}

	return d.reopen("")
}

func (d *db) InsertTableAsSelect(ctx context.Context, name, query string, opts *InsertTableOptions) error {
	d.logger.Debug("insert table", slog.String("name", name), slog.Group("option", "by_name", opts.ByName, "strategy", string(opts.Strategy), "unique_key", opts.UniqueKey))
	d.writeMu.Lock()
	defer d.writeMu.Unlock()
	conn, release, err := d.acquireWriteConn(ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = release()
	}()
	return d.insertTableAsSelect(ctx, conn, release, name, query, opts)
}

func (d *db) insertTableAsSelect(ctx context.Context, conn *sqlx.Conn, releaseConn func() error, name, query string, opts *InsertTableOptions) error {
	// Get current table version
	oldVersion, oldVersionExists, err := tableVersion(d.localPath, name)
	if err != nil || !oldVersionExists {
		return fmt.Errorf("table %q does not exist", name)
	}

	d.writeDirty = true
	// Execute the insert
	err = execIncrementalInsert(ctx, conn, fmt.Sprintf("%s.default", safeSQLName(dbName(name))), query, opts)
	if err != nil {
		return fmt.Errorf("insert: insert into table %q failed: %w", name, err)
	}

	// rename db directory
	newVersion := newVersion()
	oldVersionDir := filepath.Join(d.localPath, name, oldVersion)
	err = os.Rename(oldVersionDir, filepath.Join(d.localPath, name, newVersion))
	if err != nil {
		return fmt.Errorf("insert: update version %q failed: %w", newVersion, err)
	}

	// update version.txt
	err = os.WriteFile(filepath.Join(d.localPath, name, "version.txt"), []byte(newVersion), fs.ModePerm)
	if err != nil {
		return fmt.Errorf("insert: write version file failed: %w", err)
	}

	err = releaseConn()
	if err != nil {
		return err
	}
	// replicate
	err = d.syncBackupWithLocal(ctx, name, oldVersion)
	if err != nil {
		return fmt.Errorf("insert: replicate failed: %w", err)
	}
	// both backups and write are now in sync
	d.writeDirty = false

	// Delete the old version (ignoring errors since source the new data has already been correctly inserted)
	_ = os.RemoveAll(oldVersionDir)
	return d.reopen("")
}

// DropTable implements DB.
func (d *db) DropTable(ctx context.Context, name string) error {
	d.logger.Debug("drop table", slog.String("name", name))
	d.writeMu.Lock()
	defer d.writeMu.Unlock()
	_, release, err := d.acquireWriteConn(ctx) // we don't need the handle but need to sync the write
	if err != nil {
		return err
	}
	defer func() {
		_ = release()
	}()

	return d.dropTable(ctx, name)
}

func (d *db) dropTable(ctx context.Context, name string) error {
	_, exist, _ := tableVersion(d.localPath, name)
	if !exist {
		return fmt.Errorf("drop: table %q not found", name)
	}

	d.writeDirty = true

	// drop the table from backup location
	err := d.deleteBackup(ctx, name, "")
	if err != nil {
		return fmt.Errorf("drop: unable to drop table %q from backup: %w", name, err)
	}
	d.writeDirty = false

	// reopen db handle
	err = d.reopen(name)
	if err != nil {
		return fmt.Errorf("drop: unable to reopen: %w", err)
	}

	err = d.deleteLocalTable(name, "")
	if err != nil {
		d.logger.Debug("drop: unable to delete local table data", slog.String("table", name), slog.String("error", err.Error()))
	}
	return nil
}

func (d *db) RenameTable(ctx context.Context, oldName, newName string) error {
	d.logger.Debug("rename table", slog.String("from", oldName), slog.String("to", newName))
	if strings.EqualFold(oldName, newName) {
		return fmt.Errorf("rename: Table with name %q already exists", newName)
	}
	d.writeMu.Lock()
	defer d.writeMu.Unlock()
	_, release, err := d.acquireWriteConn(ctx) // we don't need the handle but need to sync the write
	if err != nil {
		return err
	}
	defer func() {
		_ = release()
	}()
	return d.renameTable(ctx, oldName, newName)
}

func (d *db) renameTable(ctx context.Context, oldName, newName string) error {
	oldVersion, exist, err := d.tableVersion(oldName, false)
	if err != nil {
		return err
	}
	if !exist {
		return fmt.Errorf("rename: Table %q not found", oldName)
	}

	oldVersionInNewTable, replaceInNewTable, _ := d.tableVersion(newName, false)

	d.writeDirty = true
	// copy the old table version to new table version
	version := newVersion()
	err = copyDir(filepath.Join(d.localPath, newName, version), filepath.Join(d.localPath, oldName, oldVersion))
	if err != nil {
		return fmt.Errorf("rename: copy table failed: %w", err)
	}

	// update version.txt
	err = d.setTableVersion(newName, version)
	if err != nil {
		return fmt.Errorf("rename: write version file failed: %w", err)
	}

	// sync the new table and new version
	if err := d.syncBackupWithLocal(ctx, newName, oldVersionInNewTable); err != nil {
		return fmt.Errorf("rename: unable to replicate new table: %w", err)
	}

	// drop the old table in backup
	err = d.deleteBackup(ctx, oldName, "")
	if err != nil {
		// at this point both is inconsistent
		// has both old table and new table
		return fmt.Errorf("rename: unable to delete old table %q from backup: %w", oldName, err)
	}

	// reopen db handle ignoring old name
	err = d.reopen(oldName)
	if err != nil {
		return fmt.Errorf("rename: unable to reopen: %w", err)
	}

	d.inconsistent = false

	if replaceInNewTable {
		_ = d.deleteLocalTable(newName, oldVersionInNewTable)
	}

	// delete old table from local
	err = d.deleteLocalTable(oldName, "")
	if err != nil {
		d.logger.Debug("rename: unable to delete old table", slog.String("table", oldName), slog.String("error", err.Error()))
	}

	d.writeDirty = false
	return nil
}

func (d *db) AddTableColumn(ctx context.Context, tableName, columnName, typ string) error {
	d.logger.Debug("AddTableColumn", slog.String("table", tableName), slog.String("column", columnName), slog.String("typ", typ))
	d.writeMu.Lock()
	defer d.writeMu.Unlock()
	conn, release, err := d.acquireWriteConn(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = release()
	}()

	return d.addTableColumn(ctx, conn, release, tableName, columnName, typ)
}

func (d *db) addTableColumn(ctx context.Context, conn *sqlx.Conn, releaseConn func() error, tableName, columnName, typ string) error {
	oldVersion, exist, err := tableVersion(d.localPath, tableName)
	if err != nil {
		return err
	}

	if !exist {
		return fmt.Errorf("table %q does not exist", tableName)
	}

	newVersion := newVersion()
	err = copyDir(filepath.Join(d.localPath, tableName, newVersion), filepath.Join(d.localPath, tableName, oldVersion))
	if err != nil {
		return err
	}

	// detach old db
	_, err = conn.ExecContext(ctx, fmt.Sprintf("DETACH DATABASE %s", safeSQLName(dbName(tableName))))
	if err != nil {
		return err
	}

	// reattach new db
	_, err = conn.ExecContext(ctx, fmt.Sprintf("ATTACH %s AS %s", safeSQLString(filepath.Join(d.localPath, tableName, newVersion, "data.db")), safeSQLName(dbName(tableName))))
	if err != nil {
		return err
	}

	_, err = conn.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s.default ADD COLUMN %s %s", safeSQLName(dbName(tableName)), safeSQLName(columnName), typ))
	if err != nil {
		return err
	}

	// update version.txt
	d.writeDirty = true
	err = d.setTableVersion(tableName, newVersion)
	if err != nil {
		return err
	}

	err = releaseConn()
	if err != nil {
		return err
	}

	// replicate
	err = d.syncBackupWithLocal(ctx, tableName, oldVersion)
	if err != nil {
		return err
	}
	d.writeDirty = false

	// remove old local version
	_ = d.deleteLocalTable(tableName, oldVersion)

	return d.reopen("")
}

// AlterTableColumn implements drivers.OLAPStore.
func (d *db) AlterTableColumn(ctx context.Context, tableName, columnName, newType string) error {
	d.logger.Debug("AlterTableColumn", slog.String("table", tableName), slog.String("column", columnName), slog.String("typ", newType))
	d.writeMu.Lock()
	defer d.writeMu.Unlock()
	conn, release, err := d.acquireWriteConn(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = release()
	}()

	return d.alterTableColumn(ctx, conn, release, tableName, columnName, newType)
}

func (d *db) alterTableColumn(ctx context.Context, conn *sqlx.Conn, releaseConn func() error, tableName, columnName, newType string) error {
	oldVersion, exist, err := tableVersion(d.localPath, tableName)
	if err != nil {
		return err
	}

	if !exist {
		return fmt.Errorf("table %q does not exist", tableName)
	}

	newVersion := newVersion()
	err = copyDir(filepath.Join(d.localPath, tableName, newVersion), filepath.Join(d.localPath, tableName, oldVersion))
	if err != nil {
		return err
	}

	// detach old db
	_, err = conn.ExecContext(ctx, fmt.Sprintf("DETACH DATABASE %s", safeSQLName(dbName(tableName))))
	if err != nil {
		return err
	}

	// reattach new db
	_, err = conn.ExecContext(ctx, fmt.Sprintf("ATTACH %s AS %s", safeSQLString(filepath.Join(d.localPath, tableName, newVersion, "data.db")), safeSQLName(dbName(tableName))))
	if err != nil {
		return err
	}

	_, err = conn.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s.default ALTER %s TYPE %s", safeSQLName(dbName(tableName)), safeSQLName(columnName), newType))
	if err != nil {
		return err
	}

	// update version.txt
	d.writeDirty = true
	err = d.setTableVersion(tableName, newVersion)
	if err != nil {
		return err
	}

	err = releaseConn()
	if err != nil {
		return err
	}

	// replicate
	err = d.syncBackupWithLocal(ctx, tableName, oldVersion)
	if err != nil {
		return err
	}
	d.writeDirty = false

	// remove old local version
	_ = d.deleteLocalTable(tableName, oldVersion)

	return d.reopen("")
}

func (d *db) reopen(deletedTable string) error {
	// reopen should ignore context cancellations since cancellation errors can leave read inconsistent from write
	// Also it is expected to be a fast operation so should be okay to ignore context cancellations
	// extensions are already downloaded in NewDB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	handle, err := d.openDBAndAttach(ctx, true, deletedTable)
	if err != nil {
		return err
	}

	var oldDBHandle *sqlx.DB
	d.readMu.Lock()
	// swap read handle
	oldDBHandle = d.readHandle
	d.readHandle = handle
	d.readMu.Unlock()

	// close old read handle
	if oldDBHandle != nil {
		err = oldDBHandle.Close()
		if err != nil {
			d.logger.Warn("error in closing old read handle", slog.String("error", err.Error()))
		}
	}
	return nil
}

func (d *db) Size() int64 {
	var paths []string
	entries, err := os.ReadDir(d.localPath)
	if err != nil { // ignore error
		return 0
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// this is to avoid counting temp tables during source ingestion
		// in certain cases we only want to compute the size of the serving db files
		// TODO :: remove this when removing staged table concepts
		if strings.HasPrefix(entry.Name(), "__rill_tmp_") {
			continue
		}
		path := filepath.Join(d.localPath, entry.Name())
		version, exist, _ := d.tableVersion(entry.Name(), true)
		if !exist {
			continue
		}
		paths = append(paths, filepath.Join(path, fmt.Sprintf("%s.db", version)))
	}
	return fileSize(paths)
}

// acquireWriteConn syncs the write database, initializes the write handle and returns a write connection.
// The release function should be called to release the connection.
// It should be called with the writeMu locked.
func (d *db) acquireWriteConn(ctx context.Context) (*sqlx.Conn, func() error, error) {
	err := d.syncLocalWithBackup(ctx)
	if err != nil {
		return nil, nil, err
	}

	db, err := d.openDBAndAttach(ctx, false, "")
	if err != nil {
		return nil, nil, err
	}
	conn, err := db.Connx(ctx)
	if err != nil {
		_ = db.Close()
		return nil, nil, err
	}
	return conn, func() error {
		_ = conn.Close()
		err = db.Close()
		return err
	}, nil
}

func (d *db) openDBAndAttach(ctx context.Context, read bool, ignoreTable string) (*sqlx.DB, error) {
	// open the db
	var settings map[string]string
	dsn, err := url.Parse("") // in-memory
	if err != nil {
		return nil, err
	}
	if read {
		settings = d.opts.ReadSettings
	} else {
		settings = d.opts.WriteSettings
	}
	query := dsn.Query()
	for k, v := range settings {
		query.Set(k, v)
	}
	dsn.RawQuery = query.Encode()
	connector, err := duckdb.NewConnector(dsn.String(), func(execer driver.ExecerContext) error {
		for _, qry := range d.opts.InitQueries {
			_, err := execer.ExecContext(context.Background(), qry, nil)
			if err != nil && strings.Contains(err.Error(), "Failed to download extension") {
				// Retry using another mirror. Based on: https://github.com/duckdb/duckdb/issues/9378
				_, err = execer.ExecContext(context.Background(), qry+" FROM 'http://nightly-extensions.duckdb.org'", nil)
			}
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	db := sqlx.NewDb(otelsql.OpenDB(connector), "duckdb")

	err = otelsql.RegisterDBStatsMetrics(db.DB, otelsql.WithAttributes(d.opts.OtelAttributes...))
	if err != nil {
		return nil, fmt.Errorf("registering db stats metrics: %w", err)
	}

	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	err = d.attachDBs(ctx, db, read, ignoreTable)
	if err != nil {
		db.Close()
		return nil, err
	}

	// 2023-12-11: Hail mary for solving this issue: https://github.com/duckdblabs/rilldata/issues/6.
	// Forces DuckDB to create catalog entries for the information schema up front (they are normally created lazily).
	// Can be removed if the issue persists.
	_, err = db.ExecContext(context.Background(), `
		select
			coalesce(t.table_catalog, current_database()) as "database",
			t.table_schema as "schema",
			t.table_name as "name",
			t.table_type as "type", 
			array_agg(c.column_name order by c.ordinal_position) as "column_names",
			array_agg(c.data_type order by c.ordinal_position) as "column_types",
			array_agg(c.is_nullable = 'YES' order by c.ordinal_position) as "column_nullable"
		from information_schema.tables t
		join information_schema.columns c on t.table_schema = c.table_schema and t.table_name = c.table_name
		group by 1, 2, 3, 4
		order by 1, 2, 3, 4
	`)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (d *db) attachDBs(ctx context.Context, db *sqlx.DB, read bool, ignoreTable string) error {
	entries, err := os.ReadDir(d.localPath)
	if err != nil {
		return err
	}

	var views []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if entry.Name() == ignoreTable {
			continue
		}

		version, exist, _ := d.tableVersion(entry.Name(), false)
		if !exist {
			continue
		}
		versionPath := filepath.Join(d.localPath, entry.Name(), version)

		// read meta file
		isView := true
		f, err := os.ReadFile(filepath.Join(versionPath, "meta.json"))
		if err != nil {
			pathErr := &fs.PathError{}
			if !errors.As(err, &pathErr) {
				_ = os.RemoveAll(versionPath)
				d.logger.Warn("error in reading meta file", slog.String("table", entry.Name()), slog.Any("error", err))
				return err
			}
			isView = false
		}
		if isView {
			var meta meta
			err = json.Unmarshal(f, &meta)
			if err != nil {
				_ = os.RemoveAll(versionPath)
				d.logger.Warn("error in unmarshalling meta file", slog.String("table", entry.Name()), slog.Any("error", err))
				return err
			}
			// table is a view
			views = append(views, fmt.Sprintf("CREATE OR REPLACE VIEW %s AS (%s\n)", safeSQLName(entry.Name()), meta.ViewSQL))
			continue
		}
		dbName := dbName(entry.Name())
		var readMode string
		if read {
			readMode = " (READ_ONLY)"
		}
		_, err = db.ExecContext(ctx, fmt.Sprintf("ATTACH %s AS %s %s", safeSQLString(filepath.Join(versionPath, "data.db")), safeSQLName(dbName), readMode))
		if err != nil {
			d.logger.Error("error in attaching db", slog.String("table", entry.Name()), slog.Any("error", err))
			_ = os.RemoveAll(filepath.Join(d.localPath, entry.Name()))
			return err
		}

		_, err = db.ExecContext(ctx, fmt.Sprintf("CREATE OR REPLACE VIEW %s AS SELECT * FROM %s.default", safeSQLName(entry.Name()), safeSQLName(dbName)))
		if err != nil {
			return err
		}
	}

	// create views after attaching all the DBs since views can depend on other tables
	for _, view := range views {
		_, err := db.ExecContext(ctx, view)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *db) tableVersion(name string, read bool) (string, bool, error) {
	if read {
		return tableVersion(d.localPath, name)
	}
	return tableVersion(d.localPath, name)
}

func (d *db) setTableVersion(name, version string) error {
	return os.WriteFile(filepath.Join(d.localPath, name, "version.txt"), []byte(version), fs.ModePerm)
}

func (d *db) deleteLocalTable(table, version string) error {
	var path string
	if version == "" {
		path = filepath.Join(d.localPath, table)
	} else {
		path = filepath.Join(d.localPath, table, version)
	}
	return os.RemoveAll(path)
}

func execIncrementalInsert(ctx context.Context, conn *sqlx.Conn, safeTableName, query string, opts *InsertTableOptions) error {
	var byNameClause string
	if opts.ByName {
		byNameClause = "BY NAME"
	}

	if opts.Strategy == IncrementalStrategyAppend {
		_, err := conn.ExecContext(ctx, fmt.Sprintf("INSERT INTO %s %s (%s\n)", safeTableName, byNameClause, query))
		return err
	}

	if opts.Strategy == IncrementalStrategyMerge {
		// Create a temporary table with the new data
		tmp := uuid.New().String()
		_, err := conn.ExecContext(ctx, fmt.Sprintf("CREATE TEMPORARY TABLE %s AS (%s\n)", safeSQLName(tmp), query))
		if err != nil {
			return err
		}

		// check the count of the new data
		// skip if the count is 0
		// if there was no data in the empty file then the detected schema can be different from the current schema which leads to errors or performance issues
		res := conn.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) == 0 FROM %s", safeSQLName(tmp)))
		var empty bool
		if err := res.Scan(&empty); err != nil {
			return err
		}

		if empty {
			return nil
		}

		// Drop the rows from the target table where the unique key is present in the temporary table
		where := ""
		for i, key := range opts.UniqueKey {
			key = safeSQLName(key)
			if i != 0 {
				where += " AND "
			}
			where += fmt.Sprintf("base.%s IS NOT DISTINCT FROM tmp.%s", key, key)
		}
		_, err = conn.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s base WHERE EXISTS (SELECT 1 FROM %s tmp WHERE %s)", safeTableName, safeSQLName(tmp), where))
		if err != nil {
			return err
		}

		// Insert the new data into the target table
		_, err = conn.ExecContext(ctx, fmt.Sprintf("INSERT INTO %s %s SELECT * FROM %s", safeTableName, byNameClause, safeSQLName(tmp)))
		return err
	}

	return fmt.Errorf("incremental insert strategy %q not supported", opts.Strategy)
}

func tableVersion(path, name string) (string, bool, error) {
	pathToFile := filepath.Join(path, name, "version.txt")
	contents, err := os.ReadFile(pathToFile)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", false, nil
		}
		return "", false, err
	}
	return strings.TrimSpace(string(contents)), true, nil
}

func newVersion() string {
	return strconv.FormatInt(time.Now().UnixMilli(), 10)
}

type meta struct {
	ViewSQL string
}

func writeMeta(path string, meta *meta) error {
	if meta == nil {
		return nil
	}
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("create: marshal meta failed: %w", err)
	}
	err = os.WriteFile(filepath.Join(path, "meta.json"), metaBytes, fs.ModePerm)
	if err != nil {
		return fmt.Errorf("create: write meta failed: %w", err)
	}
	return nil
}

func dbName(name string) string {
	return fmt.Sprintf("%s__data__db", name)
}

type settings struct {
	MaxMemory string `mapstructure:"max_memory"`
	Threads   string `mapstructure:"threads"`
	// Can be more settings
}

// Regex to parse human-readable size returned by DuckDB
// nolint
var humanReadableSizeRegex = regexp.MustCompile(`^([\d.]+)\s*(\S+)$`)

// Reversed logic of StringUtil::BytesToHumanReadableString
// see https://github.com/cran/duckdb/blob/master/src/duckdb/src/common/string_util.cpp#L157
// Examples: 1 bytes, 2 bytes, 1KB, 1MB, 1TB, 1PB
// nolint
func humanReadableSizeToBytes(sizeStr string) (float64, error) {
	var multiplier float64

	match := humanReadableSizeRegex.FindStringSubmatch(sizeStr)

	if match == nil {
		return 0, fmt.Errorf("invalid size format: '%s'", sizeStr)
	}

	sizeFloat, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return 0, err
	}

	switch match[2] {
	case "byte", "bytes":
		multiplier = 1
	case "KB":
		multiplier = 1000
	case "MB":
		multiplier = 1000 * 1000
	case "GB":
		multiplier = 1000 * 1000 * 1000
	case "TB":
		multiplier = 1000 * 1000 * 1000 * 1000
	case "PB":
		multiplier = 1000 * 1000 * 1000 * 1000 * 1000
	case "KiB":
		multiplier = 1024
	case "MiB":
		multiplier = 1024 * 1024
	case "GiB":
		multiplier = 1024 * 1024 * 1024
	case "TiB":
		multiplier = 1024 * 1024 * 1024 * 1024
	case "PiB":
		multiplier = 1024 * 1024 * 1024 * 1024 * 1024
	default:
		return 0, fmt.Errorf("unknown size unit '%s' in '%s'", match[2], sizeStr)
	}

	return sizeFloat * multiplier, nil
}
