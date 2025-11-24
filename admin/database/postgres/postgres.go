package postgres

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/admin/database"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"

	// Load postgres driver
	_ "github.com/jackc/pgx/v4/stdlib"
)

func init() {
	database.Register("postgres", driver{})
}

type driver struct{}

func (d driver) Open(dsn string, encKeyring []*database.EncryptionKey) (database.DB, error) {
	db, err := otelsql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = otelsql.RegisterDBStatsMetrics(db, otelsql.WithAttributes(semconv.DBSystemPostgreSQL))
	if err != nil {
		return nil, err
	}

	dbx := sqlx.NewDb(db, "pgx")
	return &connection{db: dbx, encKeyring: encKeyring}, nil
}

type connection struct {
	db         *sqlx.DB
	encKeyring []*database.EncryptionKey
}

func (c *connection) Close() error {
	return c.db.Close()
}

func (c *connection) FindOrganizations(ctx context.Context, afterName string, limit int) ([]*database.Organization, error) {
	var res []*database.Organization
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT * FROM orgs WHERE lower(name) > lower($1) ORDER BY lower(name) LIMIT $2", afterName, limit)
	if err != nil {
		return nil, parseErr("orgs", err)
	}
	return res, nil
}

func (c *connection) FindOrganizationsForUser(ctx context.Context, userID, afterName string, limit int) ([]*database.Organization, error) {
	var res []*database.Organization
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT o.* FROM orgs o
		WHERE o.id IN (SELECT uor.org_id FROM users_orgs_roles uor WHERE uor.user_id = $1)
		AND lower(o.name) > lower($2) ORDER BY lower(o.name) LIMIT $3
	`, userID, afterName, limit)
	if err != nil {
		return nil, parseErr("orgs", err)
	}
	return res, nil
}

func (c *connection) FindOrganization(ctx context.Context, orgID string) (*database.Organization, error) {
	res := &database.Organization{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM orgs WHERE id = $1", orgID).StructScan(res)
	if err != nil {
		return nil, parseErr("org", err)
	}
	return res, nil
}

func (c *connection) FindOrganizationByName(ctx context.Context, name string) (*database.Organization, error) {
	res := &database.Organization{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM orgs WHERE lower(name)=lower($1)", name).StructScan(res)
	if err != nil {
		return nil, parseErr("org", err)
	}
	return res, nil
}

func (c *connection) CountProjectsForOrganization(ctx context.Context, orgID string) (int, error) {
	var count int
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT COUNT(*) FROM projects WHERE org_id=$1", orgID).Scan(&count)
	if err != nil {
		return 0, parseErr("projects", err)
	}
	return count, nil
}

func (c *connection) FindOrganizationByCustomDomain(ctx context.Context, domain string) (*database.Organization, error) {
	res := &database.Organization{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM orgs WHERE lower(custom_domain)=lower($1)", domain).StructScan(res)
	if err != nil {
		return nil, parseErr("org", err)
	}
	return res, nil
}

func (c *connection) CheckOrganizationHasPublicProjects(ctx context.Context, orgID string) (bool, error) {
	var res bool
	err := c.getDB(ctx).QueryRowxContext(ctx,
		"SELECT EXISTS (SELECT 1 FROM projects p WHERE p.org_id = $1 AND p.public = true limit 1)", orgID).Scan(&res)
	if err != nil {
		return false, parseErr("check", err)
	}
	return res, nil
}

func (c *connection) InsertOrganization(ctx context.Context, opts *database.InsertOrganizationOptions) (*database.Organization, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.Organization{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `INSERT INTO orgs(name, display_name, description, logo_asset_id, favicon_asset_id, thumbnail_asset_id, custom_domain, default_project_role_id, quota_projects, quota_deployments, quota_slots_total, quota_slots_per_deployment, quota_outstanding_invites, quota_storage_limit_bytes_per_deployment, billing_customer_id, payment_customer_id, billing_email, created_by_user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18) RETURNING *`,
		opts.Name, opts.DisplayName, opts.Description, opts.LogoAssetID, opts.FaviconAssetID, opts.ThumbnailAssetID, opts.CustomDomain, opts.DefaultProjectRoleID, opts.QuotaProjects, opts.QuotaDeployments, opts.QuotaSlotsTotal, opts.QuotaSlotsPerDeployment, opts.QuotaOutstandingInvites, opts.QuotaStorageLimitBytesPerDeployment, opts.BillingCustomerID, opts.PaymentCustomerID, opts.BillingEmail, opts.CreatedByUserID).StructScan(res)
	if err != nil {
		return nil, parseErr("org", err)
	}
	return res, nil
}

func (c *connection) DeleteOrganization(ctx context.Context, name string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM orgs WHERE lower(name)=lower($1)", name)
	return checkDeleteRow("org", res, err)
}

func (c *connection) UpdateOrganization(ctx context.Context, id string, opts *database.UpdateOrganizationOptions) (*database.Organization, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.Organization{}
	err := c.getDB(ctx).QueryRowxContext(ctx,
		`UPDATE orgs SET name=$1, display_name=$2, description=$3, logo_asset_id=$4, favicon_asset_id=$5, thumbnail_asset_id=$6, custom_domain=$7, default_project_role_id=$8, quota_projects=$9, quota_deployments=$10, quota_slots_total=$11, quota_slots_per_deployment=$12, quota_outstanding_invites=$13, quota_storage_limit_bytes_per_deployment=$14, billing_customer_id=$15, payment_customer_id=$16, billing_email=$17, created_by_user_id=$18, billing_plan_name=$19, billing_plan_display_name=$20, updated_on=now() WHERE id=$21 RETURNING *`,
		opts.Name, opts.DisplayName, opts.Description, opts.LogoAssetID, opts.FaviconAssetID, opts.ThumbnailAssetID, opts.CustomDomain, opts.DefaultProjectRoleID, opts.QuotaProjects, opts.QuotaDeployments, opts.QuotaSlotsTotal, opts.QuotaSlotsPerDeployment, opts.QuotaOutstandingInvites, opts.QuotaStorageLimitBytesPerDeployment, opts.BillingCustomerID, opts.PaymentCustomerID, opts.BillingEmail, opts.CreatedByUserID, opts.BillingPlanName, opts.BillingPlanDisplayName, id).StructScan(res)
	if err != nil {
		return nil, parseErr("org", err)
	}
	return res, nil
}

func (c *connection) FindOrganizationWhitelistedDomainForOrganizationWithJoinedRoleNames(ctx context.Context, orgID string) ([]*database.OrganizationWhitelistedDomainWithJoinedRoleNames, error) {
	var res []*database.OrganizationWhitelistedDomainWithJoinedRoleNames
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT oad.domain, r.name FROM orgs_autoinvite_domains oad JOIN org_roles r ON r.id = oad.org_role_id WHERE oad.org_id=$1", orgID)
	if err != nil {
		return nil, parseErr("org whitelist domains", err)
	}
	return res, nil
}

func (c *connection) FindOrganizationWhitelistedDomainsForDomain(ctx context.Context, domain string) ([]*database.OrganizationWhitelistedDomain, error) {
	var res []*database.OrganizationWhitelistedDomain
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT * FROM orgs_autoinvite_domains WHERE lower(domain)=lower($1)", domain)
	if err != nil {
		return nil, parseErr("org whitelist domains", err)
	}
	return res, nil
}

func (c *connection) FindOrganizationWhitelistedDomain(ctx context.Context, orgID, domain string) (*database.OrganizationWhitelistedDomain, error) {
	res := &database.OrganizationWhitelistedDomain{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM orgs_autoinvite_domains WHERE org_id=$1 AND lower(domain)=lower($2)", orgID, domain).StructScan(res)
	if err != nil {
		return nil, parseErr("org whitelist domain", err)
	}
	return res, nil
}

func (c *connection) InsertOrganizationWhitelistedDomain(ctx context.Context, opts *database.InsertOrganizationWhitelistedDomainOptions) (*database.OrganizationWhitelistedDomain, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.OrganizationWhitelistedDomain{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `INSERT INTO orgs_autoinvite_domains(org_id, org_role_id, domain) VALUES ($1, $2, $3) RETURNING *`, opts.OrgID, opts.OrgRoleID, opts.Domain).StructScan(res)
	if err != nil {
		return nil, parseErr("org whitelist domain", err)
	}
	return res, nil
}

func (c *connection) DeleteOrganizationWhitelistedDomain(ctx context.Context, id string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM orgs_autoinvite_domains WHERE id=$1", id)
	return checkDeleteRow("org whitelist domain", res, err)
}

func (c *connection) FindInactiveOrganizations(ctx context.Context) ([]*database.Organization, error) {
	// TODO: This definition may change, but for now, we are considering an organization as inactive if it has no users
	res := []*database.Organization{}
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT o.* FROM orgs o
		WHERE now() - o.updated_on > INTERVAL '1 DAY'
		AND NOT EXISTS ( SELECT 1 FROM users_orgs_roles uor WHERE uor.org_id = o.id )
	`)
	if err != nil {
		return nil, parseErr("orgs", err)
	}
	return res, nil
}

func (c *connection) FindProjects(ctx context.Context, afterName string, limit int) ([]*database.Project, error) {
	var res []*projectDTO
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT p.* FROM projects p WHERE lower(name) > lower($1) ORDER BY lower(p.name) LIMIT $2", afterName, limit)
	if err != nil {
		return nil, parseErr("projects", err)
	}
	return c.projectsFromDTOs(res)
}

func (c *connection) FindProjectsByVersion(ctx context.Context, version, afterName string, limit int) ([]*database.Project, error) {
	var res []*projectDTO
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT p.* FROM projects p WHERE p.prod_version = $1 AND lower(name) > lower($2) ORDER BY lower(p.name) LIMIT $3", version, afterName, limit)
	if err != nil {
		return nil, parseErr("projects", err)
	}
	return c.projectsFromDTOs(res)
}

func (c *connection) FindProjectPathsByPattern(ctx context.Context, namePattern, afterName string, limit int) ([]string, error) {
	var res []string
	err := c.getDB(ctx).SelectContext(ctx, &res, `SELECT concat(o.name,'/',p.name) as project_name FROM projects p JOIN orgs o ON p.org_id = o.id
	WHERE concat(o.name,'/',p.name) ilike $1 AND concat(o.name,'/',p.name) > $2
	ORDER BY project_name
	LIMIT $3`, namePattern, afterName, limit)
	if err != nil {
		return nil, parseErr("projects", err)
	}
	return res, nil
}

func (c *connection) FindProjectPathsByPatternAndAnnotations(ctx context.Context, namePattern, afterName string, annotationKeys []string, annotationPairs map[string]string, limit int) ([]string, error) {
	if annotationKeys == nil {
		annotationKeys = []string{}
	}
	if annotationPairs == nil {
		annotationPairs = map[string]string{}
	}

	var res []string
	err := c.getDB(ctx).SelectContext(ctx, &res, `SELECT concat(o.name,'/',p.name) as project_name FROM projects p JOIN orgs o ON p.org_id = o.id
	WHERE concat(o.name,'/',p.name) ilike $1 AND concat(o.name,'/',p.name) > $2 AND p.annotations ?& $3 AND p.annotations @> $4
	ORDER BY project_name
	LIMIT $5`, namePattern, afterName, annotationKeys, annotationPairs, limit)
	if err != nil {
		return nil, parseErr("projects", err)
	}
	return res, nil
}

func (c *connection) FindProjectsForUser(ctx context.Context, userID string) ([]*database.Project, error) {
	var res []*projectDTO
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT * FROM projects
		WHERE id IN (
			SELECT upr.project_id FROM users_projects_roles upr WHERE upr.user_id = $1
			UNION
			SELECT ugpr.project_id FROM usergroups_projects_roles ugpr JOIN usergroups_users ugu ON ugpr.usergroup_id = ugu.usergroup_id WHERE ugu.user_id = $1
		)
	`, userID)
	if err != nil {
		return nil, parseErr("projects", err)
	}
	return c.projectsFromDTOs(res)
}

// FindProjectsForUserAndFingerprint returns projects for the user based on fingerprint.
// The fingerprint is simply git_remote + subpath for git based projects.
// For archive projects it is directory_name.
func (c *connection) FindProjectsForUserAndFingerprint(ctx context.Context, userID, directoryName, gitRemote, subpath, rillMgdRemote string) ([]*database.Project, error) {
	// Shouldn't happen, but just to be safe and not return all projects.
	if directoryName == "" && gitRemote == "" {
		return nil, nil
	}

	args := []any{userID, directoryName, gitRemote, subpath, rillMgdRemote}
	qry := `
		SELECT p.* FROM projects p
		WHERE p.id IN (
			SELECT upr.project_id FROM users_projects_roles upr WHERE upr.user_id = $1
			UNION
			SELECT ugpr.project_id FROM usergroups_projects_roles ugpr JOIN usergroups_users ugu ON ugpr.usergroup_id = ugu.usergroup_id WHERE ugu.user_id = $1
		)
		AND (
			(p.archive_asset_id IS NULL AND p.git_remote = $3 AND p.subpath = $4)
			OR
			(p.archive_asset_id IS NULL AND p.git_remote = $5)
			OR
			(p.archive_asset_id IS NOT NULL AND p.directory_name = $2)
		)
	`

	var res []*projectDTO
	err := c.getDB(ctx).SelectContext(ctx, &res, qry, args...)
	if err != nil {
		return nil, parseErr("projects", err)
	}
	return c.projectsFromDTOs(res)
}

func (c *connection) FindProjectsForOrganization(ctx context.Context, orgID, afterProjectName string, limit int) ([]*database.Project, error) {
	var res []*projectDTO
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT p.* FROM projects p
		WHERE p.org_id=$1 AND lower(p.name) > lower($2)
		ORDER BY lower(p.name) LIMIT $3
	`, orgID, afterProjectName, limit)
	if err != nil {
		return nil, parseErr("projects", err)
	}
	return c.projectsFromDTOs(res)
}

func (c *connection) FindProjectsForOrgAndUser(ctx context.Context, orgID, userID string, includePublic bool, afterProjectName string, limit int) ([]*database.Project, error) {
	var qry strings.Builder
	qry.WriteString("SELECT p.* FROM projects p WHERE p.org_id = $1 AND lower(p.name) > lower($2) AND (")
	if includePublic {
		qry.WriteString("p.public = true OR ")
	}
	qry.WriteString(`p.id IN (
		SELECT upr.project_id FROM users_projects_roles upr WHERE upr.user_id = $3
		UNION
		SELECT ugpr.project_id FROM usergroups_projects_roles ugpr JOIN usergroups_users uug ON ugpr.usergroup_id = uug.usergroup_id WHERE uug.user_id = $3
	)`)
	qry.WriteString(") ORDER BY lower(p.name) LIMIT $4")

	var res []*projectDTO
	err := c.getDB(ctx).SelectContext(ctx, &res, qry.String(), orgID, afterProjectName, userID, limit)
	if err != nil {
		return nil, parseErr("projects", err)
	}
	return c.projectsFromDTOs(res)
}

func (c *connection) FindPublicProjectsInOrganization(ctx context.Context, orgID, afterProjectName string, limit int) ([]*database.Project, error) {
	var res []*projectDTO
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT p.* FROM projects p
		WHERE p.org_id = $1 AND p.public = true AND lower(p.name) > lower($2)
		ORDER BY lower(p.name) LIMIT $3
	`, orgID, afterProjectName, limit)
	if err != nil {
		return nil, parseErr("projects", err)
	}
	return c.projectsFromDTOs(res)
}

func (c *connection) FindProjectsByGitRemote(ctx context.Context, remote string) ([]*database.Project, error) {
	var res []*projectDTO
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT p.* FROM projects p WHERE lower(p.git_remote)=lower($1)", remote)
	if err != nil {
		return nil, parseErr("projects", err)
	}
	return c.projectsFromDTOs(res)
}

func (c *connection) FindProjectsByGithubInstallationID(ctx context.Context, id int64) ([]*database.Project, error) {
	var res []*projectDTO
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT p.* FROM projects p WHERE p.github_installation_id=$1", id)
	if err != nil {
		return nil, parseErr("projects", err)
	}
	return c.projectsFromDTOs(res)
}

func (c *connection) FindProject(ctx context.Context, id string) (*database.Project, error) {
	res := &projectDTO{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM projects WHERE id=$1", id).StructScan(res)
	if err != nil {
		return nil, parseErr("project", err)
	}
	return c.projectFromDTO(res)
}

func (c *connection) FindProjectByName(ctx context.Context, orgName, name string) (*database.Project, error) {
	res := &projectDTO{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT p.* FROM projects p JOIN orgs o ON p.org_id = o.id WHERE lower(p.name)=lower($1) AND lower(o.name)=lower($2)", name, orgName).StructScan(res)
	if err != nil {
		return nil, parseErr("project", err)
	}
	return c.projectFromDTO(res)
}

func (c *connection) FindProjectsByNameAndUser(ctx context.Context, name, userID string) ([]*database.Project, error) {
	var res []*projectDTO
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT * FROM projects
		WHERE id IN (
			SELECT upr.project_id FROM users_projects_roles upr WHERE upr.user_id = $1
			UNION
			SELECT ugpr.project_id FROM usergroups_projects_roles ugpr JOIN usergroups_users ugu ON ugpr.usergroup_id = ugu.usergroup_id WHERE ugu.user_id = $1
		) AND lower(name)=lower($2)
	`, userID, name)
	if err != nil {
		return nil, parseErr("projects", err)
	}
	return c.projectsFromDTOs(res)
}

