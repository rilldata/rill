package resolvers

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"
	"testing"

	"github.com/joho/godotenv"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/metricsview/metricssql"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestFileYAML is the structure of a test file.
// One runtime instance will created for each test file, with the provided file contents and connectors initialized.
// Each test in the file will be run in sequence against that runtime instance.
// The available connectors are defined in runtime/testruntime/connectors.go.
type TestFileYAML struct {
	Skip         bool                 `yaml:"skip,omitempty"`
	Expensive    bool                 `yaml:"expensive,omitempty"`
	Connectors   []string             `yaml:"connectors,omitempty"`
	Variables    map[string]string    `yaml:"variables,omitempty"`
	DataFiles    map[string]string    `yaml:"data_files,omitempty"`
	ProjectFiles map[string]yaml.Node `yaml:"project_files"`
	Tests        []*TestYAML          `yaml:"tests"`
}

// TestYAML is the structure of a single resolver test executed against the runtime instance defined in TestFileYAML.
type TestYAML struct {
	Name            string    `yaml:"name"`
	Resolver        string    `yaml:"resolver"`
	Properties      yaml.Node `yaml:"properties,omitempty"`      // Expects map[string]any, but using yaml.Node to preserve order for -update.
	Args            yaml.Node `yaml:"args,omitempty"`            // Expects map[string]any, but using yaml.Node to preserve order for -update.
	UserAttributes  yaml.Node `yaml:"user_attributes,omitempty"` // Expects map[string]any, but using yaml.Node to preserve order for -update.
	AdditionalRules []struct {
		RowFilter        string `yaml:"row_filter,omitempty"` // Expects a metrics SQL expression
		TransitiveAccess []struct {
			Kind string `yaml:"kind,omitempty"`
			Name string `yaml:"name,omitempty"`
		} `yaml:"transitive_access,omitempty"`
	} `yaml:"additional_rules,omitempty"`
	SkipSecurityChecks bool             `yaml:"skip_security_checks,omitempty"`
	Result             []map[string]any `yaml:"result,omitempty"`
	ResultCSV          string           `yaml:"result_csv,omitempty"`
	ErrorContains      string           `yaml:"error_contains,omitempty"`
}

// update is a flag that denotes whether to overwrite the results in the test files instead of checking them.
var update = flag.Bool("update", false, "Update test results")

