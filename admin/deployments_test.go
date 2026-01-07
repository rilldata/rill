package admin_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/jobs/river"
	"github.com/rilldata/rill/admin/testadmin"
	"github.com/rilldata/rill/cli/testcli"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/drivers"
	riverqueue "github.com/riverqueue/river"
	"github.com/stretchr/testify/require"
)

func TestRuntimeDeployments(t *testing.T) {
	start := time.Now()
	fmt.Printf("\n[TIMING] Test started at %v\n", start)

	fmt.Printf("[TIMING] Creating admin with runtime...\n")
	stepStart := time.Now()
	adm := testadmin.NewWithOptionalRuntime(t, true)
	fmt.Printf("[TIMING] Admin creation took: %v\n", time.Since(stepStart))

	fmt.Printf("[TIMING] Creating new user...\n")
	stepStart = time.Now()
	_, c := adm.NewUser(t)
	u1 := testcli.New(t, adm, c.Token)
	fmt.Printf("[TIMING] User creation took: %v\n", time.Since(stepStart))

	fmt.Printf("[TIMING] Creating org...\n")
	stepStart = time.Now()
	result := u1.Run(t, "org", "create", "reload-configs-test")
	require.Equal(t, 0, result.ExitCode)
	fmt.Printf("[TIMING] Org creation took: %v\n", time.Since(stepStart))

	// deploy the project
	fmt.Printf("[TIMING] Initializing Rill project...\n")
	stepStart = time.Now()
	tempDir := initRillProject(t)
	fmt.Printf("[TIMING] Project initialization took: %v\n", time.Since(stepStart))

	fmt.Printf("[TIMING] Deploying project...\n")
	stepStart = time.Now()
	result = u1.Run(t, "project", "deploy", "--interactive=false", "--org=reload-configs-test", "--project=rill-mgd-deploy", "--path="+tempDir)
	require.Equal(t, 0, result.ExitCode, result.Output)
	fmt.Printf("[TIMING] Project deployment took: %v\n", time.Since(stepStart))

	// manually trigger deployment
	fmt.Printf("[TIMING] Triggering deployment #1...\n")
	stepStart = time.Now()
	depl := triggerDeployment(t, adm, "reload-configs-test", "rill-mgd-deploy")
	fmt.Printf("[TIMING] Deployment trigger #1 took: %v\n", time.Since(stepStart))

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
	fmt.Printf("[TIMING] Checking model output (eventually)...\n")
	stepStart = time.Now()
	require.Eventually(t, func() bool {
		modelOutputFn, _ := checkModelOutput()
		return modelOutputFn == 1
	}, 10*time.Second, 100*time.Millisecond, "unexpected model output")
	fmt.Printf("[TIMING] Model output check took: %v\n", time.Since(stepStart))

	// set env via `rill env set limit 10`
	fmt.Printf("[TIMING] Setting env limit=10...\n")
	stepStart = time.Now()
	result = u1.Run(t, "env", "set", "limit", "10", "--org=reload-configs-test", "--project=rill-mgd-deploy")
	require.Equal(t, 0, result.ExitCode, result.Output)
	fmt.Printf("[TIMING] Env set took: %v\n", time.Since(stepStart))

	// manually trigger deployment
	fmt.Printf("[TIMING] Triggering deployment #2...\n")
	stepStart = time.Now()
	depl = triggerDeployment(t, adm, "reload-configs-test", "rill-mgd-deploy")
	fmt.Printf("[TIMING] Deployment trigger #2 took: %v\n", time.Since(stepStart))

	// query the model and verify env variable is applied
	fmt.Printf("[TIMING] Checking model output after env set (eventually)...\n")
	stepStart = time.Now()
	require.Eventually(t, func() bool {
		modelOutputFn, _ := checkModelOutput()
		return modelOutputFn == 10
	}, 10*time.Second, 100*time.Millisecond, "unexpected model output after env set")
	fmt.Printf("[TIMING] Model output check after env set took: %v\n", time.Since(stepStart))

	// stop the deployment - rill project deployments stop main
	fmt.Printf("[TIMING] Stopping deployment...\n")
	stepStart = time.Now()
	result = u1.Run(t, "project", "deployments", "stop", "main", "--org=reload-configs-test", "--project=rill-mgd-deploy")
	require.Equal(t, 0, result.ExitCode, result.Output)
	fmt.Printf("[TIMING] Stopping deployment took: %v\n", time.Since(stepStart))

	// manually trigger deployment
	fmt.Printf("[TIMING] Triggering deployment #3...\n")
	stepStart = time.Now()
	depl = triggerDeployment(t, adm, "reload-configs-test", "rill-mgd-deploy")
	fmt.Printf("[TIMING] Deployment trigger #3 took: %v\n", time.Since(stepStart))

	// verify deployment is stopped
	fmt.Printf("[TIMING] Verifying deployment is stopped...\n")
	stepStart = time.Now()
	deploymentsResp, err := c.ListDeployments(t.Context(), &adminv1.ListDeploymentsRequest{
		Org:     "reload-configs-test",
		Project: "rill-mgd-deploy",
	})
	require.NoError(t, err)
	require.Len(t, deploymentsResp.Deployments, 1)
	require.Equal(t, adminv1.DeploymentStatus_DEPLOYMENT_STATUS_STOPPED, deploymentsResp.Deployments[0].Status)
	fmt.Printf("[TIMING] Deployment verification took: %v\n", time.Since(stepStart))

	// modify the env to set limit to 20
	fmt.Printf("[TIMING] Setting env limit=20...\n")
	stepStart = time.Now()
	result = u1.Run(t, "env", "set", "limit", "20", "--org=reload-configs-test", "--project=rill-mgd-deploy")
	require.Equal(t, 0, result.ExitCode, result.Output)
	fmt.Printf("[TIMING] Env set to 20 took: %v\n", time.Since(stepStart))

	// restart the deployment - use the api direclty since the CLI commands wait for deployment to be running which is not possible without river workers
	fmt.Printf("[TIMING] Restarting deployment...\n")
	stepStart = time.Now()
	_, err = c.StartDeployment(t.Context(), &adminv1.StartDeploymentRequest{
		DeploymentId: deploymentsResp.Deployments[0].Id,
	})
	require.NoError(t, err)
	fmt.Printf("[TIMING] Deployment restart took: %v\n", time.Since(stepStart))

	// manually trigger deployment
	fmt.Printf("[TIMING] Triggering deployment #4...\n")
	stepStart = time.Now()
	depl = triggerDeployment(t, adm, "reload-configs-test", "rill-mgd-deploy")
	fmt.Printf("[TIMING] Deployment trigger #4 took: %v\n", time.Since(stepStart))

	// query the model and verify env variable is applied
	fmt.Printf("[TIMING] Checking model output after restart (eventually)...\n")
	stepStart = time.Now()
	require.Eventually(t, func() bool {
		modelOutputFn, _ := checkModelOutput()
		return modelOutputFn == 20
	}, 10*time.Second, 100*time.Millisecond, "unexpected model output after env set post restart")
	fmt.Printf("[TIMING] Model output check after restart took: %v\n", time.Since(stepStart))

	fmt.Printf("[TIMING] Test completed in: %v\n\n", time.Since(start))
}

