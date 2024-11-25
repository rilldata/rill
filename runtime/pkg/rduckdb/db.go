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
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/jmoiron/sqlx"
	"github.com/marcboeker/go-duckdb"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/pkg/ctxsync"
	"go.opentelemetry.io/otel/attribute"
	"gocloud.dev/blob"
	"golang.org/x/sync/semaphore"
)

var errNotFound = errors.New("not found")

type DB interface {
	// Close closes the database.
	Close() error

	// AcquireReadConnection returns a connection to the database for reading.
	// Once done the connection should be released by calling the release function.
	// This connection must only be used for select queries or for creating and working with temporary tables.
	AcquireReadConnection(ctx context.Context) (conn *sqlx.Conn, release func() error, err error)

	// Size returns the size of the database in bytes.
	// It is currently implemented as sum of the size of all serving `.db` files.
	Size() int64

	// CRUD APIs

	// CreateTableAsSelect creates a new table by name from the results of the given SQL query.
	CreateTableAsSelect(ctx context.Context, name string, sql string, opts *CreateTableOptions) error

	// MutateTable allows mutating a table in the database by calling the mutateFn.
	MutateTable(ctx context.Context, name string, mutateFn func(ctx context.Context, conn *sqlx.Conn) error) error

	// DropTable removes a table from the database.
	DropTable(ctx context.Context, name string) error

	// RenameTable renames a table in the database.
	RenameTable(ctx context.Context, oldName, newName string) error
}

