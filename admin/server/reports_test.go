package server_test

import (
	"crypto/sha256"
	"encoding/hex"
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
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TestEmbedReports tests that GetIFrame stores the identity of embed users with an external user ID,
// and that reports created from an embed run as that identity.
func TestEmbedReports(t *testing.T) {
	testmode.Expensive(t)

	adm := testadmin.NewWithOptionalRuntime(t, true)

	// Create test users. Note: the first user becomes a superuser, so we let the second one own the org/project to avoid confounding effects.
	_, _ = adm.NewUser(t)
	u1, u1Client := adm.NewUser(t)
	_, u2Client := adm.NewUser(t)

	// Deploy an empty test project
	projectDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "rill.yaml"), []byte("olap_connector: duckdb\n"), 0644))
	owner := testcli.New(t, adm, u1Client.Token)
	res := owner.Run(t, "org", "create", "report-test")
	require.Equal(t, 0, res.ExitCode, res.Output)
	res = owner.Run(t, "project", "deploy", "--interactive=false", "--org=report-test", "--project=report-project", "--path="+projectDir)
	require.Equal(t, 0, res.ExitCode, res.Output)
	proj, err := adm.Admin.DB.FindProjectByName(t.Context(), "report-test", "report-project")
	require.NoError(t, err)

	// embedUserID mirrors the hash the admin service uses to identify external users.
	embedUserID := func(externalUserID string) string {
		hash := sha256.Sum256([]byte(externalUserID + proj.ID))
		return "ext_" + hex.EncodeToString(hash[:])
	}

	// getIFrame calls GetIFrame and returns the admin token in the iframe URL (empty if none).
	getIFrame := func(t *testing.T, req *adminv1.GetIFrameRequest) string {
		req.Org = "report-test"
		req.Project = "report-project"
		req.Resource = "e1"
		resp, err := u1Client.GetIFrame(t.Context(), req)
		require.NoError(t, err)
		u, err := url.Parse(resp.IframeSrc)
		require.NoError(t, err)
		return u.Query().Get("admin_token")
	}

	attributes := func(t *testing.T, attrs map[string]any) *adminv1.GetIFrameRequest_Attributes {
		s, err := structpb.NewStruct(attrs)
		require.NoError(t, err)
		return &adminv1.GetIFrameRequest_Attributes{Attributes: s}
	}

	// createReport creates the virtual file of a report created from an embed.
	createReport := func(t *testing.T, name string, ownerID *string) {
		annotationOwnerID := embedUserID("deleted")
		if ownerID != nil {
			annotationOwnerID = *ownerID
		}
		err := adm.Admin.DB.UpsertVirtualFile(t.Context(), &database.InsertVirtualFileOptions{
			ProjectID:    proj.ID,
			Environment:  "prod",
			Path:         fmt.Sprintf("reports/%s.yaml", name),
			Data:         fmt.Appendf(nil, "type: report\nannotations:\n  admin_managed: true\n  admin_embed_owner_id: %s\n  web_open_mode: creator\n", annotationOwnerID),
			EmbedOwnerID: ownerID,
		})
		require.NoError(t, err)
	}

	getReportMeta := func(t *testing.T, name string, recipients ...string) (*adminv1.GetReportMetaResponse, error) {
		return u1Client.GetReportMeta(t.Context(), &adminv1.GetReportMetaRequest{
			ProjectId:       proj.ID,
			Report:          name,
			ExecutionTime:   timestamppb.Now(),
			EmailRecipients: recipients,
			WebOpenMode:     "creator",
		})
	}

	t.Run("no external user ID", func(t *testing.T) {
		tkn := getIFrame(t, &adminv1.GetIFrameRequest{For: &adminv1.GetIFrameRequest_UserEmail{UserEmail: "someone@customer.com"}})
		require.Empty(t, tkn)
	})

	t.Run("user ID with external user ID is rejected", func(t *testing.T) {
		_, err := u1Client.GetIFrame(t.Context(), &adminv1.GetIFrameRequest{
			Org:            "report-test",
			Project:        "report-project",
			Resource:       "e1",
			ExternalUserId: "rejected",
			For:            &adminv1.GetIFrameRequest_UserId{UserId: u1.ID},
		})
		require.Error(t, err)
		_, err = adm.Admin.DB.FindEmbedUser(t.Context(), embedUserID("rejected"))
		require.ErrorIs(t, err, database.ErrNotFound)
	})

	t.Run("external user ID with attributes", func(t *testing.T) {
		id := embedUserID("tenant1:user1")
		tkn := getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "tenant1:user1", For: attributes(t, map[string]any{"tenant": "t1"})})
		require.NotEmpty(t, tkn)
		embedTkn, err := adm.Admin.ValidateEmbedToken(tkn)
		require.NoError(t, err)
		require.Equal(t, id, embedTkn.EmbedUserID)
		require.Equal(t, []database.ResourceName{{Type: "rill.runtime.v1.Explore", Name: "e1"}}, embedTkn.Resources)

		u, err := adm.Admin.DB.FindEmbedUser(t.Context(), id)
		require.NoError(t, err)
		require.Equal(t, "tenant1:user1", u.ExternalUserID)
		require.Equal(t, map[string]any{"tenant": "t1"}, u.Attributes)

		// The report runs as the embed user, and every recipient gets an unsubscribe link but no other links
		createReport(t, "r-attrs", &id)
		resp, err := getReportMeta(t, "r-attrs", "a@customer.com", "b@customer.com")
		require.NoError(t, err)
		require.Len(t, resp.DeliveryMeta, 2)
		for _, d := range resp.DeliveryMeta {
			require.Empty(t, d.UserId)
			require.Equal(t, map[string]any{"tenant": "t1", "embed": true}, d.UserAttrs.AsMap())
			require.NotEmpty(t, d.UnsubscribeUrl)
			require.Empty(t, d.OpenUrl)
			require.Empty(t, d.ExportUrl)
			require.Empty(t, d.EditUrl)
		}

		// A later iframe with different attributes updates the identity the report runs as
		getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "tenant1:user1", For: attributes(t, map[string]any{"tenant": "t2"})})
		resp, err = getReportMeta(t, "r-attrs", "a@customer.com")
		require.NoError(t, err)
		require.Equal(t, map[string]any{"tenant": "t2", "embed": true}, resp.DeliveryMeta["a@customer.com"].UserAttrs.AsMap())

		// The lookup API returns the external user ID to users who can manage reports
		eu, err := u1Client.GetEmbedUser(t.Context(), &adminv1.GetEmbedUserRequest{Org: "report-test", Project: "report-project", Id: id})
		require.NoError(t, err)
		require.Equal(t, "tenant1:user1", eu.ExternalUserId)
		_, err = u2Client.GetEmbedUser(t.Context(), &adminv1.GetEmbedUserRequest{Org: "report-test", Project: "report-project", Id: id})
		require.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	t.Run("embed attribute supplied by the embedder", func(t *testing.T) {
		id := embedUserID("custom-embed")
		getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "custom-embed", For: attributes(t, map[string]any{"embed": "custom"})})
		createReport(t, "r-custom-embed", &id)
		resp, err := getReportMeta(t, "r-custom-embed", "a@customer.com")
		require.NoError(t, err)
		require.Equal(t, map[string]any{"embed": "custom"}, resp.DeliveryMeta["a@customer.com"].UserAttrs.AsMap())
	})

	t.Run("external user ID without identity", func(t *testing.T) {
		id := embedUserID("anonymous")
		require.NotEmpty(t, getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "anonymous"}))
		createReport(t, "r-anonymous", &id)
		resp, err := getReportMeta(t, "r-anonymous", "a@customer.com")
		require.NoError(t, err)
		require.Equal(t, map[string]any{"embed": true}, resp.DeliveryMeta["a@customer.com"].UserAttrs.AsMap())
	})

	t.Run("external user ID with email of non-member", func(t *testing.T) {
		id := embedUserID("non-member")
		getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "non-member", For: &adminv1.GetIFrameRequest_UserEmail{UserEmail: "someone@customer.com"}})
		createReport(t, "r-non-member", &id)
		resp, err := getReportMeta(t, "r-non-member", "a@customer.com")
		require.NoError(t, err)
		d := resp.DeliveryMeta["a@customer.com"]
		require.Empty(t, d.UserId)
		require.Equal(t, map[string]any{"email": "someone@customer.com", "domain": "customer.com", "admin": false, "embed": true}, d.UserAttrs.AsMap())
	})

	t.Run("external user ID with email of member", func(t *testing.T) {
		id := embedUserID("member")
		getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "member", For: &adminv1.GetIFrameRequest_UserEmail{UserEmail: u1.Email}})
		createReport(t, "r-member", &id)
		resp, err := getReportMeta(t, "r-member", "a@customer.com")
		require.NoError(t, err)
		d := resp.DeliveryMeta["a@customer.com"]
		require.Equal(t, u1.ID, d.UserId)
		require.Equal(t, u1.Email, d.UserAttrs.AsMap()["email"])
		require.Equal(t, true, d.UserAttrs.AsMap()["embed"])
	})

	t.Run("missing embed user fails closed", func(t *testing.T) {
		createReport(t, "r-missing", nil)
		_, err := getReportMeta(t, "r-missing", "a@customer.com")
		require.Equal(t, codes.FailedPrecondition, status.Code(err))
	})
}
