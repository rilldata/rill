package token_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/rilldata/rill/admin/testadmin"
	"github.com/rilldata/rill/cli/testcli"
	"github.com/stretchr/testify/require"
)

func TestToken(t *testing.T) {
	adm := testadmin.New(t)
	u1 := testcli.NewWithUser(t, adm)

	// Issue a plain token
	res := u1.Run(t, "token", "issue", "--display-name", "Test Token")
	require.Equal(t, 0, res.ExitCode)
	require.Contains(t, res.Output, "Token: ")

	// Check the token works
	token := strings.TrimSpace(strings.ReplaceAll(res.Output, "Token: ", ""))
	uTmp := testcli.New(t, adm, token)
	res = uTmp.Run(t, "whoami")
	require.Equal(t, 0, res.ExitCode)
	require.Contains(t, res.Output, "Email: ")

	// Issue a token with a description
	res = u1.Run(t, "token", "issue", "--display-name", "Foo")
	require.Equal(t, 0, res.ExitCode)
	require.Contains(t, res.Output, "Token: ")

	// Issue a token with an expiration
	res = u1.Run(t, "token", "issue", "--display-name", "Test Token", "--ttl-minutes", "1")
	require.Equal(t, 0, res.ExitCode)
	require.Contains(t, res.Output, "Token: ")

	// List tokens with limit
	res = u1.Run(t, "token", "list", "--page-size", "1")
	require.Equal(t, 0, res.ExitCode)
	require.Contains(t, res.Output, "Next page token: ")

	// List tokens and count
	res = u1.Run(t, "token", "list", "--format", "json")
	require.Equal(t, 0, res.ExitCode)
	var rows []map[string]any
	err := json.Unmarshal([]byte(res.Output), &rows)
	require.NoError(t, err)
	require.Equal(t, 4, len(rows)) // 3 created above and 1 from testcli.NewWithUser

	// One should have description "Foo" and one should have an expiration
	var foundFoo, foundExpiration bool
	for _, row := range rows {
		if row["description"] == "Foo" {
			foundFoo = true
		}
		if row["expires_on"].(string) != "" {
			foundExpiration = true
		}
	}
	require.True(t, foundFoo)
	require.True(t, foundExpiration)

	// Find an ID for one of the tokens
	var tokenID string
	for _, row := range rows {
		if row["description"] == "Foo" { // Description we set above
			tokenID = row["id"].(string)
			break
		}
	}

	// Revoke the token
	res = u1.Run(t, "token", "revoke", tokenID)
	require.Equal(t, 0, res.ExitCode)

	// Check the token no longer is there
	res = u1.Run(t, "token", "list")
	require.Equal(t, 0, res.ExitCode)
	require.NotContains(t, res.Output, tokenID)
}
