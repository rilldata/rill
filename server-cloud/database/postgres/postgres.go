package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/rilldata/rill/server-cloud/database"
)

func init() {
	database.Register("postgres", driver{})
}

type driver struct{}

func (d driver) Open(dsn string) (database.DB, error) {
	db, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return nil, err
	}

	return &connection{db: db}, nil
}

type connection struct {
	db *pgx.Conn
}

func (c *connection) Close() error {
	return c.db.Close(context.Background())
}

func (c *connection) FindMigrationVersion(ctx context.Context) (int, error) {
	var version int
	err := c.db.QueryRow(ctx, fmt.Sprintf("select version from %s", migrationVersionTable)).Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}
