package server_test

import (
	"testing"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/testadmin"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateProjectSlots(t *testing.T) {
	fix := testadmin.New(t)
	_, client := fix.NewUser(t)
	org, err := client.CreateOrganization(t.Context(), &adminv1.CreateOrganizationRequest{Name: randomName()})
	require.NoError(t, err)

	for _, tt := range []struct {
		name              string
		prod, dev         int64
		wantProd, wantDev int64
	}{
		{name: "defaults", wantProd: 2, wantDev: 2},
		{name: "explicit", prod: 1, dev: 3, wantProd: 1, wantDev: 3},
		{name: "default-prod", dev: 1, wantProd: 2, wantDev: 1},
		{name: "default-dev", prod: 3, wantProd: 3, wantDev: 2},
	} {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.CreateProject(t.Context(), &adminv1.CreateProjectRequest{
				Org: org.Organization.Name, Project: tt.name,
				ProdSlots: tt.prod, DevSlots: tt.dev, SkipDeploy: true,
			})
			require.NoError(t, err)
			require.Equal(t, tt.wantProd, resp.Project.ProdSlots)
			require.Equal(t, tt.wantDev, resp.Project.DevSlots)
		})
	}
}

func TestProjectVariables(t *testing.T) {
	fix := testadmin.New(t)

	t.Run("Set, get and unset variables as a user", func(t *testing.T) {
		u1, c1 := fix.NewUser(t)

		r1, err := c1.CreateOrganization(t.Context(), &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		r2, err := c1.CreateProject(t.Context(), &adminv1.CreateProjectRequest{
			Org:        r1.Organization.Name,
			Project:    "proj1",
			ProdSlots:  1,
			SkipDeploy: true,
		})
		require.NoError(t, err)

		// Set a shared variable, a prod-only variable, and a dev-only variable.
		_, err = c1.UpdateProjectVariables(t.Context(), &adminv1.UpdateProjectVariablesRequest{
			Org:       r1.Organization.Name,
			Project:   r2.Project.Name,
			Variables: map[string]string{"FOO": "shared"},
		})
		require.NoError(t, err)
		_, err = c1.UpdateProjectVariables(t.Context(), &adminv1.UpdateProjectVariablesRequest{
			Org:         r1.Organization.Name,
			Project:     r2.Project.Name,
			Environment: "prod",
			Variables:   map[string]string{"BAR": "prod-value"},
		})
		require.NoError(t, err)
		_, err = c1.UpdateProjectVariables(t.Context(), &adminv1.UpdateProjectVariablesRequest{
			Org:         r1.Organization.Name,
			Project:     r2.Project.Name,
			Environment: "dev",
			Variables:   map[string]string{"BAZ": "dev-value"},
		})
		require.NoError(t, err)

		// Get shared variables only.
		r3, err := c1.GetProjectVariables(t.Context(), &adminv1.GetProjectVariablesRequest{
			Org:     r1.Organization.Name,
			Project: r2.Project.Name,
		})
		require.NoError(t, err)
		require.Len(t, r3.Variables, 1)
		require.Equal(t, "FOO", r3.Variables[0].Name)
		require.Equal(t, "shared", r3.Variables[0].Value)
		require.Equal(t, "", r3.Variables[0].Environment)
		require.Equal(t, u1.ID, r3.Variables[0].UpdatedByUserId)

		// Get shared and prod variables.
		r4, err := c1.GetProjectVariables(t.Context(), &adminv1.GetProjectVariablesRequest{
			Org:         r1.Organization.Name,
			Project:     r2.Project.Name,
			Environment: "prod",
		})
		require.NoError(t, err)
		require.Len(t, r4.Variables, 2)
		require.Equal(t, "FOO", r4.Variables[0].Name)
		require.Equal(t, "shared", r4.Variables[0].Value)
		require.Equal(t, "BAR", r4.Variables[1].Name)
		require.Equal(t, "prod-value", r4.Variables[1].Value)

		// Get all variables across environments.
		r5, err := c1.GetProjectVariables(t.Context(), &adminv1.GetProjectVariablesRequest{
			Org:                r1.Organization.Name,
			Project:            r2.Project.Name,
			ForAllEnvironments: true,
		})
		require.NoError(t, err)
		require.Len(t, r5.Variables, 3)
		require.Equal(t, "FOO", r5.Variables[0].Name)
		require.Equal(t, "BAR", r5.Variables[1].Name)
		require.Equal(t, "BAZ", r5.Variables[2].Name)

		// Unset a variable.
		_, err = c1.UpdateProjectVariables(t.Context(), &adminv1.UpdateProjectVariablesRequest{
			Org:            r1.Organization.Name,
			Project:        r2.Project.Name,
			UnsetVariables: []string{"FOO"},
		})
		require.NoError(t, err)

		// Check that the variable is gone.
		r6, err := c1.GetProjectVariables(t.Context(), &adminv1.GetProjectVariablesRequest{
			Org:                r1.Organization.Name,
			Project:            r2.Project.Name,
			ForAllEnvironments: true,
		})
		require.NoError(t, err)
		require.Len(t, r6.Variables, 2)
		require.Equal(t, "BAR", r6.Variables[0].Name)
		require.Equal(t, "BAZ", r6.Variables[1].Name)
	})

	t.Run("Set and get variables using a service token", func(t *testing.T) {
		_, c1 := fix.NewUser(t)

		// Create org and project
		r1, err := c1.CreateOrganization(t.Context(), &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		r2, err := c1.CreateProject(t.Context(), &adminv1.CreateProjectRequest{
			Org:        r1.Organization.Name,
			Project:    "proj1",
			ProdSlots:  1,
			SkipDeploy: true,
		})
		require.NoError(t, err)

		// Create a service with project admin role and issue an auth token for it.
		r3, err := c1.CreateService(t.Context(), &adminv1.CreateServiceRequest{
			Name:            "service1",
			Org:             r1.Organization.Name,
			Project:         r2.Project.Name,
			ProjectRoleName: database.ProjectRoleNameAdmin,
		})
		require.NoError(t, err)
		r4, err := c1.IssueServiceAuthToken(t.Context(), &adminv1.IssueServiceAuthTokenRequest{
			Org:         r1.Organization.Name,
			ServiceName: r3.Service.Name,
		})
		require.NoError(t, err)
		c2 := fix.NewClient(t, r4.Token)

		// Set a variable using the service token.
		_, err = c2.UpdateProjectVariables(t.Context(), &adminv1.UpdateProjectVariablesRequest{
			Org:       r1.Organization.Name,
			Project:   r2.Project.Name,
			Variables: map[string]string{"FOO": "bar"},
		})
		require.NoError(t, err)

		// Read it back using the service token.
		r5, err := c2.GetProjectVariables(t.Context(), &adminv1.GetProjectVariablesRequest{
			Org:     r1.Organization.Name,
			Project: r2.Project.Name,
		})
		require.NoError(t, err)
		require.Len(t, r5.Variables, 1)
		require.Equal(t, "FOO", r5.Variables[0].Name)
		require.Equal(t, "bar", r5.Variables[0].Value)
		require.Equal(t, "", r5.Variables[0].UpdatedByUserId)
	})
}

func TestProjectProvisioner(t *testing.T) {
	fix := testadmin.New(t)
	_, sc := fix.NewSuperuser(t)
	_, uc := fix.NewUser(t)

	// The user creates an org, which makes them an org admin and lets them create projects.
	org, err := uc.CreateOrganization(t.Context(), &adminv1.CreateOrganizationRequest{Name: randomName()})
	require.NoError(t, err)

	t.Run("non-superuser cannot set the provisioner", func(t *testing.T) {
		// "static" is the provisioner configured by testadmin.
		_, err := uc.CreateProject(t.Context(), &adminv1.CreateProjectRequest{
			Org:         org.Organization.Name,
			Project:     "with-provisioner",
			Provisioner: "static",
			SkipDeploy:  true,
		})
		require.Error(t, err)
		require.Equal(t, codes.PermissionDenied, status.Code(err))

		// Creating without a provisioner still works.
		proj, err := uc.CreateProject(t.Context(), &adminv1.CreateProjectRequest{
			Org:        org.Organization.Name,
			Project:    "proj1",
			SkipDeploy: true,
		})
		require.NoError(t, err)
		require.Empty(t, proj.Project.Provisioner)

		// The project admin cannot set it through UpdateProject either.
		provisioner := "static"
		_, err = uc.UpdateProject(t.Context(), &adminv1.UpdateProjectRequest{
			Org:         org.Organization.Name,
			Project:     "proj1",
			Provisioner: &provisioner,
		})
		require.Error(t, err)
		require.Equal(t, codes.PermissionDenied, status.Code(err))

		// An ordinary update by the project admin still works and does not touch the provisioner.
		desc := "updated description"
		res, err := uc.UpdateProject(t.Context(), &adminv1.UpdateProjectRequest{
			Org:         org.Organization.Name,
			Project:     "proj1",
			Description: &desc,
		})
		require.NoError(t, err)
		require.Equal(t, desc, res.Project.Description)
		require.Empty(t, res.Project.Provisioner)
	})

	t.Run("superuser can set the provisioner", func(t *testing.T) {
		provisioner := "static"
		res, err := sc.UpdateProject(t.Context(), &adminv1.UpdateProjectRequest{
			Org:                  org.Organization.Name,
			Project:              "proj1",
			Provisioner:          &provisioner,
			SuperuserForceAccess: true,
		})
		require.NoError(t, err)
		require.Equal(t, "static", res.Project.Provisioner)

		// An ordinary update by the project admin does not clear it.
		desc := "another description"
		res, err = uc.UpdateProject(t.Context(), &adminv1.UpdateProjectRequest{
			Org:         org.Organization.Name,
			Project:     "proj1",
			Description: &desc,
		})
		require.NoError(t, err)
		require.Equal(t, "static", res.Project.Provisioner)

		// An unknown provisioner is rejected even for superusers.
		unknown := "nonexistent"
		_, err = sc.UpdateProject(t.Context(), &adminv1.UpdateProjectRequest{
			Org:                  org.Organization.Name,
			Project:              "proj1",
			Provisioner:          &unknown,
			SuperuserForceAccess: true,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), `provisioner "nonexistent" is not configured`)

		// The superuser can clear it again.
		empty := ""
		res, err = sc.UpdateProject(t.Context(), &adminv1.UpdateProjectRequest{
			Org:                  org.Organization.Name,
			Project:              "proj1",
			Provisioner:          &empty,
			SuperuserForceAccess: true,
		})
		require.NoError(t, err)
		require.Empty(t, res.Project.Provisioner)

		// A superuser can also set it at creation time (in an org they can create projects in).
		suOrg, err := sc.CreateOrganization(t.Context(), &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		proj, err := sc.CreateProject(t.Context(), &adminv1.CreateProjectRequest{
			Org:         suOrg.Organization.Name,
			Project:     "proj2",
			Provisioner: "static",
			SkipDeploy:  true,
		})
		require.NoError(t, err)
		require.Equal(t, "static", proj.Project.Provisioner)
	})
}

func TestProjectTags(t *testing.T) {
	fix := testadmin.New(t)
	_, su := fix.NewSuperuser(t)

	// The user creates an org, which makes them an org admin.
	_, admin := fix.NewUser(t)
	org, err := admin.CreateOrganization(t.Context(), &adminv1.CreateOrganizationRequest{Name: randomName()})
	require.NoError(t, err)
	orgName := org.Organization.Name

	t.Run("create with tags", func(t *testing.T) {
		res, err := admin.CreateProject(t.Context(), &adminv1.CreateProjectRequest{
			Org:        orgName,
			Project:    "created-with-tags",
			Tags:       []string{"Sandbox", "team-a", "sandbox"},
			SkipDeploy: true,
		})
		require.NoError(t, err)
		require.Equal(t, []string{"sandbox", "team-a"}, res.Project.Tags)

		get, err := admin.GetProject(t.Context(), &adminv1.GetProjectRequest{Org: orgName, Project: "created-with-tags"})
		require.NoError(t, err)
		require.Equal(t, []string{"sandbox", "team-a"}, get.Project.Tags)

		// Invalid tags are rejected before the project is created.
		_, err = admin.CreateProject(t.Context(), &adminv1.CreateProjectRequest{
			Org:        orgName,
			Project:    "invalid-tags",
			Tags:       []string{"data science"},
			SkipDeploy: true,
		})
		require.Equal(t, codes.InvalidArgument, status.Code(err))
		_, err = admin.GetProject(t.Context(), &adminv1.GetProjectRequest{Org: orgName, Project: "invalid-tags"})
		require.Equal(t, codes.NotFound, status.Code(err))
	})

	t.Run("add and remove tags", func(t *testing.T) {
		_, err := admin.CreateProject(t.Context(), &adminv1.CreateProjectRequest{
			Org:        orgName,
			Project:    "edited-tags",
			Tags:       []string{"sandbox", "team-a"},
			SkipDeploy: true,
		})
		require.NoError(t, err)
		update := func(req *adminv1.UpdateProjectRequest) (*adminv1.Project, error) {
			req.Org = orgName
			req.Project = "edited-tags"
			res, err := admin.UpdateProject(t.Context(), req)
			if err != nil {
				return nil, err
			}
			return res.Project, nil
		}

		// Adding and removing in one request keeps the other tags.
		proj, err := update(&adminv1.UpdateProjectRequest{AddTags: []string{"Stage"}, RemoveTags: []string{"SANDBOX"}})
		require.NoError(t, err)
		require.Equal(t, []string{"stage", "team-a"}, proj.Tags)

		// Adding an existing tag and removing a missing one are no-ops.
		proj, err = update(&adminv1.UpdateProjectRequest{AddTags: []string{"stage"}, RemoveTags: []string{"missing"}})
		require.NoError(t, err)
		require.Equal(t, []string{"stage", "team-a"}, proj.Tags)

		// Tags can be updated together with other fields.
		desc := "Revenue review"
		proj, err = update(&adminv1.UpdateProjectRequest{Description: &desc, RemoveTags: []string{"team-a"}})
		require.NoError(t, err)
		require.Equal(t, desc, proj.Description)
		require.Equal(t, []string{"stage"}, proj.Tags)

		// A tag can't be both added and removed.
		_, err = update(&adminv1.UpdateProjectRequest{AddTags: []string{"prod"}, RemoveTags: []string{"Prod"}})
		require.Equal(t, codes.InvalidArgument, status.Code(err))

		// The limit applies to the tags after the update, and a rejected update changes nothing.
		_, err = update(&adminv1.UpdateProjectRequest{Description: new("ignored"), AddTags: []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}})
		require.Equal(t, codes.InvalidArgument, status.Code(err))
		require.ErrorContains(t, err, "at most 10 tags")
		proj, err = update(&adminv1.UpdateProjectRequest{AddTags: []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"}})
		require.NoError(t, err)
		require.Len(t, proj.Tags, 10)
		require.Equal(t, desc, proj.Description)
	})

	t.Run("viewers can read but not edit tags, and tags don't touch annotations", func(t *testing.T) {
		_, err := admin.CreateProject(t.Context(), &adminv1.CreateProjectRequest{
			Org:        orgName,
			Project:    "viewer-tags",
			Tags:       []string{"prod"},
			SkipDeploy: true,
		})
		require.NoError(t, err)
		_, err = su.SudoUpdateAnnotations(t.Context(), &adminv1.SudoUpdateAnnotationsRequest{
			Org:         orgName,
			Project:     "viewer-tags",
			Annotations: map[string]string{"internal": "true"},
		})
		require.NoError(t, err)

		viewerUser, viewer := fix.NewUser(t)
		_, err = admin.AddProjectMemberUser(t.Context(), &adminv1.AddProjectMemberUserRequest{
			Org:     orgName,
			Project: "viewer-tags",
			Email:   viewerUser.Email,
			Role:    database.ProjectRoleNameViewer,
		})
		require.NoError(t, err)

		get, err := viewer.GetProject(t.Context(), &adminv1.GetProjectRequest{Org: orgName, Project: "viewer-tags"})
		require.NoError(t, err)
		require.Equal(t, []string{"prod"}, get.Project.Tags)
		_, err = viewer.UpdateProject(t.Context(), &adminv1.UpdateProjectRequest{Org: orgName, Project: "viewer-tags", AddTags: []string{"stage"}})
		require.Equal(t, codes.PermissionDenied, status.Code(err))

		res, err := admin.UpdateProject(t.Context(), &adminv1.UpdateProjectRequest{Org: orgName, Project: "viewer-tags", AddTags: []string{"stage"}})
		require.NoError(t, err)
		require.Equal(t, []string{"prod", "stage"}, res.Project.Tags)
		get, err = su.GetProject(t.Context(), &adminv1.GetProjectRequest{Org: orgName, Project: "viewer-tags", SuperuserForceAccess: true})
		require.NoError(t, err)
		require.Equal(t, map[string]string{"internal": "true"}, get.Project.Annotations)
		require.Equal(t, []string{"prod", "stage"}, get.Project.Tags)
	})

	t.Run("list projects by tag", func(t *testing.T) {
		// A separate org, so the listed projects are known.
		listOrg, err := admin.CreateOrganization(t.Context(), &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		listOrgName := listOrg.Organization.Name
		for _, p := range []*adminv1.CreateProjectRequest{
			{Project: "prod", Tags: []string{"prod"}, Public: true},
			{Project: "stage", Tags: []string{"stage"}},
			{Project: "sandbox", Tags: []string{"sandbox"}},
			{Project: "untagged"},
		} {
			p.Org = listOrgName
			p.SkipDeploy = true
			_, err := admin.CreateProject(t.Context(), p)
			require.NoError(t, err)
		}

		// A project member who isn't an org member, and a user who isn't a member of anything.
		memberUser, member := fix.NewUser(t)
		_, err = admin.AddProjectMemberUser(t.Context(), &adminv1.AddProjectMemberUserRequest{
			Org:     listOrgName,
			Project: "sandbox",
			Email:   memberUser.Email,
			Role:    database.ProjectRoleNameViewer,
		})
		require.NoError(t, err)
		_, outsider := fix.NewUser(t)

		list := func(c *client.Client, tags ...string) ([]string, error) {
			res, err := c.ListProjectsForOrganization(t.Context(), &adminv1.ListProjectsForOrganizationRequest{Org: listOrgName, Tags: tags})
			if err != nil {
				return nil, err
			}
			names := make([]string, len(res.Projects))
			for i, p := range res.Projects {
				names[i] = p.Name
			}
			return names, nil
		}

		names, err := list(admin)
		require.NoError(t, err)
		require.Equal(t, []string{"prod", "sandbox", "stage", "untagged"}, names)
		names, err = list(admin, "Stage", "sandbox")
		require.NoError(t, err)
		require.Equal(t, []string{"sandbox", "stage"}, names)

		names, err = list(member)
		require.NoError(t, err)
		require.Equal(t, []string{"prod", "sandbox"}, names)
		names, err = list(member, "sandbox", "stage")
		require.NoError(t, err)
		require.Equal(t, []string{"sandbox"}, names)
		names, err = list(member, "stage")
		require.NoError(t, err)
		require.Empty(t, names)

		names, err = list(outsider, "prod")
		require.NoError(t, err)
		require.Equal(t, []string{"prod"}, names)
		names, err = list(outsider, "sandbox")
		require.NoError(t, err)
		require.Empty(t, names)

		_, err = list(admin, "not valid")
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})
}
