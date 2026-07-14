package embed_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rilldata/rill/admin/testadmin"
	"github.com/rilldata/rill/cli/testcli"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
)

func TestEmbedOpen(t *testing.T) {
	testmode.Expensive(t)

	adm := testadmin.NewWithOptionalRuntime(t, true)

	// First user is automatically a superuser.
	_, u1Client := adm.NewUser(t)
	u1 := testcli.New(t, adm, u1Client.Token)

	// Second user: regular, owns the org and project.
	_, u2Client := adm.NewUser(t)
	u2 := testcli.New(t, adm, u2Client.Token)

	// Third user: no access
	_, u3Client := adm.NewUser(t)
	u3 := testcli.New(t, adm, u3Client.Token)

	// u2 creates an org and deploys a project.
	res := u2.Run(t, "org", "create", "embed-test")
	require.Equal(t, 0, res.ExitCode, res.Output)

	tempDir := t.TempDir()
	putFiles(t, tempDir, map[string]string{
		"rill.yaml": `olap_connector: duckdb`,
	})
	res = u2.Run(t, "project", "deploy", "--interactive=false", "--org=embed-test", "--project=embed-project", "--path="+tempDir)
	require.Equal(t, 0, res.ExitCode, res.Output)

	// Reconcile the deployment so it has a runtime host and instance.
	adm.TriggerDeployment(t, "embed-test", "embed-project")

	// Superuser can get an embed URL for the project.
	res = u1.Run(t, "sudo", "embed", "open", "embed-test", "embed-project", "--no-open", "--navigation")
	require.Equal(t, 0, res.ExitCode, res.Output)
	require.Contains(t, res.Output, "Open browser at:")

	// Normal user cannot get an embed URL for the project.
	res = u3.Run(t, "sudo", "embed", "open", "embed-test", "embed-project", "--no-open", "--navigation")
	require.NotEqual(t, 0, res.ExitCode)
	require.Contains(t, res.Output, "does not have permission")
}

func putFiles(t *testing.T, baseDir string, files map[string]string) {
	t.Helper()
	for path, content := range files {
		path = filepath.Join(baseDir, path)
		dir := filepath.Dir(path)
		require.NoError(t, os.MkdirAll(dir, 0755))
		require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	}
}
