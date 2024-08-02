package resolvers

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/maxatome/go-testdeep/td"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type Project struct {
	Sources    map[string]yaml.Node
	Models     map[string]yaml.Node
	Dashboards map[string]yaml.Node
	APIs       map[string]yaml.Node
}
type Resolvers struct {
	Project    Project
	Connectors map[string]*testruntime.InstanceOptions
	Tests      map[string]*Test
}

type Test struct {
	Resolver string
	Options  runtime.ResolveOptions
	Result   []map[string]any
}

func TestReolvers(t *testing.T) {
	entries, err := os.ReadDir("./")
	require.NoError(t, err)
	var reg = regexp.MustCompile(`^(.*)_resolvers_test.yaml$`)
	for _, e := range entries {
		if reg.MatchString(e.Name()) {
			fmt.Println("Running ", e.Name())
			yamlFile, err := os.ReadFile(e.Name())
			require.NoError(t, err)
			var r Resolvers
			err = yaml.Unmarshal(yamlFile, &r)
			require.NoError(t, err)

			files := make(map[string]string)
			for name, node := range r.Project.Sources {
				abs := filepath.Join("sources", name)
				bytes, err := yaml.Marshal(&node)
				require.NoError(t, err)
				files[abs] = string(bytes)

				// copy the data file to the temporary folder as well
				var props map[string]string
				require.NoError(t, node.Decode(&props))
				if props["connector"] == "local_file" {
					abs := props["path"]
					bs, err := os.ReadFile(props["path"])
					require.NoError(t, err)
					files[abs] = string(bs)
				}
			}
			for name, node := range r.Project.Models {
				abs := filepath.Join("models", name)
				var bytes []byte
				// reading SQL without '|' from YAML
				if strings.HasSuffix(name, "sql") {
					bytes = []byte(node.Value)
				} else {
					bytes, err = yaml.Marshal(&node)
					require.NoError(t, err)
				}
				files[abs] = string(bytes)

			}
			for name, node := range r.Project.Dashboards {
				abs := filepath.Join("dashboards", name)
				bytes, err := yaml.Marshal(&node)
				require.NoError(t, err)
				files[abs] = string(bytes)
			}
			for name, node := range r.Project.APIs {
				abs := filepath.Join("apis", name)
				bytes, err := yaml.Marshal(&node)
				require.NoError(t, err)
				files[abs] = string(bytes)
			}

			for ct, opts := range r.Connectors {
				fmt.Println("Running with ", ct)
				if opts == nil {
					opts = &testruntime.InstanceOptions{}
				}
				if opts.Files == nil {
					opts.Files = map[string]string{"rill.yaml": ""}
				}

				switch ct {
				case "druid":
					opts.OLAPDriver = "druid"
				case "clickhouse":
					opts.OLAPDriver = "clickhouse"
				}

				maps.Copy(opts.Files, files)
				rt, instanceID := testruntime.NewInstanceWithOptions(t, *opts)
				for testName, test := range r.Tests {
					t.Run(testName, func(t *testing.T) {
						t.Log("======================")
						t.Log("Running ", testName)
						testruntime.RequireParseErrors(t, rt, instanceID, nil)
						api, err := rt.APIForName(context.Background(), instanceID, test.Resolver)
						require.NoError(t, err)

						o := test.Options
						o.InstanceID = instanceID
						o.Resolver = api.Spec.Resolver
						o.ResolverProperties = api.Spec.ResolverProperties.AsMap()

						res, err := rt.Resolve(context.Background(), &o)
						require.NoError(t, err)
						var rows []map[string]interface{}
						require.NoError(t, json.Unmarshal(res.Data, &rows), string(res.Data))

						require.Equal(t, rows, test.Result)
						t.Log("======================")
					})
				}
			}
		}
	}
}