type DBOptions struct {
	// LocalPath is the path where local db files will be stored. Should be unique for each database.
	LocalPath string
	// Remote is the blob storage bucket where the database files will be stored. This is the source of truth.
	// The local db will be eventually synced with the remote.
	Remote *blob.Bucket

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

	if readThread == 0 && writeThread == 0 {
		connector, err := duckdb.NewConnector("", nil)
		if err != nil {
			return fmt.Errorf("unable to create duckdb connector: %w", err)
		}
		defer connector.Close()
		db := sql.OpenDB(connector)
		defer db.Close()

		row := db.QueryRow("SELECT value FROM duckdb_settings() WHERE name = 'threads'")
		var threads int
		err = row.Scan(&threads)
		if err != nil {
			return fmt.Errorf("unable to get threads: %w", err)
		}

		read.Threads = strconv.Itoa((threads + 1) / 2)
		write.Threads = strconv.Itoa(threads / 2)
	}

	if readThread == 0 != (writeThread == 0) {
		// only one is defined
		var threads int
		if readThread != 0 {
			threads = readThread
		} else {
			threads = writeThread
		}

		read.Threads = strconv.Itoa((threads + 1) / 2)
		if threads <= 3 {
			write.Threads = "1"
		} else {
			write.Threads = strconv.Itoa(threads / 2)
		}
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
	// InitSQL is the SQL to run before creating the table.
	InitSQL string
}

// NewDB creates a new DB instance.
// dbIdentifier is a unique identifier for the database reported in metrics.
func NewDB(ctx context.Context, opts *DBOptions) (DB, error) {
	err := opts.ValidateSettings()
	if err != nil {
		return nil, err
	}

	bgctx, cancel := context.WithCancel(context.Background())
	db := &db{
		opts:       opts,
		localPath:  opts.LocalPath,
		remote:     opts.Remote,
		readMu:     ctxsync.NewRWMutex(),
		writeSem:   semaphore.NewWeighted(1),
		metaSem:    semaphore.NewWeighted(1),
		localDirty: true,
		ticker:     time.NewTicker(5 * time.Minute),
		logger:     opts.Logger,
		ctx:        bgctx,
		cancel:     cancel,
	}
	// catalog
	db.catalog = newCatalog(
		db.removeTableVersion,
		db.removeSnapshot,
	)

	// create local path
	err = os.MkdirAll(db.localPath, fs.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("unable to create read path: %w", err)
	}

	// sync local data
	err = db.pullFromRemote(ctx)
	if err != nil {
		return nil, err
	}

	// create read handle
	db.dbHandle, err = db.openDBAndAttach(ctx, "", "", true)
	if err != nil {
		if strings.Contains(err.Error(), "Symbol not found") {
			fmt.Printf("Your version of macOS is not supported. Please upgrade to the latest major release of macOS. See this link for details: https://support.apple.com/en-in/macos/upgrade")
			os.Exit(1)
		}
		return nil, err
	}
	go db.localDBMonitor()
	return db, nil
}

type db struct {
	opts *DBOptions

	localPath string
	remote    *blob.Bucket

	// dbHandle serves executes meta queries and serves read queries
	dbHandle *sqlx.DB
	// readMu controls access to readHandle
	readMu ctxsync.RWMutex
	// writeSem ensures only one write operation is allowed at a time
	writeSem *semaphore.Weighted
	// metaSem enures only one meta operation can run on a duckb handle.
	// Meta operations are attach, detach, create view queries done on the db handle
	metaSem *semaphore.Weighted
	// localDirty is set to true when a change is committed to the remote but not yet reflected in the local db
	localDirty bool
	// ticker to peroiodically check if local db is in sync with remote
	ticker  *time.Ticker
	catalog *catalog

	logger *slog.Logger

	// ctx and cancel to cancel background operations
	ctx    context.Context
	cancel context.CancelFunc
}

var _ DB = &db{}

func (d *db) Close() error {
	// close background operations
	d.cancel()
	d.ticker.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_ = d.writeSem.Acquire(ctx, 1)
	defer d.writeSem.Release(1)

	err := d.readMu.Lock(ctx)
	if err != nil {
		return err
	}
	defer d.readMu.Unlock()

	err = d.dbHandle.Close()
	d.dbHandle = nil
	return err
}

func (d *db) AcquireReadConnection(ctx context.Context) (*sqlx.Conn, func() error, error) {
	if err := d.readMu.RLock(ctx); err != nil {
		return nil, nil, err
	}

	// acquire a connection
	snapshot, err := d.catalog.acquireSnapshot(ctx)
	if err != nil {
		d.readMu.RUnlock()
		return nil, nil, err
	}

	conn, err := d.dbHandle.Connx(ctx)
	if err != nil {
		d.readMu.RUnlock()
		return nil, nil, err
	}

	err = d.prepareSnapshot(ctx, conn, snapshot)
	if err != nil {
		_ = conn.Close()
		d.readMu.RUnlock()
		return nil, nil, err
	}

	release := func() error {
		err = d.catalog.releaseSnapshot(ctx, snapshot)
		err = errors.Join(err, conn.Close())
		d.readMu.RUnlock()
		return err
	}
	return conn, release, nil
}

func (d *db) CreateTableAsSelect(ctx context.Context, name, query string, opts *CreateTableOptions) error {
	d.logger.Debug("create table", slog.String("name", name), slog.Bool("view", opts.View))
	err := d.writeSem.Acquire(ctx, 1)
	if err != nil {
		return err
	}
	defer d.writeSem.Release(1)

	// pull latest changes from remote
	err = d.pullFromRemote(ctx)
	if err != nil {
		return err
	}

	// check if some older version exists
	// We can also use catalog to get the latest version
	// but we are not using it here since pullFromRemote should have already updated the catalog
	// and we need meta.json contents
	oldMeta, _ := d.tableMeta(name)
	if oldMeta != nil {
		d.logger.Debug("old version", slog.String("version", oldMeta.Version))
	}

	// create new version directory
	newVersion := newVersion()
	newMeta := &tableMeta{
		Name:           name,
		Version:        newVersion,
		CreatedVersion: newVersion,
	}
	var dsn string
	if opts.View {
		dsn = ""
		newMeta.SQL = query
		// special handling to ensure that if a view is recreated with the same name and schema then any views on top of this view still works
		if oldMeta != nil && oldMeta.Type == "VIEW" {
			newMeta.CreatedVersion = oldMeta.CreatedVersion
		}
		err = os.MkdirAll(filepath.Join(d.localPath, name), fs.ModePerm)
		if err != nil {
			return fmt.Errorf("create: unable to create dir %q: %w", name, err)
		}
	} else {
		newVersionDir := filepath.Join(d.localPath, name, newVersion)
		err = os.MkdirAll(newVersionDir, fs.ModePerm)
		if err != nil {
			return fmt.Errorf("create: unable to create dir %q: %w", name, err)
		}
		dsn = filepath.Join(newVersionDir, "data.db")
	}

	// need to attach existing table so that any views dependent on this table are correctly attached
	conn, release, err := d.acquireWriteConn(ctx, dsn, name, true)
	if err != nil {
		return err
	}
	defer func() {
		_ = release()
	}()

	safeName := safeSQLName(name)
	var typ string
	if opts.View {
		typ = "VIEW"
		newMeta.Type = "VIEW"
	} else {
		typ = "TABLE"
		newMeta.Type = "TABLE"
	}
	if opts.InitSQL != "" {
		_, err = conn.ExecContext(ctx, opts.InitSQL, nil)
		if err != nil {
			return fmt.Errorf("create: init sql failed: %w", err)
		}
	}
	// ingest data
	_, err = conn.ExecContext(ctx, fmt.Sprintf("CREATE OR REPLACE %s %s AS (%s\n)", typ, safeName, query), nil)
	if err != nil {
		return fmt.Errorf("create: create %s %q failed: %w", typ, name, err)
	}

	// close write handle before syncing read so that temp files or wal files are removed
	err = release()
	if err != nil {
		return err
	}

	d.localDirty = true
	// update remote data and metadata
	if err := d.pushToRemote(ctx, name, oldMeta, newMeta); err != nil {
		return fmt.Errorf("create: replicate failed: %w", err)
	}
	d.logger.Debug("remote table updated", slog.String("name", name))
	// no errors after this point since background goroutine will eventually sync the local db

	// update local metadata
	err = d.writeTableMeta(name, newMeta)
	if err != nil {
		d.logger.Debug("create: error in writing table meta", slog.String("name", name), slog.String("error", err.Error()))
		return nil
	}

	err = d.catalog.addTableVersion(ctx, name, newMeta)
	if err != nil {
		d.logger.Debug("create: error in adding version", slog.String("table", name), slog.String("version", newMeta.Version), slog.String("error", err.Error()))
		return nil
	}
	d.localDirty = false
	return nil
}

func (d *db) MutateTable(ctx context.Context, name string, mutateFn func(ctx context.Context, conn *sqlx.Conn) error) error {
	d.logger.Debug("mutate table", slog.String("name", name))
	err := d.writeSem.Acquire(ctx, 1)
	if err != nil {
		return err
	}
	defer d.writeSem.Release(1)

	// pull latest changes from remote
	err = d.pullFromRemote(ctx)
	if err != nil {
		return err
	}

	oldMeta, err := d.tableMeta(name)
	if err != nil {
		if errors.Is(err, errNotFound) {
			return fmt.Errorf("mutate: Table %q not found", name)
		}
		return fmt.Errorf("mutate: unable to get table meta: %w", err)
	}

	// create new version directory
	newVersion := newVersion()
	newVersionDir := filepath.Join(d.localPath, name, newVersion)
	err = os.MkdirAll(newVersionDir, fs.ModePerm)
	if err != nil {
		return fmt.Errorf("mutate: unable to create dir %q: %w", name, err)
	}

	err = copyDir(newVersionDir, filepath.Join(d.localPath, name, oldMeta.Version))
	if err != nil {
		return fmt.Errorf("mutate: copy table failed: %w", err)
	}

	// acquire write connection
	// need to ignore attaching table since it is already present in the db file
	conn, release, err := d.acquireWriteConn(ctx, filepath.Join(newVersionDir, "data.db"), name, false)
	if err != nil {
		return err
	}

	err = mutateFn(ctx, conn)
	if err != nil {
		_ = release()
		return fmt.Errorf("mutate: mutate failed: %w", err)
	}

	// push to remote
	err = release()
	if err != nil {
		return fmt.Errorf("mutate: failed to close connection: %w", err)
	}
	d.localDirty = true
	meta := &tableMeta{
		Name:           name,
		Version:        newVersion,
		CreatedVersion: oldMeta.CreatedVersion,
		Type:           oldMeta.Type,
		SQL:            oldMeta.SQL,
	}
	err = d.pushToRemote(ctx, name, oldMeta, meta)
	if err != nil {
		return fmt.Errorf("mutate: replicate failed: %w", err)
	}
	// no errors after this point since background goroutine will eventually sync the local db

	// update local meta
	err = d.writeTableMeta(name, meta)
	if err != nil {
		d.logger.Debug("mutate: error in writing table meta", slog.String("name", name), slog.String("error", err.Error()))
		return nil
	}

	err = d.catalog.addTableVersion(ctx, name, meta)
	if err != nil {
		d.logger.Debug("mutate: error in adding version", slog.String("table", name), slog.String("version", meta.Version), slog.String("error", err.Error()))
		return nil
	}
	d.localDirty = false
	return nil
}

// DropTable implements DB.
func (d *db) DropTable(ctx context.Context, name string) error {
	d.logger.Debug("drop table", slog.String("name", name))
	err := d.writeSem.Acquire(ctx, 1)
	if err != nil {
		return err
	}
	defer d.writeSem.Release(1)

	// pull latest changes from remote
	err = d.pullFromRemote(ctx)
	if err != nil {
		return fmt.Errorf("drop: unable to pull from remote: %w", err)
	}

	// check if table exists
	_, err = d.tableMeta(name)
	if err != nil {
		if errors.Is(err, errNotFound) {
			return fmt.Errorf("drop: Table %q not found", name)
		}
		return fmt.Errorf("drop: unable to get table meta: %w", err)
	}

	// drop the table from remote
	d.localDirty = true
	err = d.deleteRemote(ctx, name, "")
	if err != nil {
		return fmt.Errorf("drop: unable to drop table %q from remote: %w", name, err)
	}
	// no errors after this point since background goroutine will eventually sync the local db

	err = d.catalog.removeTable(ctx, name)
	if err != nil {
		d.logger.Debug("drop: error in removing table", slog.String("name", name), slog.String("error", err.Error()))
		return nil
	}
	d.localDirty = false
	return nil
}

func (d *db) RenameTable(ctx context.Context, oldName, newName string) error {
	d.logger.Debug("rename table", slog.String("from", oldName), slog.String("to", newName))
	if strings.EqualFold(oldName, newName) {
		return fmt.Errorf("rename: Table with name %q already exists", newName)
	}
	err := d.writeSem.Acquire(ctx, 1)
	if err != nil {
		return err
	}
	defer d.writeSem.Release(1)

	// pull latest changes from remote
	err = d.pullFromRemote(ctx)
	if err != nil {
		return fmt.Errorf("rename: unable to pull from remote: %w", err)
	}

	oldMeta, err := d.tableMeta(oldName)
	if err != nil {
		if errors.Is(err, errNotFound) {
			return fmt.Errorf("rename: Table %q not found", oldName)
		}
		return fmt.Errorf("rename: unable to get table meta: %w", err)
	}

	// copy the old table to new table
	newVersion := newVersion()
	err = copyDir(filepath.Join(d.localPath, newName, newVersion), filepath.Join(d.localPath, oldName, oldMeta.Version))
	if err != nil {
		return fmt.Errorf("rename: copy table failed: %w", err)
	}

	// rename the underlying table
	err = renameTable(ctx, filepath.Join(d.localPath, newName, newVersion, "data.db"), oldName, newName)
	if err != nil {
		return fmt.Errorf("rename: rename table failed: %w", err)
	}

	d.localDirty = true
	// sync the new table and new version
	meta := &tableMeta{
		Name:           newName,
		Version:        newVersion,
		CreatedVersion: newVersion,
		Type:           oldMeta.Type,
		SQL:            oldMeta.SQL,
	}
	if err := d.pushToRemote(ctx, newName, oldMeta, meta); err != nil {
		return fmt.Errorf("rename: unable to replicate new table: %w", err)
	}

	// TODO :: fix this
	// at this point db is inconsistent
	// has both old table and new table

	// drop the old table in remote
	err = d.deleteRemote(ctx, oldName, "")
	if err != nil {
		return fmt.Errorf("rename: unable to delete old table %q from remote: %w", oldName, err)
	}

	// no errors after this point since background goroutine will eventually sync the local db

	// update local meta for new table
	err = d.writeTableMeta(newName, meta)
	if err != nil {
		d.logger.Debug("rename: error in writing table meta", slog.String("name", newName), slog.String("error", err.Error()))
		return nil
	}

	// remove old table from local db
	err = d.catalog.removeTable(ctx, oldName)
	if err != nil {
		d.logger.Debug("rename: error in removing table", slog.String("name", oldName), slog.String("error", err.Error()))
		return nil
	}
	err = d.catalog.addTableVersion(ctx, newName, meta)
	if err != nil {
		d.logger.Debug("rename: error in adding version", slog.String("table", newName), slog.String("version", newVersion), slog.String("error", err.Error()))
		return nil
	}
	d.localDirty = false
	return nil
}

func (d *db) localDBMonitor() {
	for {
		select {
		case <-d.ctx.Done():
			return
		case <-d.ticker.C:
			err := d.writeSem.Acquire(d.ctx, 1)
			if err != nil && !errors.Is(err, context.Canceled) {
				d.logger.Error("localDBMonitor: error in acquiring write sem", slog.String("error", err.Error()))
				continue
			}
			if !d.localDirty {
				// all good
				continue
			}
			err = d.pullFromRemote(d.ctx)
			if err != nil && !errors.Is(err, context.Canceled) {
				d.logger.Error("localDBMonitor: error in pulling from remote", slog.String("error", err.Error()))
			}
		}
	}
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
		meta, _ := d.tableMeta(entry.Name())
		if meta != nil {
			paths = append(paths, filepath.Join(d.localPath, entry.Name(), meta.Version, "data.db"))
		}
	}
	return fileSize(paths)
}

