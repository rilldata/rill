package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

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
	err := c.db.QueryRowxContext(ctx, "UPDATE organizations SET description=$1, updated_on=now() WHERE name=$2 RETURNING *", description, name).StructScan(res)
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

func (c *connection) FindProjectByGithubURL(ctx context.Context, githubURL string) (*database.Project, error) {
	res := &database.Project{}
	err := c.db.QueryRowxContext(ctx, "SELECT p.* FROM projects p WHERE p.github_url=lower($1)", githubURL).StructScan(res)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, database.ErrNotFound
		}
		return nil, err
	}
	return res, nil
}

func (c *connection) CreateProject(ctx context.Context, orgID string, p *database.Project) (*database.Project, error) {
	res := &database.Project{}
	err := c.db.QueryRowxContext(ctx, `
		INSERT INTO projects (organization_id, name, description, public, production_branch, github_url, github_installation_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *`,
		orgID, p.Name, p.Description, p.Public, p.ProductionBranch, p.GithubURL, p.GithubInstallationID,
	).StructScan(res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *connection) UpdateProject(ctx context.Context, p *database.Project) (*database.Project, error) {
	res := &database.Project{}
	err := c.db.QueryRowxContext(ctx, `
		UPDATE projects SET description=$1, public=$2, production_branch=$3, github_url=$4, github_installation_id=$5, updated_on=now() 
		WHERE id=$6 RETURNING *`,
		p.Description, p.Public, p.ProductionBranch, p.GithubURL, p.GithubInstallationID, p.ID,
	).StructScan(res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *connection) DeleteProject(ctx context.Context, id string) error {
	_, err := c.db.ExecContext(ctx, "DELETE FROM projects WHERE id=$1", id)
	return err
}

func (c *connection) FindUsers(ctx context.Context) ([]*database.User, error) {
	var res []*database.User
	err := c.db.Select(&res, "SELECT u.* FROM users u")
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *connection) FindUser(ctx context.Context, id string) (*database.User, error) {
	res := &database.User{}
	err := c.db.QueryRowxContext(ctx, "SELECT u.* FROM users u WHERE u.id=$1", id).StructScan(res)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, database.ErrNotFound
		}
		return nil, err
	}
	return res, nil
}

func (c *connection) FindUserByEmail(ctx context.Context, email string) (*database.User, error) {
	res := &database.User{}
	err := c.db.QueryRowxContext(ctx, "SELECT u.* FROM users u WHERE lower(u.email)=lower($1)", email).StructScan(res)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, database.ErrNotFound
		}
		return nil, err
	}
	return res, nil
}

func (c *connection) CreateUser(ctx context.Context, email, displayName, photoURL string) (*database.User, error) {
	res := &database.User{}
	err := c.db.QueryRowxContext(ctx, "INSERT INTO users (email, display_name, photo_url) VALUES ($1, $2, $3) RETURNING *", email, displayName, photoURL).StructScan(res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *connection) UpdateUser(ctx context.Context, id, displayName, photoURL string) (*database.User, error) {
	res := &database.User{}
	err := c.db.QueryRowxContext(ctx, "UPDATE users SET display_name=$1, photo_url=$2, updated_on=now() WHERE id=$3 RETURNING *", displayName, photoURL, id).StructScan(res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *connection) DeleteUser(ctx context.Context, id string) error {
	_, err := c.db.ExecContext(ctx, "DELETE FROM users WHERE id=$1", id)
	return err
}

func (c *connection) FindUserAuthTokens(ctx context.Context, userID string) ([]*database.UserAuthToken, error) {
	var res []*database.UserAuthToken
	err := c.db.Select(&res, "SELECT t.* FROM user_auth_tokens t WHERE t.user_id=$1", userID)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *connection) FindUserAuthToken(ctx context.Context, id string) (*database.UserAuthToken, error) {
	res := &database.UserAuthToken{}
	err := c.db.QueryRowxContext(ctx, "SELECT t.* FROM user_auth_tokens t WHERE t.id=$1", id).StructScan(res)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, database.ErrNotFound
		}
		return nil, err
	}
	return res, nil
}

