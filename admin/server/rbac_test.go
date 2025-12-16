package server_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/testadmin"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestRBAC(t *testing.T) {
	ctx := context.Background()
	fix := testadmin.New(t)

	t.Run("Adding org and project members", func(t *testing.T) {
		// Create users
		u1, c1 := fix.NewUser(t)
		u2, c2 := fix.NewUser(t)

		// Create org and project
		r1, err := c1.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		r4, err := c1.CreateProject(ctx, &adminv1.CreateProjectRequest{
			Org:        r1.Organization.Name,
			Project:    "proj1",
			ProdSlots:  1,
			SkipDeploy: true,
		})
		require.NoError(t, err)

		// Check u2 cannot get or list the org or project
		_, err = c2.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Org: r1.Organization.Name})
		require.Error(t, err)
		_, err = c2.GetProject(ctx, &adminv1.GetProjectRequest{Org: r1.Organization.Name, Project: r4.Project.Name})
		require.Error(t, err)
		r5, err := c2.ListOrganizations(ctx, &adminv1.ListOrganizationsRequest{})
		require.NoError(t, err)
		require.Len(t, r5.Organizations, 0)

		// Add u2 to the project
		_, err = c1.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
			Org:     r1.Organization.Name,
			Project: r4.Project.Name,
			Email:   u2.Email,
			Role:    database.ProjectRoleNameAdmin,
		})
		require.NoError(t, err)

		// Check u2 can get and list the org and project
		_, err = c2.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Org: r1.Organization.Name})
		require.NoError(t, err)
		_, err = c2.GetProject(ctx, &adminv1.GetProjectRequest{Org: r1.Organization.Name, Project: r4.Project.Name})
		require.NoError(t, err)
		r7, err := c2.ListOrganizations(ctx, &adminv1.ListOrganizationsRequest{})
		require.NoError(t, err)
		require.Len(t, r7.Organizations, 1)
		r8, err := c2.ListProjectsForOrganization(ctx, &adminv1.ListProjectsForOrganizationRequest{Org: r1.Organization.Name})
		require.NoError(t, err)
		require.Len(t, r8.Projects, 1)

		// Check u2 is a guest in the org
		r9, err := c1.ListOrganizationMemberUsers(ctx, &adminv1.ListOrganizationMemberUsersRequest{Org: r1.Organization.Name})
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
			Org:     r1.Organization.Name,
			Project: r4.Project.Name,
		})
		require.NoError(t, err)
		require.Len(t, r10.Members, 2)
		hasGuest = false
		for _, m := range r10.Members {
			require.Equal(t, database.ProjectRoleNameAdmin, m.RoleName)
			if m.OrgRoleName == database.OrganizationRoleNameGuest {
				hasGuest = true
			}
		}
		require.True(t, hasGuest)

		// Check we can't add u2 to the org (since they are already in it as a guest)
		_, err = c1.AddOrganizationMemberUser(ctx, &adminv1.AddOrganizationMemberUserRequest{
			Org:   r1.Organization.Name,
			Email: u2.Email,
			Role:  database.OrganizationRoleNameViewer,
		})
		require.Error(t, err)

		// Check we can change u2's role in the org
		_, err = c1.SetOrganizationMemberUserRole(ctx, &adminv1.SetOrganizationMemberUserRoleRequest{
			Org:   r1.Organization.Name,
			Email: u2.Email,
			Role:  database.OrganizationRoleNameViewer,
		})
		require.NoError(t, err)

		// Remove u2 from the org
		_, err = c1.RemoveOrganizationMemberUser(ctx, &adminv1.RemoveOrganizationMemberUserRequest{
			Org:   r1.Organization.Name,
			Email: u2.Email,
		})
		require.NoError(t, err)

		// Check they are removed from the list of project members as well
		r11, err := c1.ListProjectMemberUsers(ctx, &adminv1.ListProjectMemberUsersRequest{Org: r1.Organization.Name, Project: r4.Project.Name})
		require.NoError(t, err)
		require.Len(t, r11.Members, 1)
		require.Equal(t, u1.Email, r11.Members[0].UserEmail)

		// Check they can't get or list the org or project
		_, err = c2.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Org: r1.Organization.Name})
		require.Error(t, err)
		_, err = c2.GetProject(ctx, &adminv1.GetProjectRequest{Org: r1.Organization.Name, Project: r4.Project.Name})
		require.Error(t, err)
		r12, err := c2.ListOrganizations(ctx, &adminv1.ListOrganizationsRequest{})
		require.NoError(t, err)
		require.Len(t, r12.Organizations, 0)

		// Check the last admin can't leave the org
		_, err = c1.RemoveOrganizationMemberUser(ctx, &adminv1.RemoveOrganizationMemberUserRequest{
			Org:   r1.Organization.Name,
			Email: u1.Email,
		})
		require.Error(t, err)
		_, err = c1.LeaveOrganization(ctx, &adminv1.LeaveOrganizationRequest{Org: r1.Organization.Name})
		require.Error(t, err)
	})

	t.Run("Ability to filter by role in member listings", func(t *testing.T) {
		// Create org, project and usergroup
		_, c1 := fix.NewUser(t)
		r1, err := c1.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		r2, err := c1.CreateProject(ctx, &adminv1.CreateProjectRequest{
			Org:        r1.Organization.Name,
			Project:    "proj1",
			ProdSlots:  1,
			SkipDeploy: true,
		})
		require.NoError(t, err)
		r3, err := c1.CreateUsergroup(ctx, &adminv1.CreateUsergroupRequest{
			Org:  r1.Organization.Name,
			Name: "group1",
		})
		require.NoError(t, err)

		// Add a user as viewer to the org and project, and to the usergroup
		u2, _ := fix.NewUser(t)
		_, err = c1.AddOrganizationMemberUser(ctx, &adminv1.AddOrganizationMemberUserRequest{
			Org:   r1.Organization.Name,
			Email: u2.Email,
			Role:  database.OrganizationRoleNameViewer,
		})
		require.NoError(t, err)
		_, err = c1.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
			Org:     r1.Organization.Name,
			Project: r2.Project.Name,
			Email:   u2.Email,
			Role:    database.ProjectRoleNameViewer,
		})
		require.NoError(t, err)
		_, err = c1.AddUsergroupMemberUser(ctx, &adminv1.AddUsergroupMemberUserRequest{
			Org:       r1.Organization.Name,
			Usergroup: r3.Usergroup.GroupName,
			Email:     u2.Email,
		})
		require.NoError(t, err)

		// Add the usergroup as viewer on the org and project
		_, err = c1.AddOrganizationMemberUsergroup(ctx, &adminv1.AddOrganizationMemberUsergroupRequest{
			Org:       r1.Organization.Name,
			Usergroup: r3.Usergroup.GroupName,
			Role:      database.OrganizationRoleNameViewer,
		})
		require.NoError(t, err)
		_, err = c1.AddProjectMemberUsergroup(ctx, &adminv1.AddProjectMemberUsergroupRequest{
			Org:       r1.Organization.Name,
			Project:   r2.Project.Name,
			Usergroup: r3.Usergroup.GroupName,
			Role:      database.ProjectRoleNameViewer,
		})
		require.NoError(t, err)

		// Check listing counts for various role filters
		cases := []struct {
			roleName           string
			orgUserCount       int
			projUserCount      int
			orgUsergroupCount  int
			projUsergroupCount int
		}{
			{"", 2, 2, 4, 2},
			{database.OrganizationRoleNameAdmin, 1, 1, 0, 0},
			{database.OrganizationRoleNameEditor, 0, 0, 0, 0},
			{database.OrganizationRoleNameViewer, 1, 1, 1, 2},
		}
		for _, c := range cases {
			t.Run(c.roleName, func(t *testing.T) {
				r4, err := c1.ListOrganizationMemberUsers(ctx, &adminv1.ListOrganizationMemberUsersRequest{
					Org:  r1.Organization.Name,
					Role: c.roleName,
				})
				require.NoError(t, err)
				require.Len(t, r4.Members, c.orgUserCount)
				r5, err := c1.ListProjectMemberUsers(ctx, &adminv1.ListProjectMemberUsersRequest{
					Org:     r1.Organization.Name,
					Project: r2.Project.Name,
					Role:    c.roleName,
				})
				require.NoError(t, err)
				require.Len(t, r5.Members, c.projUserCount)
				r6, err := c1.ListOrganizationMemberUsergroups(ctx, &adminv1.ListOrganizationMemberUsergroupsRequest{
					Org:  r1.Organization.Name,
					Role: c.roleName,
				})
				require.NoError(t, err)
				require.Len(t, r6.Members, c.orgUsergroupCount)
				r7, err := c1.ListProjectMemberUsergroups(ctx, &adminv1.ListProjectMemberUsergroupsRequest{
					Org:     r1.Organization.Name,
					Project: r2.Project.Name,
					Role:    c.roleName,
				})
				require.NoError(t, err)
				require.Len(t, r7.Members, c.projUsergroupCount)
			})
		}
	})

	t.Run("Ability to include project and usergroup counts in org member listings", func(t *testing.T) {
		// Create org, two projects and two usergroup
		u1, c1 := fix.NewUser(t)
		r1, err := c1.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		r2, err := c1.CreateProject(ctx, &adminv1.CreateProjectRequest{
			Org:        r1.Organization.Name,
			Project:    "proj1",
			ProdSlots:  1,
			SkipDeploy: true,
		})
		require.NoError(t, err)
		_, err = c1.CreateProject(ctx, &adminv1.CreateProjectRequest{
			Org:        r1.Organization.Name,
			Project:    "proj2",
			ProdSlots:  1,
			SkipDeploy: true,
		})
		require.NoError(t, err)
		r4, err := c1.CreateUsergroup(ctx, &adminv1.CreateUsergroupRequest{
			Org:  r1.Organization.Name,
			Name: "group1",
		})
		require.NoError(t, err)
		_, err = c1.CreateUsergroup(ctx, &adminv1.CreateUsergroupRequest{
			Org:  r1.Organization.Name,
			Name: "group2",
		})
		require.NoError(t, err)

		// Add a user to the org, one of the usergroups, and one of the projects
		u2, _ := fix.NewUser(t)
		_, err = c1.AddOrganizationMemberUser(ctx, &adminv1.AddOrganizationMemberUserRequest{
			Org:   r1.Organization.Name,
			Email: u2.Email,
			Role:  database.OrganizationRoleNameViewer,
		})
		require.NoError(t, err)
		_, err = c1.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
			Org:     r1.Organization.Name,
			Project: r2.Project.Name,
			Email:   u2.Email,
			Role:    database.ProjectRoleNameViewer,
		})
		require.NoError(t, err)
		_, err = c1.AddUsergroupMemberUser(ctx, &adminv1.AddUsergroupMemberUserRequest{
			Org:       r1.Organization.Name,
			Usergroup: r4.Usergroup.GroupName,
			Email:     u2.Email,
		})
		require.NoError(t, err)

		// Check the counts for the user
		r6, err := c1.ListOrganizationMemberUsers(ctx, &adminv1.ListOrganizationMemberUsersRequest{
			Org:           r1.Organization.Name,
			IncludeCounts: true,
		})
		require.NoError(t, err)
		require.Len(t, r6.Members, 2)
		for _, m := range r6.Members {
			m.CreatedOn = nil
			m.UpdatedOn = nil
		}
		require.Contains(t, r6.Members, &adminv1.OrganizationMemberUser{
			UserId:          u1.ID,
			UserEmail:       u1.Email,
			UserName:        u1.DisplayName,
			RoleName:        database.OrganizationRoleNameAdmin,
			ProjectsCount:   2,
			UsergroupsCount: 0, // Does not include the autogroups
		})
		require.Contains(t, r6.Members, &adminv1.OrganizationMemberUser{
			UserId:          u2.ID,
			UserEmail:       u2.Email,
			UserName:        u2.DisplayName,
			RoleName:        database.OrganizationRoleNameViewer,
			ProjectsCount:   2, // Through the autogroup:member being added by default
			UsergroupsCount: 1, // Only the one added
		})
	})

	t.Run("Ability to include users counts in usergroup listings", func(t *testing.T) {
		// Create org and usergroup
		_, c1 := fix.NewUser(t)
		r1, err := c1.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		r2, err := c1.CreateUsergroup(ctx, &adminv1.CreateUsergroupRequest{
			Org:  r1.Organization.Name,
			Name: "group1",
		})
		require.NoError(t, err)

		// Create a project and add the usergroup to it
		r3, err := c1.CreateProject(ctx, &adminv1.CreateProjectRequest{
			Org:        r1.Organization.Name,
			Project:    "proj1",
			ProdSlots:  1,
			SkipDeploy: true,
		})
		require.NoError(t, err)
		_, err = c1.AddProjectMemberUsergroup(ctx, &adminv1.AddProjectMemberUsergroupRequest{
			Org:       r1.Organization.Name,
			Project:   r3.Project.Name,
			Usergroup: r2.Usergroup.GroupName,
			Role:      database.ProjectRoleNameViewer,
		})
		require.NoError(t, err)

		// Add a user to the usergroup
		u2, _ := fix.NewUser(t)
		_, err = c1.AddOrganizationMemberUser(ctx, &adminv1.AddOrganizationMemberUserRequest{
			Org:   r1.Organization.Name,
			Email: u2.Email,
			Role:  database.OrganizationRoleNameGuest,
		})
		require.NoError(t, err)
		_, err = c1.AddUsergroupMemberUser(ctx, &adminv1.AddUsergroupMemberUserRequest{
			Org:       r1.Organization.Name,
			Usergroup: r2.Usergroup.GroupName,
			Email:     u2.Email,
		})
		require.NoError(t, err)

		// Check the counts for the usergroup
		r4, err := c1.ListOrganizationMemberUsergroups(ctx, &adminv1.ListOrganizationMemberUsergroupsRequest{
			Org:           r1.Organization.Name,
			IncludeCounts: true,
		})
		require.NoError(t, err)
		require.Len(t, r4.Members, 4) // There are three system-managed autogroups and the one we added
		for _, m := range r4.Members {
			m.GroupId = ""
			m.CreatedOn = nil
			m.UpdatedOn = nil
		}
		require.Contains(t, r4.Members, &adminv1.MemberUsergroup{
			GroupName:    database.UsergroupNameAutogroupUsers,
			GroupManaged: true,
			UsersCount:   2,
		})
		require.Contains(t, r4.Members, &adminv1.MemberUsergroup{
			GroupName:    database.UsergroupNameAutogroupMembers,
			GroupManaged: true,
			UsersCount:   1,
		})
		require.Contains(t, r4.Members, &adminv1.MemberUsergroup{
			GroupName:    database.UsergroupNameAutogroupGuests,
			GroupManaged: true,
			UsersCount:   1,
		})
		require.Contains(t, r4.Members, &adminv1.MemberUsergroup{
			GroupName:    r2.Usergroup.GroupName,
			GroupManaged: false,
			UsersCount:   1,
		})

		// Check the counts for the project usergroup listing
		r5, err := c1.ListProjectMemberUsergroups(ctx, &adminv1.ListProjectMemberUsergroupsRequest{
			Org:           r1.Organization.Name,
			Project:       r3.Project.Name,
			IncludeCounts: true,
		})
		require.NoError(t, err)
		require.Len(t, r5.Members, 2)
		for _, m := range r5.Members {
			m.GroupId = ""
			m.CreatedOn = nil
			m.UpdatedOn = nil
		}
		require.Contains(t, r5.Members, &adminv1.MemberUsergroup{
			GroupName:    database.UsergroupNameAutogroupMembers,
			GroupManaged: true,
			RoleName:     database.ProjectRoleNameViewer,
			UsersCount:   1,
		})
		require.Contains(t, r5.Members, &adminv1.MemberUsergroup{
			GroupName:    r2.Usergroup.GroupName,
			GroupManaged: false,
			RoleName:     database.ProjectRoleNameViewer,
			UsersCount:   1,
		})
	})

	t.Run("Visibility of public projects", func(t *testing.T) {
		// Create users
		_, c1 := fix.NewUser(t)
		_, c2 := fix.NewUser(t)

		// Create org and public and private projects
		r1, err := c1.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		r2, err := c1.CreateProject(ctx, &adminv1.CreateProjectRequest{
			Org:        r1.Organization.Name,
			Project:    "public",
			ProdSlots:  1,
			SkipDeploy: true,
			Public:     true,
		})
		require.NoError(t, err)
		r3, err := c1.CreateProject(ctx, &adminv1.CreateProjectRequest{
			Org:        r1.Organization.Name,
			Project:    "private",
			ProdSlots:  1,
			SkipDeploy: true,
			Public:     false,
		})
		require.NoError(t, err)

		// Check u2 can access the org and the public project
		_, err = c2.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Org: r1.Organization.Name})
		require.NoError(t, err)
		_, err = c2.GetProject(ctx, &adminv1.GetProjectRequest{Org: r1.Organization.Name, Project: r2.Project.Name})
		require.NoError(t, err)

		// Check u2 can't access or list the private project
		_, err = c2.GetProject(ctx, &adminv1.GetProjectRequest{Org: r1.Organization.Name, Project: r3.Project.Name})
		require.Error(t, err)
		r4, err := c2.ListProjectsForOrganization(ctx, &adminv1.ListProjectsForOrganizationRequest{Org: r1.Organization.Name})
		require.NoError(t, err)
		require.Len(t, r4.Projects, 1)
		require.Equal(t, r2.Project.Name, r4.Projects[0].Name)

		// Delete the public project
		_, err = c1.DeleteProject(ctx, &adminv1.DeleteProjectRequest{
			Org:     r1.Organization.Name,
			Project: r2.Project.Name,
		})
		require.NoError(t, err)

		// Check u2 can't access the org any more
		_, err = c2.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Org: r1.Organization.Name})
		require.Error(t, err)
	})

	t.Run("Inviting users who have not signed up yet", func(t *testing.T) {
		// Create admin user with org, project and group
		_, c1 := fix.NewUser(t)
		org1, err := c1.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		proj1, err := c1.CreateProject(ctx, &adminv1.CreateProjectRequest{
			Org:        org1.Organization.Name,
			Project:    "proj1",
			ProdSlots:  1,
			SkipDeploy: true,
		})
		require.NoError(t, err)
		group1, err := c1.CreateUsergroup(ctx, &adminv1.CreateUsergroupRequest{
			Org:  org1.Organization.Name,
			Name: "group1",
		})
		require.NoError(t, err)

		// Invite a user that doesn't exist to the org, project and group
		userEmail := randomName() + "@example.com"
		_, err = c1.AddOrganizationMemberUser(ctx, &adminv1.AddOrganizationMemberUserRequest{
			Org:   org1.Organization.Name,
			Email: userEmail,
			Role:  database.OrganizationRoleNameViewer,
		})
		require.NoError(t, err)
		_, err = c1.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
			Org:     org1.Organization.Name,
			Project: proj1.Project.Name,
			Email:   userEmail,
			Role:    database.ProjectRoleNameAdmin,
		})
		require.NoError(t, err)
		_, err = c1.AddUsergroupMemberUser(ctx, &adminv1.AddUsergroupMemberUserRequest{
			Org:       org1.Organization.Name,
			Usergroup: group1.Usergroup.GroupName,
			Email:     userEmail,
		})
		require.NoError(t, err)

		// Check that two emails were sent (org and project addition)
		sender := fix.Admin.Email.Sender.(*email.TestSender)
		var count int
		for _, email := range sender.Emails {
			if email.ToEmail == userEmail {
				count++
			}
		}
		require.Equal(t, 2, count)

		// Check that we can list the invites
		orgInvites, err := c1.ListOrganizationInvites(ctx, &adminv1.ListOrganizationInvitesRequest{Org: org1.Organization.Name})
		require.NoError(t, err)
		require.Len(t, orgInvites.Invites, 1)
		require.Equal(t, userEmail, orgInvites.Invites[0].Email)
		require.Equal(t, database.OrganizationRoleNameViewer, orgInvites.Invites[0].RoleName)
		projInvites, err := c1.ListProjectInvites(ctx, &adminv1.ListProjectInvitesRequest{
			Org:     org1.Organization.Name,
			Project: proj1.Project.Name,
		})
		require.NoError(t, err)
		require.Len(t, projInvites.Invites, 1)
		require.Equal(t, userEmail, projInvites.Invites[0].Email)
		require.Equal(t, database.ProjectRoleNameAdmin, projInvites.Invites[0].RoleName)
		require.Equal(t, database.OrganizationRoleNameViewer, projInvites.Invites[0].OrgRoleName)

		// Create the user and check they can access the org and project, and check they are in the list of members
		_, c2 := fix.NewUserWithEmail(t, userEmail)
		_, err = c2.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Org: org1.Organization.Name})
		require.NoError(t, err)
		_, err = c2.GetProject(ctx, &adminv1.GetProjectRequest{Org: org1.Organization.Name, Project: proj1.Project.Name})
		require.NoError(t, err)
		orgMembers, err := c1.ListOrganizationMemberUsers(ctx, &adminv1.ListOrganizationMemberUsersRequest{Org: org1.Organization.Name})
		require.NoError(t, err)
		require.Len(t, orgMembers.Members, 2)
		projMembers, err := c2.ListProjectMemberUsers(ctx, &adminv1.ListProjectMemberUsersRequest{
			Org:     org1.Organization.Name,
			Project: proj1.Project.Name,
		})
		require.NoError(t, err)
		require.Len(t, projMembers.Members, 2)
		groupMembers, err := c1.ListUsergroupMemberUsers(ctx, &adminv1.ListUsergroupMemberUsersRequest{
			Org:       org1.Organization.Name,
			Usergroup: group1.Usergroup.GroupName,
		})
		require.NoError(t, err)
		require.Len(t, groupMembers.Members, 1)

		// Check that the invites were deleted
		orgInvites, err = c1.ListOrganizationInvites(ctx, &adminv1.ListOrganizationInvitesRequest{Org: org1.Organization.Name})
		require.NoError(t, err)
		require.Len(t, orgInvites.Invites, 0)
		projInvites, err = c1.ListProjectInvites(ctx, &adminv1.ListProjectInvitesRequest{
			Org:     org1.Organization.Name,
			Project: proj1.Project.Name,
		})
		require.NoError(t, err)
		require.Len(t, projInvites.Invites, 0)

		// Check that the user was added to the managed groups
		allUsers, err := c1.ListUsergroupMemberUsers(ctx, &adminv1.ListUsergroupMemberUsersRequest{
			Org:       org1.Organization.Name,
			Usergroup: database.UsergroupNameAutogroupUsers,
		})
		require.NoError(t, err)
		require.Len(t, allUsers.Members, 2)
		allGuests, err := c1.ListUsergroupMemberUsers(ctx, &adminv1.ListUsergroupMemberUsersRequest{
			Org:       org1.Organization.Name,
			Usergroup: database.UsergroupNameAutogroupGuests,
		})
		require.NoError(t, err)
		require.Len(t, allGuests.Members, 0)
		allMembers, err := c1.ListUsergroupMemberUsers(ctx, &adminv1.ListUsergroupMemberUsersRequest{
			Org:       org1.Organization.Name,
			Usergroup: database.UsergroupNameAutogroupMembers,
		})
		require.NoError(t, err)
		require.Len(t, allMembers.Members, 2)
	})

	t.Run("Inviting project users who haven't signed up yet become org guests", func(t *testing.T) {
		// Create admin user with org and project
		_, c1 := fix.NewUser(t)
		org1, err := c1.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		proj1, err := c1.CreateProject(ctx, &adminv1.CreateProjectRequest{
			Org:        org1.Organization.Name,
			Project:    "proj1",
			ProdSlots:  1,
			SkipDeploy: true,
		})
		require.NoError(t, err)

		// Invite a user that doesn't exist to the project
		userEmail := randomName() + "@example.com"
		_, err = c1.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
			Org:     org1.Organization.Name,
			Project: proj1.Project.Name,
			Email:   userEmail,
			Role:    database.ProjectRoleNameAdmin,
		})
		require.NoError(t, err)

		// Sign up the user
		_, _ = fix.NewUserWithEmail(t, userEmail)

		// Check that the user is a guest in the org
		orgMembers, err := c1.ListOrganizationMemberUsers(ctx, &adminv1.ListOrganizationMemberUsersRequest{Org: org1.Organization.Name})
		require.NoError(t, err)
		require.Len(t, orgMembers.Members, 2)
		hasGuest := false
		for _, m := range orgMembers.Members {
			if m.RoleName == database.OrganizationRoleNameGuest {
				hasGuest = true
			}
		}
		require.True(t, hasGuest)

		// Check that the user is in the autogroup:users group
		allUsers, err := c1.ListUsergroupMemberUsers(ctx, &adminv1.ListUsergroupMemberUsersRequest{
			Org:       org1.Organization.Name,
			Usergroup: database.UsergroupNameAutogroupUsers,
		})
		require.NoError(t, err)
		require.Len(t, allUsers.Members, 2)
	})

	t.Run("Project invites are connected to org invites", func(t *testing.T) {
		// Create an org and project
		_, c1 := fix.NewUser(t)
		org1, err := c1.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		proj1, err := c1.CreateProject(ctx, &adminv1.CreateProjectRequest{
			Org:        org1.Organization.Name,
			Project:    "proj1",
			ProdSlots:  1,
			SkipDeploy: true,
		})
		require.NoError(t, err)

		// Invite a user to the project
		userEmail := randomName() + "@example.com"
		_, err = c1.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
			Org:     org1.Organization.Name,
			Project: proj1.Project.Name,
			Email:   userEmail,
			Role:    database.ProjectRoleNameAdmin,
		})
		require.NoError(t, err)

		// Check invites were created for both the org and project
		orgInvites, err := c1.ListOrganizationInvites(ctx, &adminv1.ListOrganizationInvitesRequest{Org: org1.Organization.Name})
		require.NoError(t, err)
		require.Len(t, orgInvites.Invites, 1)
		require.Equal(t, userEmail, orgInvites.Invites[0].Email)
		require.Equal(t, database.OrganizationRoleNameGuest, orgInvites.Invites[0].RoleName)
		projInvites, err := c1.ListProjectInvites(ctx, &adminv1.ListProjectInvitesRequest{
			Org:     org1.Organization.Name,
			Project: proj1.Project.Name,
		})
		require.NoError(t, err)
		require.Len(t, projInvites.Invites, 1)
		require.Equal(t, userEmail, projInvites.Invites[0].Email)
		require.Equal(t, database.ProjectRoleNameAdmin, projInvites.Invites[0].RoleName)
		require.Equal(t, database.OrganizationRoleNameGuest, projInvites.Invites[0].OrgRoleName)

		// Delete the org invite
		_, err = c1.RemoveOrganizationMemberUser(ctx, &adminv1.RemoveOrganizationMemberUserRequest{
			Org:   org1.Organization.Name,
			Email: userEmail,
		})
		require.NoError(t, err)

		// Check that both invites were deleted
		orgInvites, err = c1.ListOrganizationInvites(ctx, &adminv1.ListOrganizationInvitesRequest{Org: org1.Organization.Name})
		require.NoError(t, err)
		require.Len(t, orgInvites.Invites, 0)
		projInvites, err = c1.ListProjectInvites(ctx, &adminv1.ListProjectInvitesRequest{
			Org:     org1.Organization.Name,
			Project: proj1.Project.Name,
		})
		require.NoError(t, err)
		require.Len(t, projInvites.Invites, 0)

		// Signup the user and check they are not added to the org
		_, c2 := fix.NewUserWithEmail(t, userEmail)
		orgs, err := c2.ListOrganizations(ctx, &adminv1.ListOrganizationsRequest{})
		require.NoError(t, err)
		require.Len(t, orgs.Organizations, 0)
	})

	t.Run("Adding org and project members with a service account", func(t *testing.T) {
		// Create users
		_, c1 := fix.NewUser(t)
		u2, _ := fix.NewUser(t)

		// Create org and project
		r1, err := c1.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		r2, err := c1.CreateProject(ctx, &adminv1.CreateProjectRequest{
			Org:        r1.Organization.Name,
			Project:    "proj1",
			ProdSlots:  1,
			SkipDeploy: true,
		})
		require.NoError(t, err)

		// Create service
		r3, err := c1.CreateService(ctx, &adminv1.CreateServiceRequest{
			Name:        "service1",
			Org:         r1.Organization.Name,
			OrgRoleName: database.OrganizationRoleNameAdmin,
		})
		require.NoError(t, err)
		r4, err := c1.IssueServiceAuthToken(ctx, &adminv1.IssueServiceAuthTokenRequest{
			Org:         r1.Organization.Name,
			ServiceName: r3.Service.Name,
		})
		c3 := fix.NewClient(t, r4.Token)

		// Add u2 to the project using the service account
		_, err = c3.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
			Org:     r1.Organization.Name,
			Project: r2.Project.Name,
			Email:   u2.Email,
			Role:    database.ProjectRoleNameAdmin,
		})
		require.NoError(t, err)

		// Invite a user that doesn't exist to the org, project and group
		userEmail := randomName() + "@example.com"
		_, err = c3.AddOrganizationMemberUser(ctx, &adminv1.AddOrganizationMemberUserRequest{
			Org:   r1.Organization.Name,
			Email: userEmail,
			Role:  database.OrganizationRoleNameViewer,
		})
		require.NoError(t, err)
		_, err = c3.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
			Org:     r1.Organization.Name,
			Project: r2.Project.Name,
			Email:   userEmail,
			Role:    database.ProjectRoleNameAdmin,
		})
		require.NoError(t, err)

		// List org and project members
		r5, err := c3.ListOrganizationMemberUsers(ctx, &adminv1.ListOrganizationMemberUsersRequest{Org: r1.Organization.Name})
		require.NoError(t, err)
		require.Len(t, r5.Members, 2) // u1, u2
		r6, err := c3.ListProjectMemberUsers(ctx, &adminv1.ListProjectMemberUsersRequest{Org: r1.Organization.Name, Project: r2.Project.Name})
		require.NoError(t, err)
		require.Len(t, r6.Members, 2) // u1, u2

		// List org and project invites
		r7, err := c3.ListOrganizationInvites(ctx, &adminv1.ListOrganizationInvitesRequest{Org: r1.Organization.Name})
		require.NoError(t, err)
		require.Len(t, r7.Invites, 1)
		r8, err := c3.ListProjectInvites(ctx, &adminv1.ListProjectInvitesRequest{Org: r1.Organization.Name, Project: r2.Project.Name})
		require.NoError(t, err)
		require.Len(t, r8.Invites, 1)
	})

	t.Run("Whitelisting domains on orgs", func(t *testing.T) {
		// Create admin user with four orgs
		u1, c1 := fix.NewUserWithDomain(t, "whitelist-orgs.test")
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
		_, _ = fix.NewUserWithEmail(t, userEmail)

		// Whitelist one domain on org1 and org2, another on org3, and none on org4
		_, err = c1.CreateWhitelistedDomain(ctx, &adminv1.CreateWhitelistedDomainRequest{
			Org:    org1.Organization.Name,
			Domain: "whitelist-orgs.test",
			Role:   database.OrganizationRoleNameViewer,
		})
		require.NoError(t, err)
		_, err = c1.CreateWhitelistedDomain(ctx, &adminv1.CreateWhitelistedDomainRequest{
			Org:    org2.Organization.Name,
			Domain: "whitelist-orgs.test",
			Role:   database.OrganizationRoleNameGuest,
		})
		require.NoError(t, err)
		// Since normal admins can only whitelist their own domain, we need to create and add a separate user on the other domain to whitelist it.
		uTemp, cTemp := fix.NewUserWithDomain(t, "whitelist-orgs2.test")
		_, err = c1.AddOrganizationMemberUser(ctx, &adminv1.AddOrganizationMemberUserRequest{
			Org:   org3.Organization.Name,
			Email: uTemp.Email,
			Role:  database.OrganizationRoleNameAdmin,
		})
		require.NoError(t, err)
		_, err = cTemp.CreateWhitelistedDomain(ctx, &adminv1.CreateWhitelistedDomainRequest{
			Org:    org3.Organization.Name,
			Domain: "whitelist-orgs2.test",
			Role:   database.OrganizationRoleNameAdmin,
		})
		require.NoError(t, err)

		// Check we can't whitelist the same domain on the same org again
		_, err = c1.CreateWhitelistedDomain(ctx, &adminv1.CreateWhitelistedDomainRequest{
			Org:    org1.Organization.Name,
			Domain: "whitelist-orgs.test",
			Role:   database.OrganizationRoleNameAdmin,
		})
		require.Error(t, err)

		// Check that the domains are whitelisted
		org1Domains, err := c1.ListWhitelistedDomains(ctx, &adminv1.ListWhitelistedDomainsRequest{Org: org1.Organization.Name})
		require.NoError(t, err)
		require.Len(t, org1Domains.Domains, 1)
		require.Equal(t, "whitelist-orgs.test", org1Domains.Domains[0].Domain)
		org4Domains, err := c1.ListWhitelistedDomains(ctx, &adminv1.ListWhitelistedDomainsRequest{Org: org4.Organization.Name})
		require.NoError(t, err)
		require.Len(t, org4Domains.Domains, 0)

		// Create a user matching a domain AFTER whitelisting
		userEmail2 := randomName() + "@whitelist-orgs.test"
		_, _ = fix.NewUserWithEmail(t, userEmail2)

		// Utils for checking org and group members
		checkOrgMember := func(email string, orgName string, role string, totalMembers int) {
			// Get the org members
			orgMembers, err := c1.ListOrganizationMemberUsers(ctx, &adminv1.ListOrganizationMemberUsersRequest{Org: orgName})
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
				Org:       orgName,
				Usergroup: groupName,
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
		checkOrgMember(adminEmail, org3.Organization.Name, database.OrganizationRoleNameAdmin, 2)
		checkOrgMember(adminEmail, org4.Organization.Name, database.OrganizationRoleNameAdmin, 1)

		checkOrgMember(userEmail, org1.Organization.Name, database.OrganizationRoleNameViewer, 3)
		checkGroupMember(userEmail, org1.Organization.Name, database.UsergroupNameAutogroupUsers, 3)
		checkGroupMember(userEmail, org1.Organization.Name, database.UsergroupNameAutogroupMembers, 3)
		checkOrgMember(userEmail2, org1.Organization.Name, database.OrganizationRoleNameViewer, 3)
		checkGroupMember(userEmail2, org1.Organization.Name, database.UsergroupNameAutogroupUsers, 3)
		checkGroupMember(userEmail2, org1.Organization.Name, database.UsergroupNameAutogroupMembers, 3)

		checkOrgMember(userEmail, org2.Organization.Name, database.OrganizationRoleNameGuest, 3)
		checkGroupMember(userEmail, org2.Organization.Name, database.UsergroupNameAutogroupUsers, 3)
		checkGroupMember(userEmail, org2.Organization.Name, database.UsergroupNameAutogroupGuests, 2)
		checkOrgMember(userEmail2, org2.Organization.Name, database.OrganizationRoleNameGuest, 3)
		checkGroupMember(userEmail2, org2.Organization.Name, database.UsergroupNameAutogroupUsers, 3)
		checkGroupMember(userEmail2, org2.Organization.Name, database.UsergroupNameAutogroupGuests, 2)
	})

	t.Run("Whitelisting domains on projects", func(t *testing.T) {
		// Create an admin user and two orgs with a project each
		u1, c1 := fix.NewUserWithDomain(t, "whitelist-projs.test")
		adminEmail := u1.Email
		org1, err := c1.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		proj1, err := c1.CreateProject(ctx, &adminv1.CreateProjectRequest{
			Org:        org1.Organization.Name,
			Project:    "proj1",
			ProdSlots:  1,
			SkipDeploy: true,
		})
		require.NoError(t, err)
		org2, err := c1.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		proj2, err := c1.CreateProject(ctx, &adminv1.CreateProjectRequest{
			Org:        org2.Organization.Name,
			Project:    "proj2",
			ProdSlots:  1,
			SkipDeploy: true,
		})
		require.NoError(t, err)

		// Create two users before adding the whitelist
		userEmail1 := randomName() + "@whitelist-projs.test"
		_, _ = fix.NewUserWithEmail(t, userEmail1)
		userEmail2 := randomName() + "@whitelist-projs.test"
		_, _ = fix.NewUserWithEmail(t, userEmail2)

		// Add one of the users to the org
		_, err = c1.AddOrganizationMemberUser(ctx, &adminv1.AddOrganizationMemberUserRequest{
			Org:   org1.Organization.Name,
			Email: userEmail1,
			Role:  database.OrganizationRoleNameViewer,
		})
		require.NoError(t, err)

		// Add the whitelist to the project
		_, err = c1.CreateProjectWhitelistedDomain(ctx, &adminv1.CreateProjectWhitelistedDomainRequest{
			Org:     org1.Organization.Name,
			Project: proj1.Project.Name,
			Domain:  "whitelist-projs.test",
			Role:    database.ProjectRoleNameAdmin,
		})
		require.NoError(t, err)

		// Check we can't whitelist the same domain on the same project again
		_, err = c1.CreateProjectWhitelistedDomain(ctx, &adminv1.CreateProjectWhitelistedDomainRequest{
			Org:     org1.Organization.Name,
			Project: proj1.Project.Name,
			Domain:  "whitelist-projs.test",
			Role:    database.ProjectRoleNameAdmin,
		})
		require.Error(t, err)

		// Check that the domain is whitelisted
		proj1Domains, err := c1.ListProjectWhitelistedDomains(ctx, &adminv1.ListProjectWhitelistedDomainsRequest{
			Org:     org1.Organization.Name,
			Project: proj1.Project.Name,
		})
		require.NoError(t, err)
		require.Len(t, proj1Domains.Domains, 1)
		require.Equal(t, "whitelist-projs.test", proj1Domains.Domains[0].Domain)

		// Invite a non-existing user to the org and project
		userEmail3 := randomName() + "@whitelist-projs.test"
		_, err = c1.AddOrganizationMemberUser(ctx, &adminv1.AddOrganizationMemberUserRequest{
			Org:   org1.Organization.Name,
			Email: userEmail3,
			Role:  database.OrganizationRoleNameViewer,
		})
		require.NoError(t, err)
		_, err = c1.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
			Org:     org1.Organization.Name,
			Project: proj1.Project.Name,
			Email:   userEmail3,
			Role:    database.OrganizationRoleNameViewer,
		})
		require.NoError(t, err)

		// Create two users matching the domain, one of whom matches the org-level invite
		_, _ = fix.NewUserWithEmail(t, userEmail3)
		userEmail4 := randomName() + "@whitelist-projs.test"
		_, _ = fix.NewUserWithEmail(t, userEmail4)

		// Utils for checking org, group and project members
		checkOrgMember := func(email string, orgName string, role string, totalMembers int) {
			// Get the org members
			orgMembers, err := c1.ListOrganizationMemberUsers(ctx, &adminv1.ListOrganizationMemberUsersRequest{Org: orgName})
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
				Org:       orgName,
				Usergroup: groupName,
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
				Org:     orgName,
				Project: projName,
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
		checkGroupMember(adminEmail, org1.Organization.Name, database.UsergroupNameAutogroupUsers, 5)
		checkGroupMember(adminEmail, org1.Organization.Name, database.UsergroupNameAutogroupMembers, 3)
		checkGroupMember(userEmail1, org1.Organization.Name, database.UsergroupNameAutogroupMembers, 3)
		checkGroupMember(userEmail2, org1.Organization.Name, database.UsergroupNameAutogroupGuests, 2)
		checkGroupMember(userEmail3, org1.Organization.Name, database.UsergroupNameAutogroupMembers, 3)
		checkGroupMember(userEmail4, org1.Organization.Name, database.UsergroupNameAutogroupGuests, 2)

		// Check that project-level memberships match expectations
		checkProjMember(adminEmail, org1.Organization.Name, proj1.Project.Name, database.ProjectRoleNameAdmin, 5)
		checkProjMember(adminEmail, org2.Organization.Name, proj2.Project.Name, database.ProjectRoleNameAdmin, 1)
		checkProjMember(userEmail1, org1.Organization.Name, proj1.Project.Name, database.ProjectRoleNameAdmin, 5)
		checkProjMember(userEmail2, org1.Organization.Name, proj1.Project.Name, database.ProjectRoleNameAdmin, 5)
		checkProjMember(userEmail3, org1.Organization.Name, proj1.Project.Name, database.ProjectRoleNameViewer, 5) // Because explicit invite role takes precedence over domain whitelist role
		checkProjMember(userEmail4, org1.Organization.Name, proj1.Project.Name, database.ProjectRoleNameAdmin, 5)
	})

	t.Run("Managed usergroup memberships", func(t *testing.T) {
		// Create an admin user and an org
		u1, c1 := fix.NewUser(t)
		org1, err := c1.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)

		// Util function to check usergroup members
		checkGroupMember := func(group, email string, totalMembers int) {
			groupMembers, err := c1.ListUsergroupMemberUsers(ctx, &adminv1.ListUsergroupMemberUsersRequest{
				Org:       org1.Organization.Name,
				Usergroup: group,
			})
			require.NoError(t, err)
			require.Len(t, groupMembers.Members, totalMembers)

			if email == "" {
				return
			}
			found := false
			for _, m := range groupMembers.Members {
				if m.UserEmail == email {
					found = true
				}
			}
			require.True(t, found)
		}

		// Check that the user is in the right managed groups
		checkGroupMember(database.UsergroupNameAutogroupUsers, u1.Email, 1)
		checkGroupMember(database.UsergroupNameAutogroupMembers, u1.Email, 1)
		checkGroupMember(database.UsergroupNameAutogroupGuests, "", 0)

		// Create another user, add them to the org, and check memberships
		u2, _ := fix.NewUser(t)
		_, err = c1.AddOrganizationMemberUser(ctx, &adminv1.AddOrganizationMemberUserRequest{
			Org:   org1.Organization.Name,
			Email: u2.Email,
			Role:  database.OrganizationRoleNameGuest,
		})
		require.NoError(t, err)
		checkGroupMember(database.UsergroupNameAutogroupUsers, u2.Email, 2)
		checkGroupMember(database.UsergroupNameAutogroupMembers, "", 1)
		checkGroupMember(database.UsergroupNameAutogroupGuests, u2.Email, 1)

		// Change the user to be a non-guest member and check memberships
		_, err = c1.SetOrganizationMemberUserRole(ctx, &adminv1.SetOrganizationMemberUserRoleRequest{
			Org:   org1.Organization.Name,
			Email: u2.Email,
			Role:  database.OrganizationRoleNameViewer,
		})
		require.NoError(t, err)
		checkGroupMember(database.UsergroupNameAutogroupUsers, u2.Email, 2)
		checkGroupMember(database.UsergroupNameAutogroupMembers, u2.Email, 2)
		checkGroupMember(database.UsergroupNameAutogroupGuests, "", 0)

		// Remove the user from the org and check memberships
		_, err = c1.RemoveOrganizationMemberUser(ctx, &adminv1.RemoveOrganizationMemberUserRequest{
			Org:   org1.Organization.Name,
			Email: u2.Email,
		})
		require.NoError(t, err)
		checkGroupMember(database.UsergroupNameAutogroupUsers, u1.Email, 1)
		checkGroupMember(database.UsergroupNameAutogroupMembers, u1.Email, 1)
		checkGroupMember(database.UsergroupNameAutogroupGuests, "", 0)
	})

	t.Run("Editors can manage non-admin users only", func(t *testing.T) {
		// Create an org, project and usergroup
		_, c1 := fix.NewUser(t)
		org1, err := c1.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		proj1, err := c1.CreateProject(ctx, &adminv1.CreateProjectRequest{
			Org:        org1.Organization.Name,
			Project:    "proj1",
			ProdSlots:  1,
			SkipDeploy: true,
		})
		require.NoError(t, err)
		group1, err := c1.CreateUsergroup(ctx, &adminv1.CreateUsergroupRequest{
			Org:  org1.Organization.Name,
			Name: "group1",
		})
		require.NoError(t, err)

		// Create an editor user and add them to the org and project
		u2, c2 := fix.NewUser(t)
		_, err = c1.AddOrganizationMemberUser(ctx, &adminv1.AddOrganizationMemberUserRequest{
			Org:   org1.Organization.Name,
			Email: u2.Email,
			Role:  database.OrganizationRoleNameEditor,
		})
		require.NoError(t, err)
		_, err = c1.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
			Org:     org1.Organization.Name,
			Project: proj1.Project.Name,
			Email:   u2.Email,
			Role:    database.ProjectRoleNameEditor,
		})
		require.NoError(t, err)

		// Check that the editor can add a user and usergroup to the org and project
		u3, _ := fix.NewUser(t)
		_, err = c2.AddOrganizationMemberUser(ctx, &adminv1.AddOrganizationMemberUserRequest{
			Org:   org1.Organization.Name,
			Email: u3.Email,
			Role:  database.OrganizationRoleNameEditor,
		})
		require.NoError(t, err)
		_, err = c2.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
			Org:     org1.Organization.Name,
			Project: proj1.Project.Name,
			Email:   u3.Email,
			Role:    database.ProjectRoleNameEditor,
		})
		require.NoError(t, err)
		_, err = c2.AddOrganizationMemberUsergroup(ctx, &adminv1.AddOrganizationMemberUsergroupRequest{
			Org:       org1.Organization.Name,
			Usergroup: group1.Usergroup.GroupName,
			Role:      database.OrganizationRoleNameEditor,
		})
		require.NoError(t, err)
		_, err = c2.AddProjectMemberUsergroup(ctx, &adminv1.AddProjectMemberUsergroupRequest{
			Org:       org1.Organization.Name,
			Project:   proj1.Project.Name,
			Usergroup: group1.Usergroup.GroupName,
			Role:      database.ProjectRoleNameEditor,
		})
		require.NoError(t, err)

		// Check that the editor can add a member to the usergroup
		_, err = c2.AddUsergroupMemberUser(ctx, &adminv1.AddUsergroupMemberUserRequest{
			Org:       org1.Organization.Name,
			Usergroup: group1.Usergroup.GroupName,
			Email:     u3.Email,
		})
		require.NoError(t, err)

		// Check that the editor can change the role of a user and usergroup in the org and project
		_, err = c2.SetOrganizationMemberUserRole(ctx, &adminv1.SetOrganizationMemberUserRoleRequest{
			Org:   org1.Organization.Name,
			Email: u3.Email,
			Role:  database.OrganizationRoleNameViewer,
		})
		require.NoError(t, err)
		viewerRole := database.ProjectRoleNameViewer
		_, err = c2.SetProjectMemberUserRole(ctx, &adminv1.SetProjectMemberUserRoleRequest{
			Org:     org1.Organization.Name,
			Project: proj1.Project.Name,
			Email:   u3.Email,
			Role:    &viewerRole,
		})
		require.NoError(t, err)
		_, err = c2.SetOrganizationMemberUsergroupRole(ctx, &adminv1.SetOrganizationMemberUsergroupRoleRequest{
			Org:       org1.Organization.Name,
			Usergroup: group1.Usergroup.GroupName,
			Role:      database.OrganizationRoleNameViewer,
		})
		require.NoError(t, err)
		_, err = c2.SetProjectMemberUsergroupRole(ctx, &adminv1.SetProjectMemberUsergroupRoleRequest{
			Org:       org1.Organization.Name,
			Project:   proj1.Project.Name,
			Usergroup: group1.Usergroup.GroupName,
			Role:      &viewerRole,
		})
		require.NoError(t, err)

		// Check that the editor can remove a user and usergroup from the org and project
		_, err = c2.RemoveProjectMemberUser(ctx, &adminv1.RemoveProjectMemberUserRequest{
			Org:     org1.Organization.Name,
			Project: proj1.Project.Name,
			Email:   u3.Email,
		})
		require.NoError(t, err)
		_, err = c2.RemoveProjectMemberUsergroup(ctx, &adminv1.RemoveProjectMemberUsergroupRequest{
			Org:       org1.Organization.Name,
			Project:   proj1.Project.Name,
			Usergroup: group1.Usergroup.GroupName,
		})
		require.NoError(t, err)
		_, err = c2.RemoveOrganizationMemberUsergroup(ctx, &adminv1.RemoveOrganizationMemberUsergroupRequest{
			Org:       org1.Organization.Name,
			Usergroup: group1.Usergroup.GroupName,
		})
		require.NoError(t, err)
		_, err = c2.RemoveOrganizationMemberUser(ctx, &adminv1.RemoveOrganizationMemberUserRequest{
			Org:   org1.Organization.Name,
			Email: u3.Email,
		})
		require.NoError(t, err)

		// Check that the editor can't add a user or usergroup to the org or project with an admin role
		_, err = c2.AddOrganizationMemberUser(ctx, &adminv1.AddOrganizationMemberUserRequest{
			Org:   org1.Organization.Name,
			Email: u3.Email,
			Role:  database.OrganizationRoleNameAdmin,
		})
		require.ErrorContains(t, err, "non-admin")
		_, err = c2.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
			Org:     org1.Organization.Name,
			Project: proj1.Project.Name,
			Email:   u3.Email,
			Role:    database.ProjectRoleNameAdmin,
		})
		require.ErrorContains(t, err, "non-admin")
		_, err = c2.AddOrganizationMemberUsergroup(ctx, &adminv1.AddOrganizationMemberUsergroupRequest{
			Org:       org1.Organization.Name,
			Usergroup: group1.Usergroup.GroupName,
			Role:      database.OrganizationRoleNameAdmin,
		})
		require.ErrorContains(t, err, "non-admin")
		_, err = c2.AddProjectMemberUsergroup(ctx, &adminv1.AddProjectMemberUsergroupRequest{
			Org:       org1.Organization.Name,
			Project:   proj1.Project.Name,
			Usergroup: group1.Usergroup.GroupName,
			Role:      database.ProjectRoleNameAdmin,
		})
		require.ErrorContains(t, err, "non-admin")

		// Use the admin user to add an admin user and usergroup to the org and project
		_, err = c1.AddOrganizationMemberUser(ctx, &adminv1.AddOrganizationMemberUserRequest{
			Org:   org1.Organization.Name,
			Email: u3.Email,
			Role:  database.OrganizationRoleNameAdmin,
		})
		require.NoError(t, err)
		_, err = c1.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
			Org:     org1.Organization.Name,
			Project: proj1.Project.Name,
			Email:   u3.Email,
			Role:    database.ProjectRoleNameAdmin,
		})
		require.NoError(t, err)
		_, err = c1.AddOrganizationMemberUsergroup(ctx, &adminv1.AddOrganizationMemberUsergroupRequest{
			Org:       org1.Organization.Name,
			Usergroup: group1.Usergroup.GroupName,
			Role:      database.OrganizationRoleNameAdmin,
		})
		require.NoError(t, err)
		_, err = c1.AddProjectMemberUsergroup(ctx, &adminv1.AddProjectMemberUsergroupRequest{
			Org:       org1.Organization.Name,
			Project:   proj1.Project.Name,
			Usergroup: group1.Usergroup.GroupName,
			Role:      database.ProjectRoleNameAdmin,
		})
		require.NoError(t, err)

		// Check that the editor can't add a member to the usergroup now that it has an admin role
		u4, _ := fix.NewUser(t)
		_, err = c2.AddOrganizationMemberUser(ctx, &adminv1.AddOrganizationMemberUserRequest{
			Org:   org1.Organization.Name,
			Email: u4.Email,
			Role:  database.OrganizationRoleNameViewer,
		})
		require.NoError(t, err)
		_, err = c2.AddUsergroupMemberUser(ctx, &adminv1.AddUsergroupMemberUserRequest{
			Org:       org1.Organization.Name,
			Usergroup: group1.Usergroup.GroupName,
			Email:     u4.Email,
		})
		require.ErrorContains(t, err, "non-admin")

		// Check that the editor can't change the role of a user or usergroup in the org or project to an admin role
		_, err = c2.SetOrganizationMemberUserRole(ctx, &adminv1.SetOrganizationMemberUserRoleRequest{
			Org:   org1.Organization.Name,
			Email: u3.Email,
			Role:  database.OrganizationRoleNameViewer,
		})
		require.ErrorContains(t, err, "non-admin")
		_, err = c2.SetProjectMemberUserRole(ctx, &adminv1.SetProjectMemberUserRoleRequest{
			Org:     org1.Organization.Name,
			Project: proj1.Project.Name,
			Email:   u3.Email,
			Role:    &viewerRole,
		})
		require.ErrorContains(t, err, "non-admin")
		_, err = c2.SetOrganizationMemberUsergroupRole(ctx, &adminv1.SetOrganizationMemberUsergroupRoleRequest{
			Org:       org1.Organization.Name,
			Usergroup: group1.Usergroup.GroupName,
			Role:      database.OrganizationRoleNameViewer,
		})
		require.ErrorContains(t, err, "non-admin")
		_, err = c2.SetProjectMemberUsergroupRole(ctx, &adminv1.SetProjectMemberUsergroupRoleRequest{
			Org:       org1.Organization.Name,
			Project:   proj1.Project.Name,
			Usergroup: group1.Usergroup.GroupName,
			Role:      &viewerRole,
		})
		require.ErrorContains(t, err, "non-admin")

		// Check that the editor can't remove a user or usergroup from the org or project with an admin role
		_, err = c2.RemoveOrganizationMemberUser(ctx, &adminv1.RemoveOrganizationMemberUserRequest{
			Org:   org1.Organization.Name,
			Email: u3.Email,
		})
		require.ErrorContains(t, err, "non-admin")
		_, err = c2.RemoveProjectMemberUser(ctx, &adminv1.RemoveProjectMemberUserRequest{
			Org:     org1.Organization.Name,
			Project: proj1.Project.Name,
			Email:   u3.Email,
		})
		require.ErrorContains(t, err, "non-admin")
		_, err = c2.RemoveOrganizationMemberUsergroup(ctx, &adminv1.RemoveOrganizationMemberUsergroupRequest{
			Org:       org1.Organization.Name,
			Usergroup: group1.Usergroup.GroupName,
		})
		require.ErrorContains(t, err, "non-admin")
		_, err = c2.RemoveProjectMemberUsergroup(ctx, &adminv1.RemoveProjectMemberUsergroupRequest{
			Org:       org1.Organization.Name,
			Project:   proj1.Project.Name,
			Usergroup: group1.Usergroup.GroupName,
		})
		require.ErrorContains(t, err, "non-admin")
	})

	t.Run("Organization admins can inspect the projects and usergroups of a user", func(t *testing.T) {
		// Create an org, project and usergroup
		_, c1 := fix.NewUser(t)
		org1, err := c1.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		proj1, err := c1.CreateProject(ctx, &adminv1.CreateProjectRequest{
			Org:        org1.Organization.Name,
			Project:    "proj1",
			ProdSlots:  1,
			SkipDeploy: true,
		})
		require.NoError(t, err)
		group1, err := c1.CreateUsergroup(ctx, &adminv1.CreateUsergroupRequest{
			Org:  org1.Organization.Name,
			Name: "group1",
		})
		require.NoError(t, err)

		// Create a user and add them to the org
		u2, _ := fix.NewUser(t)
		_, err = c1.AddOrganizationMemberUser(ctx, &adminv1.AddOrganizationMemberUserRequest{
			Org:   org1.Organization.Name,
			Email: u2.Email,
			Role:  database.OrganizationRoleNameViewer,
		})
		require.NoError(t, err)

		// Introspect the user's projects and usergroups
		projects, err := c1.ListProjectsForOrganizationAndUser(ctx, &adminv1.ListProjectsForOrganizationAndUserRequest{
			Org:    org1.Organization.Name,
			UserId: u2.ID,
		})
		require.NoError(t, err)
		require.Len(t, projects.Projects, 1) // Through the default group access for autogroup:members
		require.Equal(t, proj1.Project.Name, projects.Projects[0].Name)
		usergroups, err := c1.ListUsergroupsForOrganizationAndUser(ctx, &adminv1.ListUsergroupsForOrganizationAndUserRequest{
			Org:    org1.Organization.Name,
			UserId: u2.ID,
		})
		require.NoError(t, err)
		require.Len(t, usergroups.Usergroups, 2) // The default autogroups
		require.Equal(t, database.UsergroupNameAutogroupMembers, usergroups.Usergroups[0].GroupName)
		require.Equal(t, database.UsergroupNameAutogroupUsers, usergroups.Usergroups[1].GroupName)

		// Add the user explicitly to the project and usergroup
		_, err = c1.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
			Org:     org1.Organization.Name,
			Project: proj1.Project.Name,
			Email:   u2.Email,
			Role:    database.ProjectRoleNameViewer,
		})
		require.NoError(t, err)
		_, err = c1.AddUsergroupMemberUser(ctx, &adminv1.AddUsergroupMemberUserRequest{
			Org:       org1.Organization.Name,
			Usergroup: group1.Usergroup.GroupName,
			Email:     u2.Email,
		})
		require.NoError(t, err)

		// Check that the user has the project and usergroup
		projects, err = c1.ListProjectsForOrganizationAndUser(ctx, &adminv1.ListProjectsForOrganizationAndUserRequest{
			Org:    org1.Organization.Name,
			UserId: u2.ID,
		})
		require.NoError(t, err)
		require.Len(t, projects.Projects, 1)
		require.Equal(t, proj1.Project.Name, projects.Projects[0].Name)
		usergroups, err = c1.ListUsergroupsForOrganizationAndUser(ctx, &adminv1.ListUsergroupsForOrganizationAndUserRequest{
			Org:    org1.Organization.Name,
			UserId: u2.ID,
		})
		require.NoError(t, err)
		require.Len(t, usergroups.Usergroups, 3)
		require.Equal(t, group1.Usergroup.GroupName, usergroups.Usergroups[2].GroupName)
	})

	t.Run("User attributes", func(t *testing.T) {
		// Create users
		u1, c1 := fix.NewUserWithEmail(t, "attr_user@example.com")
		u2, c2 := fix.NewUser(t)

		// Create org
		r1, err := c1.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)

		// Add user2 as member
		_, err = c1.AddOrganizationMemberUser(ctx, &adminv1.AddOrganizationMemberUserRequest{
			Org:   r1.Organization.Name,
			Email: u2.Email,
			Role:  database.OrganizationRoleNameViewer,
		})
		require.NoError(t, err)

		// Set custom attributes for user1
		attrs := map[string]interface{}{
			"restaurant_id": "123",
			"department":    "engineering",
		}
		attrStruct, err := structpb.NewStruct(attrs)
		require.NoError(t, err)

		// Test setting attributes
		_, err = c1.UpdateOrganizationMemberUserAttributes(ctx, &adminv1.UpdateOrganizationMemberUserAttributesRequest{
			Org:        r1.Organization.Name,
			Email:      u1.Email,
			Attributes: attrStruct,
		})
		require.NoError(t, err)

		// Test permission check: user2 cannot set attributes for user1
		_, err = c2.UpdateOrganizationMemberUserAttributes(ctx, &adminv1.UpdateOrganizationMemberUserAttributesRequest{
			Org:        r1.Organization.Name,
			Email:      u1.Email,
			Attributes: attrStruct,
		})
		require.Error(t, err) // Should fail due to insufficient permissions

		// Verify attributes were set
		resp, err := c1.GetOrganizationMemberUser(ctx, &adminv1.GetOrganizationMemberUserRequest{
			Org:   r1.Organization.Name,
			Email: u1.Email,
		})
		require.NoError(t, err)
		require.Equal(t, attrs, resp.Member.Attributes.AsMap())

		// Test updating attributes
		newAttrs := map[string]interface{}{
			"restaurant_id": "456",
			"team":          "platform",
		}
		newAttrStruct, err := structpb.NewStruct(newAttrs)
		require.NoError(t, err)

		_, err = c1.UpdateOrganizationMemberUserAttributes(ctx, &adminv1.UpdateOrganizationMemberUserAttributesRequest{
			Org:        r1.Organization.Name,
			Email:      u1.Email,
			Attributes: newAttrStruct,
		})
		require.NoError(t, err)

		// Verify updated attributes
		updatedResp, err := c1.GetOrganizationMemberUser(ctx, &adminv1.GetOrganizationMemberUserRequest{
			Org:   r1.Organization.Name,
			Email: u1.Email,
		})
		require.NoError(t, err)
		require.Equal(t, newAttrs, updatedResp.Member.Attributes.AsMap())
	})

	t.Run("Project member role updates preserve resource restrictions when omitted", func(t *testing.T) {
		_, admin := fix.NewUser(t) // ignore first user as it will be superuser
		user, userClient := fix.NewUser(t)

		org, err := admin.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		project, err := admin.CreateProject(ctx, &adminv1.CreateProjectRequest{
			Org:        org.Organization.Name,
			Project:    "proj-scope",
			ProdSlots:  1,
			SkipDeploy: true,
		})
		require.NoError(t, err)

		depl, err := admin.CreateDeployment(ctx, &adminv1.CreateDeploymentRequest{
			Org:         org.Organization.Name,
			Project:     project.Project.Name,
			Environment: "prod",
		})
		require.NoError(t, err)

		// directly update deployment to running with runtime info especially audience for correct jwt generation
		_, err = fix.Admin.DB.UpdateDeploymentUnsafe(ctx, depl.Deployment.Id, &database.UpdateDeploymentUnsafeOptions{
			RuntimeHost:       "rill-example-host",
			RuntimeInstanceID: "rill-example-instance",
			RuntimeAudience:   "http://example.org",
			Status:            database.DeploymentStatusRunning,
			StatusMessage:     "Running",
		})
		require.NoError(t, err)

		restrict := true
		_, err = admin.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
			Org:               org.Organization.Name,
			Project:           project.Project.Name,
			Email:             user.Email,
			Role:              database.ProjectRoleNameEditor,
			RestrictResources: &restrict,
			Resources:         []*adminv1.ResourceName{{Type: "rill.runtime.v1.Explore", Name: "explore"}},
		})
		require.NoError(t, err)

		member, err := admin.GetProjectMemberUser(ctx, &adminv1.GetProjectMemberUserRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
			Email:   user.Email,
		})
		require.NoError(t, err)
		require.True(t, member.Member.RestrictResources)
		require.NotEmpty(t, member.Member.Resources)
		require.Equal(t, "rill.runtime.v1.Explore", member.Member.Resources[0].Type)
		require.Equal(t, "explore", member.Member.Resources[0].Name)

		role := database.ProjectRoleNameViewer
		_, err = admin.SetProjectMemberUserRole(ctx, &adminv1.SetProjectMemberUserRoleRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
			Email:   user.Email,
			Role:    &role,
		})
		require.NoError(t, err)

		member, err = admin.GetProjectMemberUser(ctx, &adminv1.GetProjectMemberUserRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
			Email:   user.Email,
		})
		require.NoError(t, err)
		require.Equal(t, database.ProjectRoleNameViewer, member.Member.RoleName)
		require.True(t, member.Member.RestrictResources)
		require.NotEmpty(t, member.Member.Resources)
		require.Equal(t, "rill.runtime.v1.Explore", member.Member.Resources[0].Type)
		require.Equal(t, "explore", member.Member.Resources[0].Name)

		proj, err := userClient.GetProject(ctx, &adminv1.GetProjectRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
		})
		require.NoError(t, err)
		// parse jwt
		require.NotNil(t, proj.Jwt)
		claims, err := fix.Audience.ParseAndValidate(proj.Jwt)
		require.NoError(t, err)
		require.NotNil(t, claims)
		rules := claims.Claims("").AdditionalRules
		require.Len(t, rules, 1)
		require.NotNil(t, rules[0].GetTransitiveAccess())
		require.Equal(t, "rill.runtime.v1.Explore", rules[0].GetTransitiveAccess().Resource.Kind)
		require.Equal(t, "explore", rules[0].GetTransitiveAccess().Resource.Name)
	})

	t.Run("Project usergroup role updates preserve resource restrictions when omitted", func(t *testing.T) {
		_, admin := fix.NewUser(t) // ignore first user as it will be superuser
		user, userClient := fix.NewUser(t)

		org, err := admin.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		project, err := admin.CreateProject(ctx, &adminv1.CreateProjectRequest{
			Org:        org.Organization.Name,
			Project:    "proj-group-scope",
			ProdSlots:  1,
			SkipDeploy: true,
		})
		require.NoError(t, err)

		depl, err := admin.CreateDeployment(ctx, &adminv1.CreateDeploymentRequest{
			Org:         org.Organization.Name,
			Project:     project.Project.Name,
			Environment: "prod",
		})
		require.NoError(t, err)

		// directly update deployment to running with runtime info especially audience for correct jwt generation
		_, err = fix.Admin.DB.UpdateDeploymentUnsafe(ctx, depl.Deployment.Id, &database.UpdateDeploymentUnsafeOptions{
			RuntimeHost:       "rill-example-host",
			RuntimeInstanceID: "rill-example-instance",
			RuntimeAudience:   "http://example.org",
			Status:            database.DeploymentStatusRunning,
			StatusMessage:     "Running",
		})
		require.NoError(t, err)

		group, err := admin.CreateUsergroup(ctx, &adminv1.CreateUsergroupRequest{
			Org:  org.Organization.Name,
			Name: "group-scope",
		})
		require.NoError(t, err)

		_, err = admin.AddOrganizationMemberUser(ctx, &adminv1.AddOrganizationMemberUserRequest{
			Org:   org.Organization.Name,
			Email: user.Email,
			Role:  database.OrganizationRoleNameViewer,
		})
		require.NoError(t, err)
		_, err = admin.AddUsergroupMemberUser(ctx, &adminv1.AddUsergroupMemberUserRequest{
			Org:       org.Organization.Name,
			Usergroup: group.Usergroup.GroupName,
			Email:     user.Email,
		})
		require.NoError(t, err)

		restrict := true
		_, err = admin.AddProjectMemberUsergroup(ctx, &adminv1.AddProjectMemberUsergroupRequest{
			Org:               org.Organization.Name,
			Project:           project.Project.Name,
			Usergroup:         group.Usergroup.GroupName,
			Role:              database.ProjectRoleNameEditor,
			RestrictResources: &restrict,
		})
		require.NoError(t, err)

		groups, err := admin.ListUsergroupsForProjectAndUser(ctx, &adminv1.ListUsergroupsForProjectAndUserRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
			Email:   user.Email,
		})
		require.NoError(t, err)
		require.NotEmpty(t, groups.Usergroups)
		found := false
		for _, g := range groups.Usergroups {
			if g.GroupName == group.Usergroup.GroupName {
				require.True(t, g.RestrictResources)
				require.Equal(t, database.ProjectRoleNameEditor, g.RoleName)
				found = true
				break
			}
		}
		require.True(t, found)

		role := database.ProjectRoleNameViewer
		_, err = admin.SetProjectMemberUsergroupRole(ctx, &adminv1.SetProjectMemberUsergroupRoleRequest{
			Org:       org.Organization.Name,
			Project:   project.Project.Name,
			Usergroup: group.Usergroup.GroupName,
			Role:      &role,
		})
		require.NoError(t, err)

		groups, err = admin.ListUsergroupsForProjectAndUser(ctx, &adminv1.ListUsergroupsForProjectAndUserRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
			Email:   user.Email,
		})
		require.NoError(t, err)
		require.NotEmpty(t, groups.Usergroups)
		found = false
		for _, g := range groups.Usergroups {
			if g.GroupName == group.Usergroup.GroupName {
				require.True(t, g.RestrictResources)
				require.Equal(t, database.ProjectRoleNameViewer, g.RoleName)
				found = true
				break
			}
		}
		require.True(t, found)

		// it will still not be restricted as all org level users are part of autogroup:members and it is unrestricted
		proj, err := userClient.GetProject(ctx, &adminv1.GetProjectRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
		})
		require.NoError(t, err)
		// parse jwt
		require.NotNil(t, proj.Jwt)
		claims, err := fix.Audience.ParseAndValidate(proj.Jwt)
		require.NoError(t, err)
		require.NotNil(t, claims)
		rules := claims.Claims("").AdditionalRules
		require.Len(t, rules, 0)

		// restrict autogroup:members so that usergroup restriction takes effect
		restrict = true
		_, err = admin.SetProjectMemberUsergroupRole(ctx, &adminv1.SetProjectMemberUsergroupRoleRequest{
			Org:               org.Organization.Name,
			Project:           project.Project.Name,
			Usergroup:         database.UsergroupNameAutogroupMembers,
			RestrictResources: &restrict,
		})
		require.NoError(t, err)
		// fetch again and check it is restricted now
		proj, err = userClient.GetProject(ctx, &adminv1.GetProjectRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
		})
		require.NoError(t, err)
		// parse jwt
		require.NotNil(t, proj.Jwt)
		claims, err = fix.Audience.ParseAndValidate(proj.Jwt)
		require.NoError(t, err)
		require.NotNil(t, claims)
		rules = claims.Claims("").AdditionalRules
		require.Len(t, rules, 1)
		require.NotNil(t, rules[0].GetAccess())
		require.Equal(t, false, rules[0].GetAccess().Allow)
	})

	t.Run("Project member restrictions can be added then cleared", func(t *testing.T) {
		_, admin := fix.NewUser(t) // ignore first user as it will be superuser
		user, userClient := fix.NewUser(t)

		org, err := admin.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		project, err := admin.CreateProject(ctx, &adminv1.CreateProjectRequest{
			Org:        org.Organization.Name,
			Project:    "proj-member-add-remove",
			ProdSlots:  1,
			SkipDeploy: true,
		})
		require.NoError(t, err)

		depl, err := admin.CreateDeployment(ctx, &adminv1.CreateDeploymentRequest{
			Org:         org.Organization.Name,
			Project:     project.Project.Name,
			Environment: "prod",
		})
		require.NoError(t, err)

		// directly update deployment to running with runtime info especially audience for correct jwt generation
		_, err = fix.Admin.DB.UpdateDeploymentUnsafe(ctx, depl.Deployment.Id, &database.UpdateDeploymentUnsafeOptions{
			RuntimeHost:       "rill-example-host",
			RuntimeInstanceID: "rill-example-instance",
			RuntimeAudience:   "http://example.org",
			Status:            database.DeploymentStatusRunning,
			StatusMessage:     "Running",
		})
		require.NoError(t, err)

		_, err = admin.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
			Email:   user.Email,
			Role:    database.ProjectRoleNameEditor,
		})
		require.NoError(t, err)

		member, err := admin.GetProjectMemberUser(ctx, &adminv1.GetProjectMemberUserRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
			Email:   user.Email,
		})
		require.NoError(t, err)
		require.False(t, member.Member.RestrictResources)
		require.Empty(t, member.Member.Resources)

		restrictParam := true
		_, err = admin.SetProjectMemberUserRole(ctx, &adminv1.SetProjectMemberUserRoleRequest{
			Org:               org.Organization.Name,
			Project:           project.Project.Name,
			Email:             user.Email,
			RestrictResources: &restrictParam,
			Resources:         []*adminv1.ResourceName{{Type: "rill.runtime.v1.Explore", Name: "orders"}},
		})
		require.NoError(t, err)

		member, err = admin.GetProjectMemberUser(ctx, &adminv1.GetProjectMemberUserRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
			Email:   user.Email,
		})
		require.NoError(t, err)
		require.True(t, member.Member.RestrictResources)
		require.Len(t, member.Member.Resources, 1)

		proj, err := userClient.GetProject(ctx, &adminv1.GetProjectRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
		})
		require.NoError(t, err)
		// parse jwt
		require.NotNil(t, proj.Jwt)
		claims, err := fix.Audience.ParseAndValidate(proj.Jwt)
		require.NoError(t, err)
		require.NotNil(t, claims)
		rules := claims.Claims("").AdditionalRules
		require.Len(t, rules, 1)
		require.NotNil(t, rules[0].GetTransitiveAccess())
		require.Equal(t, "rill.runtime.v1.Explore", rules[0].GetTransitiveAccess().Resource.Kind)
		require.Equal(t, "orders", rules[0].GetTransitiveAccess().Resource.Name)

		restrictParam = false
		_, err = admin.SetProjectMemberUserRole(ctx, &adminv1.SetProjectMemberUserRoleRequest{
			Org:               org.Organization.Name,
			Project:           project.Project.Name,
			Email:             user.Email,
			RestrictResources: &restrictParam,
			Resources:         nil,
		})
		require.NoError(t, err)
		require.NoError(t, err)

		member, err = admin.GetProjectMemberUser(ctx, &adminv1.GetProjectMemberUserRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
			Email:   user.Email,
		})
		require.NoError(t, err)
		require.False(t, member.Member.RestrictResources)
		require.Empty(t, member.Member.Resources)

		proj, err = userClient.GetProject(ctx, &adminv1.GetProjectRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
		})
		require.NoError(t, err)
		// parse jwt
		require.NotNil(t, proj.Jwt)
		claims, err = fix.Audience.ParseAndValidate(proj.Jwt)
		require.NoError(t, err)
		require.NotNil(t, claims)
		rules = claims.Claims("").AdditionalRules
		require.Len(t, rules, 0)
	})

	t.Run("Resource restrictions flag behaviour", func(t *testing.T) {
		_, admin := fix.NewUser(t) // ignore first user as it will be superuser
		user, userClient := fix.NewUser(t)

		org, err := admin.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		project, err := admin.CreateProject(ctx, &adminv1.CreateProjectRequest{
			Org:        org.Organization.Name,
			Project:    "proj",
			ProdSlots:  1,
			SkipDeploy: true,
		})
		require.NoError(t, err)
		depl, err := admin.CreateDeployment(ctx, &adminv1.CreateDeploymentRequest{
			Org:         org.Organization.Name,
			Project:     project.Project.Name,
			Environment: "prod",
		})
		require.NoError(t, err)

		// directly update deployment to running with runtime info especially audience for correct jwt generation
		_, err = fix.Admin.DB.UpdateDeploymentUnsafe(ctx, depl.Deployment.Id, &database.UpdateDeploymentUnsafeOptions{
			RuntimeHost:       "rill-example-host",
			RuntimeInstanceID: "rill-example-instance",
			RuntimeAudience:   "http://example.org",
			Status:            database.DeploymentStatusRunning,
			StatusMessage:     "Running",
		})
		require.NoError(t, err)

		_, err = admin.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
			Email:   user.Email,
			Role:    database.ProjectRoleNameEditor,
		})
		require.NoError(t, err)

		// now don't restrict the user but add resources
		restrictFlag := false
		_, err = admin.SetProjectMemberUserRole(ctx, &adminv1.SetProjectMemberUserRoleRequest{
			Org:               org.Organization.Name,
			Project:           project.Project.Name,
			Email:             user.Email,
			RestrictResources: &restrictFlag, // will be ignored as resources are provided
			Resources:         []*adminv1.ResourceName{{Type: "metrics_view", Name: "mv1"}},
		})
		require.NoError(t, err)

		memResp, err := admin.GetProjectMemberUser(ctx, &adminv1.GetProjectMemberUserRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
			Email:   user.Email,
		})
		require.NoError(t, err)
		require.True(t, memResp.Member.RestrictResources) // restricted flag ignored when resources list not empty in the request
		require.Len(t, memResp.Member.Resources, 1)
		require.Equal(t, &adminv1.ResourceName{Type: "metrics_view", Name: "mv1"}, memResp.Member.Resources[0])

		proj, err := userClient.GetProject(ctx, &adminv1.GetProjectRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
		})
		require.NoError(t, err)
		// parse jwt
		require.NotNil(t, proj.Jwt)
		claims, err := fix.Audience.ParseAndValidate(proj.Jwt)
		require.NoError(t, err)
		require.NotNil(t, claims)
		rules := claims.Claims("").AdditionalRules
		require.Len(t, rules, 1)
		require.NotNil(t, rules[0].GetTransitiveAccess())
		require.Equal(t, "metrics_view", rules[0].GetTransitiveAccess().Resource.Kind)
		require.Equal(t, "mv1", rules[0].GetTransitiveAccess().Resource.Name)
	})

	t.Run("User and usergroup should all be restricted to take affect", func(t *testing.T) {
		_, admin := fix.NewUser(t) // ignore first user as it will be superuser
		user, userClient := fix.NewUser(t)

		org, err := admin.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		project, err := admin.CreateProject(ctx, &adminv1.CreateProjectRequest{
			Org:        org.Organization.Name,
			Project:    "proj-group-inherit",
			ProdSlots:  1,
			SkipDeploy: true,
		})
		require.NoError(t, err)

		depl, err := admin.CreateDeployment(ctx, &adminv1.CreateDeploymentRequest{
			Org:         org.Organization.Name,
			Project:     project.Project.Name,
			Environment: "prod",
		})
		require.NoError(t, err)

		// directly update deployment to running with runtime info especially audience for correct jwt generation
		_, err = fix.Admin.DB.UpdateDeploymentUnsafe(ctx, depl.Deployment.Id, &database.UpdateDeploymentUnsafeOptions{
			RuntimeHost:       "rill-example-host",
			RuntimeInstanceID: "rill-example-instance",
			RuntimeAudience:   "http://example.org",
			Status:            database.DeploymentStatusRunning,
			StatusMessage:     "Running",
		})
		require.NoError(t, err)

		group, err := admin.CreateUsergroup(ctx, &adminv1.CreateUsergroupRequest{
			Org:  org.Organization.Name,
			Name: "group-inherit",
		})
		require.NoError(t, err)

		_, err = admin.AddOrganizationMemberUser(ctx, &adminv1.AddOrganizationMemberUserRequest{
			Org:   org.Organization.Name,
			Email: user.Email,
			Role:  database.OrganizationRoleNameViewer,
		})
		require.NoError(t, err)
		_, err = admin.AddUsergroupMemberUser(ctx, &adminv1.AddUsergroupMemberUserRequest{
			Org:       org.Organization.Name,
			Usergroup: group.Usergroup.GroupName,
			Email:     user.Email,
		})
		require.NoError(t, err)
		_, err = admin.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
			Email:   user.Email,
			Role:    database.ProjectRoleNameEditor,
		})
		require.NoError(t, err)
		// restrict resources through usergroup
		_, err = admin.AddProjectMemberUsergroup(ctx, &adminv1.AddProjectMemberUsergroupRequest{
			Org:       org.Organization.Name,
			Project:   project.Project.Name,
			Usergroup: group.Usergroup.GroupName,
			Role:      database.ProjectRoleNameEditor,
			Resources: []*adminv1.ResourceName{{Type: "metrics_view", Name: "mv1"}},
		})
		require.NoError(t, err)

		// restrict autogroup:members so that usergroup restriction takes effect
		restrict := true
		_, err = admin.SetProjectMemberUsergroupRole(ctx, &adminv1.SetProjectMemberUsergroupRoleRequest{
			Org:               org.Organization.Name,
			Project:           project.Project.Name,
			Usergroup:         database.UsergroupNameAutogroupMembers,
			RestrictResources: &restrict,
		})
		require.NoError(t, err)

		// just to check usergroup restriction applied
		ugResp, err := admin.ListUsergroupsForProjectAndUser(ctx, &adminv1.ListUsergroupsForProjectAndUserRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
			Email:   user.Email,
		})
		require.NoError(t, err)
		var memUg *adminv1.MemberUsergroup
		for _, g := range ugResp.Usergroups {
			if g.GroupName == group.Usergroup.GroupName {
				require.True(t, g.RestrictResources)
				require.Len(t, g.Resources, 1)
				require.Equal(t, &adminv1.ResourceName{Type: "metrics_view", Name: "mv1"}, g.Resources[0])
				memUg = g
				break
			}
		}
		require.NotNil(t, memUg)

		// initially user should have full access since user itself is not restricted
		proj, err := userClient.GetProject(ctx, &adminv1.GetProjectRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
		})
		require.NoError(t, err)
		// parse jwt
		require.NotNil(t, proj.Jwt)
		claims, err := fix.Audience.ParseAndValidate(proj.Jwt)
		require.NoError(t, err)
		require.NotNil(t, claims)
		rules := claims.Claims("").AdditionalRules
		require.Len(t, rules, 0) // since user has a full access role

		// Now also restrict the user directly
		_, err = admin.SetProjectMemberUserRole(ctx, &adminv1.SetProjectMemberUserRoleRequest{
			Org:       org.Organization.Name,
			Project:   project.Project.Name,
			Email:     user.Email,
			Resources: []*adminv1.ResourceName{{Type: "metrics_view", Name: "mv2"}},
		})
		require.NoError(t, err)

		// check both user restrictions are also in place
		memResp, err := admin.GetProjectMemberUser(ctx, &adminv1.GetProjectMemberUserRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
			Email:   user.Email,
		})
		require.NoError(t, err)
		require.True(t, memResp.Member.RestrictResources)
		require.True(t, memUg.RestrictResources)
		require.Len(t, memResp.Member.Resources, 1)
		require.Len(t, memUg.Resources, 1)
		require.Equal(t, &adminv1.ResourceName{Type: "metrics_view", Name: "mv2"}, memResp.Member.Resources[0])

		// now user should see both restrictions applied
		proj, err = userClient.GetProject(ctx, &adminv1.GetProjectRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
		})
		require.NoError(t, err)
		// parse jwt
		require.NotNil(t, proj.Jwt)
		claims, err = fix.Audience.ParseAndValidate(proj.Jwt)
		require.NoError(t, err)
		require.NotNil(t, claims)
		rules = claims.Claims("").AdditionalRules
		require.Len(t, rules, 2)
		resourceSet := map[string]bool{}
		for _, r := range rules {
			require.NotNil(t, r.GetTransitiveAccess())
			resourceSet[r.GetTransitiveAccess().Resource.Name] = true
		}
		require.Len(t, resourceSet, 2)
		require.Contains(t, resourceSet, "mv1")
		require.Contains(t, resourceSet, "mv2")

		// Now clear restrictions from user and usergroup
		restrictFlag := false
		_, err = admin.SetProjectMemberUserRole(ctx, &adminv1.SetProjectMemberUserRoleRequest{
			Org:               org.Organization.Name,
			Project:           project.Project.Name,
			Email:             user.Email,
			RestrictResources: &restrictFlag,
		})
		require.NoError(t, err)
		_, err = admin.SetProjectMemberUsergroupRole(ctx, &adminv1.SetProjectMemberUsergroupRoleRequest{
			Org:               org.Organization.Name,
			Project:           project.Project.Name,
			Usergroup:         group.Usergroup.GroupName,
			RestrictResources: &restrictFlag,
		})
		require.NoError(t, err)

		// Now user should have full access again and all resources cleared
		proj, err = userClient.GetProject(ctx, &adminv1.GetProjectRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
		})
		require.NoError(t, err)
		// parse jwt
		require.NotNil(t, proj.Jwt)
		claims, err = fix.Audience.ParseAndValidate(proj.Jwt)
		require.NoError(t, err)
		require.NotNil(t, claims)
		rules = claims.Claims("").AdditionalRules
		require.Len(t, rules, 0)

		// check role remains intact
		member, err := admin.GetProjectMemberUser(ctx, &adminv1.GetProjectMemberUserRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
			Email:   user.Email,
		})
		require.NoError(t, err)
		require.Equal(t, database.ProjectRoleNameEditor, member.Member.RoleName)
		require.False(t, member.Member.RestrictResources)
		require.Empty(t, member.Member.Resources)

		// check usergroup role remains intact
		groups, err := admin.ListUsergroupsForProjectAndUser(ctx, &adminv1.ListUsergroupsForProjectAndUserRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
			Email:   user.Email,
		})
		require.NoError(t, err)
		found := false
		for _, g := range groups.Usergroups {
			if g.GroupName == group.Usergroup.GroupName {
				require.Equal(t, database.ProjectRoleNameEditor, g.RoleName)
				require.False(t, g.RestrictResources)
				found = true
				break
			}
		}
		require.True(t, found)

		// now just restrict the user without any resources
		restrictFlag = true
		_, err = admin.SetProjectMemberUserRole(ctx, &adminv1.SetProjectMemberUserRoleRequest{
			Org:               org.Organization.Name,
			Project:           project.Project.Name,
			Email:             user.Email,
			RestrictResources: &restrictFlag,
		})
		require.NoError(t, err)

		proj, err = userClient.GetProject(ctx, &adminv1.GetProjectRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
		})
		require.NoError(t, err)
		// parse jwt
		require.NotNil(t, proj.Jwt)
		claims, err = fix.Audience.ParseAndValidate(proj.Jwt)
		require.NoError(t, err)
		require.NotNil(t, claims)
		rules = claims.Claims("").AdditionalRules
		require.Len(t, rules, 0) // since user is still part of usergroup with no resource restrictions

		// remove user from usergroup to see restriction take effect
		_, err = admin.RemoveUsergroupMemberUser(ctx, &adminv1.RemoveUsergroupMemberUserRequest{
			Org:       org.Organization.Name,
			Usergroup: group.Usergroup.GroupName,
			Email:     user.Email,
		})
		require.NoError(t, err)

		// deny all since user has restriction but no resources
		proj, err = userClient.GetProject(ctx, &adminv1.GetProjectRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
		})
		require.NoError(t, err)
		// parse jwt
		require.NotNil(t, proj.Jwt)
		claims, err = fix.Audience.ParseAndValidate(proj.Jwt)
		require.NoError(t, err)
		require.NotNil(t, claims)
		rules = claims.Claims("").AdditionalRules
		require.Len(t, rules, 1)
		require.NotNil(t, rules[0].GetAccess())
		require.Equal(t, false, rules[0].GetAccess().Allow)
	})

	t.Run("Project invites keep and update resource restrictions", func(t *testing.T) {
		_, admin := fix.NewUser(t)

		org, err := admin.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		project, err := admin.CreateProject(ctx, &adminv1.CreateProjectRequest{
			Org:        org.Organization.Name,
			Project:    "proj-invite-resources",
			ProdSlots:  1,
			SkipDeploy: true,
		})
		require.NoError(t, err)

		email := randomName() + "@example.com"
		restrict := true
		_, err = admin.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
			Org:               org.Organization.Name,
			Project:           project.Project.Name,
			Email:             email,
			Role:              database.ProjectRoleNameViewer,
			RestrictResources: &restrict,
		})
		require.NoError(t, err)

		invites, err := admin.ListProjectInvites(ctx, &adminv1.ListProjectInvitesRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
		})
		require.NoError(t, err)
		require.Len(t, invites.Invites, 1)
		require.Equal(t, database.ProjectRoleNameViewer, invites.Invites[0].RoleName)
		require.True(t, invites.Invites[0].RestrictResources)
		require.Empty(t, invites.Invites[0].Resources)

		// update role restriction should be intact
		_, err = admin.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
			Email:   email,
			Role:    database.ProjectRoleNameEditor,
		})
		require.NoError(t, err)

		invites, err = admin.ListProjectInvites(ctx, &adminv1.ListProjectInvitesRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
		})
		require.NoError(t, err)
		require.Len(t, invites.Invites, 1)
		require.Equal(t, database.ProjectRoleNameEditor, invites.Invites[0].RoleName)
		require.True(t, invites.Invites[0].RestrictResources)
		require.Empty(t, invites.Invites[0].Resources)

		// update to add resources
		_, err = admin.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
			Org:       org.Organization.Name,
			Project:   project.Project.Name,
			Email:     email,
			Role:      database.ProjectRoleNameEditor,
			Resources: []*adminv1.ResourceName{{Type: "metrics_view", Name: "mv_invite"}},
		})
		require.NoError(t, err)

		invites, err = admin.ListProjectInvites(ctx, &adminv1.ListProjectInvitesRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
		})
		require.NoError(t, err)
		require.Len(t, invites.Invites, 1)
		require.Equal(t, database.ProjectRoleNameEditor, invites.Invites[0].RoleName)
		require.True(t, invites.Invites[0].RestrictResources)
		require.Len(t, invites.Invites[0].Resources, 1)
		require.Equal(t, "metrics_view", invites.Invites[0].Resources[0].Type)
		require.Equal(t, "mv_invite", invites.Invites[0].Resources[0].Name)

		// Now clear restrictions
		restrictFlag := false
		_, err = admin.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
			Org:               org.Organization.Name,
			Project:           project.Project.Name,
			Email:             email,
			Role:              database.ProjectRoleNameEditor,
			RestrictResources: &restrictFlag,
		})
		require.NoError(t, err)

		invites, err = admin.ListProjectInvites(ctx, &adminv1.ListProjectInvitesRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
		})
		require.NoError(t, err)
		require.Len(t, invites.Invites, 1)
		require.Equal(t, database.ProjectRoleNameEditor, invites.Invites[0].RoleName)
		require.False(t, invites.Invites[0].RestrictResources)
		require.Empty(t, invites.Invites[0].Resources)

		// using SetProjectMemberUserRole should update the invite as well
		restrictFlag = true
		_, err = admin.SetProjectMemberUserRole(ctx, &adminv1.SetProjectMemberUserRoleRequest{
			Org:               org.Organization.Name,
			Project:           project.Project.Name,
			Email:             email,
			RestrictResources: &restrictFlag,
			// role omitted to keep same
		})
		require.NoError(t, err)
		invites, err = admin.ListProjectInvites(ctx, &adminv1.ListProjectInvitesRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
		})
		require.NoError(t, err)
		require.Len(t, invites.Invites, 1)
		require.Equal(t, database.ProjectRoleNameEditor, invites.Invites[0].RoleName)
		require.True(t, invites.Invites[0].RestrictResources)
		require.Empty(t, invites.Invites[0].Resources)

		// now update role and keep restrictions
		role := database.ProjectRoleNameViewer
		_, err = admin.SetProjectMemberUserRole(ctx, &adminv1.SetProjectMemberUserRoleRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
			Email:   email,
			Role:    &role,
			// restrict omitted to keep same
		})
		require.NoError(t, err)
		invites, err = admin.ListProjectInvites(ctx, &adminv1.ListProjectInvitesRequest{
			Org:     org.Organization.Name,
			Project: project.Project.Name,
		})
		require.NoError(t, err)
		require.Len(t, invites.Invites, 1)
		require.Equal(t, database.ProjectRoleNameViewer, invites.Invites[0].RoleName)
		require.True(t, invites.Invites[0].RestrictResources)
		require.Empty(t, invites.Invites[0].Resources)
	})
}

func randomName() string {
	id := make([]byte, 16)
	_, err := rand.Read(id)
	if err != nil {
		panic(err)
	}
	return "test_" + hex.EncodeToString(id)
}
