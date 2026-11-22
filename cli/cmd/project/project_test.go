package project_test

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/rilldata/rill/admin/testadmin"
	"github.com/rilldata/rill/cli/testcli"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/stretchr/testify/require"
)

func TestProjectTags(t *testing.T) {
	adm := testadmin.New(t)
	_, c := adm.NewUser(t)
	u1 := testcli.New(t, adm, c.Token)

	org := randomName()
	res := u1.Run(t, "org", "create", org)
	require.Equal(t, 0, res.ExitCode, res.Output)
	for _, p := range []*adminv1.CreateProjectRequest{
		{Org: org, Project: "revenue", Tags: []string{"prod"}, SkipDeploy: true},
		{Org: org, Project: "revenue-review", Tags: []string{"sandbox"}, SkipDeploy: true},
	} {
		_, err := c.CreateProject(t.Context(), p)
		require.NoError(t, err)
	}

	// Tags can be repeated or comma-separated, and are normalized
	res = u1.Run(t, "project", "edit", "revenue-review", "--add-tag", "Stage,team-a", "--add-tag", "eu", "--remove-tag", "sandbox")
	require.Equal(t, 0, res.ExitCode, res.Output)
	require.Contains(t, res.Output, "eu, stage, team-a")

	res = u1.Run(t, "project", "edit", "revenue-review", "--add-tag", "not valid")
	require.Equal(t, 1, res.ExitCode)
	require.Contains(t, res.Output, "must start with a letter or digit")

	// Filter by any of the tags
	res = u1.Run(t, "project", "list", "--tag", "stage,prod", "--format", "json")
	require.Equal(t, 0, res.ExitCode, res.Output)
	require.Contains(t, res.Output, `"name": "revenue"`)
	require.Contains(t, res.Output, `"tags": "eu, stage, team-a"`)

	res = u1.Run(t, "project", "list", "--tag", "prod", "--format", "json")
	require.Equal(t, 0, res.ExitCode, res.Output)
	require.Contains(t, res.Output, `"name": "revenue"`)
	require.NotContains(t, res.Output, "revenue-review")

	res = u1.Run(t, "project", "list", "--tag", "missing")
	require.Equal(t, 0, res.ExitCode, res.Output)
	require.Contains(t, res.Output, "No projects found")
}

func randomName() string {
	id := make([]byte, 16)
	_, err := rand.Read(id)
	if err != nil {
		panic(err)
	}
	return "test_" + hex.EncodeToString(id)
}