func triggerDeployment(t *testing.T, adm *testadmin.Fixture, org, project string) *database.Deployment {
	proj, err := adm.Admin.DB.FindProjectByName(t.Context(), org, project)
	require.NoError(t, err)
	depl, err := adm.Admin.DB.FindDeploymentsForProject(t.Context(), proj.ID, "", "")
	require.NoError(t, err)
	require.Len(t, depl, 1)
	err = river.NewReconcileDeploymentWorker(adm.Admin).Work(t.Context(), &riverqueue.Job[river.ReconcileDeploymentArgs]{
		Args: river.ReconcileDeploymentArgs{
			DeploymentID: depl[0].ID,
		},
	})
	require.NoError(t, err)
	depl, err = adm.Admin.DB.FindDeploymentsForProject(t.Context(), proj.ID, "", "")
	require.NoError(t, err)
	require.Len(t, depl, 1)
	return depl[0]
}

func initRillProject(t *testing.T) string {
	tempDir := t.TempDir()
	putFiles(t, tempDir, map[string]string{"rill.yaml": `compiler: rillv1
display_name: Untitled Rill Project
olap_connector: duckdb
vars:
  limit: 1`,
	})
	putFiles(t, tempDir, map[string]string{"models/model.sql": "SELECT {{ .env.limit }} AS lmt"})
	return tempDir
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