func (c *connection) InsertProject(ctx context.Context, opts *database.InsertProjectOptions) (*database.Project, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &projectDTO{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		INSERT INTO projects (
			org_id,
			name,
			description,
			public,
			created_by_user_id,
			directory_name,
			provisioner,
			prod_slots,
			subpath,
			prod_branch,
			archive_asset_id,
			git_remote,
			github_installation_id,
			github_repo_id,
			managed_git_repo_id,
			prod_ttl_seconds,
			prod_version,
			dev_slots,
			dev_ttl_seconds
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19) RETURNING *`,
		opts.OrganizationID,
		opts.Name,
		opts.Description,
		opts.Public,
		opts.CreatedByUserID,
		opts.DirectoryName,
		opts.Provisioner,
		opts.ProdSlots,
		opts.Subpath,
		opts.ProdBranch,
		opts.ArchiveAssetID,
		opts.GitRemote,
		opts.GithubInstallationID,
		opts.GithubRepoID,
		opts.ManagedGitRepoID,
		opts.ProdTTLSeconds,
		opts.ProdVersion,
		opts.DevSlots,
		opts.DevTTLSeconds,
	).StructScan(res)
	if err != nil {
		return nil, parseErr("project", err)
	}
	return c.projectFromDTO(res)
}

func (c *connection) DeleteProject(ctx context.Context, id string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM projects WHERE id=$1", id)
	return checkDeleteRow("project", res, err)
}

func (c *connection) UpdateProject(ctx context.Context, id string, opts *database.UpdateProjectOptions) (*database.Project, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}
	if opts.Annotations == nil {
		opts.Annotations = make(map[string]string, 0)
	}

	res := &projectDTO{}
	err := c.getDB(ctx).QueryRowxContext(
		ctx,
		`
		UPDATE projects
		SET
			name = $1,
			description = $2,
			public = $3,
			directory_name = $4,
			prod_branch = $5,
			git_remote = $6,
			github_installation_id = $7,
			github_repo_id = $8,
			managed_git_repo_id = $9,
			archive_asset_id = $10,
			prod_deployment_id = $11,
			provisioner = $12,
			prod_slots = $13,
			subpath = $14,
			prod_ttl_seconds = $15,
			annotations = $16,
			prod_version = $17,
			dev_slots = $18,
			dev_ttl_seconds = $19,
			updated_on = now()
		WHERE id = $20
		RETURNING *
		`,
		opts.Name,
		opts.Description,
		opts.Public,
		opts.DirectoryName,
		opts.ProdBranch,
		opts.GitRemote,
		opts.GithubInstallationID,
		opts.GithubRepoID,
		opts.ManagedGitRepoID,
		opts.ArchiveAssetID,
		opts.ProdDeploymentID,
		opts.Provisioner,
		opts.ProdSlots,
		opts.Subpath,
		opts.ProdTTLSeconds,
		opts.Annotations,
		opts.ProdVersion,
		opts.DevSlots,
		opts.DevTTLSeconds,
		id,
	).StructScan(res)
	if err != nil {
		return nil, parseErr("project", err)
	}
	return c.projectFromDTO(res)
}

func (c *connection) CountProjectsQuotaUsage(ctx context.Context, orgID string) (*database.ProjectsQuotaUsage, error) {
	res := &database.ProjectsQuotaUsage{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		WITH t1 AS (SELECT * FROM projects WHERE org_id = $1)
		SELECT
			(SELECT COUNT(*) FROM t1) AS projects,
			(SELECT COUNT(*) FROM deployments d WHERE d.project_id IN (SELECT id FROM t1)) AS deployments,
			(SELECT COALESCE(SUM(prod_slots), 0) FROM t1) AS slots
	`, orgID).StructScan(res)
	if err != nil {
		return nil, parseErr("projects quota usage", err)
	}
	return res, nil
}

func (c *connection) FindProjectWhitelistedDomainForProjectWithJoinedRoleNames(ctx context.Context, projectID string) ([]*database.ProjectWhitelistedDomainWithJoinedRoleNames, error) {
	var res []*database.ProjectWhitelistedDomainWithJoinedRoleNames
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT pad.domain, r.name FROM projects_autoinvite_domains pad JOIN project_roles r ON r.id = pad.project_role_id WHERE pad.project_id=$1", projectID)
	if err != nil {
		return nil, parseErr("project whitelist domains", err)
	}
	return res, nil
}

func (c *connection) FindProjectWhitelistedDomainsForDomain(ctx context.Context, domain string) ([]*database.ProjectWhitelistedDomain, error) {
	var res []*database.ProjectWhitelistedDomain
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT * FROM projects_autoinvite_domains WHERE lower(domain)=lower($1)", domain)
	if err != nil {
		return nil, parseErr("project whitelist domains", err)
	}
	return res, nil
}

func (c *connection) FindProjectWhitelistedDomain(ctx context.Context, projectID, domain string) (*database.ProjectWhitelistedDomain, error) {
	res := &database.ProjectWhitelistedDomain{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM projects_autoinvite_domains WHERE project_id=$1 AND lower(domain)=lower($2)", projectID, domain).StructScan(res)
	if err != nil {
		return nil, parseErr("project whitelist domain", err)
	}
	return res, nil
}

func (c *connection) InsertProjectWhitelistedDomain(ctx context.Context, opts *database.InsertProjectWhitelistedDomainOptions) (*database.ProjectWhitelistedDomain, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.ProjectWhitelistedDomain{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `INSERT INTO projects_autoinvite_domains(project_id, project_role_id, domain) VALUES ($1, $2, $3) RETURNING *`, opts.ProjectID, opts.ProjectRoleID, opts.Domain).StructScan(res)
	if err != nil {
		return nil, parseErr("project whitelist domain", err)
	}
	return res, nil
}

func (c *connection) DeleteProjectWhitelistedDomain(ctx context.Context, id string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM projects_autoinvite_domains WHERE id=$1", id)
	return checkDeleteRow("project whitelist domain", res, err)
}

func (c *connection) FindDeployments(ctx context.Context, afterID string, limit int) ([]*database.Deployment, error) {
	var qry strings.Builder
	var args []any
	qry.WriteString("SELECT d.* FROM deployments d ")
	if afterID != "" {
		qry.WriteString("WHERE d.id > $1 ORDER BY d.id LIMIT $2")
		args = []any{afterID, limit}
	} else {
		qry.WriteString("ORDER BY d.id LIMIT $1")
		args = []any{limit}
	}
	var res []*database.Deployment
	err := c.getDB(ctx).SelectContext(ctx, &res, qry.String(), args...)
	if err != nil {
		return nil, parseErr("deployments", err)
	}
	return res, nil
}

// FindExpiredDeployments returns all the deployments which are expired as per ttl
func (c *connection) FindExpiredDeployments(ctx context.Context) ([]*database.Deployment, error) {
	var res []*database.Deployment
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT d.* FROM deployments d
		JOIN projects p ON d.project_id = p.id
		WHERE d.status != $1
		AND ((p.prod_ttl_seconds IS NOT NULL AND d.used_on + p.prod_ttl_seconds * interval '1 second' < now())
		OR (d.environment = 'dev' AND p.dev_ttl_seconds IS NOT NULL AND d.used_on + p.dev_ttl_seconds * interval '1 second' < now()))
	`, database.DeploymentStatusStopped)
	if err != nil {
		return nil, parseErr("deployments", err)
	}
	return res, nil
}

func (c *connection) FindDeploymentsForProject(ctx context.Context, projectID string) ([]*database.Deployment, error) {
	var res []*database.Deployment
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT * FROM deployments d WHERE d.project_id=$1", projectID)
	if err != nil {
		return nil, parseErr("deployments", err)
	}
	return res, nil
}

func (c *connection) FindDeployment(ctx context.Context, id string) (*database.Deployment, error) {
	res := &database.Deployment{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT d.* FROM deployments d WHERE d.id=$1", id).StructScan(res)
	if err != nil {
		return nil, parseErr("deployment", err)
	}
	return res, nil
}

func (c *connection) FindDeploymentByInstanceID(ctx context.Context, instanceID string) (*database.Deployment, error) {
	res := &database.Deployment{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM deployments d WHERE d.runtime_instance_id=$1", instanceID).StructScan(res)
	if err != nil {
		return nil, parseErr("deployment", err)
	}
	return res, nil
}

func (c *connection) InsertDeployment(ctx context.Context, opts *database.InsertDeploymentOptions) (*database.Deployment, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.Deployment{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		INSERT INTO deployments (project_id, owner_user_id, environment, branch, runtime_host, runtime_instance_id, runtime_audience, status, status_message, desired_status, desired_status_updated_on)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, now()) RETURNING *`,
		opts.ProjectID, opts.OwnerUserID, opts.Environment, opts.Branch, opts.RuntimeHost, opts.RuntimeInstanceID, opts.RuntimeAudience, opts.Status, opts.StatusMessage, opts.DesiredStatus,
	).StructScan(res)
	if err != nil {
		return nil, parseErr("deployment", err)
	}
	return res, nil
}

func (c *connection) DeleteDeployment(ctx context.Context, id string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM deployments WHERE id=$1", id)
	return checkDeleteRow("deployment", res, err)
}

func (c *connection) UpdateDeployment(ctx context.Context, id string, opts *database.UpdateDeploymentOptions) (*database.Deployment, error) {
	res := &database.Deployment{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		UPDATE deployments
		SET branch=$1, runtime_host=$2, runtime_instance_id=$3, runtime_audience=$4, status=$5, status_message=$6, updated_on=now()
		WHERE id=$7 RETURNING *`,
		opts.Branch, opts.RuntimeHost, opts.RuntimeInstanceID, opts.RuntimeAudience, opts.Status, opts.StatusMessage, id,
	).StructScan(res)
	if err != nil {
		return nil, parseErr("deployment", err)
	}
	return res, nil
}

func (c *connection) UpdateDeploymentStatus(ctx context.Context, id string, status database.DeploymentStatus, message string) (*database.Deployment, error) {
	res := &database.Deployment{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "UPDATE deployments SET status=$1, status_message=$2, updated_on=now() WHERE id=$3 RETURNING *", status, message, id).StructScan(res)
	if err != nil {
		return nil, parseErr("deployment", err)
	}
	return res, nil
}

func (c *connection) UpdateDeploymentDesiredStatus(ctx context.Context, id string, desiredStatus database.DeploymentStatus) (*database.Deployment, error) {
	res := &database.Deployment{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "UPDATE deployments SET desired_status=$1, desired_status_updated_on=now(), updated_on=now() WHERE id=$2 RETURNING *", desiredStatus, id).StructScan(res)
	if err != nil {
		return nil, parseErr("deployment", err)
	}
	return res, nil
}

func (c *connection) UpdateDeploymentUsedOn(ctx context.Context, ids []string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "UPDATE deployments SET used_on=now() WHERE id = any($1)", ids)
	if err != nil {
		return parseErr("deployment", err)
	}
	return nil
}

func (c *connection) UpsertStaticRuntimeAssignment(ctx context.Context, id, host string, slots int) error {
	// If slots is 0, delete the assignment if it exists (may not exist due to idempotence, so not checking the affected row count).
	if slots == 0 {
		_, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM static_runtime_assignments WHERE resource_id=$1", id)
		if err != nil {
			return parseErr("slots used", err)
		}
		return nil
	}

	// Upsert the assignment.
	_, err := c.getDB(ctx).ExecContext(ctx, "INSERT INTO static_runtime_assignments (resource_id, host, slots) VALUES ($1, $2, $3) ON CONFLICT (resource_id) DO UPDATE SET slots = EXCLUDED.slots", id, host, slots)
	if err != nil {
		return parseErr("slots used", err)
	}
	return nil
}

func (c *connection) DeleteStaticRuntimeAssignment(ctx context.Context, id string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM static_runtime_assignments WHERE resource_id=$1", id)
	if err != nil {
		// Not using checkDeleteRow because this is operation must be idempotent, so the row may not exist.
		return parseErr("slots used", err)
	}
	return nil
}

func (c *connection) ResolveStaticRuntimeSlotsUsed(ctx context.Context) ([]*database.StaticRuntimeSlotsUsed, error) {
	var res []*database.StaticRuntimeSlotsUsed
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT host, SUM(slots) as slots FROM static_runtime_assignments GROUP BY host ORDER BY host")
	if err != nil {
		return nil, parseErr("slots used", err)
	}
	return res, nil
}

func (c *connection) FindUsers(ctx context.Context) ([]*database.User, error) {
	var res []*database.User
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT u.* FROM users u")
	if err != nil {
		return nil, parseErr("users", err)
	}
	return res, nil
}

func (c *connection) FindUser(ctx context.Context, id string) (*database.User, error) {
	res := &database.User{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT u.* FROM users u WHERE u.id=$1", id).StructScan(res)
	if err != nil {
		return nil, parseErr("user", err)
	}
	return res, nil
}

func (c *connection) FindUserByEmail(ctx context.Context, email string) (*database.User, error) {
	res := &database.User{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT u.* FROM users u WHERE lower(u.email)=lower($1)", email).StructScan(res)
	if err != nil {
		return nil, parseErr("user", err)
	}
	return res, nil
}

func (c *connection) FindUserWithAttributes(ctx context.Context, userID, orgID string) (*database.User, map[string]any, error) {
	var dto userWithAttributesDTO
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		SELECT u.*, uor.attributes
		FROM users u
		LEFT JOIN users_orgs_roles uor ON u.id = uor.user_id AND uor.org_id = $2
		WHERE u.id = $1
	`, userID, orgID).StructScan(&dto)
	if err != nil {
		return nil, nil, parseErr("user with org attributes", err)
	}
	user, attributes, err := dto.userWithAttributesFromDTO()
	if err != nil {
		return nil, nil, err
	}
	return user, attributes, nil
}

func (c *connection) FindUsersByEmailPattern(ctx context.Context, emailPattern, afterEmail string, limit int) ([]*database.User, error) {
	var res []*database.User
	err := c.getDB(ctx).SelectContext(ctx, &res, `SELECT u.* FROM users u
	WHERE lower(u.email) LIKE lower($1) AND lower(u.email) > lower($2)
	ORDER BY lower(u.email) LIMIT $3`, emailPattern, afterEmail, limit)
	if err != nil {
		return nil, parseErr("users", err)
	}
	return res, nil
}

// SearchProjectUsers searches for users that have access to the project.
func (c *connection) SearchProjectUsers(ctx context.Context, projectID, emailQuery, afterEmail string, limit int) ([]*database.User, error) {
	var res []*database.User
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT u.* FROM users u
		WHERE u.id IN (
			SELECT upr.user_id FROM users_projects_roles upr WHERE upr.project_id=$1
			UNION
			SELECT ugu.user_id FROM usergroups_projects_roles ugpr JOIN usergroups_users ugu ON ugpr.usergroup_id = ugu.usergroup_id WHERE ugpr.project_id=$1
		)
		AND lower(u.email) LIKE lower($2)
		AND lower(u.email) > lower($3)
		ORDER BY lower(u.email) ASC LIMIT $4`, projectID, emailQuery, afterEmail, limit)
	if err != nil {
		return nil, parseErr("users", err)
	}
	return res, nil
}

func (c *connection) InsertUser(ctx context.Context, opts *database.InsertUserOptions) (*database.User, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.User{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "INSERT INTO users (email, display_name, photo_url, quota_trial_orgs, quota_singleuser_orgs, superuser) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *", opts.Email, opts.DisplayName, opts.PhotoURL, opts.QuotaTrialOrgs, opts.QuotaSingleuserOrgs, opts.Superuser).StructScan(res)
	if err != nil {
		return nil, parseErr("user", err)
	}
	return res, nil
}

func (c *connection) CheckUsersEmpty(ctx context.Context) (bool, error) {
	var res bool
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT NOT EXISTS (SELECT 1 FROM users limit 1) ").Scan(&res)
	if err != nil {
		return false, parseErr("check", err)
	}
	return res, nil
}

func (c *connection) DeleteUser(ctx context.Context, id string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM users WHERE id=$1", id)
	return checkDeleteRow("user", res, err)
}

func (c *connection) UpdateUser(ctx context.Context, id string, opts *database.UpdateUserOptions) (*database.User, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.User{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "UPDATE users SET display_name=$2, photo_url=$3, github_username=$4, github_token=$5, github_token_expires_on=$6, github_refresh_token=$7, quota_singleuser_orgs=$8, quota_trial_orgs=$9, preference_time_zone=$10, updated_on=now() WHERE id=$1 RETURNING *",
		id,
		opts.DisplayName,
		opts.PhotoURL,
		opts.GithubUsername,
		opts.GithubToken,
		opts.GithubTokenExpiresOn,
		opts.GithubRefreshToken,
		opts.QuotaSingleuserOrgs,
		opts.QuotaTrialOrgs,
		opts.PreferenceTimeZone).StructScan(res)
	if err != nil {
		return nil, parseErr("user", err)
	}
	return res, nil
}

func (c *connection) UpdateUserActiveOn(ctx context.Context, ids []string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "UPDATE users SET active_on=now() WHERE id=ANY($1)", ids)
	if err != nil {
		return parseErr("user", err)
	}
	return nil
}

func (c *connection) CheckUserIsAnOrganizationMember(ctx context.Context, userID, orgID string) (bool, error) {
	var res bool
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT EXISTS (SELECT 1 FROM users_orgs_roles WHERE user_id=$1 AND org_id=$2)", userID, orgID).Scan(&res)
	if err != nil {
		return false, parseErr("check", err)
	}
	return res, nil
}

func (c *connection) CheckUserIsAProjectMember(ctx context.Context, userID, projectID string) (bool, error) {
	var res bool
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT EXISTS (SELECT 1 FROM users_projects_roles WHERE user_id=$1 AND project_id=$2)", userID, projectID).Scan(&res)
	if err != nil {
		return false, parseErr("check", err)
	}
	return res, nil
}

func (c *connection) GetCurrentTrialOrgCount(ctx context.Context, userID string) (int, error) {
	var count int
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT current_trial_orgs_count FROM users WHERE id=$1", userID).Scan(&count)
	if err != nil {
		return 0, parseErr("org count", err)
	}
	return count, nil
}

func (c *connection) IncrementCurrentTrialOrgCount(ctx context.Context, userID string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "UPDATE users SET current_trial_orgs_count = current_trial_orgs_count + 1 WHERE id=$1", userID)
	if err != nil {
		return parseErr("org count", err)
	}
	return nil
}

func (c *connection) FindUsergroupsForOrganizationAndUser(ctx context.Context, orgID, userID, afterName string, limit int) ([]*database.Usergroup, error) {
	var res []*database.Usergroup
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT ug.* FROM usergroups ug
		WHERE ug.org_id = $1 AND ug.id IN (
			SELECT uug.usergroup_id FROM usergroups_users uug WHERE uug.user_id = $2
		) AND lower(ug.name) > lower($3)
		ORDER BY lower(ug.name) LIMIT $4
	`, orgID, userID, afterName, limit)
	if err != nil {
		return nil, parseErr("usergroups", err)
	}
	return res, nil
}

