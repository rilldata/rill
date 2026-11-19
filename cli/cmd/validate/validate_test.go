package validate_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/rilldata/rill/cli/cmd"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	"github.com/rilldata/rill/cli/pkg/local"
	"github.com/rilldata/rill/cli/pkg/version"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// TestValidateWhileProjectIsRunning re-executes the test binary with these variables to run `rill validate` in a process of its own.
	if projectPath := os.Getenv("RILL_TEST_VALIDATE_PATH"); projectPath != "" {
		ch, err := cmdutil.NewHelper(version.Version{}, os.Getenv("RILL_TEST_VALIDATE_HOME"))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		root := cmd.RootCmd(ch)
		root.SetArgs([]string{"validate", projectPath})
		os.Exit(cmd.HandleExecuteError(ch, root.ExecuteContext(context.Background())))
	}

	os.Exit(m.Run())
}

func TestValidateWhileProjectIsRunning(t *testing.T) {
	projectPath := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(projectPath, "rill.yaml"), []byte("compiler: rillv1\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(projectPath, "m1.sql"), []byte("SELECT 1 AS a"), 0o644))

	homeDir := t.TempDir()
	require.NoError(t, dotrill.New(homeDir).Set(dotrill.ConfigFilename, dotrill.AnalyticsEnabledConfigKey, "false"))

	validate := func() (string, error) {
		c := exec.CommandContext(t.Context(), os.Args[0])
		c.Env = append(os.Environ(), "RILL_TEST_VALIDATE_PATH="+projectPath, "RILL_TEST_VALIDATE_HOME="+homeDir)
		out, err := c.CombinedOutput()
		return string(out), err
	}

	// Run the project the way `rill start --reset --port 9100` does.
	// The reset must clear the state of earlier runs without giving up the project's lock.
	staleState := filepath.Join(projectPath, "tmp", "default", "stale")
	require.NoError(t, os.MkdirAll(staleState, 0o755))
	ch, err := cmdutil.NewHelper(version.Version{}, homeDir)
	require.NoError(t, err)
	app, err := local.NewApp(t.Context(), &local.AppOptions{
		Ch:          ch,
		Silent:      true,
		Reset:       true,
		Environment: "dev",
		ProjectPath: projectPath,
		LogFormat:   "console",
		LocalURL:    "http://localhost:9100",
		ServeUI:     true,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = app.Close()
		delete(drivers.Drivers, "local_admin") // Registered by NewApp; removed so the test can run more than once per process.
	})
	require.NoDirExists(t, staleState)

	// Validate refuses to run, and says where the project is running.
	out, err := validate()
	require.Error(t, err)
	require.Contains(t, out, fmt.Sprintf("Error: `rill start` is already running for this project at http://localhost:9100 (PID %d); stop it and try again", os.Getpid()))

	// The running project is unaffected: it still reconciles changes and saves them to its catalog.
	ctrl, err := app.Runtime.Controller(t.Context(), app.Instance.ID)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(projectPath, "m2.sql"), []byte("SELECT 2 AS b"), 0o644))
	require.NoError(t, ctrl.Reconcile(t.Context(), runtime.GlobalProjectParserName))
	require.Eventually(t, func() bool {
		r, err := ctrl.Get(t.Context(), &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: "m2"}, false)
		return err == nil && r.Meta.ReconcileStatus == runtimev1.ReconcileStatus_RECONCILE_STATUS_IDLE && r.Meta.ReconcileError == ""
	}, 30*time.Second, 100*time.Millisecond)
	require.NoError(t, ctrl.Flush(t.Context()))

	// Once the project has stopped, validate runs.
	require.NoError(t, app.Close())
	out, err = validate()
	require.NoError(t, err, out)
	require.Contains(t, out, "Validation completed successfully")
}
