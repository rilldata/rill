package server_test

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/testadmin"
	"github.com/rilldata/rill/cli/testcli"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

// TestEmbedAlerts tests that alerts created from an embed run as the identity of the embed user's latest iframe,
// and that only the creator's saved email is marked for the open link to the embedder's app.
func TestEmbedAlerts(t *testing.T) {
	testmode.Expensive(t)

	adm := testadmin.NewWithOptionalRuntime(t, true)

	// Create test users. Note: the first user becomes a superuser, so we let the second one own the org/project to avoid confounding effects.
	_, _ = adm.NewUser(t)
	u1, u1Client := adm.NewUser(t)

	// Deploy an empty test project
	projectDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "rill.yaml"), []byte("olap_connector: duckdb\n"), 0644))
	owner := testcli.New(t, adm, u1Client.Token)
	res := owner.Run(t, "org", "create", "alert-test")
	require.Equal(t, 0, res.ExitCode, res.Output)
	res = owner.Run(t, "project", "deploy", "--interactive=false", "--org=alert-test", "--project=alert-project", "--path="+projectDir)
	require.Equal(t, 0, res.ExitCode, res.Output)
	proj, err := adm.Admin.DB.FindProjectByName(t.Context(), "alert-test", "alert-project")
	require.NoError(t, err)

	// getIFrame calls GetIFrame and returns the embed user ID of the admin token in the iframe URL.
	getIFrame := func(t *testing.T, req *adminv1.GetIFrameRequest) string {
		req.Org = "alert-test"
		req.Project = "alert-project"
		req.Resource = "e1"
		resp, err := u1Client.GetIFrame(t.Context(), req)
		require.NoError(t, err)
		u, err := url.Parse(resp.IframeSrc)
		require.NoError(t, err)
		tkn, err := adm.Admin.ValidateEmbedToken(u.Query().Get("admin_token"))
		require.NoError(t, err)
		return tkn.EmbedUserID
	}

	// createAlert creates the virtual file of an alert created from an embed, as the admin service saves it.
	// If ownerID is nil, the embed user who created the alert was deleted.
	createAlert := func(t *testing.T, name string, ownerID *string, ownerEmail string) {
		annotationOwnerID := "ext_deleted"
		if ownerID != nil {
			annotationOwnerID = *ownerID
		}
		data := fmt.Sprintf("type: alert\nannotations:\n  admin_managed: true\n  admin_embed_owner_id: %s\n  admin_embed_resources: '[{\"type\":\"rill.runtime.v1.Explore\",\"name\":\"e1\"}]'\n", annotationOwnerID)
		if ownerEmail != "" {
			data += fmt.Sprintf("  admin_embed_owner_email: %s\n", ownerEmail)
		}
		err := adm.Admin.DB.UpsertVirtualFile(t.Context(), &database.InsertVirtualFileOptions{
			ProjectID:    proj.ID,
			Environment:  "prod",
			Path:         fmt.Sprintf("alerts/%s.yaml", name),
			Data:         []byte(data),
			EmbedOwnerID: ownerID,
		})
		require.NoError(t, err)
	}

	getAlertMeta := func(t *testing.T, name string, recipients ...string) (*adminv1.GetAlertMetaResponse, error) {
		return u1Client.GetAlertMeta(t.Context(), &adminv1.GetAlertMetaRequest{
			ProjectId: proj.ID,
			Alert:     name,
			// The runtime passes the alert's annotations.
			// GetAlertMeta only checks that the embed owner annotation is set, and reads the owner from the virtual file.
			Annotations:     map[string]string{"admin_embed_owner_id": "ext_owner"},
			EmailRecipients: recipients,
		})
	}

	t.Run("runs as the embed user", func(t *testing.T) {
		attrs, err := structpb.NewStruct(map[string]any{"tenant": "t1"})
		require.NoError(t, err)
		id := getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "tenant1:user1", For: &adminv1.GetIFrameRequest_Attributes{Attributes: attrs}})
		createAlert(t, "a-attrs", &id, "")

		resp, err := getAlertMeta(t, "a-attrs", "a@customer.com", "b@customer.com")
		require.NoError(t, err)
		require.Equal(t, map[string]any{"tenant": "t1", "embed": true}, resp.QueryForAttributes.AsMap())
		require.Empty(t, resp.QueryForUserId)
		require.True(t, resp.RestrictResources)
		require.Len(t, resp.Resources, 1)
		require.Equal(t, "e1", resp.Resources[0].Name)

		// Every recipient gets an unsubscribe link but no links to Rill Cloud
		require.Len(t, resp.RecipientUrls, 2)
		for _, u := range resp.RecipientUrls {
			require.NotEmpty(t, u.UnsubscribeUrl)
			require.Empty(t, u.OpenUrl)
			require.Empty(t, u.EditUrl)
			require.False(t, u.EmbedOwner)
		}
	})

	t.Run("embed user with email of member", func(t *testing.T) {
		id := getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "member", For: &adminv1.GetIFrameRequest_UserEmail{UserEmail: u1.Email}})
		createAlert(t, "a-member", &id, "")
		resp, err := getAlertMeta(t, "a-member", "a@customer.com")
		require.NoError(t, err)
		require.Equal(t, u1.ID, resp.QueryForUserId)
		require.Equal(t, u1.Email, resp.QueryForAttributes.AsMap()["email"])
		require.Equal(t, true, resp.QueryForAttributes.AsMap()["embed"])
	})

	t.Run("only the creator's URLs are marked", func(t *testing.T) {
		id := getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "creator"})
		createAlert(t, "a-creator", &id, "Creator@customer.com")
		resp, err := getAlertMeta(t, "a-creator", "creator@customer.com", "other@customer.com")
		require.NoError(t, err)
		require.True(t, resp.RecipientUrls["creator@customer.com"].EmbedOwner)
		require.False(t, resp.RecipientUrls["other@customer.com"].EmbedOwner)

		// The creator can't unsubscribe from their own alert, so only the other recipient gets an unsubscribe link
		require.Empty(t, resp.RecipientUrls["creator@customer.com"].UnsubscribeUrl)
		require.NotEmpty(t, resp.RecipientUrls["other@customer.com"].UnsubscribeUrl)
	})

	t.Run("missing embed user fails closed", func(t *testing.T) {
		createAlert(t, "a-missing", nil, "")
		_, err := getAlertMeta(t, "a-missing", "a@customer.com")
		require.Equal(t, codes.FailedPrecondition, status.Code(err))
	})
}