func (c *connection) FindUsergroupByName(ctx context.Context, orgName, name string) (*database.Usergroup, error) {
	res := &database.Usergroup{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		SELECT ug.* FROM usergroups ug JOIN orgs o ON ug.org_id = o.id
		WHERE lower(ug.name)=lower($1) AND lower(o.name)=lower($2)
	`, name, orgName).StructScan(res)
	if err != nil {
		return nil, parseErr("usergroup", err)
	}
	return res, nil
}

func (c *connection) CheckUsergroupExists(ctx context.Context, groupID string) (bool, error) {
	var res bool
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT EXISTS (SELECT 1 FROM usergroups WHERE id=$1)", groupID).Scan(&res)
	if err != nil {
		return false, parseErr("check", err)
	}
	return res, nil
}

func (c *connection) InsertManagedUsergroups(ctx context.Context, orgID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, `
		INSERT INTO usergroups (org_id, name, managed) VALUES
		($1, $2, true),
		($1, $3, true),
		($1, $4, true)
	`, orgID, database.UsergroupNameAutogroupUsers, database.UsergroupNameAutogroupMembers, database.UsergroupNameAutogroupGuests)
	if err != nil {
		return parseErr("managed usergroup", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 3 {
		// This should never happen.
		panic(fmt.Sprintf("expected 3 rows to be inserted, got %d", rows))
	}
	return nil
}

func (c *connection) InsertUsergroup(ctx context.Context, opts *database.InsertUsergroupOptions) (*database.Usergroup, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.Usergroup{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		INSERT INTO usergroups (org_id, name, managed) VALUES ($1, $2, $3) RETURNING *
	`, opts.OrgID, opts.Name, opts.Managed).StructScan(res)
	if err != nil {
		return nil, parseErr("usergroup", err)
	}
	return res, nil
}

func (c *connection) UpdateUsergroupName(ctx context.Context, name, groupID string) (*database.Usergroup, error) {
	res := &database.Usergroup{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "UPDATE usergroups SET name=$1, updated_on=now() WHERE id=$2 RETURNING *", name, groupID).StructScan(res)
	if err != nil {
		return nil, parseErr("usergroup", err)
	}
	return res, nil
}

func (c *connection) UpdateUsergroupDescription(ctx context.Context, description, groupID string) (*database.Usergroup, error) {
	res := &database.Usergroup{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "UPDATE usergroups SET description=$1, updated_on=now() WHERE id=$2 RETURNING *", description, groupID).StructScan(res)
	if err != nil {
		return nil, parseErr("usergroup", err)
	}
	return res, nil
}

func (c *connection) DeleteUsergroup(ctx context.Context, groupID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM usergroups WHERE id=$1", groupID)
	return checkDeleteRow("usergroup", res, err)
}

func (c *connection) FindUsergroupsForUser(ctx context.Context, userID, orgID string) ([]*database.Usergroup, error) {
	var res []*database.Usergroup
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT ug.* FROM usergroups ug JOIN usergroups_users uug ON ug.id = uug.usergroup_id
		WHERE uug.user_id = $1 AND ug.org_id = $2
	`, userID, orgID)
	if err != nil {
		return nil, parseErr("usergroup", err)
	}
	return res, nil
}

func (c *connection) FindUsergroupMemberUsers(ctx context.Context, groupID, afterEmail string, limit int) ([]*database.UsergroupMemberUser, error) {
	var res []*database.UsergroupMemberUser
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT uug.user_id as "id", u.email, u.display_name, u.photo_url FROM usergroups_users uug
		JOIN users u ON uug.user_id = u.id
		WHERE uug.usergroup_id = $1 AND lower(u.email) > lower($2)
		ORDER BY lower(u.email) LIMIT $3
	`, groupID, afterEmail, limit)
	if err != nil {
		return nil, parseErr("usergroup member", err)
	}
	return res, nil
}

func (c *connection) InsertUsergroupMemberUser(ctx context.Context, groupID, userID string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "INSERT INTO usergroups_users (user_id, usergroup_id) VALUES ($1, $2)", userID, groupID)
	if err != nil {
		return parseErr("usergroup member", err)
	}
	return nil
}

func (c *connection) DeleteUsergroupMemberUser(ctx context.Context, groupID, userID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM usergroups_users WHERE user_id = $1 AND usergroup_id = $2", userID, groupID)
	return checkDeleteRow("usergroup member", res, err)
}

func (c *connection) DeleteUsergroupsMemberUser(ctx context.Context, orgID, userID string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, `
		DELETE FROM usergroups_users WHERE user_id = $1 AND usergroup_id IN (SELECT id FROM usergroups WHERE org_id = $2)
	`, userID, orgID)
	if err != nil {
		return parseErr("usergroup member", err)
	}
	return nil
}

func (c *connection) InsertManagedUsergroupsMemberUser(ctx context.Context, orgID, userID, roleID string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, `
		INSERT INTO usergroups_users (user_id, usergroup_id)
		SELECT $1::UUID, ug.id FROM usergroups ug WHERE ug.org_id = $2 AND ug.name = $4
		UNION ALL SELECT $1::UUID, ug.id FROM usergroups ug WHERE ug.org_id = $2 AND ug.name = $5 AND EXISTS (SELECT 1 FROM org_roles ors WHERE ors.id = $3 AND NOT ors.guest)
		UNION ALL SELECT $1::UUID, ug.id FROM usergroups ug WHERE ug.org_id = $2 AND ug.name = $6 AND EXISTS (SELECT 1 FROM org_roles ors WHERE ors.id = $3 AND ors.guest)
	`, userID, orgID, roleID, database.UsergroupNameAutogroupUsers, database.UsergroupNameAutogroupMembers, database.UsergroupNameAutogroupGuests)
	if err != nil {
		return parseErr("managed usergroup member", err)
	}
	return nil
}

func (c *connection) DeleteManagedUsergroupsMemberUser(ctx context.Context, orgID, userID string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, `
		DELETE FROM usergroups_users WHERE user_id = $1 AND usergroup_id IN (
			SELECT ug.id FROM usergroups ug WHERE ug.org_id = $2 AND ug.managed
		)
	`, userID, orgID)
	if err != nil {
		return parseErr("managed usergroup member", err)
	}
	return nil
}

func (c *connection) FindUserAuthTokens(ctx context.Context, userID, afterID string, limit int, refresh *bool) ([]*database.UserAuthToken, error) {
	var qry strings.Builder
	qry.WriteString(`
		SELECT
			t.*,
			c.display_name AS auth_client_display_name
		FROM user_auth_tokens t
		LEFT JOIN auth_clients c ON t.auth_client_id = c.id
		WHERE t.user_id = $1 AND (t.expires_on IS NULL OR t.expires_on > now())
	`)
	args := []any{userID}

	// Filter by refresh token status if specified
	if refresh != nil {
		qry.WriteString(" AND t.refresh = $2")
		args = append(args, *refresh)
	}

	if afterID != "" {
		qry.WriteString(fmt.Sprintf(" AND t.id > $%d ORDER BY t.id LIMIT $%d", len(args)+1, len(args)+2))
		args = append(args, afterID, limit)
	} else {
		qry.WriteString(fmt.Sprintf(" ORDER BY t.id LIMIT $%d", len(args)+1))
		args = append(args, limit)
	}

	var res []*database.UserAuthToken
	err := c.getDB(ctx).SelectContext(ctx, &res, qry.String(), args...)
	if err != nil {
		return nil, parseErr("auth tokens", err)
	}
	return res, nil
}

func (c *connection) FindUserAuthToken(ctx context.Context, id string) (*database.UserAuthToken, error) {
	res := &database.UserAuthToken{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT t.* FROM user_auth_tokens t WHERE t.id=$1", id).StructScan(res)
	if err != nil {
		return nil, parseErr("auth token", err)
	}
	return res, nil
}

func (c *connection) InsertUserAuthToken(ctx context.Context, opts *database.InsertUserAuthTokenOptions) (*database.UserAuthToken, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.UserAuthToken{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		INSERT INTO user_auth_tokens (id, secret_hash, user_id, display_name, auth_client_id, representing_user_id, refresh, expires_on)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *`,
		opts.ID, opts.SecretHash, opts.UserID, opts.DisplayName, opts.AuthClientID, opts.RepresentingUserID, opts.Refresh, opts.ExpiresOn,
	).StructScan(res)
	if err != nil {
		return nil, parseErr("auth token", err)
	}
	return res, nil
}

func (c *connection) UpdateUserAuthTokenUsedOn(ctx context.Context, ids []string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "UPDATE user_auth_tokens SET used_on=now() WHERE id=ANY($1)", ids)
	if err != nil {
		return parseErr("auth token", err)
	}
	return nil
}

func (c *connection) DeleteUserAuthToken(ctx context.Context, id string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM user_auth_tokens WHERE id=$1", id)
	return checkDeleteRow("auth token", res, err)
}

func (c *connection) DeleteAllUserAuthTokens(ctx context.Context, userID string) (int, error) {
	qry := "DELETE FROM user_auth_tokens WHERE user_id=$1"
	args := []any{userID}

	res, err := c.getDB(ctx).ExecContext(ctx, qry, args...)
	if err != nil {
		return 0, parseErr("delete all auth tokens", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, parseErr("delete all auth tokens", err)
	}
	return int(n), nil
}

func (c *connection) DeleteUserAuthTokensByUserAndRepresentingUser(ctx context.Context, userID, representingUserID string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM user_auth_tokens WHERE user_id = $1 AND representing_user_id = $2", userID, representingUserID)
	return parseErr("auth token", err)
}

func (c *connection) DeleteExpiredUserAuthTokens(ctx context.Context, retention time.Duration) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM user_auth_tokens WHERE expires_on IS NOT NULL AND expires_on + $1 < now()", retention)
	return parseErr("auth token", err)
}

// DeleteInactiveUserAuthTokens deletes user authentication tokens that have not been used within the specified retention period.
func (c *connection) DeleteInactiveUserAuthTokens(ctx context.Context, retention time.Duration) error {
	_, err := c.getDB(ctx).ExecContext(ctx, `DELETE FROM user_auth_tokens WHERE used_on + $1 < now() AND created_on + $1 < now()`, retention)
	return parseErr("auth token", err)
}

// FindOrganizationMemberServices returns a list of services in an org.
func (c *connection) FindOrganizationMemberServices(ctx context.Context, orgID string) ([]*database.OrganizationMemberService, error) {
	var services []*organizationMemberServiceDTO
	query := `
       SELECT s.id, s.name, COALESCE(r.name, '') as role_name, EXISTS (
	       SELECT 1
           FROM service_projects_roles spr
           JOIN projects p ON p.id = spr.project_id
           WHERE spr.service_id = s.id AND p.org_id = $1
       	) AS has_project_roles, 
        s.attributes, s.created_on, s.updated_on
        FROM service s
        LEFT JOIN service_orgs_roles org_sr ON org_sr.service_id = s.id
        LEFT JOIN org_roles r ON r.id = org_sr.org_role_id
        WHERE s.org_id = $1
	`
	err := c.getDB(ctx).SelectContext(ctx, &services, query, orgID)
	if err != nil {
		return nil, parseErr("org member services", err)
	}

	// Convert DTOs to database.OrganizationMemberService
	orgMemberServices := make([]*database.OrganizationMemberService, len(services))
	for i, dto := range services {
		o, err := dto.organizationMemberServiceFromDTO()
		if err != nil {
			return nil, fmt.Errorf("failed to convert organization member service DTO: %w", err)
		}
		orgMemberServices[i] = o
	}

	return orgMemberServices, nil
}

// FindProjectMemberServices returns the services that are members of a project
func (c *connection) FindProjectMemberServices(ctx context.Context, projectID string) ([]*database.ProjectMemberService, error) {
	var services []*projectMemberServiceDTO
	query := `
		SELECT s.id, s.name, COALESCE(r.name, '') as role_name, COALESCE(org_r.name, '') as org_role_name, s.attributes, s.created_on, s.updated_on
		FROM service s
		LEFT JOIN service_projects_roles sr ON sr.service_id = s.id
		LEFT JOIN project_roles r ON r.id = sr.project_role_id
		LEFT JOIN service_orgs_roles org_sr ON org_sr.service_id = s.id
		LEFT JOIN org_roles org_r ON org_r.id = org_sr.org_role_id
		WHERE sr.project_id = $1
	`
	err := c.getDB(ctx).SelectContext(ctx, &services, query, projectID)
	if err != nil {
		return nil, parseErr("project member services", err)
	}

	// Convert DTOs to database.ProjectMemberService
	projectMemberServices := make([]*database.ProjectMemberService, len(services))
	for i, dto := range services {
		p, err := dto.projectMemberServiceFromDTO()
		if err != nil {
			return nil, fmt.Errorf("failed to convert project member service DTO: %w", err)
		}
		projectMemberServices[i] = p
	}

	return projectMemberServices, nil
}

// FindService returns a service.
func (c *connection) FindService(ctx context.Context, id string) (*database.Service, error) {
	res := &serviceDTO{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM service WHERE id=$1", id).StructScan(res)
	if err != nil {
		return nil, parseErr("service", err)
	}
	return res.serviceFromDTO()
}

// FindServiceByName returns a service.
func (c *connection) FindServiceByName(ctx context.Context, orgID, name string) (*database.Service, error) {
	res := &serviceDTO{}

	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM service WHERE org_id=$1 AND name=$2", orgID, name).StructScan(res)
	if err != nil {
		return nil, parseErr("service", err)
	}
	return res.serviceFromDTO()
}

// FindOrganizationMemberServiceForService returns the org level service details for a specific service.
func (c *connection) FindOrganizationMemberServiceForService(ctx context.Context, id string) (*database.OrganizationMemberService, error) {
	res := &organizationMemberServiceDTO{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		SELECT s.id, s.name, COALESCE(r.name, '') as role_name, EXISTS (
		   SELECT 1
		   FROM service_projects_roles spr
		   JOIN projects p ON p.id = spr.project_id
		   WHERE spr.service_id = s.id AND p.org_id = o.id
	   	) AS has_project_roles, 
		s.attributes, s.created_on, s.updated_on
		FROM service s
		JOIN orgs o ON o.id = s.org_id
		LEFT JOIN service_orgs_roles org_sr ON org_sr.service_id = s.id
		LEFT JOIN org_roles r ON r.id = org_sr.org_role_id
		WHERE s.id = $1`, id).StructScan(res)
	if err != nil {
		return nil, parseErr("organization member service for service", err)
	}
	return res.organizationMemberServiceFromDTO()
}

// FindProjectMemberServicesForService returns all projects level details that service is a member of.
func (c *connection) FindProjectMemberServicesForService(ctx context.Context, id string) ([]*database.ProjectMemberServiceWithProject, error) {
	var services []*projectMemberServiceWithProjectDTO
	query := `
		SELECT s.id, s.name, COALESCE(r.name, '') as role_name, COALESCE(org_r.name, '') as org_role_name, s.attributes, s.created_on, s.updated_on,
		p.id AS project_id, p.name AS project_name
		FROM service s
		LEFT JOIN service_projects_roles sr ON sr.service_id = s.id
		LEFT JOIN project_roles r ON r.id = sr.project_role_id
		LEFT JOIN service_orgs_roles org_sr ON org_sr.service_id = s.id
		LEFT JOIN org_roles org_r ON org_r.id = org_sr.org_role_id
		JOIN projects p ON p.id = sr.project_id
		WHERE s.id = $1
	`
	err := c.getDB(ctx).SelectContext(ctx, &services, query, id)
	if err != nil {
		return nil, parseErr("project member services for service", err)
	}

	projectMemberServices := make([]*database.ProjectMemberServiceWithProject, len(services))
	for i, dto := range services {
		p, err := dto.projectMemberServiceWithProjectFromDTO()
		if err != nil {
			return nil, fmt.Errorf("failed to convert project member service with project DTO: %w", err)
		}
		projectMemberServices[i] = p
	}

	return projectMemberServices, nil
}

// InsertService inserts a service.
func (c *connection) InsertService(ctx context.Context, opts *database.InsertServiceOptions) (*database.Service, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	if opts.Attributes == nil {
		opts.Attributes = make(map[string]any)
	}

	res := &serviceDTO{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		INSERT INTO service (org_id, name, attributes)
		VALUES ($1, $2, $3) RETURNING *`,
		opts.OrgID, opts.Name, opts.Attributes,
	).StructScan(res)
	if err != nil {
		return nil, parseErr("service", err)
	}
	return res.serviceFromDTO()
}

// UpdateService updates a service.
func (c *connection) UpdateService(ctx context.Context, id string, opts *database.UpdateServiceOptions) (*database.Service, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	if opts.Attributes == nil {
		opts.Attributes = make(map[string]any)
	}

	res := &serviceDTO{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		UPDATE service
		SET name=$1, attributes=$2
		WHERE id=$3 RETURNING *`,
		opts.Name, opts.Attributes, id,
	).StructScan(res)
	if err != nil {
		return nil, parseErr("service", err)
	}
	return res.serviceFromDTO()
}

// UpdateServiceActiceOn updates a service's active_on timestamp.
func (c *connection) UpdateServiceActiveOn(ctx context.Context, ids []string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "UPDATE service SET active_on=now() WHERE id=ANY($1)", ids)
	if err != nil {
		return parseErr("service", err)
	}
	return nil
}