// acquireWriteConn syncs the write database, initializes the write handle and returns a write connection.
// The release function should be called to release the connection.
// It should be called with the writeMu locked.
func (d *db) acquireWriteConn(ctx context.Context, dsn, table string, attachExisting bool) (*sqlx.Conn, func() error, error) {
	var ignoreTable string
	if !attachExisting {
		ignoreTable = table
	}
	db, err := d.openDBAndAttach(ctx, dsn, ignoreTable, false)
	if err != nil {
		return nil, nil, err
	}
	conn, err := db.Connx(ctx)
	if err != nil {
		_ = db.Close()
		return nil, nil, err
	}

	if attachExisting {
		_, err = conn.ExecContext(ctx, "DROP VIEW IF EXISTS "+safeSQLName(table))
		if err != nil {
			_ = conn.Close()
			_ = db.Close()
			return nil, nil, err
		}
	}

	return conn, func() error {
		_ = conn.Close()
		err = db.Close()
		return err
	}, nil
}

func (d *db) openDBAndAttach(ctx context.Context, uri, ignoreTable string, read bool) (*sqlx.DB, error) {
	d.logger.Debug("open db", slog.Bool("read", read), slog.String("uri", uri))
	// open the db
	var settings map[string]string
	dsn, err := url.Parse(uri) // in-memory
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

	conn, err := db.Connx(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	tables, err := d.catalog.listTables(ctx)
	if err != nil {
		return nil, err
	}

	err = d.attachTables(ctx, conn, tables, ignoreTable)
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
		db.Close()
		return nil, err
	}

	return db, nil
}

