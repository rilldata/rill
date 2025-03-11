package admin

import (
	"context"

	"github.com/rilldata/rill/admin/database"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
)

// OrganizationPermissionsForUser resolves organization permissions for a user.
func (s *Service) OrganizationPermissionsForUser(ctx context.Context, orgID, userID string) (*adminv1.OrganizationPermissions, error) {
	roles, err := s.DB.ResolveOrganizationRolesForUser(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}

	composite := &adminv1.OrganizationPermissions{}
	for _, role := range roles {
		composite = UnionOrgRoles(composite, role)
	}

	// If the org has a public project, all users get read access to it.
	if !composite.ReadOrg {
		ok, err := s.DB.CheckOrganizationHasPublicProjects(ctx, orgID)
		if err != nil {
			return nil, err
		}
		if ok {
			composite.Guest = true
			composite.ReadOrg = true
			composite.ReadProjects = true
		}
	}

	return composite, nil
}

// OrganizationPermissionsForService resolves organization permissions for a service.
// A service currently gets full permissions on the org they belong to.
func (s *Service) OrganizationPermissionsForService(ctx context.Context, orgID, serviceID string) (*adminv1.OrganizationPermissions, error) {
	service, err := s.DB.FindService(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	// Services get full permissions on the org they belong to
	if orgID == service.OrgID {
		return &adminv1.OrganizationPermissions{
			Admin:            true,
			Guest:            false,
			ReadOrg:          true,
			ManageOrg:        true,
			ReadProjects:     true,
			CreateProjects:   true,
			ManageProjects:   true,
			ReadOrgMembers:   true,
			ManageOrgMembers: true,
		}, nil
	}

	return &adminv1.OrganizationPermissions{}, nil
}

// OrganizationPermissionsForDeployment resolves organization permissions for a deployment.
// A deployment does not get any permissions on the org it belongs to. It only has permissions on the project it belongs to.
func (s *Service) OrganizationPermissionsForDeployment(ctx context.Context, orgID, deploymentID string) (*adminv1.OrganizationPermissions, error) {
	return &adminv1.OrganizationPermissions{}, nil
}

// OrganizationPermissionsForMagicAuthToken resolves organization permissions for a magic auth token in the specified project.
// It grants basic read access to only the org of the project the token belongs to.
func (s *Service) OrganizationPermissionsForMagicAuthToken(ctx context.Context, orgID, tokenProjectID string) (*adminv1.OrganizationPermissions, error) {
	proj, err := s.DB.FindProject(ctx, tokenProjectID)
	if err != nil {
		return nil, err
	}

	if orgID == proj.OrganizationID {
		return &adminv1.OrganizationPermissions{
			Admin:            false,
			Guest:            true,
			ReadOrg:          true,
			ManageOrg:        false,
			ReadProjects:     false,
			CreateProjects:   false,
			ManageProjects:   false,
			ReadOrgMembers:   false,
			ManageOrgMembers: false,
		}, nil
	}

	return &adminv1.OrganizationPermissions{}, nil
}

// ProjectPermissionsForUser resolves project permissions for a user.
func (s *Service) ProjectPermissionsForUser(ctx context.Context, projectID, userID string, orgPerms *adminv1.OrganizationPermissions) (*adminv1.ProjectPermissions, error) {
	// ManageProjects permission on the org gives full access to all projects in the org (only org admins have this)
	if orgPerms.ManageProjects {
		return &adminv1.ProjectPermissions{
			Admin:                      true,
			ReadProject:                true,
			ManageProject:              true,
			ReadProd:                   true,
			ReadProdStatus:             true,
			ManageProd:                 true,
			ReadDev:                    true,
			ReadDevStatus:              true,
			ManageDev:                  true,
			ReadProvisionerResources:   true,
			ManageProvisionerResources: true,
			ReadProjectMembers:         true,
			ManageProjectMembers:       true,
			CreateMagicAuthTokens:      true,
			ManageMagicAuthTokens:      true,
			CreateReports:              true,
			ManageReports:              true,
			CreateAlerts:               true,
			ManageAlerts:               true,
			CreateBookmarks:            true,
			ManageBookmarks:            true,
		}, nil
	}

	roles, err := s.DB.ResolveProjectRolesForUser(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}

	composite := &adminv1.ProjectPermissions{}
	for _, role := range roles {
		composite = UnionProjectRoles(composite, role)
	}

	return composite, nil
}

// ProjectPermissionsService resolves project permissions for a service.
// A service currently gets full permissions on all projects in the org they belong to.
func (s *Service) ProjectPermissionsForService(ctx context.Context, projectID, serviceID string, orgPerms *adminv1.OrganizationPermissions) (*adminv1.ProjectPermissions, error) {
	if orgPerms.ManageProjects {
		return &adminv1.ProjectPermissions{
			Admin:                      true,
			ReadProject:                true,
			ManageProject:              true,
			ReadProd:                   true,
			ReadProdStatus:             true,
			ManageProd:                 true,
			ReadDev:                    true,
			ReadDevStatus:              true,
			ManageDev:                  true,
			ReadProvisionerResources:   true,
			ManageProvisionerResources: true,
			ReadProjectMembers:         true,
			ManageProjectMembers:       true,
			CreateMagicAuthTokens:      true,
			ManageMagicAuthTokens:      true,
			CreateReports:              true,
			ManageReports:              true,
			CreateAlerts:               true,
			ManageAlerts:               true,
			CreateBookmarks:            true,
			ManageBookmarks:            true,
		}, nil
	}

	return &adminv1.ProjectPermissions{}, nil
}

// ProjectPermissionsForDeployment resolves project permissions for a deployment.
// A deployment currently gets full read and no write permissions on the project it belongs to.
func (s *Service) ProjectPermissionsForDeployment(ctx context.Context, projectID, deploymentID string, orgPerms *adminv1.OrganizationPermissions) (*adminv1.ProjectPermissions, error) {
	depl, err := s.DB.FindDeployment(ctx, deploymentID)
	if err != nil {
		return nil, err
	}

	// Deployments get full read and no write permissions on the project they belong to
	if projectID == depl.ProjectID {
		return &adminv1.ProjectPermissions{
			Admin:                      false,
			ReadProject:                true,
			ManageProject:              false,
			ReadProd:                   true,
			ReadProdStatus:             true,
			ManageProd:                 false,
			ReadDev:                    true,
			ReadDevStatus:              true,
			ManageDev:                  false,
			ReadProvisionerResources:   true,
			ManageProvisionerResources: true,
			ReadProjectMembers:         true,
			ManageProjectMembers:       false,
			CreateMagicAuthTokens:      false,
			ManageMagicAuthTokens:      false,
			CreateReports:              false,
			ManageReports:              false,
			CreateAlerts:               false,
			ManageAlerts:               false,
			CreateBookmarks:            false,
			ManageBookmarks:            false,
		}, nil
	}

	return &adminv1.ProjectPermissions{}, nil
}

// ProjectPermissionsForMagicAuthToken resolves project permissions for a magic auth token.
func (s *Service) ProjectPermissionsForMagicAuthToken(ctx context.Context, projectID string, tkn *database.MagicAuthToken) (*adminv1.ProjectPermissions, error) {
	// No access if the token belongs to another project
	if projectID != tkn.ProjectID {
		return &adminv1.ProjectPermissions{}, nil
	}

	// Grant basic read access to the project and its prod deployment
	return &adminv1.ProjectPermissions{
		Admin:                      false,
		ReadProject:                true,
		ManageProject:              false,
		ReadProd:                   true,
		ReadProdStatus:             false,
		ManageProd:                 false,
		ReadDev:                    false,
		ReadDevStatus:              false,
		ManageDev:                  false,
		ReadProvisionerResources:   false,
		ManageProvisionerResources: false,
		ReadProjectMembers:         false,
		ManageProjectMembers:       false,
		CreateMagicAuthTokens:      false,
		ManageMagicAuthTokens:      false,
		CreateReports:              false,
		ManageReports:              false,
		CreateAlerts:               false,
		ManageAlerts:               false,
		CreateBookmarks:            false,
		ManageBookmarks:            false,
	}, nil
}

// UnionOrgRoles merges an organization role's permissions into the given permissions object.
func UnionOrgRoles(a *adminv1.OrganizationPermissions, b *database.OrganizationRole) *adminv1.OrganizationPermissions {
	return &adminv1.OrganizationPermissions{
		Admin:            a.Admin || b.Admin,
		Guest:            a.Guest || b.Guest,
		ReadOrg:          a.ReadOrg || b.ReadOrg,
		ManageOrg:        a.ManageOrg || b.ManageOrg,
		ReadProjects:     a.ReadProjects || b.ReadProjects,
		CreateProjects:   a.CreateProjects || b.CreateProjects,
		ManageProjects:   a.ManageProjects || b.ManageProjects,
		ReadOrgMembers:   a.ReadOrgMembers || b.ReadOrgMembers,
		ManageOrgMembers: a.ManageOrgMembers || b.ManageOrgMembers,
	}
}

// UnionProjectRoles merges a project role's permissions into the given permissions object.
func UnionProjectRoles(a *adminv1.ProjectPermissions, b *database.ProjectRole) *adminv1.ProjectPermissions {
	return &adminv1.ProjectPermissions{
		Admin:                      a.Admin || b.Admin,
		ReadProject:                a.ReadProject || b.ReadProject,
		ManageProject:              a.ManageProject || b.ManageProject,
		ReadProd:                   a.ReadProd || b.ReadProd,
		ReadProdStatus:             a.ReadProdStatus || b.ReadProdStatus,
		ManageProd:                 a.ManageProd || b.ManageProd,
		ReadDev:                    a.ReadDev || b.ReadDev,
		ReadDevStatus:              a.ReadDevStatus || b.ReadDevStatus,
		ManageDev:                  a.ManageDev || b.ManageDev,
		ReadProvisionerResources:   a.ReadProvisionerResources || b.ReadProvisionerResources,
		ManageProvisionerResources: a.ManageProvisionerResources || b.ManageProvisionerResources,
		ReadProjectMembers:         a.ReadProjectMembers || b.ReadProjectMembers,
		ManageProjectMembers:       a.ManageProjectMembers || b.ManageProjectMembers,
		CreateMagicAuthTokens:      a.CreateMagicAuthTokens || b.CreateMagicAuthTokens,
		ManageMagicAuthTokens:      a.ManageMagicAuthTokens || b.ManageMagicAuthTokens,
		CreateReports:              a.CreateReports || b.CreateReports,
		ManageReports:              a.ManageReports || b.ManageReports,
		CreateAlerts:               a.CreateAlerts || b.CreateAlerts,
		ManageAlerts:               a.ManageAlerts || b.ManageAlerts,
		CreateBookmarks:            a.CreateBookmarks || b.CreateBookmarks,
		ManageBookmarks:            a.ManageBookmarks || b.ManageBookmarks,
	}
}