// DeleteService deletes a service.
func (c *connection) DeleteService(ctx context.Context, id string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM service WHERE id=$1", id)
	return checkDeleteRow("service", res, err)
}

// FindSeviceAuthTokens returns a list of service auth tokens.
func (c *connection) FindServiceAuthTokens(ctx context.Context, serviceID string) ([]*database.ServiceAuthToken, error) {
	var res []*database.ServiceAuthToken
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT t.* FROM service_auth_tokens t WHERE t.service_id=$1", serviceID)
	if err != nil {
		return nil, parseErr("service auth tokens", err)
	}
	return res, nil
}

// FindServiceAuthToken returns a service auth token.
func (c *connection) FindServiceAuthToken(ctx context.Context, id string) (*database.ServiceAuthToken, error) {
	res := &database.ServiceAuthToken{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT t.* FROM service_auth_tokens t WHERE t.id=$1", id).StructScan(res)
	if err != nil {
		return nil, parseErr("service auth token", err)
	}
	return res, nil
}

// InsertServiceAuthToken inserts a service auth token.
func (c *connection) InsertServiceAuthToken(ctx context.Context, opts *database.InsertServiceAuthTokenOptions) (*database.ServiceAuthToken, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.ServiceAuthToken{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		INSERT INTO service_auth_tokens (id, secret_hash, service_id, expires_on)
		VALUES ($1, $2, $3, $4) RETURNING *`,
		opts.ID, opts.SecretHash, opts.ServiceID, opts.ExpiresOn,
	).StructScan(res)
	if err != nil {
		return nil, parseErr("service auth token", err)
	}
	return res, nil
}

func (c *connection) UpdateServiceAuthTokenUsedOn(ctx context.Context, ids []string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "UPDATE service_auth_tokens SET used_on=now() WHERE id=ANY($1)", ids)
	if err != nil {
		return parseErr("service auth token", err)
	}
	return nil
}

// DeleteServiceAuthToken deletes a service auth token.
func (c *connection) DeleteServiceAuthToken(ctx context.Context, id string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM service_auth_tokens WHERE id=$1", id)
	return checkDeleteRow("service auth token", res, err)
}

// DeleteExpiredServiceAuthTokens deletes expired service auth tokens.
func (c *connection) DeleteExpiredServiceAuthTokens(ctx context.Context, retention time.Duration) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM service_auth_tokens WHERE expires_on IS NOT NULL AND expires_on + $1 < now()", retention)
	return parseErr("service auth token", err)
}

// DeleteInactiveServiceAuthTokens deletes service authentication tokens that have not been used within the specified retention period.
func (c *connection) DeleteInactiveServiceAuthTokens(ctx context.Context, retention time.Duration) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM service_auth_tokens WHERE used_on + $1 < now() AND created_on + $1 < now()", retention)
	return parseErr("service auth token", err)
}

func (c *connection) FindDeploymentAuthToken(ctx context.Context, id string) (*database.DeploymentAuthToken, error) {
	res := &database.DeploymentAuthToken{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT t.* FROM deployment_auth_tokens t WHERE t.id=$1", id).StructScan(res)
	if err != nil {
		return nil, parseErr("deployment auth token", err)
	}
	return res, nil
}

func (c *connection) InsertDeploymentAuthToken(ctx context.Context, opts *database.InsertDeploymentAuthTokenOptions) (*database.DeploymentAuthToken, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.DeploymentAuthToken{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		INSERT INTO deployment_auth_tokens (id, secret_hash, deployment_id, expires_on)
		VALUES ($1, $2, $3, $4) RETURNING *`,
		opts.ID, opts.SecretHash, opts.DeploymentID, opts.ExpiresOn,
	).StructScan(res)
	if err != nil {
		return nil, parseErr("deployment auth token", err)
	}
	return res, nil
}

func (c *connection) UpdateDeploymentAuthTokenUsedOn(ctx context.Context, ids []string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "UPDATE deployment_auth_tokens SET used_on=now() WHERE id=ANY($1)", ids)
	if err != nil {
		return parseErr("deployment auth token", err)
	}
	return nil
}

func (c *connection) DeleteExpiredDeploymentAuthTokens(ctx context.Context, retention time.Duration) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM deployment_auth_tokens WHERE expires_on IS NOT NULL AND expires_on + $1 < now()", retention)
	return parseErr("deployment auth token", err)
}

func (c *connection) FindMagicAuthTokensWithUser(ctx context.Context, projectID string, createdByUserID *string, afterID string, limit int) ([]*database.MagicAuthTokenWithUser, error) {
	n := 1
	where := fmt.Sprintf("t.project_id=$%d", n)
	args := []any{projectID}
	n++

	if createdByUserID != nil {
		where = fmt.Sprintf("%s AND t.created_by_user_id=$%d", where, n)
		args = append(args, *createdByUserID)
		n++
	}

	if afterID != "" {
		where = fmt.Sprintf("%s AND t.id>$%d", where, n)
		args = append(args, afterID)
		n++
	}

	where += " AND (t.expires_on IS NULL OR t.expires_on > now()) AND t.internal=false"

	qry := fmt.Sprintf("SELECT t.*, COALESCE(u.email, '') AS created_by_user_email FROM magic_auth_tokens t LEFT JOIN users u ON t.created_by_user_id=u.id WHERE %s ORDER BY t.id LIMIT $%d", where, n)
	args = append(args, limit)

	var dtos []*magicAuthTokenWithUserDTO
	err := c.getDB(ctx).SelectContext(ctx, &dtos, qry, args...)
	if err != nil {
		return nil, parseErr("magic auth tokens", err)
	}

	res := make([]*database.MagicAuthTokenWithUser, len(dtos))
	for i, dto := range dtos {
		var err error
		res[i], err = c.magicAuthTokenWithUserFromDTO(dto)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (c *connection) FindMagicAuthToken(ctx context.Context, id string, withSecret bool) (*database.MagicAuthToken, error) {
	res := &magicAuthTokenDTO{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT t.* FROM magic_auth_tokens t WHERE t.id=$1", id).StructScan(res)
	if err != nil {
		return nil, parseErr("magic auth token", err)
	}
	return c.magicAuthTokenFromDTO(res, withSecret)
}

func (c *connection) FindMagicAuthTokenWithUser(ctx context.Context, id string) (*database.MagicAuthTokenWithUser, error) {
	res := &magicAuthTokenWithUserDTO{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT t.*, COALESCE(u.email, '') AS created_by_user_email FROM magic_auth_tokens t LEFT JOIN users u ON t.created_by_user_id=u.id WHERE t.id=$1 AND t.internal=false", id).StructScan(res)
	if err != nil {
		return nil, parseErr("magic auth token", err)
	}
	return c.magicAuthTokenWithUserFromDTO(res)
}

func (c *connection) InsertMagicAuthToken(ctx context.Context, opts *database.InsertMagicAuthTokenOptions) (*database.MagicAuthToken, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	if opts.Fields == nil {
		opts.Fields = []string{}
	}

	encSecret, encKeyID, err := c.encrypt(opts.Secret)
	if err != nil {
		return nil, err
	}

	resources, err := json.Marshal(opts.Resources)
	if err != nil {
		return nil, err
	}

	res := &magicAuthTokenDTO{}
	err = c.getDB(ctx).QueryRowxContext(ctx, `
		INSERT INTO magic_auth_tokens (id, secret_hash, secret, secret_encryption_key_id, project_id, expires_on, created_by_user_id, attributes, filter_json, fields, state, display_name, internal, resources)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) RETURNING *`,
		opts.ID, opts.SecretHash, encSecret, encKeyID, opts.ProjectID, opts.ExpiresOn, opts.CreatedByUserID, opts.Attributes, opts.FilterJSON, opts.Fields, opts.State, opts.DisplayName, opts.Internal, resources,
	).StructScan(res)
	if err != nil {
		return nil, parseErr("magic auth token", err)
	}
	return c.magicAuthTokenFromDTO(res, true)
}

func (c *connection) UpdateMagicAuthTokenUsedOn(ctx context.Context, ids []string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "UPDATE magic_auth_tokens SET used_on=now() WHERE id=ANY($1)", ids)
	if err != nil {
		return parseErr("magic auth token", err)
	}
	return nil
}

func (c *connection) DeleteMagicAuthToken(ctx context.Context, id string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM magic_auth_tokens WHERE id=$1", id)
	return checkDeleteRow("magic auth token", res, err)
}

func (c *connection) DeleteMagicAuthTokens(ctx context.Context, ids []string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM magic_auth_tokens WHERE id=ANY($1)", ids)
	return parseErr("magic auth token", err)
}

func (c *connection) DeleteExpiredMagicAuthTokens(ctx context.Context, retention time.Duration) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM magic_auth_tokens WHERE expires_on IS NOT NULL AND expires_on + $1 < now()", retention)
	return parseErr("magic auth token", err)
}

func (c *connection) FindNotificationTokens(ctx context.Context, resourceKind, resourceName string) ([]*database.NotificationToken, error) {
	var res []*database.NotificationToken
	err := c.getDB(ctx).SelectContext(ctx, &res, `SELECT * FROM notification_tokens WHERE resource_kind=$1 AND resource_name=$2`, resourceKind, resourceName)
	if err != nil {
		return nil, parseErr("notification tokens", err)
	}
	return res, nil
}

func (c *connection) FindNotificationTokensWithSecret(ctx context.Context, resourceKind, resourceName string) ([]*database.NotificationTokenWithSecret, error) {
	var res []*notificationTokenWithSecretDTO
	err := c.getDB(ctx).SelectContext(ctx, &res, `SELECT t.*, m.secret as magic_auth_token_secret, m.secret_encryption_key_id FROM notification_tokens t JOIN magic_auth_tokens m ON t.magic_auth_token_id=m.id WHERE t.resource_kind=$1 AND t.resource_name=$2`, resourceKind, resourceName)
	if err != nil {
		return nil, parseErr("notification tokens", err)
	}

	ret := make([]*database.NotificationTokenWithSecret, len(res))
	for i, dto := range res {
		ret[i], err = c.notificationTokenWithSecretFromDTO(dto)
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

func (c *connection) FindNotificationTokenForMagicAuthToken(ctx context.Context, magicAuthTokenID string) (*database.NotificationToken, error) {
	res := &database.NotificationToken{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `SELECT * FROM notification_tokens WHERE magic_auth_token_id=$1`, magicAuthTokenID).StructScan(res)
	if err != nil {
		return nil, parseErr("notification token", err)
	}
	return res, nil
}

func (c *connection) InsertNotificationToken(ctx context.Context, opts *database.InsertNotificationTokenOptions) (*database.NotificationToken, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.NotificationToken{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `INSERT INTO notification_tokens (resource_kind, resource_name, recipient_email, magic_auth_token_id) VALUES ($1, $2, $3, $4) RETURNING *`, opts.ResourceKind, opts.ResourceName, opts.RecipientEmail, opts.MagicAuthTokenID).StructScan(res)
	if err != nil {
		return nil, parseErr("notification token", err)
	}
	return res, nil
}

func (c *connection) FindDeviceAuthCodeByDeviceCode(ctx context.Context, deviceCode string) (*database.DeviceAuthCode, error) {
	authCode := &database.DeviceAuthCode{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM device_auth_codes WHERE device_code = $1", deviceCode).StructScan(authCode)
	if err != nil {
		return nil, parseErr("device auth code", err)
	}
	return authCode, nil
}

func (c *connection) FindPendingDeviceAuthCodeByUserCode(ctx context.Context, userCode string) (*database.DeviceAuthCode, error) {
	authCode := &database.DeviceAuthCode{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM device_auth_codes WHERE user_code = $1 AND expires_on > now() AND approval_state = 0", userCode).StructScan(authCode)
	if err != nil {
		return nil, parseErr("device auth code", err)
	}
	return authCode, nil
}

func (c *connection) InsertDeviceAuthCode(ctx context.Context, deviceCode, userCode, clientID string, expiresOn time.Time) (*database.DeviceAuthCode, error) {
	res := &database.DeviceAuthCode{}
	err := c.getDB(ctx).QueryRowxContext(ctx,
		`INSERT INTO device_auth_codes (device_code, user_code, expires_on, approval_state, client_id)
		VALUES ($1, $2, $3, $4, $5)  RETURNING *`, deviceCode, userCode, expiresOn, database.DeviceAuthCodeStatePending, clientID).StructScan(res)
	if err != nil {
		return nil, parseErr("device auth code", err)
	}
	return res, nil
}

func (c *connection) DeleteDeviceAuthCode(ctx context.Context, deviceCode string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM device_auth_codes WHERE device_code=$1", deviceCode)
	return checkDeleteRow("device auth code", res, err)
}

func (c *connection) UpdateDeviceAuthCode(ctx context.Context, id, userID string, approvalState database.DeviceAuthCodeState) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "UPDATE device_auth_codes SET approval_state=$1, user_id=$2, updated_on=now() WHERE id=$3", approvalState, userID, id)
	return checkUpdateRow("device auth code", res, err)
}

func (c *connection) DeleteExpiredDeviceAuthCodes(ctx context.Context, retention time.Duration) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM device_auth_codes WHERE expires_on + $1 < now()", retention)
	return parseErr("device auth code", err)
}

func (c *connection) FindAuthorizationCode(ctx context.Context, code string) (*database.AuthorizationCode, error) {
	authCode := &database.AuthorizationCode{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM authorization_codes WHERE code = $1", code).StructScan(authCode)
	if err != nil {
		return nil, parseErr("authorization code", err)
	}
	return authCode, nil
}

func (c *connection) InsertAuthorizationCode(ctx context.Context, code, userID, clientID, redirectURI, codeChallenge, codeChallengeMethod string, expiration time.Time) (*database.AuthorizationCode, error) {
	res := &database.AuthorizationCode{}
	err := c.getDB(ctx).QueryRowxContext(ctx,
		`INSERT INTO authorization_codes (code, user_id, client_id, redirect_uri, code_challenge, code_challenge_method, expires_on)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *`, code, userID, clientID, redirectURI, codeChallenge, codeChallengeMethod, expiration).StructScan(res)
	if err != nil {
		return nil, parseErr("authorization code", err)
	}
	return res, nil
}

func (c *connection) DeleteAuthorizationCode(ctx context.Context, code string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM authorization_codes WHERE code=$1", code)
	return checkDeleteRow("authorization code", res, err)
}

func (c *connection) DeleteExpiredAuthorizationCodes(ctx context.Context, retention time.Duration) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM authorization_codes WHERE expires_on + $1 < now()", retention)
	return parseErr("authorization code", err)
}

func (c *connection) InsertAuthClient(ctx context.Context, displayName, scope string, grantTypes []string) (*database.AuthClient, error) {
	if grantTypes == nil {
		grantTypes = []string{}
	}
	dto := &authClientDTO{}
	err := c.getDB(ctx).QueryRowxContext(ctx,
		`INSERT INTO auth_clients (display_name, scope, grant_types) VALUES ($1, $2, $3) RETURNING *`,
		displayName, scope, grantTypes).StructScan(dto)
	if err != nil {
		return nil, parseErr("auth client", err)
	}
	return dto.AsModel()
}

func (c *connection) FindAuthClient(ctx context.Context, id string) (*database.AuthClient, error) {
	dto := &authClientDTO{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM auth_clients WHERE id = $1", id).StructScan(dto)
	if err != nil {
		return nil, parseErr("auth client", err)
	}
	return dto.AsModel()
}

func (c *connection) UpdateAuthClientUsedOn(ctx context.Context, ids []string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "UPDATE auth_clients SET used_on=now() WHERE id=ANY($1)", ids)
	if err != nil {
		return parseErr("auth client", err)
	}
	return nil
}

func (c *connection) FindOrganizationRoles(ctx context.Context) ([]*database.OrganizationRole, error) {
	var res []*database.OrganizationRole
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT * FROM org_roles")
	if err != nil {
		return nil, parseErr("org roles", err)
	}
	return res, nil
}

func (c *connection) FindOrganizationRole(ctx context.Context, name string) (*database.OrganizationRole, error) {
	role := &database.OrganizationRole{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM org_roles WHERE lower(name)=lower($1)", name).StructScan(role)
	if err != nil {
		return nil, parseErr("org role", err)
	}
	return role, nil
}

func (c *connection) FindProjectRoles(ctx context.Context) ([]*database.ProjectRole, error) {
	var res []*database.ProjectRole
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT * FROM project_roles")
	if err != nil {
		return nil, parseErr("project roles", err)
	}
	return res, nil
}

func (c *connection) FindProjectRole(ctx context.Context, name string) (*database.ProjectRole, error) {
	role := &database.ProjectRole{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM project_roles WHERE lower(name)=lower($1)", name).StructScan(role)
	if err != nil {
		return nil, parseErr("project role", err)
	}
	return role, nil
}

func (c *connection) ResolveOrganizationRolesForUser(ctx context.Context, userID, orgID string) ([]*database.OrganizationRole, error) {
	var res []*database.OrganizationRole
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT r.* FROM users_orgs_roles uor
		JOIN org_roles r ON uor.org_role_id = r.id
		WHERE uor.user_id = $1 AND uor.org_id = $2
		UNION
		SELECT * FROM org_roles WHERE id IN (
			SELECT org_role_id FROM usergroups_orgs_roles uor JOIN usergroups_users uug
			ON uor.usergroup_id = uug.usergroup_id WHERE uug.user_id = $1 AND uor.org_id = $2
		)`, userID, orgID)
	if err != nil {
		return nil, parseErr("org roles", err)
	}
	return res, nil
}

func (c *connection) ResolveProjectRolesForUser(ctx context.Context, userID, projectID string) ([]*database.ProjectRole, error) {
	var res []*database.ProjectRole
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT r.* FROM users_projects_roles upr
		JOIN project_roles r ON upr.project_role_id = r.id
		WHERE upr.user_id = $1 AND upr.project_id = $2
		UNION
		SELECT * FROM project_roles WHERE id IN (
			SELECT project_role_id FROM usergroups_projects_roles upr JOIN usergroups_users uug
			ON upr.usergroup_id = uug.usergroup_id WHERE uug.user_id = $1 AND upr.project_id = $2
		)`, userID, projectID)
	if err != nil {
		return nil, parseErr("project roles", err)
	}
	return res, nil
}

// ResolveOrganizationRoleForService returns the organization role for the service
func (c *connection) ResolveOrganizationRoleForService(ctx context.Context, serviceID, orgID string) (*database.OrganizationRole, error) {
	var role database.OrganizationRole
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		SELECT r.*
		FROM org_roles r
		JOIN service_orgs_roles sr ON sr.org_role_id = r.id
		WHERE sr.service_id = $1 AND sr.org_id = $2
	`, serviceID, orgID).StructScan(&role)
	if err != nil {
		return nil, parseErr("service org role", err)
	}
	return &role, nil
}

// ResolveProjectRolesForService returns the project roles for a service
func (c *connection) ResolveProjectRolesForService(ctx context.Context, serviceID, projectID string) ([]*database.ProjectRole, error) {
	var roles []*database.ProjectRole
	err := c.getDB(ctx).SelectContext(ctx, &roles, `
		SELECT r.*
		FROM project_roles r
		JOIN service_projects_roles sr ON sr.project_role_id = r.id
		WHERE sr.service_id = $1 AND sr.project_id = $2
	`, serviceID, projectID)
	if err != nil {
		return nil, parseErr("service project roles", err)
	}
	return roles, nil
}

func (c *connection) FindOrganizationMemberUsers(ctx context.Context, orgID, filterRoleID string, withCounts bool, afterEmail string, limit int, searchPattern string) ([]*database.OrganizationMemberUser, error) {
	args := []any{orgID, afterEmail, limit}
	var qry strings.Builder
	qry.WriteString("SELECT u.id, u.email, u.display_name, u.photo_url, u.created_on, u.updated_on, r.name as role_name, uor.attributes")
	if withCounts {
		qry.WriteString(`,
			(
				SELECT COUNT(*) FROM projects p WHERE p.org_id = $1 AND p.id IN (
					SELECT upr.project_id FROM users_projects_roles upr WHERE upr.user_id = u.id
					UNION
					SELECT ugpr.project_id FROM usergroups_projects_roles ugpr JOIN usergroups_users uug ON ugpr.usergroup_id = uug.usergroup_id WHERE uug.user_id = u.id
				)
			) as projects_count,
			(
				SELECT COUNT(*)
				FROM usergroups_users uus
				JOIN usergroups ugu ON uus.usergroup_id = ugu.id
				WHERE ugu.org_id = $1 AND uus.user_id = u.id
			) as usergroups_count
		`)
	}
	qry.WriteString(`
		FROM users u
		JOIN users_orgs_roles uor ON u.id = uor.user_id
		JOIN org_roles r ON r.id = uor.org_role_id
		WHERE uor.org_id=$1
	`)
	if filterRoleID != "" {
		qry.WriteString(" AND uor.org_role_id=$4")
		args = append(args, filterRoleID)
	}
	if searchPattern != "" {
		qry.WriteString(fmt.Sprintf(" AND (lower(u.email) ILIKE $%d OR lower(u.display_name) ILIKE $%d)", len(args)+1, len(args)+1))
		args = append(args, searchPattern)
	}
	qry.WriteString(" AND lower(u.email) > lower($2) ORDER BY lower(u.email) LIMIT $3")

	var res []*database.OrganizationMemberUser
	var dtos []*organizationMemberUserDTO
	err := c.getDB(ctx).SelectContext(ctx, &dtos, qry.String(), args...)
	if err != nil {
		return nil, parseErr("org members", err)
	}
	for _, dto := range dtos {
		user, err := dto.organizationMemberUserFromDTO()
		if err != nil {
			return nil, err
		}
		res = append(res, user)
	}
	return res, nil
}

func (c *connection) CountOrganizationMemberUsers(ctx context.Context, orgID, filterRoleID, searchPattern string) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM users_orgs_roles uor JOIN users u ON u.id = uor.user_id WHERE uor.org_id=$1"
	args := []any{orgID}

	if filterRoleID != "" {
		query += " AND uor.org_role_id=$2"
		args = append(args, filterRoleID)
	}

	if searchPattern != "" {
		query += fmt.Sprintf(" AND (lower(u.email) ILIKE $%d OR lower(u.display_name) ILIKE $%d)", len(args)+1, len(args)+1)
		args = append(args, searchPattern)
	}

	err := c.getDB(ctx).QueryRowxContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, parseErr("org members count", err)
	}
	return count, nil
}

func (c *connection) FindOrganizationMemberUsersByRole(ctx context.Context, orgID, roleID string) ([]*database.User, error) {
	var res []*database.User
	err := c.getDB(ctx).SelectContext(
		ctx, &res, "SELECT u.* FROM users u JOIN users_orgs_roles uor on u.id = uor.user_id WHERE uor.org_id=$1 AND uor.org_role_id=$2", orgID, roleID)
	if err != nil {
		return nil, parseErr("org members", err)
	}
	return res, nil
}

func (c *connection) FindOrganizationMemberUser(ctx context.Context, orgID, userID string) (*database.OrganizationMemberUser, error) {
	qry := `SELECT u.id, u.email, u.display_name, u.photo_url, u.created_on, u.updated_on, r.name as role_name, uor.attributes,
			(
				SELECT COUNT(*) FROM projects p WHERE p.org_id = $1 AND p.id IN (
					SELECT upr.project_id FROM users_projects_roles upr WHERE upr.user_id = u.id
					UNION
					SELECT ugpr.project_id FROM usergroups_projects_roles ugpr JOIN usergroups_users uug ON ugpr.usergroup_id = uug.usergroup_id WHERE uug.user_id = u.id
				)
			) as projects_count,
			(
				SELECT COUNT(*)
				FROM usergroups_users uus
				JOIN usergroups ugu ON uus.usergroup_id = ugu.id
				WHERE ugu.org_id = $1 AND uus.user_id = u.id
			) as usergroups_count
		FROM users u
		JOIN users_orgs_roles uor ON u.id = uor.user_id
		JOIN org_roles r ON r.id = uor.org_role_id
		WHERE uor.org_id = $1 AND uor.user_id = $2`

	var dto organizationMemberUserDTO
	err := c.getDB(ctx).QueryRowxContext(ctx, qry, orgID, userID).StructScan(&dto)
	if err != nil {
		return nil, parseErr("org member", err)
	}

	user, err := dto.organizationMemberUserFromDTO()
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (c *connection) FindOrganizationMemberUserAdminStatus(ctx context.Context, orgID, userID string) (isAdmin, isLastAdmin bool, err error) {
	err = c.getDB(ctx).QueryRowxContext(ctx, `
		SELECT
			r.admin,
			NOT EXISTS (SELECT 1 FROM users_orgs_roles uor JOIN org_roles r ON r.id = uor.org_role_id WHERE uor.org_id=$1 AND r.admin=true AND uor.user_id != $2 LIMIT 1)
		FROM users_orgs_roles uor
		JOIN org_roles r ON r.id = uor.org_role_id
		WHERE uor.org_id=$1 AND uor.user_id=$2
	`, orgID, userID).Scan(&isAdmin, &isLastAdmin)
	if err != nil {
		return false, false, parseErr("org member admin status", err)
	}
	return isAdmin, isLastAdmin, nil
}

func (c *connection) InsertOrganizationMemberUser(ctx context.Context, orgID, userID, roleID string, attributes map[string]interface{}, ifNotExists bool) (bool, error) {
	attrs, err := c.validateAttributes(attributes)
	if err != nil {
		return false, err
	}

	if !ifNotExists {
		res, err := c.getDB(ctx).ExecContext(ctx, "INSERT INTO users_orgs_roles (user_id, org_id, org_role_id, attributes) VALUES ($1, $2, $3, $4)", userID, orgID, roleID, attrs)
		if err != nil {
			return false, parseErr("org member", err)
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return false, err
		}
		if rows == 0 {
			return false, fmt.Errorf("no rows affected when adding user to organization")
		}
		return true, nil
	}

	res, err := c.getDB(ctx).ExecContext(ctx, `
		INSERT INTO users_orgs_roles (user_id, org_id, org_role_id, attributes)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, org_id) DO NOTHING
	`, userID, orgID, roleID, attributes)
	if err != nil {
		return false, parseErr("org member", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	if rows > 1 {
		panic(fmt.Errorf("expected to update 0 or 1 row, but updated %d", rows))
	}
	return rows != 0, nil
}

func (c *connection) DeleteOrganizationMemberUser(ctx context.Context, orgID, userID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM users_orgs_roles WHERE user_id = $1 AND org_id = $2", userID, orgID)
	return checkDeleteRow("org member", res, err)
}

func (c *connection) UpdateOrganizationMemberUserRole(ctx context.Context, orgID, userID, roleID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, `UPDATE users_orgs_roles SET org_role_id = $1 WHERE user_id = $2 AND org_id = $3`, roleID, userID, orgID)
	return checkUpdateRow("org member", res, err)
}

func (c *connection) UpdateOrganizationMemberUserAttributes(ctx context.Context, orgID, userID string, attributes map[string]any) (bool, error) {
	attrs, err := c.validateAttributes(attributes)
	if err != nil {
		return false, err
	}

	res, err := c.getDB(ctx).ExecContext(ctx, `
		UPDATE users_orgs_roles
		SET attributes = $1
		WHERE user_id = $2 AND org_id = $3
	`, attrs, userID, orgID)
	if err != nil {
		return false, parseErr("org member attributes", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	if rows > 1 {
		panic(fmt.Errorf("expected to update 0 or 1 row, but updated %d", rows))
	}
	return rows != 0, nil
}

func (c *connection) CountSingleuserOrganizationsForMemberUser(ctx context.Context, userID string) (int, error) {
	var count int
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		SELECT COALESCE(SUM(total_count), 0) as total_count FROM (
			SELECT CASE WHEN COUNT(*) = 1 THEN 1 ELSE 0 END as total_count FROM users_orgs_roles WHERE org_id IN (
				SELECT org_id FROM users_orgs_roles WHERE user_id = $1
			) GROUP BY org_id
		) as subquery
	`, userID).Scan(&count)
	if err != nil {
		return 0, parseErr("singleuser orgs count", err)
	}
	return count, nil
}

func (c *connection) FindOrganizationMembersWithManageUsersRole(ctx context.Context, orgID string) ([]*database.OrganizationMemberUser, error) {
	var res []*database.OrganizationMemberUser
	var dtos []*organizationMemberUserDTO
	err := c.getDB(ctx).SelectContext(ctx, &dtos, `
		SELECT u.id, u.email, u.display_name, u.photo_url, u.created_on, u.updated_on, r.name as role_name, uor.attributes
		FROM users u
		JOIN users_orgs_roles uor ON u.id = uor.user_id
		JOIN org_roles r ON r.id = uor.org_role_id
		WHERE uor.org_id=$1 AND r.manage_org_members=true
		ORDER BY lower(u.email)
	`, orgID)
	if err != nil {
		return nil, parseErr("org members", err)
	}
	for _, dto := range dtos {
		user, err := dto.organizationMemberUserFromDTO()
		if err != nil {
			return nil, err
		}
		res = append(res, user)
	}
	return res, nil
}

// InsertOrganizationMemberService adds a service to an organization with a role
func (c *connection) InsertOrganizationMemberService(ctx context.Context, serviceID, orgID, roleID string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, `
		INSERT INTO service_orgs_roles (service_id, org_id, org_role_id)
		VALUES ($1, $2, $3)
	`, serviceID, orgID, roleID)
	if err != nil {
		return parseErr("service org member", err)
	}
	return nil
}

// UpdateOrganizationMemberServiceRole updates the role of a service in an organization
func (c *connection) UpdateOrganizationMemberServiceRole(ctx context.Context, serviceID, orgID, roleID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, `
		UPDATE service_orgs_roles
		SET org_role_id = $3
		WHERE service_id = $1 AND org_id = $2
	`, serviceID, orgID, roleID)
	return checkUpdateRow("service org member", res, err)
}

