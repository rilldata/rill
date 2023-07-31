package auth

import (
	"context"
	"fmt"
	"sync"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/authtoken"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
)

// OwnerType is an enum of types of claim owners
type OwnerType string

const (
	OwnerTypeAnon OwnerType = "anon"
	OwnerTypeUser OwnerType = "user"
)

// Claims resolves permissions for a requester.
type Claims interface {
	OwnerType() OwnerType
	OwnerID() string
	AuthTokenID() string
	Superuser(ctx context.Context) bool
	OrganizationPermissions(ctx context.Context, orgID string) *adminv1.OrganizationPermissions
	ProjectPermissions(ctx context.Context, orgID, projectID string) *adminv1.ProjectPermissions
}

// claimsContextKey is used to set and get Claims on a request context.
type claimsContextKey struct{}

// GetClaims retrieves Claims from a request context.
// It should only be used in handlers intercepted by UnaryServerInterceptor or StreamServerInterceptor.
func GetClaims(ctx context.Context) Claims {
	claims, ok := ctx.Value(claimsContextKey{}).(Claims)
	if !ok {
		return nil
	}

	return claims
}

// anonClaims represents claims for an unauthenticated user.
type anonClaims struct{}

// ensure anonClaims implements Claims
var _ Claims = anonClaims{}

func (c anonClaims) OwnerType() OwnerType {
	return OwnerTypeAnon
}

func (c anonClaims) OwnerID() string {
	return ""
}

func (c anonClaims) AuthTokenID() string {
	return ""
}

func (c anonClaims) Superuser(ctx context.Context) bool {
	return false
}

func (c anonClaims) OrganizationPermissions(ctx context.Context, orgID string) *adminv1.OrganizationPermissions {
	return &adminv1.OrganizationPermissions{}
}

func (c anonClaims) ProjectPermissions(ctx context.Context, orgID, projectID string) *adminv1.ProjectPermissions {
	return &adminv1.ProjectPermissions{}
}

// authTokenClaims represents claims for an admin.AuthToken.
type authTokenClaims struct {
	sync.Mutex
	token                   admin.AuthToken
	admin                   *admin.Service
	orgPermissionsCache     map[string]*adminv1.OrganizationPermissions
	projectPermissionsCache map[string]*adminv1.ProjectPermissions
	superuser               *bool
}

func newAuthTokenClaims(token admin.AuthToken, adminService *admin.Service) Claims {
	return &authTokenClaims{
		token:                   token,
		admin:                   adminService,
		orgPermissionsCache:     make(map[string]*adminv1.OrganizationPermissions),
		projectPermissionsCache: make(map[string]*adminv1.ProjectPermissions),
		superuser:               nil,
	}
}

// ensure *authTokenClaims implements Claims
var _ Claims = &authTokenClaims{}

func (c *authTokenClaims) OwnerType() OwnerType {
	t := c.token.Token().Type
	switch t {
	case authtoken.TypeUser:
		return OwnerTypeUser
	default:
		panic(fmt.Errorf("unexpected token type %q", t))
	}
}

func (c *authTokenClaims) OwnerID() string {
	return c.token.OwnerID()
}

func (c *authTokenClaims) AuthTokenID() string {
	return c.token.Token().ID.String()
}

func (c *authTokenClaims) Superuser(ctx context.Context) bool {
	switch c.token.Token().Type {
	case authtoken.TypeUser:
		// continue
	case authtoken.TypeService:
		// services can't be superusers
		return false
	default:
		panic(fmt.Errorf("unexpected token type %q", c.token.Token().Type))
	}

	c.Lock()
	defer c.Unlock()

	if c.superuser != nil {
		return *c.superuser
	}

	user, err := c.admin.DB.FindUser(ctx, c.token.OwnerID())
	if err != nil {
		panic(fmt.Errorf("failed to get user info: %w", err))
	}

	c.superuser = &user.Superuser

	return *c.superuser
}

func (c *authTokenClaims) OrganizationPermissions(ctx context.Context, orgID string) *adminv1.OrganizationPermissions {
	c.Lock()
	defer c.Unlock()

	return c.organizationPermissionsUnsafe(ctx, orgID)
}

func (c *authTokenClaims) ProjectPermissions(ctx context.Context, orgID, projectID string) *adminv1.ProjectPermissions {
	c.Lock()
	defer c.Unlock()

	perm, ok := c.projectPermissionsCache[projectID]
	if ok {
		return perm
	}

	switch c.token.Token().Type {
	case authtoken.TypeUser:
		perm = c.projectPermissionsUser(ctx, orgID, projectID)
	case authtoken.TypeService:
		perm = c.projectPermissionsService(ctx, orgID, projectID)
	default:
		panic(fmt.Errorf("unexpected token type %q", c.token.Token().Type))
	}

	c.projectPermissionsCache[projectID] = perm
	return perm
}

// organizationPermissionsUnsafe resolves organization permissions.
// organizationPermissionsUnsafe accesses the cache without locking, so it should only be called from a function that already has a lock.
func (c *authTokenClaims) organizationPermissionsUnsafe(ctx context.Context, orgID string) *adminv1.OrganizationPermissions {
	perm, ok := c.orgPermissionsCache[orgID]
	if ok {
		return perm
	}

	switch c.token.Token().Type {
	case authtoken.TypeUser:
		perm = c.organizationPermissionsUser(ctx, orgID)
	case authtoken.TypeService:
		perm = c.organizationPermissionsService(ctx, orgID)
	default:
		panic(fmt.Errorf("unexpected token type %q", c.token.Token().Type))
	}

	c.orgPermissionsCache[orgID] = perm
	return perm
}

