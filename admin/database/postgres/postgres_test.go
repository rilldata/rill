package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/pgtestcontainer"
	"github.com/stretchr/testify/require"
)

// TestPostgres starts Postgres using testcontainers and runs all other tests in
// this file as sub-tests (to prevent spawning many clusters).
func TestPostgres(t *testing.T) {
	ctx := context.Background()

	pg := pgtestcontainer.New(t)
	defer pg.Terminate(t)

	encKeyRing, err := database.NewRandomKeyring()
	require.NoError(t, err)
	conf, err := json.Marshal(encKeyRing)
	require.NoError(t, err)

	db, err := database.Open("postgres", pg.DatabaseURL, string(conf))
	require.NoError(t, err)
	require.NotNil(t, db)

	require.NoError(t, db.Migrate(ctx))

	t.Run("TestOrganizations", func(t *testing.T) { testOrganizations(t, db) })
	t.Run("TestOrgsWithPagination", func(t *testing.T) { testOrgsWithPagination(t, db) })
	t.Run("TestProjects", func(t *testing.T) { testProjects(t, db) })
	t.Run("TestProjectsWithAnnotations", func(t *testing.T) { testProjectsWithAnnotations(t, db) })
	t.Run("TestProjectsWithPagination", func(t *testing.T) { testProjectsWithPagination(t, db) })
	t.Run("TestProjectsForUsersWithPagination", func(t *testing.T) { testProjectsForUserWithPagination(t, db) })
	t.Run("TestMembersWithPagination", func(t *testing.T) { testOrgsMembersPagination(t, db) })
	t.Run("TestUpsertProjectVariable", func(t *testing.T) { testUpsertProjectVariable(t, db) })
	t.Run("TestManagedGitRepos", func(t *testing.T) { testManagedGitRepos(t, db) })
	// Add new tests here

	require.NoError(t, db.Close())
}

func testOrganizations(t *testing.T, db database.DB) {
	ctx := context.Background()

	org, err := db.FindOrganizationByName(ctx, "foo")
	require.ErrorIs(t, err, database.ErrNotFound)
	require.Nil(t, org)

	org, err = db.InsertOrganization(ctx, &database.InsertOrganizationOptions{
		Name:        "foo",
		Description: "hello world",
	})
	require.NoError(t, err)
	require.Equal(t, "foo", org.Name)
	require.Equal(t, "hello world", org.Description)
	require.Less(t, time.Since(org.CreatedOn), 10*time.Second)
	require.Less(t, time.Since(org.UpdatedOn), 10*time.Second)

	org, err = db.InsertOrganization(ctx, &database.InsertOrganizationOptions{
		Name:        "bar",
		Description: "",
	})
	require.NoError(t, err)
	require.Equal(t, "bar", org.Name)

	orgs, err := db.FindOrganizations(ctx, "", 1000)
	require.NoError(t, err)
	require.Equal(t, "bar", orgs[0].Name)
	require.Equal(t, "foo", orgs[1].Name)

	org, err = db.FindOrganizationByName(ctx, "foo")
	require.NoError(t, err)
	require.Equal(t, "foo", org.Name)
	require.Equal(t, "hello world", org.Description)

	org, err = db.UpdateOrganization(ctx, org.ID, &database.UpdateOrganizationOptions{
		Name:        org.Name,
		Description: "",
	})
	require.NoError(t, err)
	require.Equal(t, "foo", org.Name)
	require.Equal(t, "", org.Description)

	err = db.DeleteOrganization(ctx, org.Name)
	require.NoError(t, err)

	org, err = db.FindOrganizationByName(ctx, "foo")
	require.ErrorIs(t, err, database.ErrNotFound)
	require.Nil(t, org)
}

