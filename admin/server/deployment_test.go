package server_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/rilldata/rill/admin/testadmin"
	"github.com/rilldata/rill/cli/testcli"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

// TestDeploymentJWTs boots a real admin + runtime, deploys a trivial project, and exercises the
// JWT-issuing admin endpoints (GetDeployment, GetDeploymentCredentials, GetIFrame) against it.
// More assertions will be layered on top of this scaffold.
func TestDeploymentJWTs(t *testing.T) {
	testmode.Expensive(t)

	adm := testadmin.NewWithOptionalRuntime(t, true)

	// Create test users. Note: the first user becomes a superuser, so we let the second one own the org/project to avoid confounding effects.
	_, _ = adm.NewUser(t)
	u1, u1Client := adm.NewUser(t)

	// Create empty test project
	projectDir := t.TempDir()
	require.NoError(t, os.MkdirAll(projectDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "rill.yaml"), []byte("olap_connector: duckdb\n"), 0644))

	// Deploy the test project.
	// NOTE: Using testcli since it's what we've got. TODO: Move to direct API calls when we have better test utilities here.
	owner := testcli.New(t, adm, u1Client.Token)
	res := owner.Run(t, "org", "create", "jwt-test")
	require.Equal(t, 0, res.ExitCode, res.Output)
	res = owner.Run(t, "project", "deploy", "--interactive=false", "--org=jwt-test", "--project=jwt-project", "--path="+projectDir)
	require.Equal(t, 0, res.ExitCode, res.Output)
	depl := adm.TriggerDeployment(t, "jwt-test", "jwt-project")
	require.NotEmpty(t, depl.RuntimeInstanceID)
	require.Equal(t, "prod", depl.Environment)

	t.Run("GetProject for project admin", func(t *testing.T) {
		resp, err := u1Client.GetProject(t.Context(), &adminv1.GetProjectRequest{
			Org:     "jwt-test",
			Project: "jwt-project",
		})
		require.NoError(t, err)

		c := decodeJWT(t, resp.Jwt)
		require.Equal(t, u1.ID, c.Subject)
		require.Equal(t, u1.Email, c.Attrs["email"])
		require.Equal(t, true, c.Attrs["admin"])
		require.NotContains(t, c.Attrs, "service")
		require.NotContains(t, c.Attrs, "public_url")

		perms := c.Instances[depl.RuntimeInstanceID]
		requireHasPerms(t, perms, runtime.ReadObjects, runtime.ReadMetrics, runtime.ReadAPI, runtime.UseAI)
		requireHasPerms(t, perms, runtime.ReadInstance, runtime.ReadResolvers, runtime.EditTrigger)
		// Prod deployment is non-editable, so no edit-repo permissions.
		requireLacksPerms(t, perms, runtime.EditRepo, runtime.ReadOLAP, runtime.ReadProfiling, runtime.ReadRepo)
	})

	t.Run("GetDeployment for project admin", func(t *testing.T) {
		resp, err := u1Client.GetDeployment(t.Context(), &adminv1.GetDeploymentRequest{
			DeploymentId: depl.ID,
		})
		require.NoError(t, err)

		c := decodeJWT(t, resp.AccessToken)
		require.Equal(t, u1.ID, c.Subject)
		require.Equal(t, u1.Email, c.Attrs["email"])
		require.Equal(t, true, c.Attrs["admin"])

		perms := c.Instances[depl.RuntimeInstanceID]
		requireHasPerms(t, perms, runtime.ReadInstance, runtime.ReadResolvers, runtime.EditTrigger)
	})

	t.Run("GetDeployment for non-Rill email", func(t *testing.T) {
		email := "non-rill@example.com"

		resp, err := u1Client.GetDeployment(t.Context(), &adminv1.GetDeploymentRequest{
			DeploymentId: depl.ID,
			For:          &adminv1.GetDeploymentRequest_UserEmail{UserEmail: email},
		})
		require.NoError(t, err)

		c := decodeJWT(t, resp.AccessToken)
		require.Empty(t, c.Subject)
		require.Equal(t, email, c.Attrs["email"])
		require.Equal(t, false, c.Attrs["admin"])

		perms := c.Instances[depl.RuntimeInstanceID]
		requireHasPerms(t, perms, runtime.ReadObjects, runtime.ReadMetrics, runtime.ReadAPI, runtime.UseAI)
		requireLacksPerms(t, perms, runtime.EditTrigger, runtime.ReadInstance)
	})

	// GetDeployment with external user ID and custom attributes
	t.Run("GetDeployment for external user ID and custom attributes", func(t *testing.T) {
		extID := "external-customer-42"
		customAttrs, err := structpb.NewStruct(map[string]any{
			"email":     "embed@example.com",
			"tenant_id": "foo",
		})
		require.NoError(t, err)

		resp, err := u1Client.GetDeployment(t.Context(), &adminv1.GetDeploymentRequest{
			DeploymentId:   depl.ID,
			ExternalUserId: extID,
			For:            &adminv1.GetDeploymentRequest_Attributes{Attributes: customAttrs},
		})
		require.NoError(t, err)

		c := decodeJWT(t, resp.AccessToken)
		require.True(t, strings.HasPrefix(c.Subject, "ext_"))
		require.NotEqual(t, u1.ID, c.Subject)
		require.Equal(t, "embed@example.com", c.Attrs["email"])
		require.Equal(t, "foo", c.Attrs["tenant_id"])
		require.Equal(t, true, c.Attrs["external_user"])
		require.NotContains(t, c.Attrs, "admin")

		// Subject hash must be deterministic across calls with the same external id and project.
		resp2, err := u1Client.GetDeployment(t.Context(), &adminv1.GetDeploymentRequest{
			DeploymentId:   depl.ID,
			ExternalUserId: extID,
			For:            &adminv1.GetDeploymentRequest_Attributes{Attributes: customAttrs},
		})
		require.NoError(t, err)
		require.Equal(t, c.Subject, decodeJWT(t, resp2.AccessToken).Subject)
	})

	t.Run("Service token", func(t *testing.T) {
		svc, err := u1Client.CreateService(t.Context(), &adminv1.CreateServiceRequest{
			Org:             "jwt-test",
			Name:            "embed-service",
			Project:         "jwt-project",
			ProjectRoleName: "admin",
		})
		require.NoError(t, err)
		tok, err := u1Client.IssueServiceAuthToken(t.Context(), &adminv1.IssueServiceAuthTokenRequest{
			Org:         "jwt-test",
			ServiceName: "embed-service",
		})
		require.NoError(t, err)

		svcClient := adm.NewClient(t, tok.Token)

		// The service's own token has the service ID as subject, shared by every caller, and is marked as a service.
		proj, err := svcClient.GetProject(t.Context(), &adminv1.GetProjectRequest{
			Org:     "jwt-test",
			Project: "jwt-project",
		})
		require.NoError(t, err)
		c := decodeJWT(t, proj.Jwt)
		require.Equal(t, svc.Service.Id, c.Subject)
		require.Equal(t, true, c.Attrs["service"])
		require.NotContains(t, c.Attrs, "embed")
		require.NotContains(t, c.Attrs, "external_user")
		requireHasPerms(t, c.Instances[depl.RuntimeInstanceID], runtime.UseAI)

		// An embed without an external user ID keeps the service ID as subject for every viewer, and nothing marks the token as identifying one viewer.
		resp, err := svcClient.GetIFrame(t.Context(), &adminv1.GetIFrameRequest{
			Org:      "jwt-test",
			Project:  "jwt-project",
			Resource: "foo",
		})
		require.NoError(t, err)
		c = decodeJWT(t, resp.AccessToken)
		require.Equal(t, svc.Service.Id, c.Subject)
		require.Equal(t, true, c.Attrs["embed"])
		require.Equal(t, true, c.Attrs["service"])
		require.NotContains(t, c.Attrs, "external_user")
		require.NotContains(t, c.Attrs, "public_url")
		requireHasPerms(t, c.Instances[depl.RuntimeInstanceID], runtime.UseAI)

		// With an external user ID, each viewer gets their own subject and the token says so. The service itself is no longer the principal.
		resp, err = svcClient.GetIFrame(t.Context(), &adminv1.GetIFrameRequest{
			Org:            "jwt-test",
			Project:        "jwt-project",
			Resource:       "foo",
			ExternalUserId: "external-customer-42",
		})
		require.NoError(t, err)
		c = decodeJWT(t, resp.AccessToken)
		require.True(t, strings.HasPrefix(c.Subject, "ext_"))
		require.Equal(t, true, c.Attrs["embed"])
		require.Equal(t, true, c.Attrs["external_user"])
		require.NotContains(t, c.Attrs, "service")

		// The markers are derived by the server: custom attributes cannot claim a per-viewer subject or hide the shared one.
		spoofed, err := structpb.NewStruct(map[string]any{
			"external_user": true,
			"service":       false,
			"public_url":    false,
			"tenant_id":     "foo",
		})
		require.NoError(t, err)
		resp, err = svcClient.GetIFrame(t.Context(), &adminv1.GetIFrameRequest{
			Org:      "jwt-test",
			Project:  "jwt-project",
			Resource: "foo",
			For:      &adminv1.GetIFrameRequest_Attributes{Attributes: spoofed},
		})
		require.NoError(t, err)
		c = decodeJWT(t, resp.AccessToken)
		require.Empty(t, c.Subject)
		require.Equal(t, "foo", c.Attrs["tenant_id"])
		require.Equal(t, true, c.Attrs["embed"])
		require.NotContains(t, c.Attrs, "external_user")
		require.NotContains(t, c.Attrs, "service")
		require.NotContains(t, c.Attrs, "public_url")
	})

	t.Run("GetProject for public URL", func(t *testing.T) {
		tkn, err := u1Client.IssueMagicAuthToken(t.Context(), &adminv1.IssueMagicAuthTokenRequest{
			Org:       "jwt-test",
			Project:   "jwt-project",
			Resources: []*adminv1.ResourceName{{Type: runtime.ResourceKindExplore, Name: "foo"}},
		})
		require.NoError(t, err)

		resp, err := adm.NewClient(t, tkn.Token).GetProject(t.Context(), &adminv1.GetProjectRequest{
			Org:     "jwt-test",
			Project: "jwt-project",
		})
		require.NoError(t, err)

		// Everyone with the link shares the magic token ID as subject and the creator's attributes, so the token is marked as a public URL.
		c := decodeJWT(t, resp.Jwt)
		require.NotEmpty(t, c.Subject)
		require.NotEqual(t, u1.ID, c.Subject)
		require.Equal(t, u1.Email, c.Attrs["email"])
		require.Equal(t, true, c.Attrs["public_url"])
		require.NotContains(t, c.Attrs, "embed")
		requireHasPerms(t, c.Instances[depl.RuntimeInstanceID], runtime.UseAI)
	})
}

// jwtPayload mirrors the subset of the runtime JWT payload that the assertions touch.
type jwtPayload struct {
	jwt.RegisteredClaims
	System    []runtime.Permission            `json:"sys,omitempty"`
	Instances map[string][]runtime.Permission `json:"ins,omitempty"`
	Attrs     map[string]any                  `json:"attr,omitempty"`
	Security  []json.RawMessage               `json:"sec,omitempty"`
}

func decodeJWT(t *testing.T, tok string) *jwtPayload {
	t.Helper()
	c := &jwtPayload{}
	_, _, err := jwt.NewParser().ParseUnverified(tok, c)
	require.NoError(t, err)
	return c
}

func requireHasPerms(t *testing.T, got []runtime.Permission, want ...runtime.Permission) {
	t.Helper()
	for _, p := range want {
		require.True(t, slices.Contains(got, p), "expected permission %v in %v", p, got)
	}
}

func requireLacksPerms(t *testing.T, got []runtime.Permission, unwanted ...runtime.Permission) {
	t.Helper()
	for _, p := range unwanted {
		require.False(t, slices.Contains(got, p), "did not expect permission %v in %v", p, got)
	}
}
