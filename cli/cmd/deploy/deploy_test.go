package deploy_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/go-github/v71/github"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/testadmin"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/cli/testcli"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/stretchr/testify/require"
)

func TestManagedDeploy(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	err := godotenv.Load("../../../.env")
	require.NoError(t, err)

	adm := testadmin.New(t)

	_, c := adm.NewUser(t)
	u1 := testcli.New(t, adm, c.Token)

	result := u1.Run(t, "org", "create", "github-test")
	require.Equal(t, 0, result.ExitCode)

	// deploy the project
	tempDir := t.TempDir()
	os.WriteFile(filepath.Join(tempDir, "rill.yaml"), []byte(`compiler: rillv1
display_name: Untitled Rill Project
olap_connector: duckdb`), 0644)

	result = u1.Run(t, "project", "deploy", "--interactive=false", "--org=github-test", "--project=rill-mgd-deploy", "--skip-deploy=true", "--path="+tempDir)
	require.Equal(t, 0, result.ExitCode, result.Output)

	// verify the project is correctly created
	resp, err := c.GetProject(t.Context(), &adminv1.GetProjectRequest{
		OrganizationName: "github-test",
		Name:             "rill-mgd-deploy",
	})
	require.NoError(t, err)
	require.Equal(t, "rill-mgd-deploy", resp.Project.Name)

	// get a github client
	installationID, err := adm.Admin.Github.ManagedOrgInstallationID()
	require.NoError(t, err)
	ghClient := adm.Admin.Github.InstallationClient(installationID, nil)

	// cleanup repo
	t.Cleanup(func() {
		owner, repo, ok := gitutil.SplitGithubRemote(resp.Project.GitRemote)
		require.True(t, ok, "invalid github remote: %s", resp.Project.GitRemote)
		_, err = ghClient.Repositories.Delete(context.Background(), owner, repo)
		require.NoError(t, err, "failed to delete github repo %s/%s: %w", owner, repo, err)
	})

	// redeploy the same project with changes
	err = os.Mkdir(filepath.Join(tempDir, "models"), 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tempDir, "models/model.sql"), []byte(`SELECT 1 AS one`), 0644)
	require.NoError(t, err)
	result = u1.Run(t, "deploy", "--interactive=false", "--org=github-test", "--project=rill-mgd-deploy", "--skip-deploy=true", "--path="+tempDir)
	require.Equal(t, 0, result.ExitCode, result.Output)

	// verify changes are pushed to Github repo
	verifyGithubRepoContents(t, ghClient, resp.Project.GitRemote)
}