func testOrgsWithPagination(t *testing.T, db database.DB) {
	ctx := context.Background()

	user, err := db.InsertUser(ctx, &database.InsertUserOptions{Email: "test@rilldata.com"})
	require.NoError(t, err)
	require.Equal(t, "test@rilldata.com", user.Email)

	role, err := db.FindOrganizationRole(ctx, database.OrganizationRoleNameAdmin)
	require.NoError(t, err)

	// add org and give user permission
	org, err := db.InsertOrganization(ctx, &database.InsertOrganizationOptions{Name: "alpha"})
	require.NoError(t, err)
	require.Equal(t, "alpha", org.Name)
	_, err = db.InsertOrganizationMemberUser(ctx, org.ID, user.ID, role.ID, false)
	require.NoError(t, err)

	// add org and give user permission
	org, err = db.InsertOrganization(ctx, &database.InsertOrganizationOptions{Name: "beta"})
	require.NoError(t, err)
	require.Equal(t, "beta", org.Name)
	_, err = db.InsertOrganizationMemberUser(ctx, org.ID, user.ID, role.ID, false)
	require.NoError(t, err)

	// add org only
	org, err = db.InsertOrganization(ctx, &database.InsertOrganizationOptions{Name: "gamma"})
	require.NoError(t, err)
	require.Equal(t, "gamma", org.Name)

	// fetch org without name filter
	orgs, err := db.FindOrganizationsForUser(ctx, user.ID, "", 1)
	require.NoError(t, err)
	require.Equal(t, len(orgs), 1)
	require.Equal(t, "alpha", orgs[0].Name)

	// fetch org with name filter
	orgs, err = db.FindOrganizationsForUser(ctx, user.ID, orgs[0].Name, 10)
	require.NoError(t, err)
	require.Equal(t, len(orgs), 1)
	require.Equal(t, "beta", orgs[0].Name)

	//cleanup
	require.NoError(t, db.DeleteOrganization(ctx, "alpha"))
	require.NoError(t, db.DeleteOrganization(ctx, "beta"))
	require.NoError(t, db.DeleteOrganization(ctx, "gamma"))
	require.NoError(t, db.DeleteUser(ctx, user.ID))
}

func testProjects(t *testing.T, db database.DB) {
	ctx := context.Background()

	org, err := db.InsertOrganization(ctx, &database.InsertOrganizationOptions{Name: "foo"})
	require.NoError(t, err)
	require.Equal(t, "foo", org.Name)

	proj, err := db.FindProjectByName(ctx, org.Name, "bar")
	require.ErrorIs(t, err, database.ErrNotFound)
	require.Nil(t, proj)

	proj, err = db.InsertProject(ctx, &database.InsertProjectOptions{
		OrganizationID: org.ID,
		Name:           "bar",
		Description:    "hello world",
	})
	require.NoError(t, err)
	require.Equal(t, org.ID, proj.OrganizationID)
	require.Equal(t, "bar", proj.Name)
	require.Equal(t, "hello world", proj.Description)
	require.Less(t, time.Since(proj.CreatedOn), 10*time.Second)
	require.Less(t, time.Since(proj.UpdatedOn), 10*time.Second)

	proj, err = db.FindProjectByName(ctx, org.Name, proj.Name)
	require.NoError(t, err)
	require.Equal(t, org.ID, proj.OrganizationID)
	require.Equal(t, "bar", proj.Name)
	require.Equal(t, "hello world", proj.Description)

	proj.Description = ""
	proj, err = db.UpdateProject(ctx, proj.ID, &database.UpdateProjectOptions{
		Name:        proj.Name,
		Description: proj.Description,
	})
	require.NoError(t, err)
	require.Equal(t, org.ID, proj.OrganizationID)
	require.Equal(t, "bar", proj.Name)
	require.Equal(t, "", proj.Description)

	err = db.DeleteOrganization(ctx, org.Name)
	require.ErrorContains(t, err, "constraint")

	err = db.DeleteProject(ctx, proj.ID)
	require.NoError(t, err)

	proj, err = db.FindProjectByName(ctx, org.Name, "bar")
	require.ErrorIs(t, err, database.ErrNotFound)
	require.Nil(t, proj)

	err = db.DeleteOrganization(ctx, org.Name)
	require.NoError(t, err)

	org, err = db.FindOrganizationByName(ctx, "foo")
	require.ErrorIs(t, err, database.ErrNotFound)
	require.Nil(t, org)
}

