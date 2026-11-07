package server

import (
	"strings"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	_ "github.com/rilldata/rill/runtime/drivers/file"
	"github.com/rilldata/rill/runtime/parser"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

func TestYAMLForPersonalFileOnBehalf(t *testing.T) {
	// The caller tries to take over the file and share it.
	data := `
type: canvas
display_name: Sales
rows: []
annotations:
  admin_owner_user_id: someone-else
  admin_shared: "true"
`

	t.Run("drops the caller's annotations", func(t *testing.T) {
		out, err := yamlForPersonalFileOnBehalf(data, false)
		require.NoError(t, err)

		var doc map[string]any
		require.NoError(t, yaml.Unmarshal(out, &doc))
		require.NotContains(t, doc, "annotations")
		require.Equal(t, "Sales", doc["display_name"])
		require.False(t, personalFileShared(out))
	})

	t.Run("keeps the file's sharing state", func(t *testing.T) {
		out, err := yamlForPersonalFileOnBehalf("type: canvas\nrows: []\n", true)
		require.NoError(t, err)
		require.True(t, personalFileShared(out))
	})

	t.Run("the owner comes from the request, not the YAML", func(t *testing.T) {
		out, err := yamlForPersonalFileOnBehalf(data, false)
		require.NoError(t, err)
		final, err := yamlForPersonalFile("", "the-user", "rill.runtime.v1.Canvas", string(out))
		require.NoError(t, err)

		var doc struct {
			Annotations map[string]any `yaml:"annotations"`
		}
		require.NoError(t, yaml.Unmarshal(final, &doc))
		require.Equal(t, "the-user", doc.Annotations["admin_owner_user_id"])
		require.NotContains(t, doc.Annotations, "admin_shared")
	})

	t.Run("rejects empty or invalid YAML", func(t *testing.T) {
		_, err := yamlForPersonalFileOnBehalf("", false)
		require.Error(t, err)
		_, err = yamlForPersonalFileOnBehalf("rows: [", false)
		require.Error(t, err)
	})
}

func TestYAMLForPersonalFileOnBehalfParsed(t *testing.T) {
	// Each YAML tries to change the canvas through a key the parser honors besides the annotations.
	cases := map[string]string{
		"prod override shares":           "prod:\n  annotations:\n    admin_shared: \"true\"\n",
		"env override changes the owner": "env:\n  prod:\n    annotations:\n      admin_owner_user_id: someone-else\n",
		"environment_overrides unmanage": "environment_overrides:\n  prod:\n    annotations:\n      admin_managed: \"false\"\n",
		"prod override changes the type": "prod:\n  type: metrics_view\n",
		"name":                           "name: sales\n",
		"namespace":                      "namespace: team\n",
	}
	for name, extra := range cases {
		t.Run(name, func(t *testing.T) {
			out, err := yamlForPersonalFileOnBehalf("type: canvas\ndisplay_name: Sales\nrows: []\n"+extra, false)
			require.NoError(t, err)
			final, err := yamlForPersonalFile("", "the-user", "rill.runtime.v1.Canvas", string(out))
			require.NoError(t, err)

			// Parse it as the runtime of a project's primary deployment does.
			root := t.TempDir()
			h, err := drivers.Open(t.Context(), "file", "", "default", map[string]any{"dsn": root}, storage.MustNew(root, nil), activity.NewNoopClient(), zap.NewNop())
			require.NoError(t, err)
			repo, ok := h.AsRepoStore("")
			require.True(t, ok)
			require.NoError(t, repo.Put(t.Context(), "rill.yaml", strings.NewReader("")))
			require.NoError(t, repo.Put(t.Context(), virtualFilePathForPersonalFile("sales-abc123"), strings.NewReader(string(final))))
			p, err := parser.Parse(t.Context(), repo, "default", "prod", "duckdb", false)
			require.NoError(t, err)
			require.Empty(t, p.Errors)
			require.Len(t, p.Resources, 1)

			for _, r := range p.Resources {
				require.Equal(t, "sales-abc123", r.Name.Name)
				require.NotNil(t, r.CanvasSpec)
				require.Equal(t, map[string]string{
					"admin_owner_user_id": "the-user",
					"admin_managed":       "true",
					"admin_nonce":         r.CanvasSpec.Annotations["admin_nonce"],
				}, r.CanvasSpec.Annotations)
			}
		})
	}
}
