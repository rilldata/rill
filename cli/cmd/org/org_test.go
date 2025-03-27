package org_test

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/rilldata/rill/admin/testadmin"
	"github.com/rilldata/rill/cli/testcli"
	"github.com/stretchr/testify/require"
)

func TestOrg(t *testing.T) {
	adm := testadmin.New(t)
	u1 := testcli.NewWithUser(t, adm)

	// Create an org
	org1 := randomName()
	res := u1.Run(t, "org", "create", org1)
	t.Log(res.Output)
	require.Equal(t, 0, res.ExitCode)

	// Edit the org
	desc1 := "foo bar"
	res = u1.Run(t, "org", "edit", org1, "--description", desc1)
	require.Equal(t, 0, res.ExitCode)

	// Check the org is stored in local state and can be shown
	res = u1.Run(t, "org", "show")
	require.Equal(t, 0, res.ExitCode)
	require.Contains(t, res.Output, org1)
	require.Contains(t, res.Output, desc1)

	// Create another org
	org2 := randomName()
	res = u1.Run(t, "org", "create", org2)
	require.Equal(t, 0, res.ExitCode)

	// Check the org is stored in local state and can be shown
	res = u1.Run(t, "org", "show")
	require.Equal(t, 0, res.ExitCode)
	require.Contains(t, res.Output, org2)

	// Check we can switch between orgs
	res = u1.Run(t, "org", "switch", org1)
	require.Equal(t, 0, res.ExitCode)
	res = u1.Run(t, "org", "show")
	require.Equal(t, 0, res.ExitCode)
	require.Contains(t, res.Output, org1)

	// Check both orgs can be listed
	res = u1.Run(t, "org", "list")
	require.Equal(t, 0, res.ExitCode)
	require.Contains(t, res.Output, org1)
	require.Contains(t, res.Output, org2)

	// Check it can't show an org that doesn't exist
	res = u1.Run(t, "org", "show", "nonexistent")
	require.Equal(t, 1, res.ExitCode)
	require.Contains(t, res.Output, "not found")

	// Delete the second org
	res = u1.Run(t, "org", "delete", org2, "--interactive=false")
	require.Equal(t, 0, res.ExitCode)
}

func randomName() string {
	id := make([]byte, 16)
	_, err := rand.Read(id)
	if err != nil {
		panic(err)
	}
	return "test_" + hex.EncodeToString(id)
}
