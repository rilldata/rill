package metadata

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4/source"
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
	Open(dsn string) (DB, error)
	Migrations() source.Driver
}

type DB interface {
}
