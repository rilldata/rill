package server_test

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/rilldata/rill/admin/pkg/authtoken"
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

// TestGetReportMetaQueryFor tests that GetReportMeta resolves the report's owner and data access identity from owner_id and query_for.
func TestGetReportMetaQueryFor(t *testing.T) {
	testmode.Expensive(t)

	adm := testadmin.NewWithOptionalRuntime(t, true)

	// Create test users. Note: the first user becomes a superuser, so we let the second one own the org/project to avoid confounding effects.
	_, _ = adm.NewUser(t)
	u1, u1Client := adm.NewUser(t)
	u2, _ := adm.NewUser(t)

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

	const recipient = "recipient@example.com"
	const embedUser = "embed-user@customer.com"

	getReportMeta := func(t *testing.T, req *adminv1.GetReportMetaRequest) *adminv1.GetReportMetaResponse {
		req.ProjectId = proj.ID
		req.Report = "r1"
		req.ExecutionTime = timestamppb.Now()
		req.Resources = []*adminv1.ResourceName{{Type: "rill.runtime.v1.Report", Name: "r1"}}
		if req.WebOpenMode == "" {
			req.WebOpenMode = "creator"
		}
		resp, err := u1Client.GetReportMeta(t.Context(), req)
		require.NoError(t, err)
		return resp
	}

	// tokenAttrs returns the attributes and creator of the magic token in the delivery's export URL.
	tokenAttrs := func(t *testing.T, d *adminv1.GetReportMetaResponse_DeliveryMeta) (map[string]any, *string) {
		u, err := url.Parse(d.ExportUrl)
		require.NoError(t, err)
		tkn, err := authtoken.FromString(u.Query().Get("token"))
		require.NoError(t, err)
		mgc, err := adm.Admin.DB.FindMagicAuthToken(t.Context(), tkn.ID.String(), false)
		require.NoError(t, err)
		return mgc.Attributes, mgc.CreatedByUserID
	}

	t.Run("owner only", func(t *testing.T) {
		resp := getReportMeta(t, &adminv1.GetReportMetaRequest{
			OwnerId:         u1.ID,
			EmailRecipients: []string{u1.Email, recipient},
		})

		d := resp.DeliveryMeta[recipient]
		require.Equal(t, u1.ID, d.UserId)
		require.Equal(t, u1.Email, d.UserAttrs.AsMap()["email"])
		attrs, createdBy := tokenAttrs(t, d)
		require.Equal(t, u1.Email, attrs["email"])
		require.Equal(t, u1.ID, *createdBy)

		require.NotEmpty(t, resp.DeliveryMeta[u1.Email].EditUrl)
		require.Empty(t, resp.DeliveryMeta[u1.Email].UnsubscribeUrl)
	})

	t.Run("query_for_user_id takes precedence over owner", func(t *testing.T) {
		resp := getReportMeta(t, &adminv1.GetReportMetaRequest{
			OwnerId:         u1.ID,
			EmailRecipients: []string{u1.Email, recipient},
			QueryFor:        &adminv1.GetReportMetaRequest_QueryForUserId{QueryForUserId: u2.ID},
		})

		d := resp.DeliveryMeta[recipient]
		require.Equal(t, u2.ID, d.UserId)
		require.Equal(t, u2.Email, d.UserAttrs.AsMap()["email"])
		attrs, createdBy := tokenAttrs(t, d)
		require.Equal(t, u2.Email, attrs["email"])
		require.Equal(t, u1.ID, *createdBy)

		// The explicit owner still gets the edit link
		require.NotEmpty(t, resp.DeliveryMeta[u1.Email].EditUrl)
	})

	t.Run("query_for_user_email for non-Rill user without owner", func(t *testing.T) {
		resp := getReportMeta(t, &adminv1.GetReportMetaRequest{
			EmailRecipients: []string{embedUser, recipient},
			QueryFor:        &adminv1.GetReportMetaRequest_QueryForUserEmail{QueryForUserEmail: embedUser},
		})

		d := resp.DeliveryMeta[recipient]
		require.Empty(t, d.UserId)
		require.Equal(t, embedUser, d.UserAttrs.AsMap()["email"])
		attrs, createdBy := tokenAttrs(t, d)
		require.Equal(t, embedUser, attrs["email"])
		require.Nil(t, createdBy)

		// The query_for user is treated as the owner
		require.NotEmpty(t, resp.DeliveryMeta[embedUser].EditUrl)
		require.Empty(t, resp.DeliveryMeta[embedUser].UnsubscribeUrl)
	})

	t.Run("query_for_attributes", func(t *testing.T) {
		forAttrs, err := structpb.NewStruct(map[string]any{"tenant": "acme"})
		require.NoError(t, err)
		resp := getReportMeta(t, &adminv1.GetReportMetaRequest{
			OwnerId:         u1.ID,
			EmailRecipients: []string{recipient},
			QueryFor:        &adminv1.GetReportMetaRequest_QueryForAttributes{QueryForAttributes: forAttrs},
		})

		d := resp.DeliveryMeta[recipient]
		require.Empty(t, d.UserId)
		require.Equal(t, map[string]any{"tenant": "acme"}, d.UserAttrs.AsMap())
		attrs, createdBy := tokenAttrs(t, d)
		require.Equal(t, map[string]any{"tenant": "acme"}, attrs)
		require.Equal(t, u1.ID, *createdBy)
	})

	t.Run("query_for requires creator mode", func(t *testing.T) {
		_, err := u1Client.GetReportMeta(t.Context(), &adminv1.GetReportMetaRequest{
			ProjectId:       proj.ID,
			Report:          "r1",
			ExecutionTime:   timestamppb.Now(),
			EmailRecipients: []string{recipient},
			WebOpenMode:     "recipient",
			QueryFor:        &adminv1.GetReportMetaRequest_QueryForUserId{QueryForUserId: u2.ID},
		})
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})
}
