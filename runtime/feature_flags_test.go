package runtime

import (
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func Test_ResolveFeatureFlags(t *testing.T) {
	featureFlagTemplates := map[string]string{
		"dimension_search": `{{if eq (.user.domain) "rilldata.com"}}true{{end}}`,
		"alerts":           `'{{.user.domain}}' = 'rilldata.com'`,
		"reports":          `'{{.user.domain}}' in ['rilldata.com', 'gmail.com']`,
		"chat":             `{{not .user.embed}}`,
		"dashboard_chat":   `{{.user.embed}}`,
	}

	tests := []struct {
		name         string
		userAttrs    map[string]any
		featureFlags map[string]bool
	}{
		{
			name: "rilldata user",
			userAttrs: map[string]any{
				"domain": "rilldata.com",
			},
			featureFlags: map[string]bool{
				"exports":             true,
				"cloudDataViewer":     false,
				"dimensionSearch":     true,
				"twoTieredNavigation": false,
				"rillTime":            true,
				"hidePublicUrl":       false,
				"exportHeader":        false,
				"alerts":              true,
				"reports":             true,
				"darkMode":            true,
				"chat":                true,
				"dashboardChat":       false,
				"chatCharts":          true,
				"deploy":              true,
			},
		},
		{
			name: "gmail user",
			userAttrs: map[string]any{
				"domain": "gmail.com",
			},
			featureFlags: map[string]bool{
				"exports":             true,
				"cloudDataViewer":     false,
				"dimensionSearch":     false,
				"twoTieredNavigation": false,
				"rillTime":            true,
				"hidePublicUrl":       false,
				"exportHeader":        false,
				"alerts":              false,
				"reports":             true,
				"darkMode":            true,
				"chat":                true,
				"dashboardChat":       false,
				"chatCharts":          true,
				"deploy":              true,
			},
		},
		{
			name: "yahoo user",
			userAttrs: map[string]any{
				"domain": "yahoo.com",
			},
			featureFlags: map[string]bool{
				"exports":             true,
				"cloudDataViewer":     false,
				"dimensionSearch":     false,
				"twoTieredNavigation": false,
				"rillTime":            true,
				"hidePublicUrl":       false,
				"exportHeader":        false,
				"alerts":              false,
				"reports":             false,
				"darkMode":            true,
				"chat":                true,
				"dashboardChat":       false,
				"chatCharts":          true,
				"deploy":              true,
			},
		},
		{
			name: "embedded user",
			userAttrs: map[string]any{
				"embed": true,
			},
			featureFlags: map[string]bool{
				"exports":             true,
				"cloudDataViewer":     false,
				"dimensionSearch":     false,
				"twoTieredNavigation": false,
				"rillTime":            true,
				"hidePublicUrl":       true,
				"exportHeader":        false,
				"alerts":              false,
				"reports":             false,
				"darkMode":            true,
				"chat":                false,
				"dashboardChat":       true,
				"chatCharts":          true,
				"deploy":              true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			featureFlags, err := ResolveFeatureFlags(
				&drivers.Instance{
					FeatureFlags: featureFlagTemplates,
				},
				test.userAttrs,
				true,
			)
			require.NoError(t, err)
			require.Equal(t, test.featureFlags, featureFlags)
		})
	}
}
