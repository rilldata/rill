package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/admin/database"

	// Load postgres driver
	_ "github.com/jackc/pgx/v4/stdlib"
)

func init() {
	database.Register("postgres", driver{})
}

type driver struct{}

func (d driver) Open(dsn string) (database.DB, error) {
	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, err
	}

	return &connection{db: db}, nil
}

type connection struct {
	db *sqlx.DB
}

func (c *connection) Close() error {
	return c.db.Close()
}

func (c *connection) FindMigrationVersion(ctx context.Context) (int, error) {
	var version int
	err := c.db.QueryRowxContext(ctx, fmt.Sprintf("select version from %s", migrationVersionTable)).Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

func (c *connection) FindOrganizations(ctx context.Context) ([]*database.Organization, error) {
	var res []*database.Organization
	err := c.db.Select(&res, "SELECT * FROM organizations ORDER BY name")
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *connection) FindOrganizationByName(ctx context.Context, name string) (*database.Organization, error) {
	res := &database.Organization{}
	err := c.db.QueryRowxContext(ctx, "SELECT * FROM organizations WHERE name = $1", name).StructScan(res)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, database.ErrNotFound
		}
		return nil, err
	}
	return res, nil
}

func (c *connection) CreateOrganization(ctx context.Context, name, description string) (*database.Organization, error) {
	res := &database.Organization{}
	err := c.db.QueryRowxContext(ctx, "INSERT INTO organizations(name, description) VALUES ($1, $2) RETURNING *", name, description).StructScan(res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *connection) UpdateOrganization(ctx context.Context, name, description string) (*database.Organization, error) {
	res := &database.Organization{}
	err := c.db.QueryRowxContext(ctx, "UPDATE organizations SET description=$1 WHERE name=$2 RETURNING *", description, name).StructScan(res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *connection) DeleteOrganization(ctx context.Context, name string) error {
	_, err := c.db.ExecContext(ctx, "DELETE FROM organizations WHERE name=$1", name)
	return err
}

func (c *connection) FindProjects(ctx context.Context, orgName string) ([]*database.Project, error) {
	var res []*database.Project
	err := c.db.Select(&res, "SELECT p.* FROM projects p JOIN organizations o ON p.organization_id = o.id WHERE o.name=$1 ORDER BY p.name", orgName)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *connection) FindProjectByName(ctx context.Context, orgName, name string) (*database.Project, error) {
	res := &database.Project{}
	err := c.db.QueryRowxContext(ctx, "SELECT p.* FROM projects p JOIN organizations o ON p.organization_id = o.id WHERE p.name=$1 AND o.name=$2", name, orgName).StructScan(res)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, database.ErrNotFound
		}
		return nil, err
	}
	return res, nil
}

func (c *connection) CreateProject(ctx context.Context, orgID, name, description string) (*database.Project, error) {
	res := &database.Project{}
	err := c.db.QueryRowxContext(ctx, "INSERT INTO projects(organization_id, name, description) VALUES ($1, $2, $3) RETURNING *", orgID, name, description).StructScan(res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *connection) UpdateProject(ctx context.Context, id, description string) (*database.Project, error) {
	res := &database.Project{}
	err := c.db.QueryRowxContext(ctx, "UPDATE projects SET description=$1 WHERE id=$2 RETURNING *", description, id).StructScan(res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *connection) DeleteProject(ctx context.Context, id string) error {
	_, err := c.db.ExecContext(ctx, "DELETE FROM projects WHERE id=$1", id)
	return err
}
