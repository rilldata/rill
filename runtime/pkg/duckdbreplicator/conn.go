package duckdbreplicator

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// Conn represents a single database connection.
// This is useful when running a chain of queries using a single write connection.
type Conn interface {
	// Connx returns the underlying sqlx.Conn.
	Connx() *sqlx.Conn

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

type conn struct {
	*sqlx.Conn

	db *db
}

var _ Conn = (*conn)(nil)

func (c *conn) Connx() *sqlx.Conn {
	return c.Conn
}

func (c *conn) CreateTableAsSelect(ctx context.Context, name, sql string, opts *CreateTableOptions) error {
	if opts == nil {
		opts = &CreateTableOptions{}
	}
	return c.db.createTableAsSelect(ctx, c.Conn, func() error { return nil }, name, sql, opts)
}

// InsertTableAsSelect inserts the results of the given SQL query into the table.
func (c *conn) InsertTableAsSelect(ctx context.Context, name, sql string, opts *InsertTableOptions) error {
	if opts == nil {
		opts = &InsertTableOptions{
			Strategy: IncrementalStrategyAppend,
		}
	}
	return c.db.insertTableAsSelect(ctx, c.Conn, func() error { return nil }, name, sql, opts)
}

// DropTable removes a table from the database.
func (c *conn) DropTable(ctx context.Context, name string) error {
	return c.db.dropTable(ctx, name)
}

// RenameTable renames a table in the database.
func (c *conn) RenameTable(ctx context.Context, oldName, newName string) error {
	return c.db.renameTable(ctx, oldName, newName)
}

// AddTableColumn adds a column to the table.
func (c *conn) AddTableColumn(ctx context.Context, tableName, columnName, typ string) error {
	return c.db.addTableColumn(ctx, c.Conn, func() error { return nil }, tableName, columnName, typ)
}

// AlterTableColumn alters the type of a column in the table.
func (c *conn) AlterTableColumn(ctx context.Context, tableName, columnName, newType string) error {
	return c.db.alterTableColumn(ctx, c.Conn, func() error { return nil }, tableName, columnName, newType)
}

type singledbConn struct {
	*sqlx.Conn

	db *singledb
}

var _ Conn = (*singledbConn)(nil)

func (c *singledbConn) Connx() *sqlx.Conn {
	return c.Conn
}

func (c *singledbConn) CreateTableAsSelect(ctx context.Context, name, sql string, opts *CreateTableOptions) error {
	return c.db.createTableAsSelect(ctx, c.Conn, name, sql, opts)
}

// InsertTableAsSelect inserts the results of the given SQL query into the table.
func (c *singledbConn) InsertTableAsSelect(ctx context.Context, name, sql string, opts *InsertTableOptions) error {
	if opts == nil {
		opts = &InsertTableOptions{
			Strategy: IncrementalStrategyAppend,
		}
	}
	return execIncrementalInsert(ctx, c.Conn, name, sql, opts)
}

// DropTable removes a table from the database.
func (c *singledbConn) DropTable(ctx context.Context, name string) error {
	return c.db.dropTable(ctx, c.Conn, name)
}

// RenameTable renames a table in the database.
func (c *singledbConn) RenameTable(ctx context.Context, oldName, newName string) error {
	return c.db.renameTable(ctx, c.Conn, oldName, newName)
}

// AddTableColumn adds a column to the table.
func (c *singledbConn) AddTableColumn(ctx context.Context, tableName, columnName, typ string) error {
	return c.db.addTableColumn(ctx, c.Conn, tableName, columnName, typ)
}

// AlterTableColumn alters the type of a column in the table.
func (c *singledbConn) AlterTableColumn(ctx context.Context, tableName, columnName, newType string) error {
	return c.db.alterTableColumn(ctx, c.Conn, tableName, columnName, newType)
}
