package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Drivers is a registry of drivers
var Drivers = make(map[string]Driver)

// Register registers a new driver
func Register(name string, driver Driver) {
	if Drivers[name] != nil {
		panic(fmt.Errorf("already registered infra driver with name '%s'", name))
	}
	Drivers[name] = driver
}

type Driver interface {
	Open(dsn string) (Connection, error)
}

type Connection interface {
	Execute(ctx context.Context, stmt *Statement) (*sqlx.Rows, error)
	Close() error
	InformationSchema() InformationSchema
}

type Statement struct {
	Query    string
	Args     []any
	DryRun   bool
	Priority int
}

type InformationSchema interface {
	All(ctx context.Context) ([]*Table, error)
	Lookup(ctx context.Context, name string) (*Table, error)
}

type Table struct {
	Database string
	Schema   string
	Name     string
	Type     string
	Columns  []Column
}

type Column struct {
	Name     string
	Type     string
	Nullable bool
}

// ErrClosed should be returned by Execute if the connection is closed
var ErrClosed = errors.New("infra: connection is closed")

// ErrNotFound should be returned by InformationSchema.Lookup if no resource was found
var ErrNotFound = errors.New("infra: not found")
