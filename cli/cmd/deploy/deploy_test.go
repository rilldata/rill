package deploy_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/go-github/v71/github"
	"github.com/google/uuid"
	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/testadmin"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/cli/testcli"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
)

func TestManagedDeploy(t *testing.T) {
	testmode.Expensive(t)
	adm := testadmin.New(t)

	_, c := adm.NewUser(t)
	u1 := testcli.New(t, adm, c.Token)

	result := u1.Run(t, "org", "create", "github-test")
	require.Equal(t, 0, result.ExitCode)

	// deploy the project
	tempDir := initRillProject(t)

	result = u1.Run(t, "project", "deploy", "--interactive=false", "--org=github-test", "--project=rill-mgd-deploy", "--skip-deploy=true", "--path="+tempDir)
	require.Equal(t, 0, result.ExitCode, result.Output)

	// verify the project is correctly created
	resp, err := c.GetProject(t.Context(), &adminv1.GetProjectRequest{
		Org:     "github-test",
		Project: "rill-mgd-deploy",
	})
	require.NoError(t, err)
	require.Equal(t, "rill-mgd-deploy", resp.Project.Name)

	// get a github client
	installationID, err := adm.Admin.Github.ManagedOrgInstallationID()
	require.NoError(t, err)
	ghClient := adm.Admin.Github.InstallationClient(installationID, nil)

	// cleanup repo
	t.Cleanup(func() {
		// delete github repo
		// currently github repos are deleted by a background job
		// but for test cleanup we will delete it here directly
		owner, repo, ok := gitutil.SplitGithubRemote(resp.Project.GitRemote)
		require.True(t, ok, "invalid github remote: %s", resp.Project.GitRemote)
		_, err = ghClient.Repositories.Delete(context.Background(), owner, repo)
		require.NoError(t, err, "failed to delete github repo %s/%s: %v", owner, repo, err)
	})

	// redeploy the same project with changes
	changes := map[string]string{
		"models/model.sql": `SELECT 1 AS one`,
	}
	putFiles(t, tempDir, changes)
	result = u1.Run(t, "deploy", "--interactive=false", "--org=github-test", "--project=rill-mgd-deploy", "--skip-deploy=true", "--path="+tempDir)
	require.Equal(t, 0, result.ExitCode, result.Output)

	// verify changes are pushed to Github repo
	verifyGithubRepoContents(t, ghClient, resp.Project.GitRemote, changes)
}

// This test require gh cli to be installed on the system.
// Alternatively a personal access token can be set via RILL_TEST_GH_TOKEN environment variable.
func TestGithubDeploy(t *testing.T) {
	testmode.Expensive(t)
	personalAccessToken := getGithubAuthToken(t)
	// github client
	ghClient := github.NewTokenClient(t.Context(), personalAccessToken)
	ghUser, _, err := ghClient.Users.Get(t.Context(), "")
	require.NoError(t, err)

	// test service
	adm := testadmin.New(t)
	user, c := adm.NewUser(t)
	expiry := time.Now().Add(time.Hour * 24 * 30)
	adm.Admin.DB.UpdateUser(t.Context(), user.ID, &database.UpdateUserOptions{
		DisplayName:          user.DisplayName,
		PhotoURL:             user.PhotoURL,
		GithubUsername:       *ghUser.Login,
		GithubToken:          personalAccessToken,
		GithubTokenExpiresOn: &expiry,
	})
	u1 := testcli.New(t, adm, c.Token)

	t.Run("self-hosted git deploy", func(t *testing.T) {
		testSelfHostedDeploy(t, c, ghClient, u1)
	})

	t.Run("self-hosted git deploy with monorepo", func(t *testing.T) {
		testSelfHostedMonorepoDeploy(t, c, ghClient, u1)
	})
}

