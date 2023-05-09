package auth

import (
	"context"
	"errors"
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
	OrganizationPermissions(ctx context.Context, orgID string) *adminv1.OrganizationPermissions
	ProjectPermissions(ctx context.Context, orgID, projectID string) *adminv1.ProjectPermissions
	Superuser(ctx context.Context) bool
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

func (c anonClaims) OwnerType() OwnerType {
	return OwnerTypeAnon
}

func (c anonClaims) OwnerID() string {
	return ""
}

func (c anonClaims) AuthTokenID() string {
	return ""
}

func (c anonClaims) OrganizationPermissions(ctx context.Context, orgID string) *adminv1.OrganizationPermissions {
	return &adminv1.OrganizationPermissions{}
}

func (c anonClaims) ProjectPermissions(ctx context.Context, orgID, projectID string) *adminv1.ProjectPermissions {
	return &adminv1.ProjectPermissions{}
}

func (c anonClaims) Superuser(ctx context.Context) bool {
	return false
}

// ensure anonClaims implements Claims
var _ Claims = anonClaims{}

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

func (c *authTokenClaims) OrganizationPermissions(ctx context.Context, orgID string) *adminv1.OrganizationPermissions {
	switch c.token.Token().Type {
	case authtoken.TypeUser:
		// continue
	case authtoken.TypeService:
		panic(errors.New("service tokens not supported"))
	default:
		panic(fmt.Errorf("unexpected token type %q", c.token.Token().Type))
	}

	c.Lock()
	defer c.Unlock()

	if perm, ok := c.orgPermissionsCache[orgID]; ok {
		return perm
	}

	roles, err := c.admin.DB.ResolveOrganizationRolesForUser(ctx, c.token.OwnerID(), orgID)
	if err != nil {
		panic(fmt.Errorf("failed to get org permissions: %w", err))
	}

	composite := &adminv1.OrganizationPermissions{}
	for _, role := range roles {
		composite = unionOrgRoles(composite, role)
	}

	c.orgPermissionsCache[orgID] = composite
	return composite
}

func (c *authTokenClaims) ProjectPermissions(ctx context.Context, orgID, projectID string) *adminv1.ProjectPermissions {
	switch c.token.Token().Type {
	case authtoken.TypeUser:
		// continue
	case authtoken.TypeService:
		panic(errors.New("service tokens not supported"))
	default:
		panic(fmt.Errorf("unexpected token type %q", c.token.Token().Type))
	}

	// ManageProjects permission on the org gives full access to all projects in the org (only org admins have this)
	orgPerms := c.OrganizationPermissions(ctx, orgID)
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

	c.Lock()
	defer c.Unlock()

	if perm, ok := c.projectPermissionsCache[projectID]; ok {
		return perm
	}

	roles, err := c.admin.DB.ResolveProjectRolesForUser(ctx, c.token.OwnerID(), projectID)
	if err != nil {
		panic(fmt.Errorf("failed to get project permissions: %w", err))
	}

	composite := &adminv1.ProjectPermissions{}
	for _, role := range roles {
		composite = unionProjectRoles(composite, role)
	}

	c.projectPermissionsCache[projectID] = composite
	return composite
}

func (c *authTokenClaims) Superuser(ctx context.Context) bool {
	switch c.token.Token().Type {
	case authtoken.TypeUser:
		// continue
	case authtoken.TypeService:
		panic(errors.New("service account cannot be superuser"))
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

// ensure *authTokenClaims implements Claims
var _ Claims = &authTokenClaims{}

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
