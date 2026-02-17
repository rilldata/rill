package deployment_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-github/v71/github"
	"github.com/rilldata/rill/admin/testadmin"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/cli/testcli"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
)

func TestRuntimeDeployments(t *testing.T) {
	testmode.Expensive(t)

	adm := testadmin.NewWithOptionalRuntime(t, true)
	_, c := adm.NewUser(t)
	u1 := testcli.New(t, adm, c.Token)

	result := u1.Run(t, "org", "create", "reload-configs-test")
	require.Equal(t, 0, result.ExitCode)

	// deploy the project
	tempDir := t.TempDir()
	putFiles(t, tempDir, map[string]string{"rill.yaml": `compiler: rillv1
display_name: Untitled Rill Project
olap_connector: duckdb
vars:
  limit: 1`,
	})
	putFiles(t, tempDir, map[string]string{"models/model.sql": "SELECT {{ .env.limit }} AS lmt"})
	result = u1.Run(t, "project", "deploy", "--interactive=false", "--org=reload-configs-test", "--project=rill-mgd-deploy", "--path="+tempDir)
	require.Equal(t, 0, result.ExitCode, result.Output)

	// manually trigger deployment
	depl := adm.TriggerDeployment(t, "reload-configs-test", "rill-mgd-deploy")

	// check model output
	checkModelOutput := func() (int, error) {
		olap, release, err := adm.Runtime.OLAP(t.Context(), depl.RuntimeInstanceID, "duckdb")
		if err != nil {
			return 0, err
		}
		defer release()

		rows, err := olap.Query(t.Context(), &drivers.Statement{Query: "SELECT lmt FROM model"})
		if err != nil {
			return 0, err
		}
		defer rows.Close()

		var res int
		for rows.Next() {
			if err := rows.Scan(&res); err != nil {
				return 0, err
			}
		}
		if err := rows.Err(); err != nil {
			return 0, err
		}
		return res, nil
	}
	require.Eventually(t, func() bool {
		modelOutputFn, _ := checkModelOutput()
		return modelOutputFn == 1
	}, 10*time.Second, 100*time.Millisecond, "unexpected model output")

	// set env via `rill env set limit 10`
	result = u1.Run(t, "env", "set", "limit", "10", "--org=reload-configs-test", "--project=rill-mgd-deploy")
	require.Equal(t, 0, result.ExitCode, result.Output)

	// manually trigger deployment
	depl = adm.TriggerDeployment(t, "reload-configs-test", "rill-mgd-deploy")

	// query the model and verify env variable is applied
	require.Eventually(t, func() bool {
		modelOutputFn, _ := checkModelOutput()
		return modelOutputFn == 10
	}, 10*time.Second, 100*time.Millisecond, "unexpected model output after env set")

	// stop the deployment - rill project deployments stop main
	result = u1.Run(t, "project", "deployment", "stop", "main", "--org=reload-configs-test", "--project=rill-mgd-deploy")
	require.Equal(t, 0, result.ExitCode, result.Output)

	// manually trigger deployment
	depl = adm.TriggerDeployment(t, "reload-configs-test", "rill-mgd-deploy")

	// verify deployment is stopped
	deploymentsResp, err := c.ListDeployments(t.Context(), &adminv1.ListDeploymentsRequest{
		Org:     "reload-configs-test",
		Project: "rill-mgd-deploy",
	})
	require.NoError(t, err)
	require.Len(t, deploymentsResp.Deployments, 1)
	require.Equal(t, adminv1.DeploymentStatus_DEPLOYMENT_STATUS_STOPPED, deploymentsResp.Deployments[0].Status)

	// modify the env to set limit to 20
	result = u1.Run(t, "env", "set", "limit", "20", "--org=reload-configs-test", "--project=rill-mgd-deploy")
	require.Equal(t, 0, result.ExitCode, result.Output)

	// restart the deployment - use the api directly since the CLI commands wait for deployment to be running which is not possible without river workers
	_, err = c.StartDeployment(t.Context(), &adminv1.StartDeploymentRequest{
		DeploymentId: deploymentsResp.Deployments[0].Id,
	})
	require.NoError(t, err)

	// manually trigger deployment
	depl = adm.TriggerDeployment(t, "reload-configs-test", "rill-mgd-deploy")

	// query the model and verify env variable is applied
	require.Eventually(t, func() bool {
		modelOutputFn, _ := checkModelOutput()
		return modelOutputFn == 20
	}, 10*time.Second, 100*time.Millisecond, "unexpected model output after env set post restart")
}