func (d *db) attachTables(ctx context.Context, conn *sqlx.Conn, tables []*tableMeta, ignoreTable string) error {
	// sort tables by created_version
	// this is to ensure that views/tables on which other views depend are attached first
	slices.SortFunc(tables, func(a, b *tableMeta) int {
		// all tables should be attached first and can be attached in any order
		if a.Type == "TABLE" && b.Type == "TABLE" {
			return 0
		}
		if a.Type == "TABLE" {
			return -1
		}
		if b.Type == "TABLE" {
			return 1
		}
		return strings.Compare(a.CreatedVersion, b.CreatedVersion)
	})
	for _, table := range tables {
		if table.Name == ignoreTable {
			continue
		}
		err := d.attachTable(ctx, conn, table)
		if err != nil {
			return fmt.Errorf("failed to attach table %q: %w", table.Name, err)
		}
	}
	return nil
}

func (d *db) attachTable(ctx context.Context, conn *sqlx.Conn, table *tableMeta) error {
	safeTable := safeSQLName(table.Name)
	if table.Type == "VIEW" {
		_, err := conn.ExecContext(ctx, fmt.Sprintf("CREATE OR REPLACE VIEW %s AS %s", safeTable, table.SQL))
		return err
	}

	safeDBName := safeSQLName(dbName(table.Name, table.Version))
	_, err := conn.ExecContext(ctx, fmt.Sprintf("ATTACH IF NOT EXISTS %s AS %s (READ_ONLY)", safeSQLString(filepath.Join(d.localPath, table.Name, table.Version, "data.db")), safeDBName))
	if err != nil {
		d.logger.Warn("error in attaching db", slog.String("table", table.Name), slog.Any("error", err))
		return err
	}

	_, err = conn.ExecContext(ctx, fmt.Sprintf("CREATE OR REPLACE VIEW %s AS SELECT * FROM %s.%s", safeTable, safeDBName, safeTable))
	return err
}