func (c *connection) CreateUserAuthToken(ctx context.Context, opts *database.CreateUserAuthTokenOptions) (*database.UserAuthToken, error) {
	res := &database.UserAuthToken{}
	err := c.db.QueryRowxContext(ctx, `
		INSERT INTO user_auth_tokens (id, secret_hash, user_id, display_name, auth_client_id)
		VALUES ($1, $2, $3, $4, $5) RETURNING *`,
		opts.ID, opts.SecretHash, opts.UserID, opts.DisplayName, opts.AuthClientID,
	).StructScan(res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *connection) DeleteUserAuthToken(ctx context.Context, id string) error {
	_, err := c.db.ExecContext(ctx, "DELETE FROM user_auth_tokens WHERE id=$1", id)
	return err
}

// CreateAuthCode inserts the authorization code data into the store.
func (c *connection) CreateAuthCode(ctx context.Context, deviceCode, userCode, clientID string, expiresOn time.Time) (*database.AuthCode, error) {
	res := &database.AuthCode{}
	err := c.db.QueryRowxContext(ctx,
		`INSERT INTO device_code_auth (device_code, user_code, expires_on, approval_state, client_id)
		VALUES ($1, $2, $3, $4, $5)  RETURNING *`, deviceCode, userCode, expiresOn, database.Pending, clientID).StructScan(res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// FindAuthCodeByDeviceCode retrieves the authorization code data from the store
func (c *connection) FindAuthCodeByDeviceCode(ctx context.Context, deviceCode string) (*database.AuthCode, error) {
	authCode := &database.AuthCode{}
	err := c.db.QueryRowxContext(ctx, "SELECT * FROM device_code_auth WHERE device_code = $1", deviceCode).StructScan(authCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, database.ErrNotFound
		}
		return nil, err
	}
	return authCode, nil
}

// FindAuthCodeByUserCode retrieves the authorization code data from the store
func (c *connection) FindAuthCodeByUserCode(ctx context.Context, userCode string) (*database.AuthCode, error) {
	authCode := &database.AuthCode{}
	err := c.db.QueryRowxContext(ctx, "SELECT * FROM device_code_auth WHERE user_code = $1", userCode).StructScan(authCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, database.ErrNotFound
		}
		return nil, err
	}
	return authCode, nil
}

// UpdateAuthCode updates the authorization code data in the store
func (c *connection) UpdateAuthCode(ctx context.Context, userCode, userID string, approvalState database.AuthCodeApprovalState) error {
	res, err := c.db.ExecContext(ctx, "UPDATE device_code_auth SET approval_state=$1, user_id=$2, updated_on=now() WHERE user_code=$3",
		approvalState, userID, userCode)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return database.ErrNotFound
	}
	if rows != 1 {
		return fmt.Errorf("problem in updating auth code, expected 1 row to be affected, got %d", rows)
	}
	return nil
}

// DeleteAuthCode deletes the authorization code data from the store
func (c *connection) DeleteAuthCode(ctx context.Context, deviceCode string) error {
	res, err := c.db.ExecContext(ctx, "DELETE FROM device_code_auth WHERE device_code=$1", deviceCode)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return database.ErrNotFound
	}
	if rows != 1 {
		return fmt.Errorf("problem in deleting auth code, expected 1 row to be affected, got %d", rows)
	}
	return nil
}

func (c *connection) FindUserGithubInstallation(ctx context.Context, userID string, installationID int64) (*database.UserGithubInstallation, error) {
	res := &database.UserGithubInstallation{}
	err := c.db.QueryRowxContext(ctx, "SELECT * FROM users_github_installations WHERE user_id=$1 AND installation_id=$2", userID, installationID).StructScan(res)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, database.ErrNotFound
		}
		return nil, err
	}
	return res, nil
}

func (c *connection) UpsertUserGithubInstallation(ctx context.Context, userID string, installationID int64) error {
	_, err := c.db.ExecContext(ctx, "INSERT INTO users_github_installations (user_id, installation_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", userID, installationID)
	if err != nil {
		return err
	}
	return nil
}

func (c *connection) DeleteUserGithubInstallations(ctx context.Context, installationID int64) error {
	_, err := c.db.ExecContext(ctx, "DELETE FROM users_github_installations WHERE installation_id=$1", installationID)
	return err
}
