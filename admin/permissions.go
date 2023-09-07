package admin

import (
	"context"

	"github.com/rilldata/rill/admin/database"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
)

// OrganizationPermissionsForUser resolves organization permissions for a user.
func (s *Service) OrganizationPermissionsForUser(ctx context.Context, orgID, userID string) (*adminv1.OrganizationPermissions, error) {
	roles, err := s.DB.ResolveOrganizationRolesForUser(context.Background(), userID, orgID)
	if err != nil {
		return nil, err
	}

	composite := &adminv1.OrganizationPermissions{}
	for _, role := range roles {
		composite = unionOrgRoles(composite, role)
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

// ProjectPermissionsForUser resolves project permissions for a user.
func (s *Service) ProjectPermissionsForUser(ctx context.Context, projectID, userID string, orgPerms *adminv1.OrganizationPermissions) (*adminv1.ProjectPermissions, error) {
	// ManageProjects permission on the org gives full access to all projects in the org (only org admins have this)
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
		}, nil
	}

	roles, err := s.DB.ResolveProjectRolesForUser(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}

	composite := &adminv1.ProjectPermissions{}
	for _, role := range roles {
		composite = unionProjectRoles(composite, role)
	}

	return composite, nil
}

// ProjectPermissionsService resolves project permissions for a service.
// A service currently gets full permissions on all projects in the org they belong to.
func (s *Service) ProjectPermissionsForService(ctx context.Context, projectID, serviceID string, orgPerms *adminv1.OrganizationPermissions) (*adminv1.ProjectPermissions, error) {
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
		}, nil
	}

	return &adminv1.ProjectPermissions{}, nil
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