func (d *db) tableMeta(name string) (*tableMeta, error) {
	contents, err := os.ReadFile(filepath.Join(d.localPath, name, "meta.json"))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, errNotFound
		}
		return nil, err
	}
	m := &tableMeta{}
	err = json.Unmarshal(contents, m)
	if err != nil {
		return nil, err
	}

	if m.Type == "VIEW" {
		return m, nil
	}
	// this is required because release version does not table table directory as of now
	_, err = os.Stat(filepath.Join(d.localPath, name, m.Version))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, errNotFound
		}
		return nil, err
	}
	return m, nil
}

func (d *db) writeTableMeta(name string, meta *tableMeta) error {
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("create: marshal meta failed: %w", err)
	}
	err = os.WriteFile(filepath.Join(d.localPath, name, "meta.json"), metaBytes, fs.ModePerm)
	if err != nil {
		return fmt.Errorf("create: write meta failed: %w", err)
	}
	return nil
}

// deleteLocalTableFiles delete table files for the given table name. If version is provided, only that version is deleted.
func (d *db) deleteLocalTableFiles(name, version string) error {
	var path string
	if version == "" {
		path = filepath.Join(d.localPath, name)
	} else {
		path = filepath.Join(d.localPath, name, version)
	}
	return os.RemoveAll(path)
}

