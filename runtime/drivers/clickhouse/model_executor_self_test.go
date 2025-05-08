package clickhouse_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestMaterializeType(t *testing.T) {
	truth, falsity := true, false
	cases := []struct {
		name        string
		materialize *bool
		typ         string
		wantType    string
		wantErr     bool
	}{
		{"plain", nil, "", "view", false},
		{"materialize-false", &falsity, "", "view", false},
		{"materialize-true", &truth, "", "table", false},
		{"materialize-false-view", &falsity, "view", "view", false},
		{"materialize-true-view", &truth, "view", "", true},
		{"materialize-false-table", &falsity, "table", "", true},
		{"materialize-true-table", &truth, "table", "table", false},
		{"materialize-false-dictionary", &falsity, "dictionary", "", true},
		{"materialize-true-dictionary", &truth, "dictionary", "dictionary", false},
		{"unknown-type", nil, "unknown", "", true},
		{"unknown-type-materialize-false", &falsity, "unknown", "", true},
		{"unknown-type-materialize-true", &truth, "unknown", "", true},
	}

	files := map[string]string{"rill.yaml": "olap_connector: clickhouse\n"}
	for _, c := range cases {
		data := "type: model\nsql: SELECT 1 AS id\n"
		if c.materialize != nil {
			data += fmt.Sprintf("materialize: %v\n", *c.materialize)
		}
		if c.typ != "" {
			data += fmt.Sprintf("output:\n  type: %s\n", c.typ)
			if c.typ == "dictionary" {
				data += "  primary_key: id\n"
			}
		}
		files[fmt.Sprintf("%s.yaml", c.name)] = data
	}

	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		TestConnectors: []string{"clickhouse"},
		Files:          files,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)

	for _, c := range cases {
		r := testruntime.GetResource(t, rt, id, runtime.ResourceKindModel, c.name)
		require.NotNil(t, r, c.name)
		if c.wantErr {
			require.NotEmpty(t, r.Meta.ReconcileError, c.name)
		} else {
			require.Empty(t, r.Meta.ReconcileError, c.name)
		}
		if c.wantType != "" {
			resultProps := r.GetModel().State.ResultProperties.AsMap()
			typ := strings.ToLower(resultProps["type"].(string))
			require.Equal(t, c.wantType, typ, c.name)
		}
	}
}
