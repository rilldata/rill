package server

import (
	"testing"

	"github.com/stretchr/testify/require"
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