func (c *connection) FindProjectMemberUsers(ctx context.Context, orgID, projectID, filterRoleID, afterEmail string, limit int) ([]*database.ProjectMemberUser, error) {
	args := []any{orgID, projectID, afterEmail, limit}
	var qry strings.Builder
	qry.WriteString(`
		SELECT
			-- User info
			u.id, u.email, u.display_name, u.photo_url, u.created_on, u.updated_on,
			-- Project role name
			(SELECT pr.name FROM project_roles pr WHERE pr.id = upr.project_role_id) as role_name,
			-- Org role name
			(
				SELECT orr.name
				FROM org_roles orr
				JOIN users_orgs_roles uor
				ON orr.id = uor.org_role_id
				WHERE uor.user_id = u.id AND uor.org_id = $1
			) as org_role_name
		FROM users u
		JOIN users_projects_roles upr ON upr.user_id = u.id
		WHERE upr.project_id = $2
	`)
	if filterRoleID != "" {
		qry.WriteString(" AND upr.project_role_id=$5")
		args = append(args, filterRoleID)
	}
	qry.WriteString(" AND lower(u.email) > lower($3) ORDER BY lower(u.email) LIMIT $4")

	var res []*database.ProjectMemberUser
	err := c.getDB(ctx).SelectContext(ctx, &res, qry.String(), args...)
	if err != nil {
		return nil, parseErr("project members", err)
	}
	return res, nil
}

func (c *connection) FindProjectMemberUserRole(ctx context.Context, projectID, userID string) (*database.ProjectRole, error) {
	role := &database.ProjectRole{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		SELECT r.* FROM users_projects_roles upr
		JOIN project_roles r ON r.id = upr.project_role_id
		WHERE upr.project_id=$1 AND upr.user_id=$2
	`, projectID, userID).StructScan(role)
	if err != nil {
		return nil, parseErr("project member role", err)
	}
	return role, nil
}

func (c *connection) FindSuperusers(ctx context.Context) ([]*database.User, error) {
	var res []*database.User
	err := c.getDB(ctx).SelectContext(ctx, &res, `SELECT u.* FROM users u WHERE u.superuser = true`)
	if err != nil {
		return nil, parseErr("project members", err)
	}
	return res, nil
}

func (c *connection) UpdateSuperuser(ctx context.Context, userID string, superuser bool) error {
	res, err := c.getDB(ctx).ExecContext(ctx, `UPDATE users SET superuser=$2, updated_on=now() WHERE id=$1`, userID, superuser)
	return checkUpdateRow("superuser", res, err)
}

func (c *connection) InsertProjectMemberUser(ctx context.Context, projectID, userID, roleID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "INSERT INTO users_projects_roles (user_id, project_id, project_role_id) VALUES ($1, $2, $3)", userID, projectID, roleID)
	if err != nil {
		return parseErr("project member", err)
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

// FindOrganizationMemberUsergroups returns org user groups as a collection of MemberUsergroup.
// If a user group has no org role then RoleName is empty.
func (c *connection) FindOrganizationMemberUsergroups(ctx context.Context, orgID, filterRoleID string, withCounts bool, afterName string, limit int) ([]*database.MemberUsergroup, error) {
	args := []any{orgID, afterName, limit}
	var qry strings.Builder
	qry.WriteString("SELECT ug.id, ug.name, ug.managed, ug.created_on, ug.updated_on, COALESCE(r.name, '') as role_name")
	if withCounts {
		qry.WriteString(`,
			(
				SELECT COUNT(*) FROM usergroups_users uug WHERE uug.usergroup_id = ug.id
			) as users_count
		`)
	}
	qry.WriteString(`
		FROM usergroups ug
		LEFT JOIN usergroups_orgs_roles uor ON ug.id = uor.usergroup_id
		LEFT JOIN org_roles r ON uor.org_role_id = r.id
		WHERE ug.org_id=$1
	`)
	if filterRoleID != "" {
		qry.WriteString(" AND uor.org_role_id=$4")
		args = append(args, filterRoleID)
	}
	qry.WriteString(" AND lower(ug.name) > lower($2) ORDER BY lower(ug.name) LIMIT $3")

	var res []*database.MemberUsergroup
	err := c.getDB(ctx).SelectContext(ctx, &res, qry.String(), args...)
	if err != nil {
		return nil, parseErr("org groups", err)
	}
	return res, nil
}

func (c *connection) FindOrganizationMemberUsergroupRole(ctx context.Context, groupID, orgID string) (*database.OrganizationRole, error) {
	role := &database.OrganizationRole{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		SELECT r.* FROM usergroups_orgs_roles uor
		JOIN org_roles r ON r.id = uor.org_role_id
		WHERE uor.usergroup_id=$1 AND uor.org_id=$2
	`, groupID, orgID).StructScan(role)
	if err != nil {
		return nil, parseErr("org group member role", err)
	}
	return role, nil
}

func (c *connection) InsertOrganizationMemberUsergroup(ctx context.Context, groupID, orgID, roleID string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, `
		INSERT INTO usergroups_orgs_roles (usergroup_id, org_id, org_role_id) VALUES ($1, $2, $3)
	`, groupID, orgID, roleID)
	if err != nil {
		return parseErr("org group member", err)
	}
	return nil
}

func (c *connection) UpdateOrganizationMemberUsergroup(ctx context.Context, groupID, orgID, roleID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, `
		UPDATE usergroups_orgs_roles SET org_role_id = $3 WHERE usergroup_id = $1 AND org_id = $2
	`, groupID, orgID, roleID)
	return checkUpdateRow("org group member", res, err)
}