// organizationPermissionsUser resolves organization permissions for a user.
func (c *authTokenClaims) organizationPermissionsUser(ctx context.Context, orgID string) *adminv1.OrganizationPermissions {
	roles, err := c.admin.DB.ResolveOrganizationRolesForUser(context.Background(), c.token.OwnerID(), orgID)
	if err != nil {
		panic(fmt.Errorf("failed to get org permissions: %w", err))
	}

	composite := &adminv1.OrganizationPermissions{}
	for _, role := range roles {
		composite = unionOrgRoles(composite, role)
	}

	return composite
}

// organizationPermissionsService resolves organization permissions for a service.
// A service currently gets full permissions on the org they belong to.
func (c *authTokenClaims) organizationPermissionsService(ctx context.Context, orgID string) *adminv1.OrganizationPermissions {
	service, err := c.admin.DB.FindService(ctx, c.token.OwnerID())
	if err != nil {
		panic(fmt.Errorf("failed to get service info: %w", err))
	}

	// Services get full permissions on the org they belong to
	if orgID == service.OrgID {
		return &adminv1.OrganizationPermissions{
			ReadOrg:          true,
			ManageOrg:        true,
			ReadProjects:     true,
			CreateProjects:   true,
			ManageProjects:   true,
			ReadOrgMembers:   true,
			ManageOrgMembers: true,
		}
	}

	return &adminv1.OrganizationPermissions{}
}

// projectPermissionsUser resolves project permissions for a user.
func (c *authTokenClaims) projectPermissionsUser(ctx context.Context, orgID, projectID string) *adminv1.ProjectPermissions {
	// ManageProjects permission on the org gives full access to all projects in the org (only org admins have this)
	orgPerms := c.organizationPermissionsUnsafe(ctx, orgID)
	if orgPerms.ManageProjects {
		return &adminv1.ProjectPermissions{
			ReadProject:          true,
			ManageProject:        true,
			ReadProd:             true,
			ReadProdStatus:       true,
			ManageProd:           true,
			ReadDev:              true,
			ReadDevStatus:        true,
			ManageDev:            true,
			ReadProjectMembers:   true,
			ManageProjectMembers: true,
		}
	}

	roles, err := c.admin.DB.ResolveProjectRolesForUser(ctx, c.token.OwnerID(), projectID)
	if err != nil {
		panic(fmt.Errorf("failed to get project permissions: %w", err))
	}

	composite := &adminv1.ProjectPermissions{}
	for _, role := range roles {
		composite = unionProjectRoles(composite, role)
	}

	return composite
}

// projectPermissionsService resolves project permissions for a service.
// A service currently gets full permissions on all projects in the org they belong to.
func (c *authTokenClaims) projectPermissionsService(ctx context.Context, orgID, projectID string) *adminv1.ProjectPermissions {
	orgPerms := c.organizationPermissionsUnsafe(ctx, orgID)
	if orgPerms.ManageProjects {
		return &adminv1.ProjectPermissions{
			ReadProject:          true,
			ManageProject:        true,
			ReadProd:             true,
			ReadProdStatus:       true,
			ManageProd:           true,
			ReadDev:              true,
			ReadDevStatus:        true,
			ManageDev:            true,
			ReadProjectMembers:   true,
			ManageProjectMembers: true,
		}
	}

	return &adminv1.ProjectPermissions{}
}

func unionOrgRoles(a *adminv1.OrganizationPermissions, b *database.OrganizationRole) *adminv1.OrganizationPermissions {
	return &adminv1.OrganizationPermissions{
		ReadOrg:          a.ReadOrg || b.ReadOrg,
		ManageOrg:        a.ManageOrg || b.ManageOrg,
		ReadProjects:     a.ReadProjects || b.ReadProjects,
		CreateProjects:   a.CreateProjects || b.CreateProjects,
		ManageProjects:   a.ManageProjects || b.ManageProjects,
		ReadOrgMembers:   a.ReadOrgMembers || b.ReadOrgMembers,
		ManageOrgMembers: a.ManageOrgMembers || b.ManageOrgMembers,
	}
}

func unionProjectRoles(a *adminv1.ProjectPermissions, b *database.ProjectRole) *adminv1.ProjectPermissions {
	return &adminv1.ProjectPermissions{
		ReadProject:          a.ReadProject || b.ReadProject,
		ManageProject:        a.ManageProject || b.ManageProject,
		ReadProd:             a.ReadProd || b.ReadProd,
		ReadProdStatus:       a.ReadProdStatus || b.ReadProdStatus,
		ManageProd:           a.ManageProd || b.ManageProd,
		ReadDev:              a.ReadDev || b.ReadDev,
		ReadDevStatus:        a.ReadDevStatus || b.ReadDevStatus,
		ManageDev:            a.ManageDev || b.ManageDev,
		ReadProjectMembers:   a.ReadProjectMembers || b.ReadProjectMembers,
		ManageProjectMembers: a.ManageProjectMembers || b.ManageProjectMembers,
	}
}
