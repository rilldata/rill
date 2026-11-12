package server_test

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/admin/testadmin"
	"github.com/rilldata/rill/cli/testcli"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v3"
)

// TestEmbedPersonalFiles tests that embed users identified by an external user ID can manage their own personal files,
// and that personal files are isolated between embed users and Rill users.
func TestEmbedPersonalFiles(t *testing.T) {
	testmode.Expensive(t)

	adm := testadmin.NewWithOptionalRuntime(t, true)

	// Create test users. Note: the first user becomes a superuser, so we let the second one own the org/project to avoid confounding effects.
	_, _ = adm.NewUser(t)
	_, u1Client := adm.NewUser(t)

	// Deploy an empty test project
	projectDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "rill.yaml"), []byte("olap_connector: duckdb\n"), 0644))
	owner := testcli.New(t, adm, u1Client.Token)
	res := owner.Run(t, "org", "create", "personal-test")
	require.Equal(t, 0, res.ExitCode, res.Output)
	res = owner.Run(t, "project", "deploy", "--interactive=false", "--org=personal-test", "--project=personal-project", "--path="+projectDir)
	require.Equal(t, 0, res.ExitCode, res.Output)

	// embedClient calls GetIFrame with navigation enabled and returns a client using the iframe's admin token.
	embedClient := func(t *testing.T, externalUserID string) *client.Client {
		resp, err := u1Client.GetIFrame(t.Context(), &adminv1.GetIFrameRequest{
			Org:            "personal-test",
			Project:        "personal-project",
			Navigation:     true,
			ExternalUserId: externalUserID,
		})
		require.NoError(t, err)
		u, err := url.Parse(resp.IframeSrc)
		require.NoError(t, err)
		tkn := u.Query().Get("admin_token")
		require.NotEmpty(t, tkn)
		return adm.NewClient(t, tkn)
	}

	create := func(t *testing.T, c *client.Client, data string) string {
		resp, err := c.CreatePersonalFile(t.Context(), &adminv1.CreatePersonalFileRequest{
			Org:         "personal-test",
			Project:     "personal-project",
			Kind:        runtime.ResourceKindCanvas,
			DisplayName: "My canvas",
			Yaml:        data,
		})
		require.NoError(t, err)
		return resp.Name
	}

	get := func(t *testing.T, c *client.Client, name string) (*adminv1.GetPersonalFileResponse, error) {
		return c.GetPersonalFile(t.Context(), &adminv1.GetPersonalFileRequest{Org: "personal-test", Project: "personal-project", Name: name})
	}

	edit := func(t *testing.T, c *client.Client, name, data string) error {
		_, err := c.EditPersonalFile(t.Context(), &adminv1.EditPersonalFileRequest{Org: "personal-test", Project: "personal-project", Name: name, Kind: runtime.ResourceKindCanvas, Yaml: data})
		return err
	}

	del := func(t *testing.T, c *client.Client, name string) error {
		_, err := c.DeletePersonalFile(t.Context(), &adminv1.DeletePersonalFileRequest{Org: "personal-test", Project: "personal-project", Name: name})
		return err
	}

	list := func(t *testing.T, c *client.Client) []string {
		resp, err := c.ListPersonalFiles(t.Context(), &adminv1.ListPersonalFilesRequest{Org: "personal-test", Project: "personal-project"})
		require.NoError(t, err)
		return resp.Files
	}

	annotations := func(t *testing.T, data string) map[string]any {
		var doc map[string]any
		require.NoError(t, yaml.Unmarshal([]byte(data), &doc))
		res, _ := doc["annotations"].(map[string]any)
		return res
	}

	embedA := embedClient(t, "user-a")
	embedB := embedClient(t, "user-b")

	t.Run("embed user can manage their own personal files", func(t *testing.T) {
		name := create(t, embedA, "")
		require.Contains(t, list(t, embedA), name)
		require.NotContains(t, list(t, embedB), name)
		require.NotContains(t, list(t, u1Client), name)

		resp, err := get(t, embedA, name)
		require.NoError(t, err)
		require.NoError(t, edit(t, embedA, name, resp.Yaml))

		require.NoError(t, del(t, embedA, name))
		require.NotContains(t, list(t, embedA), name)
	})

	t.Run("embed users can't access each other's personal files", func(t *testing.T) {
		name := create(t, embedA, "")
		_, err := get(t, embedB, name)
		require.Equal(t, codes.PermissionDenied, status.Code(err))
		require.Equal(t, codes.PermissionDenied, status.Code(edit(t, embedB, name, "rows: []\n")))
		require.Equal(t, codes.PermissionDenied, status.Code(del(t, embedB, name)))
	})

	t.Run("project members can't access embed users' personal files", func(t *testing.T) {
		name := create(t, embedA, "")
		_, err := get(t, u1Client, name)
		require.Equal(t, codes.PermissionDenied, status.Code(err))
		require.Equal(t, codes.PermissionDenied, status.Code(edit(t, u1Client, name, "rows: []\n")))
		require.Equal(t, codes.PermissionDenied, status.Code(del(t, u1Client, name)))
	})

	t.Run("embed users can't access project members' personal files", func(t *testing.T) {
		name := create(t, u1Client, "")
		_, err := get(t, embedA, name)
		require.Equal(t, codes.PermissionDenied, status.Code(err))
		require.Equal(t, codes.PermissionDenied, status.Code(edit(t, embedA, name, "rows: []\n")))
		require.Equal(t, codes.PermissionDenied, status.Code(del(t, embedA, name)))
	})

	t.Run("client-supplied admin annotations are stripped", func(t *testing.T) {
		data := "rows: []\nannotations:\n  admin_shared: \"true\"\n  admin_owner_user_id: someone\n  admin_embed_owner_id: someone\n  custom: value\n"

		name := create(t, embedA, data)
		resp, err := get(t, embedA, name)
		require.NoError(t, err)
		a := annotations(t, resp.Yaml)
		require.NotContains(t, a, "admin_shared")
		require.NotContains(t, a, "admin_owner_user_id")
		require.NotEqual(t, "someone", a["admin_embed_owner_id"])
		require.NotEmpty(t, a["admin_embed_owner_id"])
		require.Equal(t, "value", a["custom"])

		name = create(t, u1Client, data)
		resp, err = get(t, u1Client, name)
		require.NoError(t, err)
		a = annotations(t, resp.Yaml)
		require.Equal(t, "true", a["admin_shared"])
		require.NotContains(t, a, "admin_embed_owner_id")
		require.NotEqual(t, "someone", a["admin_owner_user_id"])
		require.NotEmpty(t, a["admin_owner_user_id"])
	})

	t.Run("environment overrides are rejected", func(t *testing.T) {
		data := "rows: []\nprod:\n  type: model\n  sql: SELECT 42\n"

		_, err := embedA.CreatePersonalFile(t.Context(), &adminv1.CreatePersonalFileRequest{
			Org:         "personal-test",
			Project:     "personal-project",
			Kind:        runtime.ResourceKindCanvas,
			DisplayName: "My canvas",
			Yaml:        data,
		})
		require.Equal(t, codes.InvalidArgument, status.Code(err))

		name := create(t, embedA, "")
		require.Equal(t, codes.InvalidArgument, status.Code(edit(t, embedA, name, data)))
	})

	t.Run("embeds with fixed resources are rejected", func(t *testing.T) {
		resp, err := u1Client.GetIFrame(t.Context(), &adminv1.GetIFrameRequest{
			Org:            "personal-test",
			Project:        "personal-project",
			Resource:       "c1",
			ExternalUserId: "fixed",
		})
		require.NoError(t, err)
		u, err := url.Parse(resp.IframeSrc)
		require.NoError(t, err)
		c := adm.NewClient(t, u.Query().Get("admin_token"))
		_, err = c.ListPersonalFiles(t.Context(), &adminv1.ListPersonalFilesRequest{Org: "personal-test", Project: "personal-project"})
		require.Equal(t, codes.FailedPrecondition, status.Code(err))
	})
}
