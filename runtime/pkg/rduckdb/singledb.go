package rduckdb

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/XSAM/otelsql"
	"github.com/jmoiron/sqlx"
	"github.com/marcboeker/go-duckdb"
	"go.opentelemetry.io/otel/attribute"
)

type singledb struct {
	db      *sqlx.DB
	writeMU *sync.Mutex // limits write queries to one at a time. Does not block read queries.
	logger  *slog.Logger
}

type SingleDBOptions struct {
	DSN         string
	Clean       bool
	InitQueries []string
	Logger      *slog.Logger
}

var _ DB = &singledb{}

// NewSingleDB creates a new DB that writes to and reads from a single DuckDB database.
// This is useful for testing only.
func NewSingleDB(ctx context.Context, opts *SingleDBOptions) (DB, error) {
	if opts.Clean {
		u, err := url.Parse(opts.DSN)
		if err != nil {
			return nil, err
		}
		if u.Path != "" {
			err = os.Remove(u.Path)
			if err != nil && !os.IsNotExist(err) {
				return nil, err
			}
		}
	}
	connector, err := duckdb.NewConnector(opts.DSN, func(execer driver.ExecerContext) error {
		for _, qry := range opts.InitQueries {
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
		if strings.Contains(err.Error(), "Symbol not found") {
			fmt.Printf("Your version of macOS is not supported. Please upgrade to the latest major release of macOS. See this link for details: https://support.apple.com/en-in/macos/upgrade")
			os.Exit(1)
		}
		return nil, err
	}

	db := sqlx.NewDb(otelsql.OpenDB(connector), "duckdb")
	err = otelsql.RegisterDBStatsMetrics(db.DB, otelsql.WithAttributes(attribute.String("db.system", "duckdb")))
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("registering db stats metrics: %w", err)
	}

	err = db.PingContext(context.Background())
	if err != nil {
		db.Close()
		return nil, err
	}
	if opts.Logger == nil {
		opts.Logger = slog.New(slog.NewJSONHandler(io.Discard, nil))
	}
	return &singledb{
		db:      db,
		writeMU: &sync.Mutex{},
		logger:  opts.Logger,
	}, nil
}

// Close implements DB.
func (s *singledb) Close() error {
	return s.db.Close()
}

// AcquireReadConnection implements DB.
func (s *singledb) AcquireReadConnection(ctx context.Context) (Conn, func() error, error) {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return nil, nil, err
	}

	return &singledbConn{
		Conn: conn,
		db:   s,
	}, conn.Close, nil
}

func (s *singledb) AcquireWriteConnection(ctx context.Context) (Conn, func() error, error) {
	s.writeMU.Lock()
	c, err := s.db.Connx(ctx)
	if err != nil {
		s.writeMU.Unlock()
		return nil, nil, err
	}

	return &singledbConn{
			Conn: c,
			db:   s,
		}, func() error {
			err := c.Close()
			s.writeMU.Unlock()
			return err
		}, nil
}

// CreateTableAsSelect implements DB.
func (s *singledb) CreateTableAsSelect(ctx context.Context, name, uery string, opts *CreateTableOptions) error {
	s.writeMU.Lock()
	defer s.writeMU.Unlock()

	conn, err := s.db.Connx(ctx)
	if err != nil {
		return err
	}

	return s.createTableAsSelect(ctx, conn, name, uery, opts)
}

func (s *singledb) createTableAsSelect(ctx context.Context, conn *sqlx.Conn, name, query string, opts *CreateTableOptions) error {
	var typ string
	if opts != nil && opts.View {
		typ = "VIEW"
	} else {
		typ = "TABLE"
	}

	_, err := conn.ExecContext(ctx, fmt.Sprintf("CREATE OR REPLACE %s %s AS (%s\n)", typ, safeSQLName(name), query))
	return err
}

// DropTable implements DB.
func (s *singledb) DropTable(ctx context.Context, name string) error {
	s.writeMU.Lock()
	defer s.writeMU.Unlock()

	conn, err := s.db.Connx(ctx)
	if err != nil {
		return err
	}

	return s.dropTable(ctx, conn, name)
}

