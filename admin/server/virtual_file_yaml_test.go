package server

import (
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v3"
)

// TestYAMLForPersonalFileBlockedKeys tests that personal files can't set properties that change the resource's identity
// or bypass the server-enforced type and annotations through environment overrides.
func TestYAMLForPersonalFileBlockedKeys(t *testing.T) {
	tt := []struct {
		name    string
		data    string
		wantErr bool
	}{
		{name: "empty", data: ""},
		{
			name: "canvas",
			data: `
display_name: Mine
tags: [a]
filters:
  enable: true
rows:
  - items:
      - markdown:
          content: hello
annotations:
  custom: value
`,
		},
		{name: "prod type override", data: "prod:\n  type: model\n  sql: SELECT 42\n", wantErr: true},
		{name: "prod annotation override", data: "rows: []\nprod:\n  annotations:\n    admin_managed: \"false\"\n", wantErr: true},
		{name: "dev", data: "rows: []\ndev:\n  display_name: x\n", wantErr: true},
		{name: "env", data: "rows: []\nenv:\n  prod:\n    type: model\n", wantErr: true},
		{name: "environment_overrides", data: "rows: []\nenvironment_overrides:\n  prod:\n    type: model\n", wantErr: true},
		{name: "merge key", data: "base: &base\n  type: model\nprod:\n  <<: *base\n", wantErr: true},
		{name: "name", data: "rows: []\nname: other\n", wantErr: true},
		{name: "namespace", data: "rows: []\nnamespace: other\n", wantErr: true},
		{name: "kind", data: "rows: []\nkind: model\n", wantErr: true},
		{name: "refs", data: "rows: []\nrefs: [other]\n", wantErr: true},
		{name: "connector", data: "rows: []\nconnector: duckdb\n", wantErr: true},
		{name: "sql", data: "rows: []\nsql: SELECT 1\n", wantErr: true},
	}

	o := &owner{embedUserID: "ext_abc"}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			out, err := yamlForPersonalFile("", o, runtime.ResourceKindCanvas, tc.data)
			if tc.wantErr {
				require.Equal(t, codes.InvalidArgument, status.Code(err))
				return
			}
			require.NoError(t, err)

			var doc map[string]any
			require.NoError(t, yaml.Unmarshal(out, &doc))
			require.Equal(t, "canvas", doc["type"])
			annotations, _ := doc["annotations"].(map[string]any)
			require.Equal(t, "true", annotations["admin_managed"])
			require.Equal(t, "ext_abc", annotations["admin_embed_owner_id"])
		})
	}
}