func testProjectsWithAnnotations(t *testing.T, db database.DB) {
	ctx := context.Background()

	org, err := db.InsertOrganization(ctx, &database.InsertOrganizationOptions{Name: "foo"})
	require.NoError(t, err)
	require.Equal(t, "foo", org.Name)

	opts := &database.InsertProjectOptions{
		OrganizationID: org.ID,
		Name:           "bar",
	}
	proj, err := db.InsertProject(ctx, opts)
	require.NoError(t, err)
	require.Empty(t, proj.Annotations)

	annotations := map[string]string{"foo": "bar", "bar": "baz"}
	_, err = db.UpdateProject(ctx, proj.ID, &database.UpdateProjectOptions{
		Name:        proj.Name,
		Annotations: annotations,
	})
	require.NoError(t, err)

	proj, err = db.FindProjectByName(ctx, org.Name, proj.Name)
	require.NoError(t, err)
	require.Equal(t, "bar", proj.Name)
	require.Equal(t, annotations, proj.Annotations)

	projs, err := db.FindProjectPathsByPatternAndAnnotations(ctx, "%", "", []string{"foo"}, nil, 10)
	require.NoError(t, err)
	require.Equal(t, "foo/bar", projs[0])

	projs, err = db.FindProjectPathsByPatternAndAnnotations(ctx, "%", "", nil, map[string]string{"foo": "bar"}, 1)
	require.NoError(t, err)
	require.Equal(t, "foo/bar", projs[0])

	projs, err = db.FindProjectPathsByPatternAndAnnotations(ctx, "%", "", nil, map[string]string{"foo": ""}, 1)
	require.NoError(t, err)
	require.Len(t, projs, 0)

	err = db.DeleteProject(ctx, proj.ID)
	require.NoError(t, err)

	err = db.DeleteOrganization(ctx, org.Name)
	require.NoError(t, err)
}

func testProjectsWithPagination(t *testing.T, db database.DB) {
	ctx := context.Background()

	// add org
	org, err := db.InsertOrganization(ctx, &database.InsertOrganizationOptions{Name: "test"})
	require.NoError(t, err)
	require.Equal(t, "test", org.Name)

	// add another org
	org2, err := db.InsertOrganization(ctx, &database.InsertOrganizationOptions{Name: "test2"})
	require.NoError(t, err)
	require.Equal(t, "test2", org2.Name)

	// add projects
	proj, err := db.InsertProject(ctx, &database.InsertProjectOptions{OrganizationID: org.ID, Name: "alpha"})
	require.NoError(t, err)
	require.Equal(t, "alpha", proj.Name)

	proj1, err := db.InsertProject(ctx, &database.InsertProjectOptions{OrganizationID: org.ID, Name: "beta"})
	require.NoError(t, err)
	require.Equal(t, "beta", proj1.Name)

	proj2, err := db.InsertProject(ctx, &database.InsertProjectOptions{OrganizationID: org.ID, Name: "gamma"})
	require.NoError(t, err)
	require.Equal(t, "gamma", proj2.Name)

	proj3, err := db.InsertProject(ctx, &database.InsertProjectOptions{OrganizationID: org2.ID, Name: "other"})
	require.NoError(t, err)
	require.Equal(t, "other", proj3.Name)

	// fetch project name without name filter
	projs, err := db.FindProjectsForOrganization(ctx, org.ID, "", 2)
	require.NoError(t, err)
	require.Equal(t, len(projs), 2)
	require.Equal(t, "alpha", projs[0].Name)
	require.Equal(t, "beta", projs[1].Name)

	// fetch project with name filter
	projs, err = db.FindProjectsForOrganization(ctx, org.ID, projs[1].Name, 2)
	require.NoError(t, err)
	require.Equal(t, len(projs), 1)
	require.Equal(t, "gamma", projs[0].Name)

	//cleanup
	require.NoError(t, db.DeleteProject(ctx, proj.ID))
	require.NoError(t, db.DeleteProject(ctx, proj1.ID))
	require.NoError(t, db.DeleteProject(ctx, proj2.ID))
	require.NoError(t, db.DeleteProject(ctx, proj3.ID))
	require.NoError(t, db.DeleteOrganization(ctx, "test"))
	require.NoError(t, db.DeleteOrganization(ctx, "test2"))
}