func testSelfHostedDeploy(t *testing.T, adminClient *client.Client, ghClient *github.Client, adm *testcli.Fixture) {
	testmode.Expensive(t)
	result := adm.Run(t, "org", "create", "github-test")
	require.Equal(t, 0, result.ExitCode)

	// create a rill project
	tempDir := initRillProject(t)

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
		require.NoError(t, err, "failed to delete github repo %s/%s: %v", owner, ghrepo, err)
	})

	author := &object.Signature{
		Name:  "Rill test user",
		Email: "test.user@rilldata.com",
	}
	err = gitutil.CommitAndPush(t.Context(), tempDir, &gitutil.Config{
		Remote:        *repo.CloneURL,
		DefaultBranch: "main",
	}, "", author)
	require.NoError(t, err, "failed to push to github repo")

	// deploy project backed by github
	result = adm.Run(t, "deploy", "--interactive=false", "--org=github-test", "--project=self-hosted-deploy", "--skip-deploy=true", "--path="+tempDir)
	require.Equal(t, 0, result.ExitCode, result.Output)

	// verify the project is correctly created
	resp, err := adminClient.GetProject(t.Context(), &adminv1.GetProjectRequest{
		Org:     "github-test",
		Project: "self-hosted-deploy",
	})
	require.NoError(t, err)
	require.Equal(t, "self-hosted-deploy", resp.Project.Name)
	require.Empty(t, resp.Project.ManagedGitId)

	// check remote configured in directory
	remote, err := gitutil.ExtractGitRemote(tempDir, "origin", false)
	require.NoError(t, err)
	require.Equal(t, *repo.CloneURL, remote.URL)

	result = adm.Run(t, "deploy", "--interactive=false", "--org=github-test", "--project=self-hosted-deploy", "--skip-deploy=true", "--path="+tempDir)
	require.Equal(t, 0, result.ExitCode, result.Output)

	changes := map[string]string{
		"models/model.sql": `SELECT 1 AS one`,
	}
	putFiles(t, tempDir, changes)
	// redeploy the same project with changes
	result = adm.Run(t, "deploy", "--interactive=false", "--org=github-test", "--project=self-hosted-deploy", "--skip-deploy=true", "--path="+tempDir)
	require.Equal(t, 0, result.ExitCode, result.Output)

	// verify changes are pushed to Github repo
	verifyGithubRepoContents(t, ghClient, resp.Project.GitRemote, changes)
}

func testSelfHostedMonorepoDeploy(t *testing.T, adminClient *client.Client, ghClient *github.Client, adm *testcli.Fixture) {
	testmode.Expensive(t)
	result := adm.Run(t, "org", "create", "github-monorepo-test")
	require.Equal(t, 0, result.ExitCode)

	// create a monorepo structure
	tempDir := initMonorepo(t)

	// create a github repo
	repo, _, err := ghClient.Repositories.Create(t.Context(), "", &github.Repository{
		Name:    github.Ptr("self-hosted-monorepo" + uuid.NewString()[:8]),
		Private: github.Ptr(true),
	})
	require.NoError(t, err)

	// cleanup repo
	t.Cleanup(func() {
		owner, ghrepo, ok := gitutil.SplitGithubRemote(*repo.CloneURL)
		require.True(t, ok, "invalid github remote: %s", *repo.CloneURL)
		_, err = ghClient.Repositories.Delete(context.Background(), owner, ghrepo)
		require.NoError(t, err, "failed to delete github repo %s/%s: %v", owner, ghrepo, err)
	})

	author := &object.Signature{
		Name:  "Rill test user",
		Email: "test.user@rilldata.com",
	}
	err = gitutil.CommitAndPush(t.Context(), tempDir, &gitutil.Config{
		Remote:        *repo.CloneURL,
		DefaultBranch: "main",
	}, "", author)
	require.NoError(t, err, "failed to push to github repo")

	// Clone two separate copies of the same repo to simulate independent working directories
	// This demonstrates that different subpaths can be worked on independently
	clone1Dir := t.TempDir()
	clone2Dir := t.TempDir()

	// Clone repo to first directory
	err = cloneRepo(t.Context(), *repo.CloneURL, clone1Dir)
	require.NoError(t, err, "failed to clone repo to first directory")

	// Clone repo to second directory
	err = cloneRepo(t.Context(), *repo.CloneURL, clone2Dir)
	require.NoError(t, err, "failed to clone repo to second directory")

	// deploy project1 from first clone
	project1Path := filepath.Join(clone1Dir, "project1")
	result = adm.Run(t, "deploy", "--interactive=false", "--org=github-monorepo-test", "--project=monorepo-project1", "--skip-deploy=true", "--path="+project1Path)
	require.Equal(t, 0, result.ExitCode, result.Output)

	// verify the first project is correctly created with subpath
	resp1, err := adminClient.GetProject(t.Context(), &adminv1.GetProjectRequest{
		Org:     "github-monorepo-test",
		Project: "monorepo-project1",
	})
	require.NoError(t, err)
	require.Equal(t, "monorepo-project1", resp1.Project.Name)
	require.Empty(t, resp1.Project.ManagedGitId)
	require.Equal(t, "project1", resp1.Project.Subpath)

	// deploy project2 from second clone (independent of first clone)
	project2Path := filepath.Join(clone2Dir, "project2")
	result = adm.Run(t, "deploy", "--interactive=false", "--org=github-monorepo-test", "--project=monorepo-project2", "--skip-deploy=true", "--path="+project2Path)
	require.Equal(t, 0, result.ExitCode, result.Output)

	// verify the second project is correctly created with different subpath
	resp2, err := adminClient.GetProject(t.Context(), &adminv1.GetProjectRequest{
		Org:     "github-monorepo-test",
		Project: "monorepo-project2",
	})
	require.NoError(t, err)
	require.Equal(t, "monorepo-project2", resp2.Project.Name)
	require.Empty(t, resp2.Project.ManagedGitId)
	require.Equal(t, "project2", resp2.Project.Subpath)
	require.Equal(t, resp1.Project.GitRemote, resp2.Project.GitRemote)

	// make changes to project1 in first clone only
	changesProject1 := map[string]string{
		"project1/models/model.sql": `SELECT 1 AS one`,
	}
	putFiles(t, clone1Dir, changesProject1)

	// redeploy project1 with changes from first clone
	result = adm.Run(t, "deploy", "--interactive=false", "--org=github-monorepo-test", "--project=monorepo-project1", "--skip-deploy=true", "--path="+project1Path)
	require.Equal(t, 0, result.ExitCode, result.Output)

	// verify changes are pushed to Github repo in project1 subpath
	verifyGithubRepoContents(t, ghClient, resp1.Project.GitRemote, changesProject1)

	// make changes to project2 in second clone independently
	changesProject2 := map[string]string{
		"project2/models/model.sql": `SELECT 2 AS two`,
	}
	putFiles(t, clone2Dir, changesProject2)

	// redeploy project2 with changes from second clone
	result = adm.Run(t, "deploy", "--interactive=false", "--org=github-monorepo-test", "--project=monorepo-project2", "--skip-deploy=true", "--path="+project2Path)
	require.Equal(t, 0, result.ExitCode, result.Output)

	// verify changes are pushed to Github repo in project2 subpath
	verifyGithubRepoContents(t, ghClient, resp2.Project.GitRemote, changesProject2)

	// verify both projects' changes exist in the repo, demonstrating independence
	allChanges := map[string]string{
		"project1/models/model.sql": `SELECT 1 AS one`,
		"project2/models/model.sql": `SELECT 2 AS two`,
	}
	verifyGithubRepoContents(t, ghClient, resp1.Project.GitRemote, allChanges)
}

