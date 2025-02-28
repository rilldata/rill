package server

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/rilldata/rill/admin/database"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/stretchr/testify/require"
)

func TestRBAC(t *testing.T) {
	ctx := context.Background()
	svr := newTestServer(t)

	t.Run("Adding org and project members", func(t *testing.T) {
		// Create users
		u1, c1 := newTestUser(t, svr)
		u2, c2 := newTestUser(t, svr)

		// Create org and project
		r1, err := c1.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		r4, err := c1.CreateProject(ctx, &adminv1.CreateProjectRequest{
			OrganizationName: r1.Organization.Name,
			Name:             "proj1",
			ProdSlots:        1,
			SkipDeploy:       true,
		})
		require.NoError(t, err)

		// Check u2 cannot get or list the org or project
		_, err = c2.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Name: r1.Organization.Name})
		require.Error(t, err)
		_, err = c2.GetProject(ctx, &adminv1.GetProjectRequest{OrganizationName: r1.Organization.Name, Name: r4.Project.Name})
		require.Error(t, err)
		r5, err := c2.ListOrganizations(ctx, &adminv1.ListOrganizationsRequest{})
		require.NoError(t, err)
		require.Len(t, r5.Organizations, 0)

		// Add u2 to the project
		_, err = c1.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
			Organization: r1.Organization.Name,
			Project:      r4.Project.Name,
			Email:        u2.Email,
			Role:         database.ProjectRoleNameAdmin,
		})
		require.NoError(t, err)

		// Check u2 can get and list the org and project
		_, err = c2.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Name: r1.Organization.Name})
		require.NoError(t, err)
		_, err = c2.GetProject(ctx, &adminv1.GetProjectRequest{OrganizationName: r1.Organization.Name, Name: r4.Project.Name})
		require.NoError(t, err)
		r7, err := c2.ListOrganizations(ctx, &adminv1.ListOrganizationsRequest{})
		require.NoError(t, err)
		require.Len(t, r7.Organizations, 1)
		r8, err := c2.ListProjectsForOrganization(ctx, &adminv1.ListProjectsForOrganizationRequest{OrganizationName: r1.Organization.Name})
		require.NoError(t, err)
		require.Len(t, r8.Projects, 1)

		// Check u2 is a guest in the org
		r9, err := c1.ListOrganizationMemberUsers(ctx, &adminv1.ListOrganizationMemberUsersRequest{Organization: r1.Organization.Name})
		require.NoError(t, err)
		require.Len(t, r9.Members, 2)
		hasGuest := false
		for _, m := range r9.Members {
			if m.RoleName == database.OrganizationRoleNameGuest {
				hasGuest = true
			}
		}
		require.True(t, hasGuest)

		// Check u2 is an admin in the project
		r10, err := c1.ListProjectMemberUsers(ctx, &adminv1.ListProjectMemberUsersRequest{
			Organization: r1.Organization.Name,
			Project:      r4.Project.Name,
		})
		require.NoError(t, err)
		require.Len(t, r10.Members, 2)
		for _, m := range r10.Members {
			require.Equal(t, database.ProjectRoleNameAdmin, m.RoleName)
		}

		// Check we can't add u2 to the org (since they are already in it as a guest)
		_, err = c1.AddOrganizationMemberUser(ctx, &adminv1.AddOrganizationMemberUserRequest{
			Organization: r1.Organization.Name,
			Email:        u2.Email,
			Role:         database.OrganizationRoleNameViewer,
		})
		require.Error(t, err)

		// Check we can change u2's role in the org
		_, err = c1.SetOrganizationMemberUserRole(ctx, &adminv1.SetOrganizationMemberUserRoleRequest{
			Organization: r1.Organization.Name,
			Email:        u2.Email,
			Role:         database.OrganizationRoleNameViewer,
		})
		require.NoError(t, err)

		// Remove u2 from the org
		_, err = c1.RemoveOrganizationMemberUser(ctx, &adminv1.RemoveOrganizationMemberUserRequest{
			Organization: r1.Organization.Name,
			Email:        u2.Email,
		})
		require.NoError(t, err)

		// Check they are removed from the list of project members as well
		r11, err := c1.ListProjectMemberUsers(ctx, &adminv1.ListProjectMemberUsersRequest{Organization: r1.Organization.Name, Project: r4.Project.Name})
		require.NoError(t, err)
		require.Len(t, r11.Members, 1)
		require.Equal(t, u1.Email, r11.Members[0].UserEmail)

		// Check they can't get or list the org or project
		_, err = c2.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Name: r1.Organization.Name})
		require.Error(t, err)
		_, err = c2.GetProject(ctx, &adminv1.GetProjectRequest{OrganizationName: r1.Organization.Name, Name: r4.Project.Name})
		require.Error(t, err)
		r12, err := c2.ListOrganizations(ctx, &adminv1.ListOrganizationsRequest{})
		require.NoError(t, err)
		require.Len(t, r12.Organizations, 0)

		// Check the last admin can't leave the org
		_, err = c1.RemoveOrganizationMemberUser(ctx, &adminv1.RemoveOrganizationMemberUserRequest{
			Organization: r1.Organization.Name,
			Email:        u1.Email,
		})
		require.Error(t, err)
		_, err = c1.LeaveOrganization(ctx, &adminv1.LeaveOrganizationRequest{Organization: r1.Organization.Name})
		require.Error(t, err)
	})

	t.Run("Visibility of public projects", func(t *testing.T) {
		// Create users
		_, c1 := newTestUser(t, svr)
		_, c2 := newTestUser(t, svr)

		// Create org and public and private projects
		r1, err := c1.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		r2, err := c1.CreateProject(ctx, &adminv1.CreateProjectRequest{
			OrganizationName: r1.Organization.Name,
			Name:             "public",
			ProdSlots:        1,
			SkipDeploy:       true,
			Public:           true,
		})
		require.NoError(t, err)
		r3, err := c1.CreateProject(ctx, &adminv1.CreateProjectRequest{
			OrganizationName: r1.Organization.Name,
			Name:             "private",
			ProdSlots:        1,
			SkipDeploy:       true,
			Public:           false,
		})
		require.NoError(t, err)

		// Check u2 can access the org and the public project
		_, err = c2.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Name: r1.Organization.Name})
		require.NoError(t, err)
		_, err = c2.GetProject(ctx, &adminv1.GetProjectRequest{OrganizationName: r1.Organization.Name, Name: r2.Project.Name})
		require.NoError(t, err)

		// Check u2 can't access or list the private project
		_, err = c2.GetProject(ctx, &adminv1.GetProjectRequest{OrganizationName: r1.Organization.Name, Name: r3.Project.Name})
		require.Error(t, err)
		r4, err := c2.ListProjectsForOrganization(ctx, &adminv1.ListProjectsForOrganizationRequest{OrganizationName: r1.Organization.Name})
		require.NoError(t, err)
		require.Len(t, r4.Projects, 1)
		require.Equal(t, r2.Project.Name, r4.Projects[0].Name)

		// Delete the public project
		_, err = c1.DeleteProject(ctx, &adminv1.DeleteProjectRequest{
			OrganizationName: r1.Organization.Name,
			Name:             r2.Project.Name,
		})
		require.NoError(t, err)

		// Check u2 can't access the org any more
		_, err = c2.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Name: r1.Organization.Name})
		require.Error(t, err)
	})

	t.Run("Inviting users who have not signed up yet", func(t *testing.T) {
		// Create admin user with org, project and group
		_, c1 := newTestUser(t, svr)
		org1, err := c1.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		proj1, err := c1.CreateProject(ctx, &adminv1.CreateProjectRequest{
			OrganizationName: org1.Organization.Name,
			Name:             "proj1",
			ProdSlots:        1,
			SkipDeploy:       true,
		})
		require.NoError(t, err)
		group1, err := c1.CreateUsergroup(ctx, &adminv1.CreateUsergroupRequest{
			Organization: org1.Organization.Name,
			Name:         "group1",
		})
		require.NoError(t, err)

		// Invite a user that doesn't exist to the org, project and group
		userEmail := randomName() + "@example.com"
		_, err = c1.AddOrganizationMemberUser(ctx, &adminv1.AddOrganizationMemberUserRequest{
			Organization: org1.Organization.Name,
			Email:        userEmail,
			Role:         database.OrganizationRoleNameViewer,
		})
		require.NoError(t, err)
		_, err = c1.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
			Organization: org1.Organization.Name,
			Project:      proj1.Project.Name,
			Email:        userEmail,
			Role:         database.ProjectRoleNameAdmin,
		})
		require.NoError(t, err)
		_, err = c1.AddUsergroupMemberUser(ctx, &adminv1.AddUsergroupMemberUserRequest{
			Organization: org1.Organization.Name,
			Usergroup:    group1.Usergroup.GroupName,
			Email:        userEmail,
		})
		require.NoError(t, err)

		// Check that two emails were sent (org and project addition)
		sender := svr.admin.Email.Sender.(*email.TestSender)
		var count int
		for _, email := range sender.Emails {
			if email.ToEmail == userEmail {
				count++
			}
		}
		require.Equal(t, 2, count)

		// Check that we can list the invites
		orgInvites, err := c1.ListOrganizationInvites(ctx, &adminv1.ListOrganizationInvitesRequest{Organization: org1.Organization.Name})
		require.NoError(t, err)
		require.Len(t, orgInvites.Invites, 1)
		require.Equal(t, userEmail, orgInvites.Invites[0].Email)
		require.Equal(t, database.OrganizationRoleNameViewer, orgInvites.Invites[0].Role)
		projInvites, err := c1.ListProjectInvites(ctx, &adminv1.ListProjectInvitesRequest{
			Organization: org1.Organization.Name,
			Project:      proj1.Project.Name,
		})
		require.NoError(t, err)
		require.Len(t, projInvites.Invites, 1)
		require.Equal(t, userEmail, projInvites.Invites[0].Email)
		require.Equal(t, database.ProjectRoleNameAdmin, projInvites.Invites[0].Role)

		// Create the user and check they can access the org and project, and check they are in the list of members
		_, c2 := newTestUserWithEmail(t, svr, userEmail)
		_, err = c2.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Name: org1.Organization.Name})
		require.NoError(t, err)
		_, err = c2.GetProject(ctx, &adminv1.GetProjectRequest{OrganizationName: org1.Organization.Name, Name: proj1.Project.Name})
		require.NoError(t, err)
		orgMembers, err := c1.ListOrganizationMemberUsers(ctx, &adminv1.ListOrganizationMemberUsersRequest{Organization: org1.Organization.Name})
		require.NoError(t, err)
		require.Len(t, orgMembers.Members, 2)
		projMembers, err := c2.ListProjectMemberUsers(ctx, &adminv1.ListProjectMemberUsersRequest{
			Organization: org1.Organization.Name,
			Project:      proj1.Project.Name,
		})
		require.NoError(t, err)
		require.Len(t, projMembers.Members, 2)
		groupMembers, err := c1.ListUsergroupMemberUsers(ctx, &adminv1.ListUsergroupMemberUsersRequest{
			Organization: org1.Organization.Name,
			Usergroup:    group1.Usergroup.GroupName,
		})
		require.NoError(t, err)
		require.Len(t, groupMembers.Members, 1)

		// Check that the invites were deleted
		orgInvites, err = c1.ListOrganizationInvites(ctx, &adminv1.ListOrganizationInvitesRequest{Organization: org1.Organization.Name})
		require.NoError(t, err)
		require.Len(t, orgInvites.Invites, 0)
		projInvites, err = c1.ListProjectInvites(ctx, &adminv1.ListProjectInvitesRequest{
			Organization: org1.Organization.Name,
			Project:      proj1.Project.Name,
		})
		require.NoError(t, err)
		require.Len(t, projInvites.Invites, 0)

		// Check that the user was added to the managed groups
		allUsers, err := c1.ListUsergroupMemberUsers(ctx, &adminv1.ListUsergroupMemberUsersRequest{
			Organization: org1.Organization.Name,
			Usergroup:    database.ManagedUsergroupNameAllUsers,
		})
		require.NoError(t, err)
		require.Len(t, allUsers.Members, 2)
		allGuests, err := c1.ListUsergroupMemberUsers(ctx, &adminv1.ListUsergroupMemberUsersRequest{
			Organization: org1.Organization.Name,
			Usergroup:    database.ManagedUsergroupNameAllGuests,
		})
		require.NoError(t, err)
		require.Len(t, allGuests.Members, 0)
		allMembers, err := c1.ListUsergroupMemberUsers(ctx, &adminv1.ListUsergroupMemberUsersRequest{
			Organization: org1.Organization.Name,
			Usergroup:    database.ManagedUsergroupNameAllMembers,
		})
		require.NoError(t, err)
		require.Len(t, allMembers.Members, 2)
	})

	t.Run("Inviting project users who haven't signed up yet become org guests", func(t *testing.T) {
		// Create admin user with org and project
		_, c1 := newTestUser(t, svr)
		org1, err := c1.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		proj1, err := c1.CreateProject(ctx, &adminv1.CreateProjectRequest{
			OrganizationName: org1.Organization.Name,
			Name:             "proj1",
			ProdSlots:        1,
			SkipDeploy:       true,
		})
		require.NoError(t, err)

		// Invite a user that doesn't exist to the project
		userEmail := randomName() + "@example.com"
		_, err = c1.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
			Organization: org1.Organization.Name,
			Project:      proj1.Project.Name,
			Email:        userEmail,
			Role:         database.ProjectRoleNameAdmin,
		})
		require.NoError(t, err)

		// Sign up the user
		_, _ = newTestUserWithEmail(t, svr, userEmail)

		// Check that the user is a guest in the org
		orgMembers, err := c1.ListOrganizationMemberUsers(ctx, &adminv1.ListOrganizationMemberUsersRequest{Organization: org1.Organization.Name})
		require.NoError(t, err)
		require.Len(t, orgMembers.Members, 2)
		hasGuest := false
		for _, m := range orgMembers.Members {
			if m.RoleName == database.OrganizationRoleNameGuest {
				hasGuest = true
			}
		}
		require.True(t, hasGuest)

		// Check that the user is in the all-users group
		allUsers, err := c1.ListUsergroupMemberUsers(ctx, &adminv1.ListUsergroupMemberUsersRequest{
			Organization: org1.Organization.Name,
			Usergroup:    database.ManagedUsergroupNameAllUsers,
		})
		require.NoError(t, err)
		require.Len(t, allUsers.Members, 2)
	})

	t.Run("Whitelisting domains on orgs", func(t *testing.T) {
		// Create admin user with four orgs
		u1, c1 := newTestUserWithDomain(t, svr, "whitelist-orgs.test")
		adminEmail := u1.Email
		org1, err := c1.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		org2, err := c1.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		org3, err := c1.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		org4, err := c1.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)

		// Create a user matching a domain BEFORE whitelisting
		userEmail := randomName() + "@whitelist-orgs.test"
		_, _ = newTestUserWithEmail(t, svr, userEmail)

		// Whitelist one domain on org1 and org2, another on org3, and none on org4
		_, err = c1.CreateWhitelistedDomain(ctx, &adminv1.CreateWhitelistedDomainRequest{
			Organization: org1.Organization.Name,
			Domain:       "whitelist-orgs.test",
			Role:         database.OrganizationRoleNameViewer,
		})
		require.NoError(t, err)
		_, err = c1.CreateWhitelistedDomain(ctx, &adminv1.CreateWhitelistedDomainRequest{
			Organization: org2.Organization.Name,
			Domain:       "whitelist-orgs.test",
			Role:         database.OrganizationRoleNameGuest,
		})
		require.NoError(t, err)
		_, err = c1.CreateWhitelistedDomain(ctx, &adminv1.CreateWhitelistedDomainRequest{
			Organization: org3.Organization.Name,
			Domain:       "whitelist-orgs2.test",
			Role:         database.OrganizationRoleNameAdmin,
		})
		require.NoError(t, err)

		// Check we can't whitelist the same domain on the same org again
		_, err = c1.CreateWhitelistedDomain(ctx, &adminv1.CreateWhitelistedDomainRequest{
			Organization: org1.Organization.Name,
			Domain:       "whitelist-orgs.test",
			Role:         database.OrganizationRoleNameAdmin,
		})
		require.Error(t, err)

		// Check that the domains are whitelisted
		org1Domains, err := c1.ListWhitelistedDomains(ctx, &adminv1.ListWhitelistedDomainsRequest{Organization: org1.Organization.Name})
		require.NoError(t, err)
		require.Len(t, org1Domains.Domains, 1)
		require.Equal(t, "whitelist-orgs.test", org1Domains.Domains[0].Domain)
		org4Domains, err := c1.ListWhitelistedDomains(ctx, &adminv1.ListWhitelistedDomainsRequest{Organization: org4.Organization.Name})
		require.NoError(t, err)
		require.Len(t, org4Domains.Domains, 0)

		// Create a user matching a domain AFTER whitelisting
		userEmail2 := randomName() + "@whitelist-orgs.test"
		_, _ = newTestUserWithEmail(t, svr, userEmail2)

		// Utils for checking org and group members
		checkOrgMember := func(email string, orgName string, role string, totalMembers int) {
			// Get the org members
			orgMembers, err := c1.ListOrganizationMemberUsers(ctx, &adminv1.ListOrganizationMemberUsersRequest{Organization: orgName})
			require.NoError(t, err)
			require.Len(t, orgMembers.Members, totalMembers)

			// Check the user is in the org members
			found := false
			for _, m := range orgMembers.Members {
				if m.UserEmail == email && m.RoleName == role {
					found = true
				}
			}
			require.True(t, found)
		}
		checkGroupMember := func(email string, orgName string, groupName string, totalMembers int) {
			// Get the group members
			groupMembers, err := c1.ListUsergroupMemberUsers(ctx, &adminv1.ListUsergroupMemberUsersRequest{
				Organization: orgName,
				Usergroup:    groupName,
			})
			require.NoError(t, err)
			require.Len(t, groupMembers.Members, totalMembers)

			// Check the user is in the group members
			found := false
			for _, m := range groupMembers.Members {
				if m.UserEmail == email {
					found = true
				}
			}
			require.True(t, found)
		}

		// Check that the users are in the orgs that match their domain and groups that match their roles
		checkOrgMember(adminEmail, org1.Organization.Name, database.OrganizationRoleNameAdmin, 3)
		checkOrgMember(adminEmail, org2.Organization.Name, database.OrganizationRoleNameAdmin, 3)
		checkOrgMember(adminEmail, org3.Organization.Name, database.OrganizationRoleNameAdmin, 1)
		checkOrgMember(adminEmail, org4.Organization.Name, database.OrganizationRoleNameAdmin, 1)

		checkOrgMember(userEmail, org1.Organization.Name, database.OrganizationRoleNameViewer, 3)
		checkGroupMember(userEmail, org1.Organization.Name, database.ManagedUsergroupNameAllUsers, 3)
		checkGroupMember(userEmail, org1.Organization.Name, database.ManagedUsergroupNameAllMembers, 3)
		checkOrgMember(userEmail2, org1.Organization.Name, database.OrganizationRoleNameViewer, 3)
		checkGroupMember(userEmail2, org1.Organization.Name, database.ManagedUsergroupNameAllUsers, 3)
		checkGroupMember(userEmail2, org1.Organization.Name, database.ManagedUsergroupNameAllMembers, 3)

		checkOrgMember(userEmail, org2.Organization.Name, database.OrganizationRoleNameGuest, 3)
		checkGroupMember(userEmail, org2.Organization.Name, database.ManagedUsergroupNameAllUsers, 3)
		checkGroupMember(userEmail, org2.Organization.Name, database.ManagedUsergroupNameAllGuests, 2)
		checkOrgMember(userEmail2, org2.Organization.Name, database.OrganizationRoleNameGuest, 3)
		checkGroupMember(userEmail2, org2.Organization.Name, database.ManagedUsergroupNameAllUsers, 3)
		checkGroupMember(userEmail2, org2.Organization.Name, database.ManagedUsergroupNameAllGuests, 2)
	})

	t.Run("Whitelisting domains on projects", func(t *testing.T) {
		// Create an admin user and two orgs with a project each
		u1, c1 := newTestUserWithDomain(t, svr, "whitelist-projs.test")
		adminEmail := u1.Email
		org1, err := c1.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		proj1, err := c1.CreateProject(ctx, &adminv1.CreateProjectRequest{
			OrganizationName: org1.Organization.Name,
			Name:             "proj1",
			ProdSlots:        1,
			SkipDeploy:       true,
		})
		require.NoError(t, err)
		org2, err := c1.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		proj2, err := c1.CreateProject(ctx, &adminv1.CreateProjectRequest{
			OrganizationName: org2.Organization.Name,
			Name:             "proj2",
			ProdSlots:        1,
			SkipDeploy:       true,
		})
		require.NoError(t, err)

		// Create two users before adding the whitelist
		userEmail1 := randomName() + "@whitelist-projs.test"
		_, _ = newTestUserWithEmail(t, svr, userEmail1)
		userEmail2 := randomName() + "@whitelist-projs.test"
		_, _ = newTestUserWithEmail(t, svr, userEmail2)

		// Add one of the users to the org
		_, err = c1.AddOrganizationMemberUser(ctx, &adminv1.AddOrganizationMemberUserRequest{
			Organization: org1.Organization.Name,
			Email:        userEmail1,
			Role:         database.OrganizationRoleNameViewer,
		})
		require.NoError(t, err)

		// Add the whitelist to the project
		_, err = c1.CreateProjectWhitelistedDomain(ctx, &adminv1.CreateProjectWhitelistedDomainRequest{
			Organization: org1.Organization.Name,
			Project:      proj1.Project.Name,
			Domain:       "whitelist-projs.test",
			Role:         database.ProjectRoleNameAdmin,
		})

		// Check we can't whitelist the same domain on the same project again
		_, err = c1.CreateProjectWhitelistedDomain(ctx, &adminv1.CreateProjectWhitelistedDomainRequest{
			Organization: org1.Organization.Name,
			Project:      proj1.Project.Name,
			Domain:       "whitelist-projs.test",
			Role:         database.ProjectRoleNameAdmin,
		})
		require.Error(t, err)

		// Check that the domain is whitelisted
		proj1Domains, err := c1.ListProjectWhitelistedDomains(ctx, &adminv1.ListProjectWhitelistedDomainsRequest{
			Organization: org1.Organization.Name,
			Project:      proj1.Project.Name,
		})
		require.NoError(t, err)
		require.Len(t, proj1Domains.Domains, 1)
		require.Equal(t, "whitelist-projs.test", proj1Domains.Domains[0].Domain)

		// Invite a non-existing user to the org and project
		userEmail3 := randomName() + "@whitelist-projs.test"
		_, err = c1.AddOrganizationMemberUser(ctx, &adminv1.AddOrganizationMemberUserRequest{
			Organization: org1.Organization.Name,
			Email:        userEmail3,
			Role:         database.OrganizationRoleNameViewer,
		})
		require.NoError(t, err)
		_, err = c1.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
			Organization: org1.Organization.Name,
			Project:      proj1.Project.Name,
			Email:        userEmail3,
			Role:         database.OrganizationRoleNameViewer,
		})
		require.NoError(t, err)

		// Create two users matching the domain, one of whom matches the org-level invite
		_, _ = newTestUserWithEmail(t, svr, userEmail3)
		userEmail4 := randomName() + "@whitelist-projs.test"
		_, _ = newTestUserWithEmail(t, svr, userEmail4)

		// Utils for checking org, group and project members
		checkOrgMember := func(email string, orgName string, role string, totalMembers int) {
			// Get the org members
			orgMembers, err := c1.ListOrganizationMemberUsers(ctx, &adminv1.ListOrganizationMemberUsersRequest{Organization: orgName})
			require.NoError(t, err)
			require.Len(t, orgMembers.Members, totalMembers)

			// Check the user is in the org members
			found := false
			for _, m := range orgMembers.Members {
				if m.UserEmail == email && m.RoleName == role {
					found = true
				}
			}
			require.True(t, found)
		}
		checkGroupMember := func(email string, orgName string, groupName string, totalMembers int) {
			// Get the group members
			groupMembers, err := c1.ListUsergroupMemberUsers(ctx, &adminv1.ListUsergroupMemberUsersRequest{
				Organization: orgName,
				Usergroup:    groupName,
			})
			require.NoError(t, err)
			require.Len(t, groupMembers.Members, totalMembers)

			// Check the user is in the group members
			found := false
			for _, m := range groupMembers.Members {
				if m.UserEmail == email {
					found = true
				}
			}
			require.True(t, found)
		}
		checkProjMember := func(email string, orgName string, projName string, role string, totalMembers int) {
			// Get the project members
			projMembers, err := c1.ListProjectMemberUsers(ctx, &adminv1.ListProjectMemberUsersRequest{
				Organization: orgName,
				Project:      projName,
			})
			require.NoError(t, err)
			require.Len(t, projMembers.Members, totalMembers)

			// Check the user is in the project members
			found := false
			for _, m := range projMembers.Members {
				if m.UserEmail == email && m.RoleName == role {
					found = true
				}
			}
			require.True(t, found)
		}

		// Check that org-level and group memberships match expectations
		checkOrgMember(adminEmail, org1.Organization.Name, database.OrganizationRoleNameAdmin, 5)
		checkOrgMember(userEmail1, org1.Organization.Name, database.OrganizationRoleNameViewer, 5)
		checkOrgMember(userEmail2, org1.Organization.Name, database.OrganizationRoleNameGuest, 5)
		checkOrgMember(userEmail3, org1.Organization.Name, database.OrganizationRoleNameViewer, 5)
		checkOrgMember(userEmail4, org1.Organization.Name, database.OrganizationRoleNameGuest, 5)
		checkGroupMember(adminEmail, org1.Organization.Name, database.ManagedUsergroupNameAllUsers, 5)
		checkGroupMember(adminEmail, org1.Organization.Name, database.ManagedUsergroupNameAllMembers, 3)
		checkGroupMember(userEmail1, org1.Organization.Name, database.ManagedUsergroupNameAllMembers, 3)
		checkGroupMember(userEmail2, org1.Organization.Name, database.ManagedUsergroupNameAllGuests, 2)
		checkGroupMember(userEmail3, org1.Organization.Name, database.ManagedUsergroupNameAllMembers, 3)
		checkGroupMember(userEmail4, org1.Organization.Name, database.ManagedUsergroupNameAllGuests, 2)

		// Check that project-level memberships match expectations
		checkProjMember(adminEmail, org1.Organization.Name, proj1.Project.Name, database.ProjectRoleNameAdmin, 5)
		checkProjMember(adminEmail, org2.Organization.Name, proj2.Project.Name, database.ProjectRoleNameAdmin, 1)
		checkProjMember(userEmail1, org1.Organization.Name, proj1.Project.Name, database.ProjectRoleNameAdmin, 5)
		checkProjMember(userEmail2, org1.Organization.Name, proj1.Project.Name, database.ProjectRoleNameAdmin, 5)
		checkProjMember(userEmail3, org1.Organization.Name, proj1.Project.Name, database.ProjectRoleNameViewer, 5) // Because explicit invite role takes precedence over domain whitelist role
		checkProjMember(userEmail4, org1.Organization.Name, proj1.Project.Name, database.ProjectRoleNameAdmin, 5)
	})

	t.Run("Managed usergroup memberships", func(t *testing.T) {

	})
}

func randomName() string {
	id := randomBytes(16)
	return "test_" + hex.EncodeToString(id)
}