func testProjectsForUserWithPagination(t *testing.T, db database.DB) {
	ctx := context.Background()

	// add user
	user, err := db.InsertUser(ctx, &database.InsertUserOptions{Email: "test@rilldata.com"})
	require.NoError(t, err)
	require.Equal(t, "test@rilldata.com", user.Email)

	// fetch role
	role, err := db.FindProjectRole(ctx, database.ProjectRoleNameEditor)
	require.NoError(t, err)

	// add org
	org, err := db.InsertOrganization(ctx, &database.InsertOrganizationOptions{Name: "test"})
	require.NoError(t, err)
	require.Equal(t, "test", org.Name)

	// add projects
	// public project
	proj, err := db.InsertProject(ctx, &database.InsertProjectOptions{OrganizationID: org.ID, Name: "alpha", Public: true})
	require.NoError(t, err)

	// user added as collaborator
	proj1, err := db.InsertProject(ctx, &database.InsertProjectOptions{OrganizationID: org.ID, Name: "beta"})
	require.NoError(t, err)
	require.Equal(t, "beta", proj1.Name)
	require.NoError(t, db.InsertProjectMemberUser(ctx, proj1.ID, user.ID, role.ID))

	// public project and user added as collaborator
	proj2, err := db.InsertProject(ctx, &database.InsertProjectOptions{OrganizationID: org.ID, Name: "gamma", Public: true})
	require.NoError(t, err)
	require.Equal(t, "gamma", proj2.Name)
	require.NoError(t, db.InsertProjectMemberUser(ctx, proj2.ID, user.ID, role.ID))

	// internal project
	proj3, err := db.InsertProject(ctx, &database.InsertProjectOptions{OrganizationID: org.ID, Name: "internal"})
	require.NoError(t, err)
	require.Equal(t, "internal", proj3.Name)

	// fetch projects without name filter
	projs, err := db.FindProjectsForOrgAndUser(ctx, org.ID, user.ID, true, "", 2)
	require.NoError(t, err)
	require.Equal(t, len(projs), 2)
	require.Equal(t, "alpha", projs[0].Name)
	require.Equal(t, "beta", projs[1].Name)

	// fetch project with name filter
	projs, err = db.FindProjectsForOrgAndUser(ctx, org.ID, user.ID, true, projs[1].Name, 2)
	require.NoError(t, err)
	require.Equal(t, len(projs), 1)
	require.Equal(t, "gamma", projs[0].Name)

	//cleanup
	require.NoError(t, db.DeleteProject(ctx, proj.ID))
	require.NoError(t, db.DeleteProject(ctx, proj1.ID))
	require.NoError(t, db.DeleteProject(ctx, proj2.ID))
	require.NoError(t, db.DeleteProject(ctx, proj3.ID))
	require.NoError(t, db.DeleteOrganization(ctx, "test"))
	require.NoError(t, db.DeleteUser(ctx, user.ID))
}

func testOrgsMembersPagination(t *testing.T, db database.DB) {
	ctx := context.Background()

	adminUser, err := db.InsertUser(ctx, &database.InsertUserOptions{Email: "test1@rilldata.com"})
	require.NoError(t, err)

	viewerUser, err := db.InsertUser(ctx, &database.InsertUserOptions{Email: "test2@rilldata.com"})
	require.NoError(t, err)

	admin, err := db.FindOrganizationRole(ctx, database.OrganizationRoleNameAdmin)
	require.NoError(t, err)
	viewer, err := db.FindOrganizationRole(ctx, database.OrganizationRoleNameViewer)
	require.NoError(t, err)

	// add org and give user permission
	org, err := db.InsertOrganization(ctx, &database.InsertOrganizationOptions{Name: "alpha"})
	require.NoError(t, err)
	_, err = db.InsertOrganizationMemberUser(ctx, org.ID, adminUser.ID, admin.ID, false)
	require.NoError(t, err)
	_, err = db.InsertOrganizationMemberUser(ctx, org.ID, viewerUser.ID, viewer.ID, false)
	require.NoError(t, err)
	require.NoError(t, db.InsertOrganizationInvite(ctx, &database.InsertOrganizationInviteOptions{Email: "test3@rilldata.com", InviterID: adminUser.ID, OrgID: org.ID, RoleID: viewer.ID}))

	// fetch members without name filter
	users, err := db.FindOrganizationMemberUsers(ctx, org.ID, "", true, "", 1)
	require.NoError(t, err)
	require.Equal(t, len(users), 1)
	require.Equal(t, "test1@rilldata.com", users[0].Email)

	// fetch members with name filter
	users, err = db.FindOrganizationMemberUsers(ctx, org.ID, "", true, users[0].Email, 1)
	require.NoError(t, err)
	require.Equal(t, len(users), 1)
	require.Equal(t, "test2@rilldata.com", users[0].Email)

	// fetch invites without name filter
	invites, err := db.FindOrganizationInvites(ctx, org.ID, "", 1)
	require.NoError(t, err)
	require.Equal(t, len(invites), 1)
	require.Equal(t, "test3@rilldata.com", invites[0].Email)

	invites, err = db.FindOrganizationInvites(ctx, org.ID, invites[0].Email, 1)
	require.NoError(t, err)
	require.Equal(t, len(invites), 0)

	//cleanup
	require.NoError(t, db.DeleteOrganization(ctx, "alpha"))
}

