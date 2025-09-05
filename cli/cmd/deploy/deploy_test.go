package deploy_test

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	"github.com/rilldata/rill/admin/testadmin"
	"github.com/rilldata/rill/cli/testcli"
	"github.com/stretchr/testify/require"
)

func TestDeploy(t *testing.T) {
	err := godotenv.Load("../../../.env")
	require.NoError(t, err)

	ok, _ := strconv.ParseBool(os.Getenv("USE_ACTUAL_GITHUB"))
	if !ok {
		t.Skip("Skipping TestDeploy since USE_ACTUAL_GITHUB is not set")
	}
	adm := testadmin.New(t)

	u1 := testcli.NewWithUser(t, adm)

	result := u1.Run(t, "org", "create", "github-test")
	require.Equal(t, 0, result.ExitCode)

	// deploy the project
	tempDir := t.TempDir()
	os.WriteFile(filepath.Join(tempDir, "rill.yaml"), []byte(`compiler: rillv1
display_name: Untitled Rill Project
olap_connector: duckdb`), 0644)

	result = u1.Run(t, "project", "deploy", "--interactive=false", "--org=github-test", "--project=rill-mgd-deploy", "--skip-deploy=true", "--path="+tempDir)
	require.Equal(t, 0, result.ExitCode)

	// verify the project is correctly created
	verifyProjects(t, u1, "rill-mgd-deploy")

	// TODO : cleanup the managed github repo
}

func verifyProjects(t *testing.T, u *testcli.Fixture, expectedProjects ...string) {
	result := u.Run(t, "project", "list", "--format=csv")
	require.Equal(t, 0, result.ExitCode)
	out := strings.TrimPrefix(strings.TrimSpace(result.Output), "Projects list\n")

	// parse csv output
	r := csv.NewReader(strings.NewReader(strings.TrimSpace(out)))
	records, err := r.ReadAll()
	require.NoError(t, err)
	var projects []string
	for i, record := range records {
		if i == 0 {
			// skip header
			continue
		}
		projects = append(projects, record[0])
	}
	require.ElementsMatch(t, expectedProjects, projects)
}
