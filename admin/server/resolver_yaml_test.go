package server

import (
	"testing"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/yaml.v3"
)

// TestYAMLForResolverProperties checks that alerts and reports written with a resolver keep its properties as plain YAML,
// so the runtime parses the same properties it was given.
func TestYAMLForResolverProperties(t *testing.T) {
	s := &Server{}
	props, err := structpb.NewStruct(map[string]any{
		"sql":   "SELECT 1 FROM mv1",
		"limit": 10,
	})
	require.NoError(t, err)
	want := map[string]any{"metrics_sql": map[string]any{"sql": "SELECT 1 FROM mv1", "limit": 10}}

	alert := &adminv1.AlertOptions{
		DisplayName:        "Alert",
		MetricsViewName:    "mv1",
		Resolver:           "metrics_sql",
		ResolverProperties: props,
		EmailRecipients:    []string{"user_1@example.com"},
	}
	report := &adminv1.ReportOptions{
		DisplayName:        "Report",
		RefreshCron:        "0 9 * * 1",
		Resolver:           "metrics_sql",
		ResolverProperties: props,
		EmailRecipients:    []string{"user_1@example.com"},
		Explore:            "e1",
		WebOpenMode:        string(WebOpenModeRecipient),
	}

	generators := map[string]func() ([]byte, error){
		"managed alert":    func() ([]byte, error) { return s.yamlForManagedAlert(alert, "owner") },
		"committed alert":  func() ([]byte, error) { return s.yamlForCommittedAlert(alert) },
		"managed report":   func() ([]byte, error) { return s.yamlForManagedReport(report, "owner", "", "", nil) },
		"committed report": func() ([]byte, error) { return s.yamlForCommittedReport(report) },
	}
	for name, generate := range generators {
		t.Run(name, func(t *testing.T) {
			data, err := generate()
			require.NoError(t, err)
			var res struct {
				Data map[string]any `yaml:"data"`
			}
			require.NoError(t, yaml.Unmarshal(data, &res))
			require.Equal(t, want, res.Data)
		})
	}
}