// TestResolvers loads the test YAML files in ./testdata and runs them. Each test file should match the schema of TestFileYAML.
//
// The test YAML files provide a compact format for initializing runtime instances for a set of project files and connectors,
// and running a series of resolvers against them and testing the results.
//
// Example: run all resolver tests:
//
//	go test -run ^TestResolvers$ ./runtime/resolvers
//
// Example: update all resolver tests:
//
//	go test -run ^TestResolvers$ ./runtime/resolvers -update
//
// Example: run a single resolver test file:
//
// go test -run ^TestResolvers/metrics_clickhouse$ ./runtime/resolvers
//
// Example: run a single resolver file test case:
//
// go test -run ^TestResolvers/metrics_clickhouse/simple$ ./runtime/resolvers
func TestResolvers(t *testing.T) {
	// Evaluate the -update flag.
	flag.Parse()
	update := update != nil && *update

	// Discover the test files.
	files, err := filepath.Glob("./testdata/*.yaml")
	require.NoError(t, err)

	// Load .env file at the repo root (if any)
	_, currentFile, _, _ := goruntime.Caller(0)
	envPath := filepath.Join(currentFile, "..", "..", "..", ".env")
	_, err = os.Stat(envPath)
	if err == nil {
		require.NoError(t, godotenv.Load(envPath))
	}
	// Run each test file as a subtest.
	for _, f := range files {
		t.Run(fileutil.Stem(f), func(t *testing.T) {
			// Load the test file.
			data, err := os.ReadFile(f)
			require.NoError(t, err)
			var tf TestFileYAML
			err = yaml.Unmarshal(data, &tf)
			require.NoError(t, err)

			// Handle skip and expensive
			if tf.Skip {
				t.Skip("skipping test because it is marked skip: true")
			}
			if tf.Expensive {
				testmode.Expensive(t)
			}

			// Create a map of project files for the runtime instance.
			projectFiles := make(map[string]string)
			projectFiles["rill.yaml"] = ""
			for name, node := range tf.ProjectFiles {
				bytes, err := yaml.Marshal(&node)
				require.NoError(t, err)
				projectFiles[name] = string(bytes)
			}

			// Add local data files to project files
			for name, data := range tf.DataFiles {
				bytes, err := os.ReadFile(data)
				require.NoError(t, err)
				projectFiles[name] = string(bytes)
			}

			// Acquire the connectors for the runtime instance.
			vars := make(map[string]string)
			for k, v := range tf.Variables {
				vars[k] = v
			}
			for _, connector := range tf.Connectors {
				acquire, ok := testruntime.Connectors[connector]
				require.True(t, ok, "unknown connector %q", connector)
				cfg := acquire(t)
				for k, v := range cfg {
					k = fmt.Sprintf("connector.%s.%s", connector, k)
					vars[k] = v
				}
			}

			// Create the test runtime instance.
			rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
				Files:     projectFiles,
				Variables: vars,
			})
			testruntime.RequireReconcileState(t, rt, instanceID, -1, 0, 0)

			// Run each test case against the test runtime instance as a subtest.
			for _, tc := range tf.Tests {
				t.Run(tc.Name, func(t *testing.T) {
					// Read mapping properties that were parsed as yaml.Node to avoid reshuffling the order when using -update.
					properties := make(map[string]any)
					err := tc.Properties.Decode(&properties)
					require.NoError(t, err, "failed to decode properties into map[string]any")

					args := make(map[string]any)
					err = tc.Args.Decode(&args)
					require.NoError(t, err, "failed to decode args into map[string]any")

					userAttributes := make(map[string]any)
					err = tc.UserAttributes.Decode(&userAttributes)
					require.NoError(t, err, "failed to decode user_attributes into map[string]any")

					// Parse additional security rules.
					var additionalRules []*runtimev1.SecurityRule
					for _, r := range tc.AdditionalRules {
						// Parse row filters.
						// NOTE: Uses metrics SQL filters, not JSON filters or raw SQL filters.
						if r.RowFilter != "" {
							expr, err := metricssql.ParseFilter(r.RowFilter)
							require.NoError(t, err, "failed to parse additional rule row filter expression: %s", r.RowFilter)

							additionalRules = append(additionalRules, &runtimev1.SecurityRule{
								Rule: &runtimev1.SecurityRule_RowFilter{
									RowFilter: &runtimev1.SecurityRuleRowFilter{
										Expression: metricsview.ExpressionToProto(expr),
									},
								},
							})
						}

						// Parse transitive access rules.
						for _, ta := range r.TransitiveAccess {
							additionalRules = append(additionalRules, &runtimev1.SecurityRule{
								Rule: &runtimev1.SecurityRule_TransitiveAccess{
									TransitiveAccess: &runtimev1.SecurityRuleTransitiveAccess{
										Resource: &runtimev1.ResourceName{
											Kind: ta.Kind,
											Name: ta.Name,
										},
									},
								},
							})
						}
					}

					// Run the resolver.
					opts := &testruntime.RequireResolveOptions{
						Resolver:           tc.Resolver,
						Properties:         properties,
						Args:               args,
						UserAttributes:     userAttributes,
						AdditionalRules:    additionalRules,
						SkipSecurityChecks: tc.SkipSecurityChecks,
						Result:             tc.Result,
						ResultCSV:          tc.ResultCSV,
						ErrorContains:      tc.ErrorContains,
						Update:             update,
					}
					testruntime.RequireResolve(t, rt, instanceID, opts)

					// If the -update flag is set, update the test case results instead of checking them.
					// The updated test case will be written back to the test file later.
					if update {
						preferCSV := tc.ResultCSV != ""
						tc.Result = opts.Result
						tc.ResultCSV = opts.ResultCSV
						tc.ErrorContains = opts.ErrorContains
						if preferCSV {
							tc.Result = nil
						} else {
							tc.ResultCSV = ""
						}
						return
					}
				})
			}

			// If the -update flag is set, the TestYAML values have been updated with the output results.
			// Write out the updated test file.
			if update {
				buf := &bytes.Buffer{}
				yamlEncoder := yaml.NewEncoder(buf)
				yamlEncoder.SetIndent(2)
				err := yamlEncoder.Encode(tf)
				require.NoError(t, err)
				require.NoError(t, os.WriteFile(f, buf.Bytes(), 0644))
			}
		})
	}

}
