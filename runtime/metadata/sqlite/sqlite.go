package sqlite

import (
	"embed"

	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"

	"github.com/rilldata/rill/runtime/metadata"
)

//go:embed migrations/*.sql
var fs embed.FS

func init() {
	metadata.Register("sqlite", driver{})
}

type driver struct{}

func (d driver) Open(dsn string) (metadata.DB, error) {
	db, err := sqlx.Connect("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	return &connection{db: db}, nil
}

func (d driver) Migrations() source.Driver {
	source, err := iofs.New(fs, "migrations")
	if err != nil {
		panic(err)
	}
	return source
}

type connection struct {
	db *sqlx.DB
}