func verifyGithubRepoContents(t *testing.T, client *github.Client, remote string, changes map[string]string) {
	owner, repo, ok := gitutil.SplitGithubRemote(remote)
	require.True(t, ok, "invalid github remote: %s", remote)

	// TODO: consider downloading the repo and checking the files locally instead of making multiple API calls
	for path, expectedContent := range changes {
		con, _, _, err := client.Repositories.GetContents(t.Context(), owner, repo, path, nil)
		require.NoError(t, err)
		contents, err := con.GetContent()
		require.NoError(t, err)
		require.Equal(t, expectedContent, contents)
	}
}

func getGithubAuthToken(t *testing.T) string {
	// check if token is set via environment variable
	if token := os.Getenv("RILL_TEST_GH_TOKEN"); token != "" {
		return token
	}
	// exec gh auth token and extract token
	// throw error if gh cli is not installed
	t.Helper()

	// Try to find gh in PATH first
	ghPath, err := exec.LookPath("gh")
	if err != nil {
		// Fallback to common installation paths
		commonPaths := []string{
			"/opt/homebrew/bin/gh",
			"/usr/local/bin/gh",
			"/usr/bin/gh",
		}
		for _, path := range commonPaths {
			if _, err := os.Stat(path); err == nil {
				ghPath = path
				break
			}
		}
		if ghPath == "" {
			t.Fatal("gh cli not found in PATH or common installation paths. For installation instructions, visit: https://github.com/cli/cli#installation")
		}
	}

	cmd := exec.CommandContext(t.Context(), ghPath, "auth", "token")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "failed to get github auth token: %s", string(output))
	return strings.TrimSpace(string(output))
}

func putFiles(t *testing.T, baseDir string, files map[string]string) {
	for path, content := range files {
		path = filepath.Join(baseDir, path)
		dir := filepath.Dir(path)
		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err)
		err = os.WriteFile(path, []byte(content), 0644)
		require.NoError(t, err)
	}
}

func initRillProject(t *testing.T) string {
	tempDir := t.TempDir()
	putFiles(t, tempDir, map[string]string{"rill.yaml": `compiler: rillv1
display_name: Untitled Rill Project
olap_connector: duckdb`,
	})
	return tempDir
}

func initMonorepo(t *testing.T) string {
	tempDir := t.TempDir()

	// Create project1 in monorepo
	putFiles(t, tempDir, map[string]string{
		"project1/rill.yaml": `compiler: rillv1
display_name: Monorepo Project 1
olap_connector: duckdb`,
		"project1/models/.gitkeep": "",
	})

	// Create project2 in monorepo
	putFiles(t, tempDir, map[string]string{
		"project2/rill.yaml": `compiler: rillv1
display_name: Monorepo Project 2
olap_connector: duckdb`,
		"project2/models/.gitkeep": "",
	})

	// Add root level README for the monorepo
	putFiles(t, tempDir, map[string]string{
		"README.md": "# Test Monorepo\nThis is a test monorepo with multiple Rill projects.",
	})

	return tempDir
}

func cloneRepo(ctx context.Context, repoURL, path string) error {
	cmd := exec.CommandContext(ctx, "git", "clone", repoURL, path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("clone repo failed with error: %s(%s)", out, err)
	}
	return nil
}