func TestPrimaryBranchChange(t *testing.T) {
	testmode.Expensive(t)

	adm := testadmin.NewWithOptionalRuntime(t, true)
	_, c := adm.NewUser(t)
	u1 := testcli.New(t, adm, c.Token)

	result := u1.Run(t, "org", "create", "branch-change-test")
	require.Equal(t, 0, result.ExitCode)

	// deploy the project on main branch
	tempDir := t.TempDir()
	putFiles(t, tempDir, map[string]string{"rill.yaml": `compiler: rillv1
display_name: Branch Change Test
olap_connector: duckdb`,
	})
	putFiles(t, tempDir, map[string]string{"models/model.sql": "SELECT 'main' AS branch"})
	result = u1.Run(t, "project", "deploy", "--interactive=false", "--org=branch-change-test", "--project=branch-test", "--path="+tempDir)
	require.Equal(t, 0, result.ExitCode, result.Output)

	// manually trigger deployment
	depl := adm.TriggerDeployment(t, "branch-change-test", "branch-test")

	// check model output from main branch
	checkModelOutput := func() (string, error) {
		olap, release, err := adm.Runtime.OLAP(t.Context(), depl.RuntimeInstanceID, "duckdb")
		if err != nil {
			return "", err
		}
		defer release()

		rows, err := olap.Query(t.Context(), &drivers.Statement{Query: "SELECT branch FROM model"})
		if err != nil {
			return "", err
		}
		defer rows.Close()

		var res string
		for rows.Next() {
			if err := rows.Scan(&res); err != nil {
				return "", err
			}
		}
		if err := rows.Err(); err != nil {
			return "", err
		}
		return res, nil
	}
	require.Eventually(t, func() bool {
		branch, _ := checkModelOutput()
		return branch == "main"
	}, 10*time.Second, 100*time.Millisecond, "expected model output to be 'main'")

	// get project to find git remote
	proj, err := c.GetProject(t.Context(), &adminv1.GetProjectRequest{
		Org:     "branch-change-test",
		Project: "branch-test",
	})
	require.NoError(t, err)

	// create a new branch in github with updated model
	installationID, err := adm.Admin.Github.ManagedOrgInstallationID()
	require.NoError(t, err)
	ghClient := adm.Admin.Github.InstallationClient(installationID, nil)

	owner, repo, ok := gitutil.SplitGithubRemote(proj.Project.GitRemote)
	require.True(t, ok, "invalid github remote: %s", proj.Project.GitRemote)

	// get the current main branch ref
	mainRef, _, err := ghClient.Git.GetRef(t.Context(), owner, repo, "refs/heads/main")
	require.NoError(t, err)

	// create new branch "feature" from main
	newBranchRef := "refs/heads/feature"
	_, _, err = ghClient.Git.CreateRef(t.Context(), owner, repo, &github.Reference{
		Ref:    &newBranchRef,
		Object: &github.GitObject{SHA: mainRef.Object.SHA},
	})
	require.NoError(t, err)

	// update model.sql in the feature branch
	fileContent := "SELECT 'feature' AS branch"
	filePath := "models/model.sql"

	// get current file to get its SHA
	fileInfo, _, _, err := ghClient.Repositories.GetContents(t.Context(), owner, repo, filePath, &github.RepositoryContentGetOptions{
		Ref: "feature",
	})
	require.NoError(t, err)

	// update file in feature branch
	_, _, err = ghClient.Repositories.UpdateFile(t.Context(), owner, repo, filePath, &github.RepositoryContentFileOptions{
		Message: github.Ptr("Update model for feature branch"),
		Content: []byte(fileContent),
		SHA:     fileInfo.SHA,
		Branch:  github.Ptr("feature"),
	})
	require.NoError(t, err)

	// change primary branch using project edit
	result = u1.Run(t, "project", "edit", "--primary-branch=feature", "--project=branch-test", "--org=branch-change-test")
	require.Equal(t, 0, result.ExitCode, result.Output)

	// verify project primary branch is updated
	proj, err = c.GetProject(t.Context(), &adminv1.GetProjectRequest{
		Org:     "branch-change-test",
		Project: "branch-test",
	})
	require.NoError(t, err)
	require.Equal(t, "feature", proj.Project.PrimaryBranch)

	// manually trigger deployment to pick up new branch
	depl = adm.TriggerDeployment(t, "branch-change-test", "branch-test")

	// verify model is updated with changes from feature branch
	require.Eventually(t, func() bool {
		branch, _ := checkModelOutput()
		return branch == "feature"
	}, 10*time.Second, 100*time.Millisecond, "expected model output to be 'feature' after branch change")
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
