package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rilldata/rill/admin/database"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestPostgres starts Postgres using testcontainers and runs all other tests in
// this file as sub-tests (to prevent spawning many clusters).
func TestPostgres(t *testing.T) {
	ctx := context.Background()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:14",
			ExposedPorts: []string{"5432/tcp"},
			WaitingFor:   wait.ForListeningPort("5432/tcp"),
			Env: map[string]string{
				"POSTGRES_USER":     "postgres",
				"POSTGRES_PASSWORD": "postgres",
				"POSTGRES_DB":       "postgres",
			},
		},
	})
	require.NoError(t, err)
	defer container.Terminate(ctx)

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)
	databaseURL := fmt.Sprintf("postgres://postgres:postgres@%s:%d/postgres", host, port.Int())

	db, err := database.Open("postgres", databaseURL)
	require.NoError(t, err)
	require.NotNil(t, db)

	require.NoError(t, db.Migrate(ctx))

	t.Run("TestOrganizations", func(t *testing.T) { testOrganizations(t, db) })
	t.Run("TestProjects", func(t *testing.T) { testProjects(t, db) })
	// Add new tests here
	t.Run("TestProjectsWithVariables", func(t *testing.T) { testProjectsWithVariables(t, db) })

	t.Run("TestOrgsWithPagination", func(t *testing.T) { testOrgsWithPagination(t, db) })
	t.Run("TestProjectsWithPagination", func(t *testing.T) { testProjectsWithPagination(t, db) })
	t.Run("TestProjectsForUsersWithPagination", func(t *testing.T) { testProjectsForUserWithPagination(t, db) })
	t.Run("TestMembersWithPagination", func(t *testing.T) { testOrgsMembersPagination(t, db) })

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

	orgs, err := db.FindOrganizations(ctx, &database.PaginationOptions{PageSize: 1000})
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

func testProjectsWithVariables(t *testing.T, db database.DB) {
	ctx := context.Background()

	org, err := db.InsertOrganization(ctx, &database.InsertOrganizationOptions{Name: "foo"})
	require.NoError(t, err)
	require.Equal(t, "foo", org.Name)

	opts := &database.InsertProjectOptions{
		OrganizationID: org.ID,
		Name:           "bar",
		Description:    "hello world",
		ProdVariables:  map[string]string{"hello": "world"},
	}
	proj, err := db.InsertProject(ctx, opts)
	require.NoError(t, err)
	require.Equal(t, database.Variables(opts.ProdVariables), proj.ProdVariables)

	proj, err = db.FindProjectByName(ctx, org.Name, proj.Name)
	require.NoError(t, err)
	require.Equal(t, database.Variables(opts.ProdVariables), proj.ProdVariables)
}