func (d *db) removeTableVersion(ctx context.Context, name, version string) error {
	err := d.metaSem.Acquire(ctx, 1)
	if err != nil {
		return err
	}
	d.metaSem.Release(1)

	_, err = d.dbHandle.ExecContext(ctx, "DETACH DATABASE IF EXISTS "+dbName(name, version))
	if err != nil {
		return err
	}
	return d.deleteLocalTableFiles(name, version)
}

func (d *db) prepareSnapshot(ctx context.Context, conn *sqlx.Conn, s *snapshot) error {
	err := d.metaSem.Acquire(ctx, 1)
	if err != nil {
		return err
	}
	defer d.metaSem.Release(1)

	if s.ready {
		return nil
	}

	_, err = conn.ExecContext(ctx, "CREATE SCHEMA IF NOT EXISTS "+schemaName(s.id))
	if err != nil {
		return err
	}

	_, err = conn.ExecContext(ctx, "USE "+schemaName(s.id))
	if err != nil {
		return err
	}

	return d.attachTables(ctx, conn, s.tables, "")
}

func (d *db) removeSnapshot(ctx context.Context, id int) error {
	err := d.metaSem.Acquire(ctx, 1)
	if err != nil {
		return err
	}
	defer d.metaSem.Release(1)

	_, err = d.dbHandle.Exec(fmt.Sprintf("DROP SCHEMA %s CASCADE", schemaName(id)))
	return err
}

type tableMeta struct {
	Name           string `json:"name"`
	Version        string `json:"version"`
	CreatedVersion string `json:"created_version"`
	Type           string `json:"type"` // either TABLE or VIEW
	SQL            string `json:"sql"`  // populated for views
}

func renameTable(ctx context.Context, dbFile, old, newName string) error {
	db, err := sql.Open("duckdb", dbFile)
	if err != nil {
		return err
	}
	defer db.Close()

	var isView bool
	err = db.QueryRowContext(ctx, "SELECT lower(table_type) = 'view' FROM INFORMATION_SCHEMA.TABLES WHERE table_name = ?", old).Scan(&isView)
	if err != nil {
		return err
	}

	var typ string
	if isView {
		typ = "VIEW"
	} else {
		typ = "TABLE"
	}

	_, err = db.ExecContext(ctx, fmt.Sprintf("ALTER %s %s RENAME TO %s", typ, old, newName))
	return err
}

func newVersion() string {
	return strconv.FormatInt(time.Now().UnixMilli(), 10)
}

func dbName(table, version string) string {
	return fmt.Sprintf("%s__%s__db", table, version)
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

func schemaName(gen int) string {
	return fmt.Sprintf("main_%v", gen)
}