func (s *singledb) dropTable(ctx context.Context, conn *sqlx.Conn, name string) error {
	view, err := isView(ctx, conn, name)
	if err != nil {
		return err
	}
	var typ string
	if view {
		typ = "VIEW"
	} else {
		typ = "TABLE"
	}

	_, err = conn.ExecContext(ctx, fmt.Sprintf("DROP %s %s", typ, safeSQLName(name)))
	return err
}

// InsertTableAsSelect implements DB.
func (s *singledb) InsertTableAsSelect(ctx context.Context, name, query string, opts *InsertTableOptions) error {
	s.writeMU.Lock()
	defer s.writeMU.Unlock()

	conn, err := s.db.Connx(ctx)
	if err != nil {
		return err
	}

	return execIncrementalInsert(ctx, conn, safeSQLName(name), query, opts)
}

// RenameTable implements DB.
func (s *singledb) RenameTable(ctx context.Context, oldName, newName string) error {
	s.writeMU.Lock()
	defer s.writeMU.Unlock()

	conn, err := s.db.Connx(ctx)
	if err != nil {
		return err
	}

	return s.renameTable(ctx, conn, oldName, newName)
}

func (s *singledb) renameTable(ctx context.Context, conn *sqlx.Conn, oldName, newName string) error {
	view, err := isView(ctx, conn, oldName)
	if err != nil {
		return err
	}

	var typ string
	if view {
		typ = "VIEW"
	} else {
		typ = "TABLE"
	}

	newNameIsView, err := isView(ctx, conn, newName)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		// The newName does not exist.
		_, err = conn.ExecContext(ctx, fmt.Sprintf("ALTER %s %s RENAME TO %s", typ, safeSQLName(oldName), safeSQLName(newName)))
		return err
	}

	// The newName is already occupied.
	var existingTyp string
	if newNameIsView {
		existingTyp = "VIEW"
	} else {
		existingTyp = "TABLE"
	}

	_, err = conn.ExecContext(ctx, fmt.Sprintf("DROP %s IF EXISTS %s", existingTyp, safeSQLName(newName)))
	if err != nil {
		return err
	}

	_, err = conn.ExecContext(ctx, fmt.Sprintf("ALTER %s %s RENAME TO %s", typ, safeSQLName(oldName), safeSQLName(newName)))
	return err
}

// AddTableColumn implements DB.
func (s *singledb) AddTableColumn(ctx context.Context, tableName, columnName, typ string) error {
	s.writeMU.Lock()
	defer s.writeMU.Unlock()

	conn, err := s.db.Connx(ctx)
	if err != nil {
		return err
	}

	return s.addTableColumn(ctx, conn, tableName, columnName, typ)
}

func (s *singledb) addTableColumn(ctx context.Context, conn *sqlx.Conn, tableName, columnName, typ string) error {
	_, err := conn.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", safeSQLString(tableName), safeSQLName(columnName), typ))
	return err
}

// AlterTableColumn implements DB.
func (s *singledb) AlterTableColumn(ctx context.Context, tableName, columnName, newType string) error {
	s.writeMU.Lock()
	defer s.writeMU.Unlock()

	conn, err := s.db.Connx(ctx)
	if err != nil {
		return err
	}

	return s.alterTableColumn(ctx, conn, tableName, columnName, newType)
}

func (s *singledb) alterTableColumn(ctx context.Context, conn *sqlx.Conn, tableName, columnName, newType string) error {
	_, err := conn.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s", safeSQLName(tableName), safeSQLName(columnName), newType))
	return err
}

// TODO :: fix by calling pragma_database_size
func (s *singledb) Size() int64 {
	return 0
}

func isView(ctx context.Context, conn *sqlx.Conn, name string) (bool, error) {
	var view bool
	err := conn.QueryRowxContext(ctx, `
		SELECT 
			UPPER(table_type) = 'VIEW' 
		FROM 
			information_schema.tables 
		WHERE 
			table_catalog = current_database() 
			AND table_schema = 'main' 
			AND LOWER(table_name) = LOWER(?)
	`, name).Scan(&view)
	if err != nil {
		return false, err
	}
	return view, nil
}
