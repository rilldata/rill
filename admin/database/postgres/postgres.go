package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
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

func (c *connection) FindOrganizations(ctx context.Context) ([]*database.Organization, error) {
	var res []*database.Organization
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT * FROM organizations ORDER BY name")
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func (c *connection) FindOrganizationByName(ctx context.Context, name string) (*database.Organization, error) {
	res := &database.Organization{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM organizations WHERE name = $1", name).StructScan(res)
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func (c *connection) FindOrganizationByID(ctx context.Context, orgID string) (*database.Organization, error) {
	res := &database.Organization{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM organizations WHERE id = $1", orgID).StructScan(res)
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func (c *connection) InsertOrganization(ctx context.Context, name, description string) (*database.Organization, error) {
	res := &database.Organization{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "INSERT INTO organizations(name, description) VALUES ($1, $2) RETURNING *", name, description).StructScan(res)
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func (c *connection) InsertOrganizationFromSeeds(ctx context.Context, nameSeeds []string, description string) (*database.Organization, error) {
	// If this is called in a transaction, we must use savepoints to avoid aborting the whole transaction when a unique constraint is violated.
	isTx := txFromContext(ctx) != nil

	for _, name := range nameSeeds {
		// Create savepoint if in tx
		if isTx {
			_, err := c.getDB(ctx).ExecContext(ctx, "SAVEPOINT bi")
			if err != nil {
				return nil, err
			}
		}

		// Try to create the org
		org, err := c.InsertOrganization(ctx, name, description)
		if err == nil {
			return org, nil
		}

		// If the error is not a name uniqueness violation, return err
		err = parseErr(err)
		if !errors.Is(err, database.ErrNotUnique) {
			return nil, err
		}

		// Name is not unique. Continue to try the next seed.

		// If in tx, rollback to the savepoint first.
		if isTx {
			_, err := c.getDB(ctx).ExecContext(ctx, "ROLLBACK TO SAVEPOINT bi")
			if err != nil {
				return nil, err
			}
		}
	}

	// No seed was unique
	return nil, database.ErrNotUnique
}

func (c *connection) UpdateOrganization(ctx context.Context, name, description string) (*database.Organization, error) {
	res := &database.Organization{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "UPDATE organizations SET description=$1, updated_on=now() WHERE name=$2 RETURNING *", description, name).StructScan(res)
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func (c *connection) DeleteOrganization(ctx context.Context, name string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM organizations WHERE name=$1", name)
	return parseErr(err)
}

func (c *connection) FindProjects(ctx context.Context, orgName string) ([]*database.Project, error) {
	var res []*database.Project
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT p.* FROM projects p JOIN organizations o ON p.organization_id = o.id WHERE o.name=$1 ORDER BY p.name", orgName)
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func (c *connection) FindProjectByName(ctx context.Context, orgName, name string) (*database.Project, error) {
	res := &database.Project{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT p.* FROM projects p JOIN organizations o ON p.organization_id = o.id WHERE p.name=$1 AND o.name=$2", name, orgName).StructScan(res)
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func (c *connection) FindProjectByGithubURL(ctx context.Context, githubURL string) (*database.Project, error) {
	res := &database.Project{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT p.* FROM projects p WHERE p.github_url=lower($1)", githubURL).StructScan(res)
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func (c *connection) InsertProject(ctx context.Context, opts *database.InsertProjectOptions) (*database.Project, error) {
	res := &database.Project{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		INSERT INTO projects (organization_id, name, description, public, region, production_olap_driver, production_olap_dsn, production_slots, production_branch, production_variables, github_url, github_installation_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING *`,
		opts.OrganizationID, opts.Name, opts.Description, opts.Public, opts.Region, opts.ProductionOLAPDriver, opts.ProductionOLAPDSN, opts.ProductionSlots, opts.ProductionBranch, database.Variables(opts.ProductionVariables), opts.GithubURL, opts.GithubInstallationID,
	).StructScan(res)
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func (c *connection) UpdateProject(ctx context.Context, id string, opts *database.UpdateProjectOptions) (*database.Project, error) {
	res := &database.Project{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		UPDATE projects SET description=$1, public=$2, production_branch=$3, production_variables=$4, github_url=$5, github_installation_id=$6, production_deployment_id=$7, updated_on=now()
		WHERE id=$8 RETURNING *`,
		opts.Description, opts.Public, opts.ProductionBranch, database.Variables(opts.ProductionVariables), opts.GithubURL, opts.GithubInstallationID, opts.ProductionDeploymentID, id,
	).StructScan(res)
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func (c *connection) DeleteProject(ctx context.Context, id string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM projects WHERE id=$1", id)
	return parseErr(err)
}

func (c *connection) FindUsers(ctx context.Context) ([]*database.User, error) {
	var res []*database.User
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT u.* FROM users u")
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func (c *connection) FindUser(ctx context.Context, id string) (*database.User, error) {
	res := &database.User{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT u.* FROM users u WHERE u.id=$1", id).StructScan(res)
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func (c *connection) FindUserByEmail(ctx context.Context, email string) (*database.User, error) {
	res := &database.User{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT u.* FROM users u WHERE lower(u.email)=lower($1)", email).StructScan(res)
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func (c *connection) InsertUser(ctx context.Context, email, displayName, photoURL string) (*database.User, error) {
	res := &database.User{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "INSERT INTO users (email, display_name, photo_url) VALUES ($1, $2, $3) RETURNING *", email, displayName, photoURL).StructScan(res)
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func (c *connection) UpdateUser(ctx context.Context, id, displayName, photoURL string) (*database.User, error) {
	res := &database.User{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "UPDATE users SET display_name=$1, photo_url=$2, updated_on=now() WHERE id=$3 RETURNING *", displayName, photoURL, id).StructScan(res)
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func (c *connection) DeleteUser(ctx context.Context, id string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM users WHERE id=$1", id)
	return parseErr(err)
}

func (c *connection) FindUserAuthTokens(ctx context.Context, userID string) ([]*database.UserAuthToken, error) {
	var res []*database.UserAuthToken
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT t.* FROM user_auth_tokens t WHERE t.user_id=$1", userID)
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func (c *connection) FindUserAuthToken(ctx context.Context, id string) (*database.UserAuthToken, error) {
	res := &database.UserAuthToken{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT t.* FROM user_auth_tokens t WHERE t.id=$1", id).StructScan(res)
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func (c *connection) InsertUserAuthToken(ctx context.Context, opts *database.InsertUserAuthTokenOptions) (*database.UserAuthToken, error) {
	res := &database.UserAuthToken{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		INSERT INTO user_auth_tokens (id, secret_hash, user_id, display_name, auth_client_id)
		VALUES ($1, $2, $3, $4, $5) RETURNING *`,
		opts.ID, opts.SecretHash, opts.UserID, opts.DisplayName, opts.AuthClientID,
	).StructScan(res)
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func (c *connection) DeleteUserAuthToken(ctx context.Context, id string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM user_auth_tokens WHERE id=$1", id)
	return parseErr(err)
}

func (c *connection) InsertAuthCode(ctx context.Context, deviceCode, userCode, clientID string, expiresOn time.Time) (*database.AuthCode, error) {
	res := &database.AuthCode{}
	err := c.getDB(ctx).QueryRowxContext(ctx,
		`INSERT INTO device_code_auth (device_code, user_code, expires_on, approval_state, client_id)
		VALUES ($1, $2, $3, $4, $5)  RETURNING *`, deviceCode, userCode, expiresOn, database.AuthCodeStatePending, clientID).StructScan(res)
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func (c *connection) FindAuthCodeByDeviceCode(ctx context.Context, deviceCode string) (*database.AuthCode, error) {
	authCode := &database.AuthCode{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM device_code_auth WHERE device_code = $1", deviceCode).StructScan(authCode)
	if err != nil {
		return nil, parseErr(err)
	}
	return authCode, nil
}

func (c *connection) FindAuthCodeByUserCode(ctx context.Context, userCode string) (*database.AuthCode, error) {
	authCode := &database.AuthCode{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM device_code_auth WHERE user_code = $1", userCode).StructScan(authCode)
	if err != nil {
		return nil, parseErr(err)
	}
	return authCode, nil
}

func (c *connection) UpdateAuthCode(ctx context.Context, userCode, userID string, approvalState database.AuthCodeState) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "UPDATE device_code_auth SET approval_state=$1, user_id=$2, updated_on=now() WHERE user_code=$3",
		approvalState, userID, userCode)
	if err != nil {
		return parseErr(err)
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

func (c *connection) DeleteAuthCode(ctx context.Context, deviceCode string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM device_code_auth WHERE device_code=$1", deviceCode)
	if err != nil {
		return parseErr(err)
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
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM users_github_installations WHERE user_id=$1 AND installation_id=$2", userID, installationID).StructScan(res)
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func (c *connection) UpsertUserGithubInstallation(ctx context.Context, userID string, installationID int64) error {
	// TODO: Handle updated_on
	_, err := c.getDB(ctx).ExecContext(ctx, "INSERT INTO users_github_installations (user_id, installation_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", userID, installationID)
	if err != nil {
		return parseErr(err)
	}
	return nil
}

func (c *connection) DeleteUserGithubInstallations(ctx context.Context, installationID int64) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM users_github_installations WHERE installation_id=$1", installationID)
	return parseErr(err)
}

func (c *connection) FindDeployments(ctx context.Context, projectID string) ([]*database.Deployment, error) {
	var res []*database.Deployment
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT * FROM deployments d WHERE d.project_id=$1", projectID)
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func (c *connection) FindDeployment(ctx context.Context, id string) (*database.Deployment, error) {
	res := &database.Deployment{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT d.* FROM deployments d WHERE d.id=$1", id).StructScan(res)
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func (c *connection) InsertDeployment(ctx context.Context, opts *database.InsertDeploymentOptions) (*database.Deployment, error) {
	res := &database.Deployment{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		INSERT INTO deployments (project_id, slots, branch, runtime_host, runtime_instance_id, runtime_audience, status, logs)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *`,
		opts.ProjectID, opts.Slots, opts.Branch, opts.RuntimeHost, opts.RuntimeInstanceID, opts.RuntimeAudience, opts.Status, opts.Logs,
	).StructScan(res)
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func (c *connection) UpdateDeploymentStatus(ctx context.Context, id string, status database.DeploymentStatus, logs string) (*database.Deployment, error) {
	res := &database.Deployment{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "UPDATE deployments SET status=$1, logs=$2, updated_on=now() WHERE id=$3 RETURNING *", status, logs, id).StructScan(res)
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func (c *connection) DeleteDeployment(ctx context.Context, id string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM deployments WHERE id=$1", id)
	return parseErr(err)
}

func (c *connection) QueryRuntimeSlotsUsed(ctx context.Context) ([]*database.RuntimeSlotsUsed, error) {
	var res []*database.RuntimeSlotsUsed
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT d.runtime_host, SUM(d.slots) AS slots_used FROM deployments d GROUP BY d.runtime_host")
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func parseErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return database.ErrNotFound
	}
	if strings.Contains(strings.ToLower(err.Error()), "violates unique constraint") {
		return database.ErrNotUnique
	}
	return err
}

func (c *connection) FindOrganizationMembers(ctx context.Context, orgID string) ([]*database.User, error) {
	var res []*database.User
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT u.* FROM users u JOIN users_orgs_roles uor on u.id = uor.user_id WHERE uor.org_id=$1", orgID)
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func (c *connection) FindOrganizationMembersByRole(ctx context.Context, orgID, roleID string) ([]*database.User, error) {
	var res []*database.User
	err := c.getDB(ctx).SelectContext(
		ctx, &res, "SELECT u.* FROM users u JOIN users_orgs_roles uor on u.id = uor.user_id WHERE uor.org_id=$1 AND uor.org_role_id=$2", orgID, roleID)
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func (c *connection) AddOrganizationMember(ctx context.Context, orgID, userID, roleID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "INSERT INTO users_orgs_roles (user_id, org_id, org_role_id) VALUES ($1, $2, $3)", userID, orgID, roleID)
	if err != nil {
		return parseErr(err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("no rows affected when adding user to organization")
	}
	return nil
}

func (c *connection) RemoveOrganizationMember(ctx context.Context, orgID, userID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM users_orgs_roles WHERE user_id = $1 AND org_id = $2", userID, orgID)
	if err != nil {
		return parseErr(err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("no rows affected when removing user from organization")
	}
	return nil
}

func (c *connection) UpdateOrganizationMemberRole(ctx context.Context, orgID, userID, roleID string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, `UPDATE users_orgs_roles SET org_role_id = $1 WHERE user_id = $2 AND org_id = $3`,
		roleID, userID, orgID)
	return parseErr(err)
}

func (c *connection) FindProjectMembers(ctx context.Context, projectID string) ([]*database.User, error) {
	var res []*database.User
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT u.* FROM users u JOIN users_projects_roles upr on u.id = upr.user_id WHERE upr.project_id=$1", projectID)
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func (c *connection) AddProjectMember(ctx context.Context, projectID, userID, roleID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "INSERT INTO users_projects_roles (user_id, project_id, project_role_id) VALUES ($1, $2, $3)", userID, projectID, roleID)
	if err != nil {
		return parseErr(err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("no rows affected when adding user to project")
	}
	return nil
}

func (c *connection) RemoveProjectMember(ctx context.Context, projectID, userID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM users_projects_roles WHERE user_id = $1 AND project_id = $2", userID, projectID)
	if err != nil {
		return parseErr(err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("no rows affected when removing user from project")
	}
	return nil
}

func (c *connection) UpdateProjectMemberRole(ctx context.Context, projectID, userID, roleID string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, `UPDATE users_projects_roles SET project_role_id = $1 WHERE user_id = $2 AND project_id = $3`,
		roleID, userID, projectID)
	return parseErr(err)
}

func (c *connection) ResolveUserOrganizationRole(ctx context.Context, userID, orgID string) (*database.OrganizationRole, error) {
	res := &database.OrganizationRole{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		SELECT r.* FROM users_orgs_roles uor
		JOIN org_roles r ON uor.org_role_id = r.id
		WHERE uor.user_id = $1 AND uor.org_id = $2
	`, userID, orgID).StructScan(res)
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func (c *connection) ResolveUserProjectRole(ctx context.Context, userID, projectID string) (*database.ProjectRole, error) {
	res := &database.ProjectRole{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		SELECT r.* FROM users_projects_roles upr
		JOIN project_roles r ON upr.project_role_id = r.id
		WHERE upr.user_id = $1 AND upr.project_id = $2
	`, userID, projectID).StructScan(res)
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func (c *connection) ResolveUserGroupOrgRoles(ctx context.Context, userID, orgID string) ([]*database.OrganizationRole, error) {
	var res []*database.OrganizationRole
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT * FROM org_roles WHERE id IN (
			SELECT org_role_id FROM usergroups_orgs_roles uor JOIN users_usergroups uug 
			ON uor.usergroup_id = uug.group_id WHERE uug.user_id = $1 AND uor.org_id = $2
	)`, userID, orgID)
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func (c *connection) ResolveUserGroupProjectRoles(ctx context.Context, userID, projectID string) ([]*database.ProjectRole, error) {
	var res []*database.ProjectRole
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT * FROM projects_roles WHERE id IN (
			SELECT project_role_id FROM usergroups_projects_roles upr JOIN users_usergroups uug 
			ON upr.usergroup_id = uug.group_id WHERE uug.user_id = $1 AND upr.project_id = $2
	)`, userID, projectID)
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func (c *connection) FindOrganizationRole(ctx context.Context, name string) (*database.OrganizationRole, error) {
	role := &database.OrganizationRole{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM org_roles WHERE name = $1", name).Scan(role)
	if err != nil {
		return nil, parseErr(err)
	}
	return role, nil
}

func (c *connection) FindProjectRole(ctx context.Context, name string) (*database.ProjectRole, error) {
	role := &database.ProjectRole{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM ptoject_roles WHERE name = $1", name).Scan(role)
	if err != nil {
		return nil, parseErr(err)
	}
	return role, nil
}

func (c *connection) InsertUserGroup(ctx context.Context, name, orgID, description string) (*database.UserGroup, error) {
	res := &database.UserGroup{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		INSERT INTO usergroups (display_name, org_id, description) VALUES ($1, $2, $3) RETURNING *
	`, name, orgID, description).StructScan(res)
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func (c *connection) UpdateOrganizationAllUserGroup(ctx context.Context, orgID, groupID string) (*database.Organization, error) {
	res := &database.Organization{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		UPDATE organizations SET all_group_id = $1 WHERE id = $2 RETURNING *
	`, groupID, orgID).StructScan(res)
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}

func (c *connection) AddUserGroupMember(ctx context.Context, userID, groupID string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "INSERT INTO users_usergroups (user_id, group_id) VALUES ($1, $2)", userID, groupID)
	if err != nil {
		return parseErr(err)
	}
	return nil
}

func (c *connection) AddUserGroupProjectRole(ctx context.Context, groupID, projectID, roleID string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "INSERT INTO usergroups_projects_roles (usergroup_id, project_id, project_role_id) VALUES ($1, $2, $3)", groupID, projectID, roleID)
	if err != nil {
		return parseErr(err)
	}
	return nil
}

func (c *connection) FindUserGroups(ctx context.Context, userID, orgID string) ([]*database.UserGroup, error) {
	var res []*database.UserGroup
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT ug.* FROM usergroups ug JOIN users_usergroups uug ON ug.id = uug.group_id
		WHERE uug.user_id = $1 AND ug.org_id = $2
	`, userID, orgID)
	if err != nil {
		return nil, parseErr(err)
	}
	return res, nil
}