func testUpsertProjectVariable(t *testing.T, db database.DB) {
	_, projectID, userID := seed(t, db)

	ctx := context.Background()
	// create project variables
	vars, err := db.UpsertProjectVariable(ctx, projectID, "", map[string]string{"foo1": "bar1", "foo2": "bar2", "foo3": "bar3"}, userID)
	require.NoError(t, err)

	require.Equal(t, len(vars), 3)

	// update some variables
	vars, err = db.UpsertProjectVariable(ctx, projectID, "prod", map[string]string{"foo1": "baz1", "foo2": "baz2"}, userID)
	require.NoError(t, err)
	require.Equal(t, len(vars), 2)

	// update some dev variables
	vars, err = db.UpsertProjectVariable(ctx, projectID, "dev", map[string]string{"foo3": "bad3"}, userID)
	require.NoError(t, err)
	require.Equal(t, len(vars), 1)

	// find all variables
	vars, err = db.FindProjectVariables(ctx, projectID, nil)
	require.NoError(t, err)
	require.Equal(t, len(vars), 6)

	// find project variables
	env := "prod"
	vars, err = db.FindProjectVariables(ctx, projectID, &env)
	require.NoError(t, err)

	require.Equal(t, len(vars), 3)
	for _, v := range vars {
		switch v.Name {
		case "foo1":
			require.Equal(t, "baz1", string(v.Value))
		case "foo2":
			require.Equal(t, "baz2", string(v.Value))
		case "foo3":
			require.Equal(t, "bar3", string(v.Value))
		}
	}

	err = db.DeleteProjectVariables(ctx, projectID, "", []string{"foo1", "foo2", "foo3", "foo4"})
	require.NoError(t, err)

	// find project variables
	vars, err = db.FindProjectVariables(ctx, projectID, &env)
	require.NoError(t, err)
	require.Equal(t, len(vars), 2)

	err = db.DeleteProjectVariables(ctx, projectID, "prod", []string{"foo1", "foo2", "foo3", "foo4"})
	require.NoError(t, err)

	// find project variables
	vars, err = db.FindProjectVariables(ctx, projectID, &env)
	require.NoError(t, err)
	require.Equal(t, len(vars), 0)

	// cleanup
	require.NoError(t, db.DeleteProject(ctx, projectID))
	require.NoError(t, db.DeleteOrganization(ctx, "alpha"))
	require.NoError(t, db.DeleteUser(ctx, userID))
}

