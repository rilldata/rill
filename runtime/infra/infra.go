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
		panic(fmt.Errorf("Already registered infra driver with name '%s'", name))
	}
	Drivers[name] = driver
}

type Driver interface {
	Open(dsn string) (Connection, error)
}

type Connection interface {
	Execute(ctx context.Context, stmt *Statement) (*sqlx.Rows, error)
	Close() error
	InformationSchema() string
}

type Statement struct {
	Query    string
	Args     []any
	DryRun   bool
	Priority int
}

// ErrClosed should be returned by Execute if the connection is closed
var ErrClosed = errors.New("infra: connection is closed")