// This test require a Github personal access token and refresh token to work
// Those can be generated using command `rill devtool gh-token` command
// The refresh token can only be used once to generate a new personal access token and refresh token so the command should be executed everytime before running this test
func TestGithubDeploy(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	err := godotenv.Load("../../../.env")
	require.NoError(t, err)

	err = godotenv.Load("../../../.github_env")
	require.NoError(t, err)

	personalAccessToken := os.Getenv("GH_ACCESS_TOKEN")
	if personalAccessToken == "" {
		t.Fatal("Personal access token not found. Run `rill devtool gh-token` to generate one")
	}
	refreshToken := os.Getenv("GH_REFRESH_TOKEN")
	if refreshToken == "" {
		t.Fatal("Refresh token not found. Run `rill devtool gh-token` to generate one")
	}

	// remove .github_env since refresh token can only be used once
	err = os.Remove("../../../.github_env")
	require.NoError(t, err)

	// github client
	ghClient := github.NewTokenClient(t.Context(), personalAccessToken)
	ghUser, _, err := ghClient.Users.Get(t.Context(), "")
	require.NoError(t, err)

	// test service
	adm := testadmin.New(t)
	user, c := adm.NewUser(t)
	adm.Admin.DB.UpdateUser(t.Context(), user.ID, &database.UpdateUserOptions{
		DisplayName:        user.DisplayName,
		PhotoURL:           user.PhotoURL,
		GithubUsername:     *ghUser.Login,
		GithubRefreshToken: refreshToken,
	})
	u1 := testcli.New(t, adm, c.Token)

	t.Run("self-hosted git deploy", func(t *testing.T) {
		testSelfHostedDeploy(t, c, ghClient, u1)
	})
}
func testSelfHostedDeploy(t *testing.T, adminClient *client.Client, ghClient *github.Client, adm *testcli.Fixture) {
	result := adm.Run(t, "org", "create", "github-test")
	require.Equal(t, 0, result.ExitCode)

	// create a rill project
	tempDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tempDir, "rill.yaml"), []byte(`compiler: rillv1
display_name: Untitled Rill Project
olap_connector: duckdb`), 0644)
	require.NoError(t, err)

	// create a github repo
	repo, _, err := ghClient.Repositories.Create(t.Context(), "", &github.Repository{
		Name:    github.Ptr("self-hosted-git-repo" + uuid.NewString()[:8]),
		Private: github.Ptr(true),
	})
	require.NoError(t, err)

	// cleanup repo
	t.Cleanup(func() {
		owner, ghrepo, ok := gitutil.SplitGithubRemote(*repo.CloneURL)
		require.True(t, ok, "invalid github remote: %s", *repo.CloneURL)
		_, err = ghClient.Repositories.Delete(context.Background(), owner, ghrepo)
		require.NoError(t, err, "failed to delete github repo %s/%s: %w", owner, ghrepo, err)
	})

	// push to github
	require.NoError(t, err)
	author := &object.Signature{
		Name:  "Rill test user",
		Email: "test.user@rilldata.com",
	}
	require.NoError(t, err)
	err = gitutil.CommitAndForcePush(t.Context(), tempDir, &gitutil.Config{
		Remote:        *repo.CloneURL,
		DefaultBranch: "main",
	}, "", author)
	require.NoError(t, err, "failed to push to github repo")

	// deploy project backed by github
	result = adm.Run(t, "deploy", "--interactive=false", "--org=github-test", "--project=self-hosted-deploy", "--skip-deploy=true", "--path="+tempDir)
	require.Equal(t, 0, result.ExitCode, result.Output)

	// verify the project is correctly created
	resp, err := adminClient.GetProject(t.Context(), &adminv1.GetProjectRequest{
		OrganizationName: "github-test",
		Name:             "self-hosted-deploy",
	})
	require.NoError(t, err)
	require.Equal(t, "self-hosted-deploy", resp.Project.Name)
	require.True(t, resp.Project.ManagedGitId == "")

	// check remote configured in directory
	remote, err := gitutil.ExtractGitRemote(tempDir, "origin", false)
	require.NoError(t, err)
	require.Equal(t, *repo.CloneURL, remote.URL)

	// redeploy the same project with changes
	err = os.Mkdir(filepath.Join(tempDir, "models"), 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tempDir, "models/model.sql"), []byte(`SELECT 1 AS one`), 0644)
	require.NoError(t, err)
	result = adm.Run(t, "deploy", "--interactive=false", "--org=github-test", "--project=self-hosted-deploy", "--skip-deploy=true", "--path="+tempDir)
	require.Equal(t, 0, result.ExitCode, result.Output)

	// verify changes are pushed to Github repo
	verifyGithubRepoContents(t, ghClient, resp.Project.GitRemote)
}

func verifyGithubRepoContents(t *testing.T, client *github.Client, remote string) {
	owner, repo, ok := gitutil.SplitGithubRemote(remote)
	require.True(t, ok, "invalid github remote: %s", remote)

	con, _, _, err := client.Repositories.GetContents(t.Context(), owner, repo, "models/model.sql", nil)
	require.NoError(t, err)
	contents, err := con.GetContent()
	require.NoError(t, err)
	require.Equal(t, "SELECT 1 AS one", contents)
}