func (c *connection) DeleteOrganizationMemberUsergroup(ctx context.Context, groupID, orgID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM usergroups_orgs_roles WHERE usergroup_id = $1 AND org_id = $2", groupID, orgID)
	return checkDeleteRow("org group member", res, err)
}

func (c *connection) FindProjectMemberUsergroups(ctx context.Context, projectID, filterRoleID string, withCounts bool, afterName string, limit int) ([]*database.MemberUsergroup, error) {
	args := []any{projectID, afterName, limit}
	var qry strings.Builder
	qry.WriteString(`SELECT ug.id, ug.name, ug.managed, ug.created_on, ug.updated_on, r.name as "role_name"`)
	if withCounts {
		qry.WriteString(`,
			(
				SELECT COUNT(*) FROM usergroups_users uug WHERE uug.usergroup_id = ug.id
			) as users_count
		`)
	}
	qry.WriteString(`
		FROM usergroups ug
		JOIN usergroups_projects_roles upr ON ug.id = upr.usergroup_id
		JOIN project_roles r ON upr.project_role_id = r.id
		WHERE upr.project_id=$1
	`)
	if filterRoleID != "" {
		qry.WriteString(" AND upr.project_role_id=$4")
		args = append(args, filterRoleID)
	}
	qry.WriteString(" AND lower(ug.name) > lower($2) ORDER BY lower(ug.name) LIMIT $3")

	var res []*database.MemberUsergroup
	err := c.getDB(ctx).SelectContext(ctx, &res, qry.String(), args...)
	if err != nil {
		return nil, parseErr("project groups", err)
	}
	return res, nil
}