func testManagedGitRepos(t *testing.T, db database.DB) {
	// create a user with random email id
	user, err := db.InsertUser(context.Background(), &database.InsertUserOptions{Email: fmt.Sprintf("user%d@rilldata.com", time.Now().UnixNano())})
	require.NoError(t, err)

	// add some orgs
	org1, err := db.InsertOrganization(context.Background(), &database.InsertOrganizationOptions{
		Name:            "test-mgd-repo-1",
		CreatedByUserID: &user.ID,
	})
	require.NoError(t, err)

	org2, err := db.InsertOrganization(context.Background(), &database.InsertOrganizationOptions{
		Name:            "test-mgd-repo-2",
		CreatedByUserID: &user.ID,
	})
	require.NoError(t, err)

	org3, err := db.InsertOrganization(context.Background(), &database.InsertOrganizationOptions{
		Name:            "test-mgd-repo-3",
		CreatedByUserID: &user.ID,
	})
	require.NoError(t, err)

	// insert some repos
	m1, err := db.InsertManagedGitRepo(context.Background(), &database.InsertManagedGitRepoOptions{
		OrgID:   org1.ID,
		Remote:  "https://github.com/rilldata/rill.git",
		OwnerID: user.ID,
	})
	require.NoError(t, err)

	m2, err := db.InsertManagedGitRepo(context.Background(), &database.InsertManagedGitRepoOptions{
		OrgID:   org2.ID,
		Remote:  "https://github.com/rilldata/rill2.git",
		OwnerID: user.ID,
	})
	require.NoError(t, err)

	// there are no unused repos because just created
	mgdRepos, err := db.FindUnusedManagedGitRepos(context.Background(), 100)
	require.NoError(t, err)
	require.Equal(t, 0, len(mgdRepos))

	m3, err := db.InsertManagedGitRepo(context.Background(), &database.InsertManagedGitRepoOptions{
		OrgID:   org3.ID,
		Remote:  "https://github.com/rilldata/rill3.git",
		OwnerID: user.ID,
	})
	require.NoError(t, err)

	// create projects using the repos
	p1, err := db.InsertProject(context.Background(), &database.InsertProjectOptions{
		OrganizationID:   org1.ID,
		Name:             "test-mgd-repo-1",
		ManagedGitRepoID: &m1.ID,
	})
	require.NoError(t, err)

	p3, err := db.InsertProject(context.Background(), &database.InsertProjectOptions{
		OrganizationID:   org3.ID,
		Name:             "test-mgd-repo-3",
		ManagedGitRepoID: &m3.ID,
	})
	require.NoError(t, err)

	// verify 3 repos exist
	repos, err := db.CountManagedGitRepos(context.Background(), org1.ID)
	require.NoError(t, err)
	require.Equal(t, 1, repos)
	repos, err = db.CountManagedGitRepos(context.Background(), org2.ID)
	require.NoError(t, err)
	require.Equal(t, 1, repos)
	repos, err = db.CountManagedGitRepos(context.Background(), org3.ID)
	require.NoError(t, err)
	require.Equal(t, 1, repos)

	// delete org
	require.NoError(t, db.DeleteProject(context.Background(), p3.ID))
	require.NoError(t, db.DeleteOrganization(context.Background(), org3.Name))

	// the mgd repo still exists but org_id is set to null
	repo, err := db.FindManagedGitRepo(context.Background(), "https://github.com/rilldata/rill3.git")
	require.NoError(t, err)
	var res *string = nil
	require.Equal(t, repo.OrgID, res)

	// there are no unused repos because just created
	mgdRepos, err = db.FindUnusedManagedGitRepos(context.Background(), 100)
	require.NoError(t, err)
	require.Equal(t, 0, len(mgdRepos))

	// manually update updated_at to old date for managed repos
	_, err = db.(*connection).db.Exec("UPDATE managed_git_repos SET updated_on = NOW() - INTERVAL '10 DAY'")
	require.NoError(t, err)

	// now we should see 2 unused repos(m2 and m3)
	mgdRepos, err = db.FindUnusedManagedGitRepos(context.Background(), 100)
	require.NoError(t, err)
	require.Equal(t, 2, len(mgdRepos))
	var ids []string
	for _, repo := range mgdRepos {
		ids = append(ids, repo.ID)
	}
	require.NotContains(t, m1.ID, ids)

	// cleanup
	require.NoError(t, db.DeleteProject(context.Background(), p1.ID))
	require.NoError(t, db.DeleteOrganization(context.Background(), org1.Name))
	require.NoError(t, db.DeleteOrganization(context.Background(), org2.Name))
	require.NoError(t, db.DeleteUser(context.Background(), user.ID))
	require.NoError(t, db.DeleteManagedGitRepos(context.Background(), []string{m1.ID, m2.ID, m3.ID}))
}

func seed(t *testing.T, db database.DB) (orgID, projectID, userID string) {
	ctx := context.Background()

	// create a user with random email id
	adminUser, err := db.InsertUser(ctx, &database.InsertUserOptions{Email: fmt.Sprintf("user%d@rilldata.com", time.Now().UnixNano())})
	require.NoError(t, err)

	admin, err := db.FindOrganizationRole(ctx, database.OrganizationRoleNameAdmin)
	require.NoError(t, err)

	// add org and give user permission
	org, err := db.InsertOrganization(ctx, &database.InsertOrganizationOptions{Name: "alpha"})
	require.NoError(t, err)
	_, err = db.InsertOrganizationMemberUser(ctx, org.ID, adminUser.ID, admin.ID, false)
	require.NoError(t, err)

	proj, err := db.InsertProject(ctx, &database.InsertProjectOptions{OrganizationID: org.ID, Name: "alpha", Public: true})
	require.NoError(t, err)

	return org.ID, proj.ID, adminUser.ID
}
