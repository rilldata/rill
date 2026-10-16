package ai_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestPersonalCanvasCheckAccess(t *testing.T) {
	enabled := map[string]string{"rill.yaml": "features:\n  personal_canvases: true\n"}
	disabled := map[string]string{"rill.yaml": "features:\n  personal_canvases: false\n"}

	user := map[string]any{"email": "user@example.com"}
	service := map[string]any{"name": "my-service"}

	tests := []struct {
		name        string
		files       map[string]string
		environment string
		userID      string
		attrs       map[string]any
		permissions []runtime.Permission
		want        bool
	}{
		{name: "user with AI in a project with personal canvases", files: enabled, userID: uuid.NewString(), attrs: user, permissions: []runtime.Permission{runtime.UseAI}, want: true},
		{name: "personal canvases disabled", files: disabled, userID: uuid.NewString(), attrs: user, permissions: []runtime.Permission{runtime.UseAI}, want: false},
		{name: "no user, as in Act runs", files: enabled, permissions: []runtime.Permission{runtime.UseAI}, want: false},
		{name: "service token", files: enabled, userID: uuid.NewString(), attrs: service, permissions: []runtime.Permission{runtime.UseAI}, want: false},
		{name: "missing AI permission", files: enabled, userID: uuid.NewString(), attrs: user, want: false},
		{name: "non-prod deployment", files: enabled, environment: "dev", userID: uuid.NewString(), attrs: user, permissions: []runtime.Permission{runtime.UseAI}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environment := tt.environment
			if environment == "" {
				environment = "prod"
			}
			rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{Files: tt.files, Environment: environment})
			s := newPersonalCanvasSession(t, rt, instanceID, tt.userID, tt.attrs, tt.permissions...)
			ctx := ai.WithSession(t.Context(), s)

			// develop_personal_canvas follows the user's access.
			tool, ok := s.Tool(ai.DevelopPersonalCanvasName)
			require.True(t, ok)
			allowed, err := tool.CheckAccess(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.want, allowed)

			// write_personal_canvas is never available outside develop_personal_canvas, so neither MCP clients nor the analyst see it.
			tool, ok = s.Tool(ai.WritePersonalCanvasName)
			require.True(t, ok)
			allowed, err = tool.CheckAccess(ctx)
			require.NoError(t, err)
			require.False(t, allowed)

			// Inside develop_personal_canvas, write_personal_canvas follows the user's access.
			_, err = s.Call(ctx, &ai.CallOptions{
				Role: ai.RoleAssistant,
				Name: ai.DevelopPersonalCanvasName,
				Args: json.RawMessage(`{}`),
				Handler: func(ctx context.Context) (any, error) {
					allowed, err := tool.CheckAccess(ctx)
					require.NoError(t, err)
					require.Equal(t, tt.want, allowed)
					return map[string]any{}, nil
				},
			})
			require.NoError(t, err)
		})
	}
}

func TestDevelopPersonalCanvasOnlyEditsOwnCanvases(t *testing.T) {
	owner := uuid.NewString()
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Environment: "prod",
		Files: map[string]string{
			"rill.yaml": "features:\n  personal_canvases: true\n",
			// A personal canvas as the admin service presents it in the runtime's repo.
			"__virtual__/personal/sales-1a2b3c4d.yaml": "type: canvas\nrows: []\nannotations:\n  admin_owner_user_id: " + owner + "\n",
		},
	})

	// Another user can't edit it, and the error doesn't reveal that it exists.
	s := newPersonalCanvasSession(t, rt, instanceID, uuid.NewString(), map[string]any{"email": "other@example.com"}, runtime.UseAI)
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.DevelopPersonalCanvasName, nil, &ai.DevelopPersonalCanvasArgs{
		Prompt: "Add a chart of revenue by country",
		Name:   "sales-1a2b3c4d",
	})
	require.ErrorContains(t, err, `personal canvas "sales-1a2b3c4d" not found`)

	// Same for a canvas that doesn't exist.
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.DevelopPersonalCanvasName, nil, &ai.DevelopPersonalCanvasArgs{
		Prompt: "Add a chart of revenue by country",
		Name:   "missing",
	})
	require.ErrorContains(t, err, `personal canvas "missing" not found`)
}

func TestPersonalCanvasStatus(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Environment: "prod",
		Files:       map[string]string{"rill.yaml": "features:\n  personal_canvases: true\n"},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, -1, 0, 0)
	s := newPersonalCanvasSession(t, rt, instanceID, uuid.NewString(), map[string]any{"email": "user@example.com"}, runtime.UseAI)
	ctx := ai.WithSession(t.Context(), s)

	repo, release, err := rt.Repo(ctx, instanceID)
	require.NoError(t, err)
	t.Cleanup(release)
	path := ai.PersonalCanvasPath("sales-1a2b3c4d")

	// The parser hasn't seen the file when the status is requested, as when the admin service saved it.
	// The parse error must still be reported, not an empty status that reads as success.
	require.NoError(t, repo.Put(ctx, path, strings.NewReader("type: canvas\nrows: [\n")))
	resources, parseErr, err := ai.PersonalCanvasStatus(ctx, rt, path)
	require.NoError(t, err)
	require.NotEmpty(t, parseErr)
	require.Empty(t, resources)

	// Once fixed, the error is gone and the canvas is reported.
	require.NoError(t, repo.Put(ctx, path, strings.NewReader("type: canvas\nrows: []\n")))
	resources, parseErr, err = ai.PersonalCanvasStatus(ctx, rt, path)
	require.NoError(t, err)
	require.Empty(t, parseErr)
	require.Len(t, resources, 1)
	require.Equal(t, runtime.ResourceKindCanvas, resources[0]["kind"])
	require.Empty(t, resources[0]["reconcile_error"])
}

func newPersonalCanvasSession(t *testing.T, rt *runtime.Runtime, instanceID, userID string, attrs map[string]any, permissions ...runtime.Permission) *ai.Session {
	t.Helper()
	r := ai.NewRunner(rt, activity.NewNoopClient())
	s, err := r.Session(t.Context(), &ai.SessionOptions{
		InstanceID: instanceID,
		Claims: &runtime.SecurityClaims{
			UserID:         userID,
			UserAttributes: attrs,
			Permissions:    permissions,
		},
		UserAgent: "mcp-client",
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, s.Flush(t.Context()))
	})
	return s
}