func (c *connection) FindProjectMemberUsergroupRole(ctx context.Context, groupID, projectID string) (*database.ProjectRole, error) {
	role := &database.ProjectRole{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		SELECT r.* FROM usergroups_projects_roles upr
		JOIN project_roles r ON r.id = upr.project_role_id
		WHERE upr.usergroup_id=$1 AND upr.project_id=$2
	`, groupID, projectID).StructScan(role)
	if err != nil {
		return nil, parseErr("project group member role", err)
	}
	return role, nil
}

func (c *connection) InsertProjectMemberUsergroup(ctx context.Context, groupID, projectID, roleID string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, `
		INSERT INTO usergroups_projects_roles (usergroup_id, project_id, project_role_id) VALUES ($1, $2, $3)
	`, groupID, projectID, roleID)
	if err != nil {
		return parseErr("project group member", err)
	}
	return nil
}

func (c *connection) UpdateProjectMemberUsergroup(ctx context.Context, groupID, projectID, roleID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, `
		UPDATE usergroups_projects_roles SET project_role_id = $3 WHERE usergroup_id = $1 AND project_id = $2
	`, groupID, projectID, roleID)
	return checkUpdateRow("project group member", res, err)
}

func (c *connection) DeleteProjectMemberUsergroup(ctx context.Context, groupID, projectID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM usergroups_projects_roles WHERE usergroup_id = $1 AND project_id = $2", groupID, projectID)
	return checkDeleteRow("project group member", res, err)
}

func (c *connection) DeleteProjectMemberUser(ctx context.Context, projectID, userID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM users_projects_roles WHERE user_id = $1 AND project_id = $2", userID, projectID)
	return checkDeleteRow("project member", res, err)
}

func (c *connection) DeleteAllProjectMemberUserForOrganization(ctx context.Context, orgID, userID string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM users_projects_roles upr WHERE upr.user_id = $1 AND upr.project_id IN (SELECT p.id FROM projects p WHERE p.org_id = $2)", userID, orgID)
	if err != nil {
		return parseErr("project member", err)
	}
	return nil
}

func (c *connection) UpdateProjectMemberUserRole(ctx context.Context, projectID, userID, roleID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, `UPDATE users_projects_roles SET project_role_id = $1 WHERE user_id = $2 AND project_id = $3`, roleID, userID, projectID)
	return checkUpdateRow("project member", res, err)
}

// UpsertProjectMemberServiceRole inserts or updates the role of a service in a project
func (c *connection) UpsertProjectMemberServiceRole(ctx context.Context, serviceID, projectID, roleID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, `
		INSERT INTO service_projects_roles (service_id, project_id, project_role_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (service_id, project_id) DO UPDATE SET project_role_id = $3
	`, serviceID, projectID, roleID)
	return checkUpdateRow("service project member", res, err)
}

func (c *connection) DeleteOrganizationMemberService(ctx context.Context, serviceID, orgID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, `
		DELETE FROM service_orgs_roles
		WHERE service_id = $1 AND org_id = $2
	`, serviceID, orgID)
	return checkDeleteRow("service org member", res, err)
}

// DeleteProjectMemberService removes a service from a project
func (c *connection) DeleteProjectMemberService(ctx context.Context, serviceID, projectID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, `
		DELETE FROM service_projects_roles
		WHERE service_id = $1 AND project_id = $2
	`, serviceID, projectID)
	return checkDeleteRow("service project member", res, err)
}

func (c *connection) FindOrganizationInvites(ctx context.Context, orgID, afterEmail string, limit int) ([]*database.OrganizationInviteWithRole, error) {
	var res []*database.OrganizationInviteWithRole
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT uoi.id, uoi.email, ur.name as role_name, u.email as invited_by
		FROM org_invites uoi
		JOIN org_roles ur ON uoi.org_role_id = ur.id
		LEFT JOIN users u ON uoi.invited_by_user_id = u.id
		WHERE uoi.org_id = $1 AND lower(uoi.email) > lower($2)
		ORDER BY lower(uoi.email) LIMIT $3
	`, orgID, afterEmail, limit)
	if err != nil {
		return nil, parseErr("org invites", err)
	}
	return res, nil
}

func (c *connection) CountOrganizationInvites(ctx context.Context, orgID string) (int, error) {
	var count int
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT COUNT(*) FROM org_invites WHERE org_id = $1", orgID).Scan(&count)
	if err != nil {
		return 0, parseErr("org invites count", err)
	}
	return count, nil
}

func (c *connection) FindOrganizationInvitesByEmail(ctx context.Context, userEmail string) ([]*database.OrganizationInvite, error) {
	var dtos []*organizationInviteDTO
	err := c.getDB(ctx).SelectContext(ctx, &dtos, "SELECT * FROM org_invites WHERE lower(email) = lower($1)", userEmail)
	if err != nil {
		return nil, parseErr("org invites", err)
	}
	res := make([]*database.OrganizationInvite, len(dtos))
	for i, dto := range dtos {
		var err error
		res[i], err = dto.AsModel()
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (c *connection) FindOrganizationInvite(ctx context.Context, orgID, userEmail string) (*database.OrganizationInvite, error) {
	dto := &organizationInviteDTO{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM org_invites WHERE lower(email) = lower($1) AND org_id = $2", userEmail, orgID).StructScan(dto)
	if err != nil {
		return nil, parseErr("org invite", err)
	}
	return dto.AsModel()
}

func (c *connection) InsertOrganizationInvite(ctx context.Context, opts *database.InsertOrganizationInviteOptions) error {
	if err := database.Validate(opts); err != nil {
		return err
	}

	var inviterID any
	if opts.InviterID != "" {
		inviterID = opts.InviterID
	}

	_, err := c.getDB(ctx).ExecContext(ctx, "INSERT INTO org_invites (email, invited_by_user_id, org_id, org_role_id) VALUES ($1, $2, $3, $4)", opts.Email, inviterID, opts.OrgID, opts.RoleID)
	if err != nil {
		return parseErr("org invite", err)
	}
	return nil
}

func (c *connection) UpdateOrganizationInviteUsergroups(ctx context.Context, id string, groupIDs []string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, `UPDATE org_invites SET usergroup_ids = $1 WHERE id = $2`, groupIDs, id)
	return checkUpdateRow("org invite", res, err)
}

func (c *connection) DeleteOrganizationInvite(ctx context.Context, id string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM org_invites WHERE id = $1", id)
	return checkDeleteRow("org invite", res, err)
}

func (c *connection) CountInvitesForOrganization(ctx context.Context, orgID string) (int, error) {
	var count int
	// count outstanding org invites as well as project invites for this org
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		SELECT COALESCE(SUM(total_count), 0) as total_count FROM (
  			SELECT COUNT(*) as total_count FROM org_invites WHERE org_id = $1
  			UNION ALL
  			SELECT COUNT(*) as total_count FROM project_invites WHERE project_id IN (SELECT id FROM projects WHERE org_id = $1)
		) as subquery
		`, orgID).Scan(&count)
	if err != nil {
		return 0, parseErr("invites count", err)
	}
	return count, nil
}

func (c *connection) UpdateOrganizationInviteRole(ctx context.Context, id, roleID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, `UPDATE org_invites SET org_role_id = $1 WHERE id = $2`, roleID, id)
	return checkUpdateRow("org invite", res, err)
}

func (c *connection) FindProjectInvites(ctx context.Context, projectID, afterEmail string, limit int) ([]*database.ProjectInviteWithRole, error) {
	var res []*database.ProjectInviteWithRole
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT upi.id, upi.email, upr.name as role_name, uor.name as org_role_name, u.email as invited_by
		FROM project_invites upi
		JOIN project_roles upr ON upi.project_role_id = upr.id
		LEFT JOIN users u ON upi.invited_by_user_id = u.id
		LEFT JOIN org_invites uoi ON upi.org_invite_id = uoi.id
		LEFT JOIN org_roles uor ON uoi.org_role_id = uor.id
		WHERE upi.project_id = $1 AND lower(upi.email) > lower($2)
		ORDER BY lower(upi.email) LIMIT $3
	`, projectID, afterEmail, limit)
	if err != nil {
		return nil, parseErr("project invites", err)
	}
	return res, nil
}

func (c *connection) FindProjectInvitesByEmail(ctx context.Context, userEmail string) ([]*database.ProjectInvite, error) {
	var res []*database.ProjectInvite
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT * FROM project_invites WHERE lower(email) = lower($1)", userEmail)
	if err != nil {
		return nil, parseErr("project invites", err)
	}
	return res, nil
}

func (c *connection) FindProjectInvite(ctx context.Context, projectID, userEmail string) (*database.ProjectInvite, error) {
	res := &database.ProjectInvite{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM project_invites WHERE lower(email) = lower($1) AND project_id = $2", userEmail, projectID).StructScan(res)
	if err != nil {
		return nil, parseErr("project invite", err)
	}
	return res, nil
}

func (c *connection) InsertProjectInvite(ctx context.Context, opts *database.InsertProjectInviteOptions) error {
	if err := database.Validate(opts); err != nil {
		return err
	}

	var inviterID any
	if opts.InviterID != "" {
		inviterID = opts.InviterID
	}

	_, err := c.getDB(ctx).ExecContext(ctx,
		`INSERT INTO project_invites (email, org_invite_id, project_id, project_role_id, invited_by_user_id) VALUES ($1, $2, $3, $4, $5)`,
		opts.Email, opts.OrgInviteID, opts.ProjectID, opts.RoleID, inviterID)
	if err != nil {
		return parseErr("project invite", err)
	}
	return nil
}

func (c *connection) DeleteProjectInvite(ctx context.Context, id string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM project_invites WHERE id = $1", id)
	return checkDeleteRow("project invite", res, err)
}

func (c *connection) UpdateProjectInviteRole(ctx context.Context, id, roleID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, `UPDATE project_invites SET project_role_id = $1 WHERE id = $2`, roleID, id)
	return checkUpdateRow("project invite", res, err)
}

func (c *connection) FindProjectAccessRequests(ctx context.Context, projectID, afterID string, limit int) ([]*database.ProjectAccessRequest, error) {
	var res []*database.ProjectAccessRequest
	var err error
	if afterID != "" {
		err = c.getDB(ctx).SelectContext(ctx, &res, `
			SELECT par.user_id
			FROM project_access_requests par
			WHERE par.project_id = $1 AND par.user_id > $2
			ORDER BY par.user_id LIMIT $3
		`, projectID, afterID, limit)
	} else {
		err = c.getDB(ctx).SelectContext(ctx, &res, `
			SELECT par.user_id
			FROM project_access_requests par
			WHERE par.project_id = $1
			ORDER BY par.user_id LIMIT $2
		`, projectID, limit)
	}
	if err != nil {
		return nil, parseErr("project access request", err)
	}
	return res, nil
}

func (c *connection) FindProjectAccessRequest(ctx context.Context, projectID, userID string) (*database.ProjectAccessRequest, error) {
	res := &database.ProjectAccessRequest{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM project_access_requests WHERE user_id = $1 AND project_id = $2", userID, projectID).StructScan(res)
	if err != nil {
		return nil, parseErr("project access request", err)
	}
	return res, nil
}

func (c *connection) FindProjectAccessRequestByID(ctx context.Context, id string) (*database.ProjectAccessRequest, error) {
	res := &database.ProjectAccessRequest{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM project_access_requests WHERE id=$1", id).StructScan(res)
	if err != nil {
		return nil, parseErr("project access request", err)
	}
	return res, nil
}

func (c *connection) InsertProjectAccessRequest(ctx context.Context, opts *database.InsertProjectAccessRequestOptions) (*database.ProjectAccessRequest, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.ProjectAccessRequest{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "INSERT INTO project_access_requests (user_id, project_id) VALUES ($1, $2) RETURNING *", opts.UserID, opts.ProjectID).StructScan(res)
	if err != nil {
		return nil, parseErr("project access request", err)
	}
	return res, nil
}

func (c *connection) DeleteProjectAccessRequest(ctx context.Context, id string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM project_access_requests WHERE id = $1", id)
	return checkDeleteRow("project access request", res, err)
}

// FindBookmarks returns a list of bookmarks for a user per project
func (c *connection) FindBookmarks(ctx context.Context, projectID, resourceKind, resourceName, userID string) ([]*database.Bookmark, error) {
	var res []*database.Bookmark
	err := c.getDB(ctx).SelectContext(ctx, &res, `SELECT * FROM bookmarks WHERE project_id = $1 and resource_kind = $2 and lower(resource_name) = lower($3) and (user_id = $4 or shared = true or "default" = true)`,
		projectID, resourceKind, resourceName, userID)
	if err != nil {
		return nil, parseErr("bookmarks", err)
	}
	return res, nil
}

// FindBookmark returns a bookmark for given bookmark id
func (c *connection) FindBookmark(ctx context.Context, bookmarkID string) (*database.Bookmark, error) {
	res := &database.Bookmark{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM bookmarks WHERE id = $1", bookmarkID).StructScan(res)
	if err != nil {
		return nil, parseErr("bookmarks", err)
	}
	return res, nil
}

func (c *connection) FindDefaultBookmark(ctx context.Context, projectID, resourceKind, resourceName string) (*database.Bookmark, error) {
	res := &database.Bookmark{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `SELECT * FROM bookmarks WHERE project_id = $1 and resource_kind = $2 and lower(resource_name) = lower($3) and "default" = true`,
		projectID, resourceKind, resourceName).StructScan(res)
	if err != nil {
		return nil, parseErr("bookmarks", err)
	}
	return res, nil
}

// InsertBookmark inserts a bookmark for a user per project
func (c *connection) InsertBookmark(ctx context.Context, opts *database.InsertBookmarkOptions) (*database.Bookmark, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.Bookmark{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `INSERT INTO bookmarks (display_name, description, url_search, resource_kind, resource_name, project_id, user_id, "default", shared)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING *`,
		opts.DisplayName, opts.Description, opts.URLSearch, opts.ResourceKind, opts.ResourceName, opts.ProjectID, opts.UserID, opts.Default, opts.Shared).StructScan(res)
	if err != nil {
		return nil, parseErr("bookmarks", err)
	}
	return res, nil
}

func (c *connection) UpdateBookmark(ctx context.Context, opts *database.UpdateBookmarkOptions) error {
	if err := database.Validate(opts); err != nil {
		return err
	}
	res, err := c.getDB(ctx).ExecContext(ctx, `UPDATE bookmarks SET display_name=$1, description=$2, url_search=$3, shared=$4 WHERE id=$5`,
		opts.DisplayName, opts.Description, opts.URLSearch, opts.Shared, opts.BookmarkID)
	return checkUpdateRow("bookmark", res, err)
}

// DeleteBookmark deletes a bookmark for a given bookmark id
func (c *connection) DeleteBookmark(ctx context.Context, bookmarkID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM bookmarks WHERE id = $1", bookmarkID)
	return checkDeleteRow("bookmarks", res, err)
}

func (c *connection) FindVirtualFiles(ctx context.Context, projectID, environment string, afterUpdatedOn time.Time, afterPath string, limit int) ([]*database.VirtualFile, error) {
	var res []*database.VirtualFile
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT path, data, deleted, updated_on
		FROM virtual_files
		WHERE project_id=$1 AND environment=$2 AND (updated_on>$3 OR updated_on=$3 AND path>$4)
		ORDER BY updated_on, path LIMIT $5
	`, projectID, environment, afterUpdatedOn, afterPath, limit)
	if err != nil {
		return nil, parseErr("virtual files", err)
	}
	return res, nil
}

func (c *connection) FindVirtualFile(ctx context.Context, projectID, environment, path string) (*database.VirtualFile, error) {
	res := &database.VirtualFile{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		SELECT path, data, deleted, updated_on
		FROM virtual_files
		WHERE project_id=$1 AND environment=$2 AND path=$3
	`, projectID, environment, path).StructScan(res)
	if err != nil {
		return nil, parseErr("virtual files", err)
	}
	return res, nil
}

func (c *connection) UpsertVirtualFile(ctx context.Context, opts *database.InsertVirtualFileOptions) error {
	if err := database.Validate(opts); err != nil {
		return err
	}

	_, err := c.getDB(ctx).ExecContext(ctx, `
		INSERT INTO virtual_files (project_id, environment, path, data, deleted)
		VALUES ($1, $2, $3, $4, FALSE)
		ON CONFLICT (project_id, environment, path) DO UPDATE SET
			data = EXCLUDED.data,
			deleted = FALSE,
			updated_on = now()
	`, opts.ProjectID, opts.Environment, opts.Path, opts.Data)
	if err != nil {
		return parseErr("virtual file", err)
	}
	return nil
}

func (c *connection) UpdateVirtualFileDeleted(ctx context.Context, projectID, environment, path string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, `
		UPDATE virtual_files SET
			data = ''::BYTEA,
			deleted = TRUE,
			updated_on = now()
		WHERE project_id=$1 AND environment=$2 AND path=$3`, projectID, environment, path)
	return checkUpdateRow("virtual file", res, err)
}

func (c *connection) DeleteExpiredVirtualFiles(ctx context.Context, retention time.Duration) error {
	_, err := c.getDB(ctx).ExecContext(ctx, `DELETE FROM virtual_files WHERE deleted AND updated_on + $1 < now()`, retention)
	return parseErr("virtual files", err)
}

func (c *connection) FindAsset(ctx context.Context, id string) (*database.Asset, error) {
	res := &database.Asset{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM assets WHERE id = $1", id).StructScan(res)
	if err != nil {
		return nil, parseErr("asset", err)
	}
	return res, nil
}

func (c *connection) InsertAsset(ctx context.Context, id, organizationID, path, ownerID string, public bool) (*database.Asset, error) {
	res := &database.Asset{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		INSERT INTO assets (id, org_id, path, owner_id, public)
		VALUES ($1, $2, $3, $4, $5) RETURNING *`,
		id, organizationID, path, ownerID, public,
	).StructScan(res)
	if err != nil {
		return nil, parseErr("asset", err)
	}
	return res, nil
}

func (c *connection) FindUnusedAssets(ctx context.Context, limit int) ([]*database.Asset, error) {
	var res []*database.Asset
	// find assets that are not associated with any project or org
	// skip assets that are less thans 7 days old to avoid deleting assets for projects
	// that were accidentally deleted and may need to be restored
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT a.* FROM assets a 
		WHERE a.created_on < now() - INTERVAL '7 DAYS'
		AND NOT EXISTS (SELECT 1 FROM projects p WHERE p.archive_asset_id = a.id)
		AND NOT EXISTS (SELECT 1 FROM orgs o WHERE o.logo_asset_id = a.id)
		AND NOT EXISTS (SELECT 1 FROM orgs o WHERE o.favicon_asset_id = a.id)
		AND NOT EXISTS (SELECT 1 FROM orgs o WHERE o.thumbnail_asset_id = a.id)
		ORDER BY a.created_on DESC LIMIT $1
	`, limit)
	if err != nil {
		return nil, parseErr("assets", err)
	}
	return res, nil
}

func (c *connection) DeleteAssets(ctx context.Context, ids []string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM assets WHERE id=ANY($1)", ids)
	return parseErr("asset", err)
}

func (c *connection) FindOrganizationIDsWithBilling(ctx context.Context) ([]string, error) {
	var res []string
	err := c.getDB(ctx).SelectContext(ctx, &res, `SELECT id FROM orgs WHERE billing_customer_id <> ''`)
	if err != nil {
		return nil, parseErr("billing orgs", err)
	}
	return res, nil
}

func (c *connection) FindOrganizationIDsWithoutBilling(ctx context.Context) ([]string, error) {
	var res []string
	err := c.getDB(ctx).SelectContext(ctx, &res, `SELECT id FROM orgs WHERE billing_customer_id = ''`)
	if err != nil {
		return nil, parseErr("billing orgs without billing or payment info", err)
	}
	return res, nil
}

func (c *connection) CountBillingProjectsForOrganization(ctx context.Context, orgID string, createdBefore time.Time) (int, error) {
	var count int
	err := c.getDB(ctx).QueryRowxContext(ctx, `SELECT COUNT(*) FROM projects WHERE org_id = $1 AND prod_deployment_id IS NOT NULL AND created_on < $2`, orgID, createdBefore).Scan(&count)
	if err != nil {
		return 0, parseErr("billing projects", err)
	}
	return count, nil
}

func (c *connection) FindBillingUsageReportedOn(ctx context.Context) (time.Time, error) {
	var usageReportedOn sql.NullTime
	err := c.getDB(ctx).QueryRowxContext(ctx, `SELECT usage_reported_on FROM billing_reporting_time`).Scan(&usageReportedOn)
	if err != nil {
		return time.Time{}, parseErr("billing usage", err)
	}
	if !usageReportedOn.Valid {
		return time.Time{}, nil
	}
	return usageReportedOn.Time, nil
}

func (c *connection) UpdateBillingUsageReportedOn(ctx context.Context, usageReportedOn time.Time) error {
	res, err := c.getDB(ctx).ExecContext(ctx, `UPDATE billing_reporting_time SET usage_reported_on=$1`, usageReportedOn)
	return checkUpdateRow("billing usage", res, err)
}

func (c *connection) FindOrganizationForPaymentCustomerID(ctx context.Context, customerID string) (*database.Organization, error) {
	res := &database.Organization{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `SELECT * FROM orgs WHERE payment_customer_id = $1`, customerID).StructScan(res)
	if err != nil {
		return nil, parseErr("billing org for payment id", err)
	}
	return res, nil
}

func (c *connection) FindOrganizationForBillingCustomerID(ctx context.Context, customerID string) (*database.Organization, error) {
	res := &database.Organization{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `SELECT * FROM orgs WHERE billing_customer_id = $1`, customerID).StructScan(res)
	if err != nil {
		return nil, parseErr("billing org for billing id", err)
	}
	return res, nil
}

func (c *connection) FindBillingIssuesForOrg(ctx context.Context, orgID string) ([]*database.BillingIssue, error) {
	var res []*billingIssueDTO
	err := c.db.SelectContext(ctx, &res, `SELECT * FROM billing_issues WHERE org_id = $1`, orgID)
	if err != nil {
		return nil, parseErr("billing issues", err)
	}

	var billingErrors []*database.BillingIssue
	for _, dto := range res {
		billingErrors = append(billingErrors, dto.AsModel())
	}
	return billingErrors, nil
}

func (c *connection) FindBillingIssueByTypeForOrg(ctx context.Context, orgID string, errorType database.BillingIssueType) (*database.BillingIssue, error) {
	res := &billingIssueDTO{}
	err := c.db.GetContext(ctx, res, `SELECT * FROM billing_issues WHERE org_id = $1 AND type = $2`, orgID, errorType)
	if err != nil {
		return nil, parseErr("billing issue", err)
	}
	return res.AsModel(), nil
}

func (c *connection) FindBillingIssueByType(ctx context.Context, errorType database.BillingIssueType) ([]*database.BillingIssue, error) {
	var res []*billingIssueDTO
	err := c.db.SelectContext(ctx, &res, `SELECT * FROM billing_issues WHERE type = $1`, errorType)
	if err != nil {
		return nil, parseErr("billing issues", err)
	}

	var billingErrors []*database.BillingIssue
	for _, dto := range res {
		billingErrors = append(billingErrors, dto.AsModel())
	}
	return billingErrors, nil
}

func (c *connection) FindBillingIssueByTypeAndOverdueProcessed(ctx context.Context, errorType database.BillingIssueType, overdueProcessed bool) ([]*database.BillingIssue, error) {
	var res []*billingIssueDTO
	err := c.db.SelectContext(ctx, &res, `SELECT * FROM billing_issues WHERE type = $1 AND overdue_processed = $2`, errorType, overdueProcessed)
	if err != nil {
		return nil, parseErr("billing issues", err)
	}

	var billingErrors []*database.BillingIssue
	for _, dto := range res {
		billingErrors = append(billingErrors, dto.AsModel())
	}
	return billingErrors, nil
}

func (c *connection) UpsertBillingIssue(ctx context.Context, opts *database.UpsertBillingIssueOptions) (*database.BillingIssue, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	metadata, err := json.Marshal(opts.Metadata)
	if err != nil {
		return nil, err
	}

	temp := &billingIssueDTO{
		OrgID:     opts.OrgID,
		Type:      opts.Type,
		Metadata:  metadata,
		EventTime: opts.EventTime,
	}

	temp.Level = temp.getBillingIssueLevel()

	res := &billingIssueDTO{}
	err = c.getDB(ctx).QueryRowxContext(ctx, `INSERT INTO billing_issues (org_id, type, level, metadata, event_time) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (org_id, type) DO UPDATE SET metadata = $4, event_time = $5 RETURNING *`, temp.OrgID, temp.Type, temp.Level, temp.Metadata, temp.EventTime).StructScan(res)
	if err != nil {
		return nil, parseErr("billing issue", err)
	}
	return res.AsModel(), nil
}

func (c *connection) UpdateBillingIssueOverdueAsProcessed(ctx context.Context, id string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, `UPDATE billing_issues SET overdue_processed = true WHERE id = $1`, id)
	return checkUpdateRow("billing issue", res, err)
}

func (c *connection) DeleteBillingIssue(ctx context.Context, id string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM billing_issues WHERE id = $1", id)
	return checkDeleteRow("billing issue", res, err)
}

func (c *connection) DeleteBillingIssueByTypeForOrg(ctx context.Context, orgID string, errorType database.BillingIssueType) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM billing_issues WHERE org_id = $1 AND type = $2", orgID, errorType)
	return checkDeleteRow("billing issue", res, err)
}

func (c *connection) FindProjectVariables(ctx context.Context, projectID string, environment *string) ([]*database.ProjectVariable, error) {
	q := `SELECT * FROM project_variables p WHERE p.project_id = $1`
	args := []interface{}{projectID}
	if environment != nil {
		// Also include variables that are not environment specific and not set for the given environment
		q += `
			AND (
				p.environment = $2
				OR (
					p.environment = ''
					AND NOT EXISTS (
						SELECT 1
						FROM project_variables p2
						WHERE p2.project_id = p.project_id
						AND p2.environment = $2
						AND lower(p2.name) = lower(p.name)
					)
				)
			)
		`
		args = append(args, environment)
	}
	var res []*database.ProjectVariable
	err := c.getDB(ctx).SelectContext(ctx, &res, q, args...)
	if err != nil {
		return nil, parseErr("project variables", err)
	}

	// Decrypt the variables
	err = c.decryptProjectVariables(res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *connection) UpsertProjectVariable(ctx context.Context, projectID, environment string, vars map[string]string, userID string) ([]*database.ProjectVariable, error) {
	query := `INSERT INTO project_variables (project_id, environment, name, value, value_encryption_key_id, updated_by_user_id, updated_on)
	VALUES %s
	ON CONFLICT (project_id, environment, lower(name)) DO UPDATE SET
		value = EXCLUDED.value,
		value_encryption_key_id = EXCLUDED.value_encryption_key_id,
		updated_by_user_id = EXCLUDED.updated_by_user_id,
		updated_on = now() RETURNING *`

	var placeholders strings.Builder
	args := []any{projectID, environment, userID}
	i := 3
	for key, value := range vars {
		// Encrypt the variables
		encryptedValue, valueEncryptionKeyID, err := c.encrypt([]byte(value))
		if err != nil {
			return nil, err
		}

		if valueEncryptionKeyID != "" {
			value = base64.StdEncoding.EncodeToString(encryptedValue)
		}
		args = append(args, key, value, valueEncryptionKeyID)
		if placeholders.Len() > 0 {
			placeholders.WriteString(", ")
		}
		fmt.Fprintf(&placeholders, "($1, $2, $%d, $%d, $%d, $3, now())", i+1, i+2, i+3) // project_id, environment, name, value, value_encryption_key_id, updated_by_user_id, updated_on
		i += 3
	}

	var res []*database.ProjectVariable
	err := c.getDB(ctx).SelectContext(ctx, &res, fmt.Sprintf(query, placeholders.String()), args...)
	if err != nil {
		return nil, parseErr("project variables", err)
	}

	// Decrypt the variables
	err = c.decryptProjectVariables(res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *connection) DeleteProjectVariables(ctx context.Context, projectID, environment string, names []string) error {
	if len(names) == 0 {
		return fmt.Errorf("no names provided to delete project variables")
	}

	placeholders := make([]string, len(names))
	args := make([]interface{}, len(names)+2)
	args[0] = projectID
	args[1] = environment
	for i, name := range names {
		placeholders[i] = fmt.Sprintf("lower($%d)", i+3)
		args[i+2] = name
	}

	query := fmt.Sprintf("DELETE FROM project_variables WHERE project_id = $1 AND environment = $2 AND lower(name) IN (%s)", strings.Join(placeholders, ","))
	_, err := c.getDB(ctx).ExecContext(ctx, query, args...)
	return err
}

func (c *connection) FindProvisionerResourcesForDeployment(ctx context.Context, deploymentID string) ([]*database.ProvisionerResource, error) {
	var res []*provisionerResourceDTO
	err := c.getDB(ctx).SelectContext(ctx, &res, `SELECT * FROM provisioner_resources WHERE deployment_id = $1`, deploymentID)
	if err != nil {
		return nil, parseErr("provisioner resources", err)
	}
	return c.provisionerResourcesFromDTOs(res)
}

func (c *connection) FindProvisionerResourceByTypeAndName(ctx context.Context, deploymentID, typ, name string) (*database.ProvisionerResource, error) {
	res := &provisionerResourceDTO{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `SELECT * FROM provisioner_resources WHERE deployment_id = $1 AND "type" = $2 AND name = $3`, deploymentID, typ, name).StructScan(res)
	if err != nil {
		return nil, parseErr("provisioner resource", err)
	}
	return c.provisionerResourceFromDTO(res)
}

func (c *connection) InsertProvisionerResource(ctx context.Context, opts *database.InsertProvisionerResourceOptions) (*database.ProvisionerResource, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	args, err := json.Marshal(opts.Args)
	if err != nil {
		return nil, err
	}
	state, err := json.Marshal(opts.State)
	if err != nil {
		return nil, err
	}
	config, err := json.Marshal(opts.Config)
	if err != nil {
		return nil, err
	}

	res := &provisionerResourceDTO{}
	err = c.getDB(ctx).QueryRowxContext(ctx, `
		INSERT INTO provisioner_resources (id, deployment_id, "type", name, status, status_message, provisioner, args_json, state_json, config_json)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING *`,
		opts.ID, opts.DeploymentID, opts.Type, opts.Name, opts.Status, opts.StatusMessage, opts.Provisioner, args, state, config,
	).StructScan(res)
	if err != nil {
		return nil, parseErr("provisioner resource", err)
	}
	return c.provisionerResourceFromDTO(res)
}

func (c *connection) UpdateProvisionerResource(ctx context.Context, id string, opts *database.UpdateProvisionerResourceOptions) (*database.ProvisionerResource, error) {
	args, err := json.Marshal(opts.Args)
	if err != nil {
		return nil, err
	}
	state, err := json.Marshal(opts.State)
	if err != nil {
		return nil, err
	}
	config, err := json.Marshal(opts.Config)
	if err != nil {
		return nil, err
	}

	res := &provisionerResourceDTO{}
	err = c.getDB(ctx).QueryRowxContext(ctx, `
		UPDATE provisioner_resources SET status = $1, status_message = $2, args_json = $3, state_json = $4, config_json = $5, updated_on = now() WHERE id = $6 RETURNING *`,
		opts.Status, opts.StatusMessage, args, state, config, id,
	).StructScan(res)
	if err != nil {
		return nil, parseErr("provisioner resource", err)
	}
	return c.provisionerResourceFromDTO(res)
}

func (c *connection) DeleteProvisionerResource(ctx context.Context, id string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM provisioner_resources WHERE id = $1", id)
	return checkDeleteRow("provisioner resource", res, err)
}

func (c *connection) FindManagedGitRepo(ctx context.Context, remote string) (*database.ManagedGitRepo, error) {
	res := &database.ManagedGitRepo{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM managed_git_repos WHERE remote = $1", remote).StructScan(res)
	if err != nil {
		return nil, parseErr("managed git repo", err)
	}
	return res, nil
}

func (c *connection) FindUnusedManagedGitRepos(ctx context.Context, pageSize int) ([]*database.ManagedGitRepo, error) {
	// find managed github repos that are not associated with any project
	// skip repos that are less than 7 days old to avoid deleting repos for projects
	// that were accidentally deleted and may need to be restored
	var res []*database.ManagedGitRepo
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT * FROM managed_git_repos m
		WHERE updated_on < now() - INTERVAL '7 DAYS'
		AND (
			m.org_id IS NULL 
			OR NOT EXISTS (SELECT 1 FROM projects p WHERE p.managed_git_repo_id = m.id)
		)
		ORDER BY updated_on DESC
		LIMIT $1
	`, pageSize)
	if err != nil {
		return nil, parseErr("managed git repo", err)
	}
	return res, nil
}

func (c *connection) CountManagedGitRepos(ctx context.Context, orgID string) (int, error) {
	var count int
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		SELECT COUNT(*)
		FROM managed_git_repos m
		WHERE org_id = $1
	`, orgID).Scan(&count)
	if err != nil {
		return 0, parseErr("managed git repo count", err)
	}
	return count, nil
}

func (c *connection) InsertManagedGitRepo(ctx context.Context, opts *database.InsertManagedGitRepoOptions) (*database.ManagedGitRepo, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.ManagedGitRepo{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		INSERT INTO managed_git_repos (org_id, remote, owner_id)
		VALUES ($1, $2, $3) RETURNING *`,
		opts.OrgID, opts.Remote, opts.OwnerID,
	).StructScan(res)
	if err != nil {
		return nil, parseErr("managed git repo", err)
	}
	return res, nil
}

func (c *connection) DeleteManagedGitRepos(ctx context.Context, ids []string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM managed_git_repos WHERE id = ANY($1)", ids)
	return parseErr("managed git repo", err)
}

func (c *connection) FindGitRepoTransfer(ctx context.Context, remote string) (*database.GitRepoTransfer, error) {
	res := &database.GitRepoTransfer{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM git_repo_transfers WHERE from_git_remote = $1", remote).StructScan(res)
	if err != nil {
		return nil, parseErr("git repo transfer", err)
	}
	return res, nil
}

func (c *connection) InsertGitRepoTransfer(ctx context.Context, fromRemote, toRemote string) (*database.GitRepoTransfer, error) {
	res := &database.GitRepoTransfer{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		INSERT INTO git_repo_transfers (from_git_remote, to_git_remote)
		VALUES ($1, $2) RETURNING *`,
		fromRemote, toRemote,
	).StructScan(res)
	if err != nil {
		return nil, parseErr("git repo transfer", err)
	}
	return res, nil
}

// projectDTO wraps database.Project, using the pgtype package to handle types that pgx can't read directly into their native Go types.
type projectDTO struct {
	*database.Project
	ProdVariables pgtype.JSON `db:"prod_variables"`
	Annotations   pgtype.JSON `db:"annotations"`
}

func (c *connection) projectFromDTO(dto *projectDTO) (*database.Project, error) {
	err := dto.Annotations.AssignTo(&dto.Project.Annotations)
	if err != nil {
		return nil, err
	}

	return dto.Project, nil
}

func (c *connection) projectsFromDTOs(dtos []*projectDTO) ([]*database.Project, error) {
	res := make([]*database.Project, len(dtos))
	for i, dto := range dtos {
		var err error
		res[i], err = c.projectFromDTO(dto)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

// provisionerResourceDTO wraps database.ProvisionerResource, using the pgtype package to handle types that pgx can't read directly into their native Go types.
type provisionerResourceDTO struct {
	*database.ProvisionerResource
	Args   pgtype.JSON `db:"args_json"`
	State  pgtype.JSON `db:"state_json"`
	Config pgtype.JSON `db:"config_json"`
}

func (c *connection) provisionerResourceFromDTO(dto *provisionerResourceDTO) (*database.ProvisionerResource, error) {
	err := dto.Args.AssignTo(&dto.ProvisionerResource.Args)
	if err != nil {
		return nil, err
	}
	err = dto.State.AssignTo(&dto.ProvisionerResource.State)
	if err != nil {
		return nil, err
	}
	err = dto.Config.AssignTo(&dto.ProvisionerResource.Config)
	if err != nil {
		return nil, err
	}
	return dto.ProvisionerResource, nil
}

func (c *connection) provisionerResourcesFromDTOs(dtos []*provisionerResourceDTO) ([]*database.ProvisionerResource, error) {
	res := make([]*database.ProvisionerResource, len(dtos))
	for i, dto := range dtos {
		var err error
		res[i], err = c.provisionerResourceFromDTO(dto)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

// magicAuthTokenDTO wraps database.MagicAuthToken, using the pgtype package to handly types that pgx can't read directly into their native Go types.
type magicAuthTokenDTO struct {
	*database.MagicAuthToken
	Attributes pgtype.JSON      `db:"attributes"`
	Fields     pgtype.TextArray `db:"fields"`
	Resources  pgtype.JSONB     `db:"resources"`
}

func (c *connection) magicAuthTokenFromDTO(dto *magicAuthTokenDTO, fetchSecret bool) (*database.MagicAuthToken, error) {
	err := dto.Attributes.AssignTo(&dto.MagicAuthToken.Attributes)
	if err != nil {
		return nil, err
	}
	err = dto.Fields.AssignTo(&dto.MagicAuthToken.Fields)
	if err != nil {
		return nil, err
	}
	err = dto.Resources.AssignTo(&dto.MagicAuthToken.Resources)
	if err != nil {
		return nil, err
	}

	if fetchSecret {
		dto.MagicAuthToken.Secret, err = c.decrypt(dto.MagicAuthToken.Secret, dto.MagicAuthToken.SecretEncryptionKeyID)
		if err != nil {
			return nil, err
		}
	} else {
		dto.MagicAuthToken.Secret = nil
		dto.MagicAuthToken.SecretEncryptionKeyID = ""
	}

	return dto.MagicAuthToken, nil
}

// magicAuthTokenWithUserDTO wraps database.MagicAuthTokenWithUser, using the pgtype package to handly types that pgx can't read directly into their native Go types.
type magicAuthTokenWithUserDTO struct {
	*database.MagicAuthTokenWithUser
	Attributes pgtype.JSON      `db:"attributes"`
	Fields     pgtype.TextArray `db:"fields"`
	Resources  pgtype.JSONB     `db:"resources"`
}

func (c *connection) magicAuthTokenWithUserFromDTO(dto *magicAuthTokenWithUserDTO) (*database.MagicAuthTokenWithUser, error) {
	err := dto.Attributes.AssignTo(&dto.MagicAuthTokenWithUser.Attributes)
	if err != nil {
		return nil, err
	}
	err = dto.Fields.AssignTo(&dto.MagicAuthToken.Fields)
	if err != nil {
		return nil, err
	}
	err = dto.Resources.AssignTo(&dto.MagicAuthToken.Resources)
	if err != nil {
		return nil, err
	}

	dto.MagicAuthTokenWithUser.Secret, err = c.decrypt(dto.MagicAuthTokenWithUser.Secret, dto.MagicAuthTokenWithUser.SecretEncryptionKeyID)
	if err != nil {
		return nil, err
	}

	return dto.MagicAuthTokenWithUser, nil
}

type notificationTokenWithSecretDTO struct {
	*database.NotificationTokenWithSecret
	SecretEncryptionKeyID string `db:"secret_encryption_key_id"`
}

func (c *connection) notificationTokenWithSecretFromDTO(dto *notificationTokenWithSecretDTO) (*database.NotificationTokenWithSecret, error) {
	if dto.SecretEncryptionKeyID == "" {
		return dto.NotificationTokenWithSecret, nil
	}
	decrypted, err := c.decrypt(dto.NotificationTokenWithSecret.MagicAuthTokenSecret, dto.SecretEncryptionKeyID)
	if err != nil {
		return nil, err
	}
	dto.NotificationTokenWithSecret.MagicAuthTokenSecret = decrypted
	return dto.NotificationTokenWithSecret, nil
}

type organizationInviteDTO struct {
	*database.OrganizationInvite
	UsergroupIDs pgtype.TextArray `db:"usergroup_ids"`
}

func (o *organizationInviteDTO) AsModel() (*database.OrganizationInvite, error) {
	err := o.UsergroupIDs.AssignTo(&o.OrganizationInvite.UsergroupIDs)
	if err != nil {
		return nil, err
	}

	return o.OrganizationInvite, nil
}

type authClientDTO struct {
	*database.AuthClient
	GrantTypes pgtype.TextArray `db:"grant_types"`
}

func (dto *authClientDTO) AsModel() (*database.AuthClient, error) {
	if dto.AuthClient == nil {
		dto.AuthClient = &database.AuthClient{}
	}
	if err := dto.GrantTypes.AssignTo(&dto.AuthClient.GrantTypes); err != nil {
		return nil, err
	}
	return dto.AuthClient, nil
}

type billingIssueDTO struct {
	ID               string                     `db:"id"`
	OrgID            string                     `db:"org_id"`
	Type             database.BillingIssueType  `db:"type"`
	Level            database.BillingIssueLevel `db:"level"`
	Metadata         json.RawMessage            `db:"metadata"`
	OverdueProcessed bool                       `db:"overdue_processed"`
	EventTime        time.Time                  `db:"event_time"`
	CreatedOn        time.Time                  `db:"created_on"`
}

func (b *billingIssueDTO) AsModel() *database.BillingIssue {
	var metadata database.BillingIssueMetadata
	switch b.Type {
	case database.BillingIssueTypeOnTrial:
		metadata = &database.BillingIssueMetadataOnTrial{}
	case database.BillingIssueTypeTrialEnded:
		metadata = &database.BillingIssueMetadataTrialEnded{}
	case database.BillingIssueTypeNoPaymentMethod:
		metadata = &database.BillingIssueMetadataNoPaymentMethod{}
	case database.BillingIssueTypeNoBillableAddress:
		metadata = &database.BillingIssueMetadataNoBillableAddress{}
	case database.BillingIssueTypePaymentFailed:
		metadata = &database.BillingIssueMetadataPaymentFailed{}
	case database.BillingIssueTypeSubscriptionCancelled:
		metadata = &database.BillingIssueMetadataSubscriptionCancelled{}
	case database.BillingIssueTypeNeverSubscribed:
		metadata = &database.BillingIssueMetadataNeverSubscribed{}
	default:
	}
	if err := json.Unmarshal(b.Metadata, &metadata); err != nil {
		return nil
	}
	return &database.BillingIssue{
		ID:        b.ID,
		OrgID:     b.OrgID,
		Type:      b.Type,
		Level:     b.Level,
		Metadata:  metadata,
		EventTime: b.EventTime,
		CreatedOn: b.CreatedOn,
	}
}

func (b *billingIssueDTO) getBillingIssueLevel() database.BillingIssueLevel {
	if b.Type == database.BillingIssueTypeUnspecified {
		return database.BillingIssueLevelUnspecified
	}
	if b.Type == database.BillingIssueTypeOnTrial {
		return database.BillingIssueLevelWarning
	}
	return database.BillingIssueLevelError
}

type serviceDTO struct {
	*database.Service
	Attributes pgtype.JSON `db:"attributes"`
}

func (d *serviceDTO) serviceFromDTO() (*database.Service, error) {
	err := d.Attributes.AssignTo(&d.Service.Attributes)
	if err != nil {
		return nil, err
	}
	return d.Service, nil
}

type organizationMemberServiceDTO struct {
	*database.OrganizationMemberService
	Attributes pgtype.JSON `db:"attributes"`
}

func (d *organizationMemberServiceDTO) organizationMemberServiceFromDTO() (*database.OrganizationMemberService, error) {
	err := d.Attributes.AssignTo(&d.OrganizationMemberService.Attributes)
	if err != nil {
		return nil, err
	}
	return d.OrganizationMemberService, nil
}

type projectMemberServiceDTO struct {
	*database.ProjectMemberService
	Attributes pgtype.JSON `db:"attributes"`
}

func (d *projectMemberServiceDTO) projectMemberServiceFromDTO() (*database.ProjectMemberService, error) {
	err := d.Attributes.AssignTo(&d.ProjectMemberService.Attributes)
	if err != nil {
		return nil, err
	}
	return d.ProjectMemberService, nil
}

type projectMemberServiceWithProjectDTO struct {
	*database.ProjectMemberServiceWithProject
	Attributes pgtype.JSON `db:"attributes"`
}

func (d *projectMemberServiceWithProjectDTO) projectMemberServiceWithProjectFromDTO() (*database.ProjectMemberServiceWithProject, error) {
	err := d.Attributes.AssignTo(&d.ProjectMemberServiceWithProject.Attributes)
	if err != nil {
		return nil, err
	}
	return d.ProjectMemberServiceWithProject, nil
}

type organizationMemberUserDTO struct {
	*database.OrganizationMemberUser
	Attributes pgtype.JSON `db:"attributes"`
}

func (dto *organizationMemberUserDTO) organizationMemberUserFromDTO() (*database.OrganizationMemberUser, error) {
	user := &database.OrganizationMemberUser{
		ID:              dto.ID,
		Email:           dto.Email,
		DisplayName:     dto.DisplayName,
		PhotoURL:        dto.PhotoURL,
		RoleName:        dto.RoleName,
		ProjectsCount:   dto.ProjectsCount,
		UsergroupsCount: dto.UsergroupsCount,
		CreatedOn:       dto.CreatedOn,
		UpdatedOn:       dto.UpdatedOn,
	}

	// Handle Attributes: Normalize NULL JSONB to empty map
	var attrs map[string]any
	if err := dto.Attributes.AssignTo(&attrs); err != nil {
		return nil, err
	}
	if attrs == nil {
		attrs = make(map[string]any)
	}
	user.Attributes = attrs

	return user, nil
}

type userWithAttributesDTO struct {
	*database.User
	Attributes pgtype.JSON `db:"attributes"`
}

func (dto *userWithAttributesDTO) userWithAttributesFromDTO() (*database.User, map[string]any, error) {
	// Handle Attributes: Normalize NULL JSONB to empty map
	var attrs map[string]any
	if err := dto.Attributes.AssignTo(&attrs); err != nil {
		return nil, nil, err
	}
	if attrs == nil {
		attrs = make(map[string]any)
	}

	return dto.User, attrs, nil
}

func (c *connection) decryptProjectVariables(res []*database.ProjectVariable) error {
	for _, v := range res {
		if v.ValueEncryptionKeyID == "" {
			continue
		}
		dec, err := base64.StdEncoding.DecodeString(v.Value)
		if err != nil {
			return err
		}

		decryptedValue, err := c.decrypt(dec, v.ValueEncryptionKeyID)
		if err != nil {
			return err
		}

		v.Value = string(decryptedValue)
	}
	return nil
}

// returns the encrypted text and the encryption key id used. The first key in the keyring is used for encryption. If the keyring is empty, the text is returned as is along with an empty key id.
func (c *connection) encrypt(text []byte) ([]byte, string, error) {
	if len(c.encKeyring) == 0 {
		return text, "", nil
	}
	// use the first key in the keyring for encryption
	encrypted, err := encrypt(text, c.encKeyring[0].Secret)
	if err != nil {
		return nil, "", err
	}
	return encrypted, c.encKeyring[0].ID, nil
}

// returns the decrypted text, using the encryption key id provided. If the key id is empty, the text is returned as is.
func (c *connection) decrypt(text []byte, encKeyID string) ([]byte, error) {
	if encKeyID == "" {
		return text, nil
	}
	var encKey *database.EncryptionKey
	for _, key := range c.encKeyring {
		if key.ID == encKeyID {
			encKey = key
			break
		}
	}
	if encKey == nil {
		return nil, fmt.Errorf("encryption key id %s not found in keyring", encKeyID)
	}
	return decrypt(text, encKey.Secret)
}

func checkUpdateRow(target string, res sql.Result, err error) error {
	if err != nil {
		return parseErr(target, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return parseErr(target, err)
	}
	if n == 0 {
		return parseErr(target, sql.ErrNoRows)
	}
	if n > 1 {
		// This should never happen
		panic(fmt.Errorf("expected to update 1 row, but updated %d", n))
	}
	return nil
}

func checkDeleteRow(target string, res sql.Result, err error) error {
	if err != nil {
		return parseErr(target, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return parseErr(target, err)
	}
	if n == 0 {
		return parseErr(target, sql.ErrNoRows)
	}
	if n > 1 {
		// This should never happen
		panic(fmt.Errorf("expected to delete 1 row, but deleted %d", n))
	}
	return nil
}

func parseErr(target string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		if target == "" {
			return database.ErrNotFound
		}
		return database.NewNotFoundError(fmt.Sprintf("%s not found", target))
	}
	var pgerr *pgconn.PgError
	if !errors.As(err, &pgerr) {
		return err
	}
	if pgerr.Code == "23505" { // unique_violation
		switch pgerr.ConstraintName {
		case "orgs_name_idx":
			return database.NewNotUniqueError("an org with that name already exists")
		case "projects_name_idx":
			return database.NewNotUniqueError("a project with that name already exists in the org")
		case "users_email_idx":
			return database.NewNotUniqueError("a user with that email already exists")
		case "usergroups_name_idx":
			return database.NewNotUniqueError("a usergroup with that name already exists in the org")
		case "usergroups_users_pkey":
			return database.NewNotUniqueError("user is already a member of the usergroup")
		case "users_orgs_roles_pkey":
			return database.NewNotUniqueError("user is already a member of the org")
		case "users_projects_roles_pkey":
			return database.NewNotUniqueError("user is already a member of the project")
		case "usergroups_orgs_roles_pkey":
			return database.NewNotUniqueError("usergroup is already a member of the org")
		case "usergroups_projects_roles_pkey":
			return database.NewNotUniqueError("usergroup is already a member of the project")
		case "org_invites_email_org_idx":
			return database.NewNotUniqueError("email has already been invited to the org")
		case "project_invites_email_project_idx":
			return database.NewNotUniqueError("email has already been invited to the project")
		case "orgs_autoinvite_domains_org_id_domain_idx":
			return database.NewNotUniqueError("domain has already been added for the org")
		case "service_name_idx":
			return database.NewNotUniqueError("a service with that name already exists in the org")
		case "virtual_files_pkey":
			return database.NewNotUniqueError("a virtual file already exists at that path")
		default:
			if target == "" {
				return database.ErrNotUnique
			}
			return database.NewNotUniqueError(fmt.Sprintf("%s already exists", target))
		}
	}
	return err
}

// encrypts plaintext using AES-GCM with the given key and returns the base64 encoded ciphertext
func encrypt(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// nonce is prepended to the ciphertext, so it can be used for decryption
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decrypts the ciphertext using AES-GCM with the given key and returns the plaintext
func decrypt(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	d, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return d, nil
}

// validateAttributes validates keys and values of an attributes map, handling nil input
func (c *connection) validateAttributes(attributes map[string]any) (map[string]any, error) {
	if attributes == nil {
		return make(map[string]any), nil
	}

	if len(attributes) > 50 {
		return nil, fmt.Errorf("too many attributes: maximum 50 allowed, got %d", len(attributes))
	}

	for key, value := range attributes {
		// Validate key format
		if !isValidAttributeKey(key) {
			return nil, fmt.Errorf("invalid attribute key '%s': must contain only alphanumeric characters and underscores", key)
		}

		// Validate value length
		if str, ok := value.(string); ok && len(str) > 256 {
			return nil, fmt.Errorf("attribute value for key '%s' too long: maximum 256 characters, got %d", key, len(str))
		}
	}
	return attributes, nil
}

// isValidAttributeKey checks if an attribute key contains only alphanumeric characters and underscores
func isValidAttributeKey(key string) bool {
	if key == "" {
		return false
	}
	for _, r := range key {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}
	return true
}
