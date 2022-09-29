package metadata

import (
	"context"
	"fmt"
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

// Open opens a new metadata DB connection
func Open(driver string, dsn string) (DB, error) {
	d, ok := Drivers[driver]
	if !ok {
		return nil, fmt.Errorf("unknown database driver: %s", driver)
	}

	db, err := d.Open(dsn)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Driver is the interface for metadata drivers
type Driver interface {
	Open(dsn string) (DB, error)
}

// DB is the interface for a metadata database
type DB interface {
	Close() error
	Migrate(ctx context.Context) error
	FindMigrationVersion(ctx context.Context) (int, error)
}