func testOrgsWithPagination(t *testing.T, db database.DB) {
	ctx := context.Background()

	user, err := db.InsertUser(ctx, &database.InsertUserOptions{Email: "test@rilldata.com"})
	require.NoError(t, err)
	require.Equal(t, "test@rilldata.com", user.Email)

	role, err := db.FindOrganizationRole(ctx, database.OrganizationRoleNameAdmin)

	// add org and give user permission
	org, err := db.InsertOrganization(ctx, &database.InsertOrganizationOptions{Name: "alpha"})
	require.NoError(t, err)
	require.Equal(t, "alpha", org.Name)
	require.NoError(t, db.InsertOrganizationMemberUser(ctx, org.ID, user.ID, role.ID))

	// add org and give user permission
	org, err = db.InsertOrganization(ctx, &database.InsertOrganizationOptions{Name: "beta"})
	require.NoError(t, err)
	require.Equal(t, "beta", org.Name)
	require.NoError(t, db.InsertOrganizationMemberUser(ctx, org.ID, user.ID, role.ID))

	// add org only
	org, err = db.InsertOrganization(ctx, &database.InsertOrganizationOptions{Name: "gamma"})
	require.NoError(t, err)
	require.Equal(t, "gamma", org.Name)

	// fetch org without name filter
	orgs, err := db.FindOrganizationsForUser(ctx, user.ID, &database.PaginationOptions{PageSize: 1})
	require.NoError(t, err)
	require.Equal(t, len(orgs), 1)
	require.Equal(t, "alpha", orgs[0].Name)

	// fetch org with name filter
	orgs, err = db.FindOrganizationsForUser(ctx, user.ID, &database.PaginationOptions{PageSize: 10, Cursor: orgs[0].Name})
	require.NoError(t, err)
	require.Equal(t, len(orgs), 1)
	require.Equal(t, "beta", orgs[0].Name)

	//cleanup
	require.NoError(t, db.DeleteOrganization(ctx, "alpha"))
	require.NoError(t, db.DeleteOrganization(ctx, "beta"))
	require.NoError(t, db.DeleteOrganization(ctx, "gamma"))
	require.NoError(t, db.DeleteUser(ctx, user.ID))
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
	projs, err := db.FindProjectsForOrganization(ctx, org.ID, &database.PaginationOptions{PageSize: 2})
	require.NoError(t, err)
	require.Equal(t, len(projs), 2)
	require.Equal(t, "alpha", projs[0].Name)
	require.Equal(t, "beta", projs[1].Name)

	// fetch project with name filter
	projs, err = db.FindProjectsForOrganization(ctx, org.ID, &database.PaginationOptions{PageSize: 2, Cursor: projs[1].Name})
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
	role, err := db.FindProjectRole(ctx, database.ProjectRoleNameCollaborator)

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
	projs, err := db.FindProjectsForOrgAndUser(ctx, org.ID, user.ID, &database.PaginationOptions{PageSize: 2})
	require.NoError(t, err)
	require.Equal(t, len(projs), 2)
	require.Equal(t, "alpha", projs[0].Name)
	require.Equal(t, "beta", projs[1].Name)

	// fetch project with name filter
	projs, err = db.FindProjectsForOrgAndUser(ctx, org.ID, user.ID, &database.PaginationOptions{PageSize: 2, Cursor: projs[1].Name})
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
	viewer, err := db.FindOrganizationRole(ctx, database.OrganizationRoleNameViewer)

	// add org and give user permission
	org, err := db.InsertOrganization(ctx, &database.InsertOrganizationOptions{Name: "alpha"})
	require.NoError(t, err)
	require.NoError(t, db.InsertOrganizationMemberUser(ctx, org.ID, adminUser.ID, admin.ID))
	require.NoError(t, db.InsertOrganizationMemberUser(ctx, org.ID, viewerUser.ID, viewer.ID))
	require.NoError(t, db.InsertOrganizationInvite(ctx, "test3@rilldata.com", org.ID, viewer.ID, adminUser.ID))

	// fetch members without name filter
	users, err := db.FindOrganizationMemberUsers(ctx, org.ID, &database.PaginationOptions{PageSize: 1})
	require.NoError(t, err)
	require.Equal(t, len(users), 1)
	require.Equal(t, "test1@rilldata.com", users[0].Email)

	// fetch members with name filter
	users, err = db.FindOrganizationMemberUsers(ctx, org.ID, &database.PaginationOptions{PageSize: 1, Cursor: users[0].Email})
	require.NoError(t, err)
	require.Equal(t, len(users), 1)
	require.Equal(t, "test2@rilldata.com", users[0].Email)

	// fetch invites without name filter
	invites, err := db.FindOrganizationInvites(ctx, org.ID, &database.PaginationOptions{PageSize: 1})
	require.NoError(t, err)
	require.Equal(t, len(invites), 1)
	require.Equal(t, "test3@rilldata.com", invites[0].Email)

	invites, err = db.FindOrganizationInvites(ctx, org.ID, &database.PaginationOptions{PageSize: 1, Cursor: invites[0].Email})
	require.NoError(t, err)
	require.Equal(t, len(invites), 0)

	//cleanup
	require.NoError(t, db.DeleteOrganization(ctx, "alpha"))
}
