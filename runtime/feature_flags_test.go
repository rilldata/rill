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
				"exports":                         true,
				"cloudDataViewer":                 false,
				"dimensionSearch":                 true,
				"twoTieredNavigation":             false,
				"rillTime":                        true,
				"hidePublicUrl":                   false,
				"exportHeader":                    false,
				"alerts":                          true,
				"reports":                         true,
				"chat":                            true,
				"chatMemory":                      true,
				"dashboardChat":                   false,
				"developerChat":                   true,
				"chatCharts":                      true,
				"deploy":                          true,
				"developerAgent":                  true,
				"stickyDashboardState":            false,
				"cloudEditing":                    false,
				"customCharts":                    false,
				"customComponents":                false,
				"personalCanvases":                false,
				"disablePersistentDashboardState": false,
			},
		},
		{
			name: "gmail user",
			userAttrs: map[string]any{
				"domain": "gmail.com",
			},
			featureFlags: map[string]bool{
				"exports":                         true,
				"cloudDataViewer":                 false,
				"dimensionSearch":                 false,
				"twoTieredNavigation":             false,
				"rillTime":                        true,
				"hidePublicUrl":                   false,
				"exportHeader":                    false,
				"alerts":                          false,
				"reports":                         true,
				"chat":                            true,
				"chatMemory":                      true,
				"dashboardChat":                   false,
				"developerChat":                   true,
				"chatCharts":                      true,
				"deploy":                          true,
				"developerAgent":                  true,
				"stickyDashboardState":            false,
				"cloudEditing":                    false,
				"customCharts":                    false,
				"customComponents":                false,
				"personalCanvases":                false,
				"disablePersistentDashboardState": false,
			},
		},
		{
			name: "yahoo user",
			userAttrs: map[string]any{
				"domain": "yahoo.com",
			},
			featureFlags: map[string]bool{
				"exports":                         true,
				"cloudDataViewer":                 false,
				"dimensionSearch":                 false,
				"twoTieredNavigation":             false,
				"rillTime":                        true,
				"hidePublicUrl":                   false,
				"exportHeader":                    false,
				"alerts":                          false,
				"reports":                         false,
				"chat":                            true,
				"chatMemory":                      true,
				"dashboardChat":                   false,
				"developerChat":                   true,
				"chatCharts":                      true,
				"deploy":                          true,
				"developerAgent":                  true,
				"stickyDashboardState":            false,
				"cloudEditing":                    false,
				"customCharts":                    false,
				"customComponents":                false,
				"personalCanvases":                false,
				"disablePersistentDashboardState": false,
			},
		},
		{
			name: "embedded user",
			userAttrs: map[string]any{
				"embed": true,
			},
			featureFlags: map[string]bool{
				"exports":                         true,
				"cloudDataViewer":                 false,
				"dimensionSearch":                 false,
				"twoTieredNavigation":             false,
				"rillTime":                        true,
				"hidePublicUrl":                   true,
				"exportHeader":                    false,
				"alerts":                          false,
				"reports":                         false,
				"chat":                            false,
				"chatMemory":                      false, // forced false because chat is false
				"dashboardChat":                   false, // forced false because chat is false
				"developerChat":                   true,
				"chatCharts":                      true,
				"deploy":                          true,
				"developerAgent":                  true,
				"stickyDashboardState":            false,
				"cloudEditing":                    false,
				"customCharts":                    false,
				"customComponents":                false,
				"personalCanvases":                false,
				"disablePersistentDashboardState": false,
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

func Test_ResolveFeatureFlags_CloudEditingDisabledAnnotation(t *testing.T) {
	tests := []struct {
		name            string
		annotationValue string
		expected        bool
	}{
		{name: "annotation absent", expected: true},
		{name: "annotation false", annotationValue: "false", expected: true},
		{name: "annotation true", annotationValue: "true", expected: false},
		{name: "annotation true case insensitive", annotationValue: "TRUE", expected: false},
		{name: "annotation invalid", annotationValue: "invalid", expected: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			annotations := map[string]string{}
			if test.annotationValue != "" {
				annotations[CloudEditingDisabledAnnotation] = test.annotationValue
			}

			for _, camelCase := range []bool{false, true} {
				featureFlags, err := ResolveFeatureFlags(&drivers.Instance{
					FeatureFlags: map[string]string{"cloud_editing": "true"},
					Annotations:  annotations,
				}, nil, camelCase)
				require.NoError(t, err)

				key := "cloud_editing"
				if camelCase {
					key = "cloudEditing"
				}
				require.Equal(t, test.expected, featureFlags[key])
			}
		})
	}
}

// Test_ResolveFeatureFlags_ChatMemoryPrincipals checks that chat_memory is off for tokens whose subject does not identify one viewer, whatever the project configures:
// public URLs, and service tokens or embeds without an external_user_id.
func Test_ResolveFeatureFlags_ChatMemoryPrincipals(t *testing.T) {
	inst := &drivers.Instance{
		FeatureFlags: map[string]string{
			"chat":           "true",
			"dashboard_chat": "true",
			"chat_memory":    "true",
		},
	}

	tests := []struct {
		name       string
		userAttrs  map[string]any
		chatMemory bool
	}{
		{"rill user", map[string]any{"email": "user@example.com"}, true},
		{"anonymous", nil, true},
		{"embed without external user", map[string]any{"embed": true, "tenant_id": "foo"}, false},
		{"embed with external user", map[string]any{"embed": true, "external_user": true, "tenant_id": "foo"}, true},
		{"ephemeral token with external user", map[string]any{"external_user": true}, true},
		{"service token", map[string]any{"name": "my-service", "admin": true, "service": true}, false},
		{"service token embed with external user", map[string]any{"name": "my-service", "admin": true, "service": true, "embed": true, "external_user": true}, true},
		{"public URL", map[string]any{"email": "creator@example.com", "public_url": true}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			featureFlags, err := ResolveFeatureFlags(inst, test.userAttrs, true)
			require.NoError(t, err)
			require.Equal(t, test.chatMemory, featureFlags["chatMemory"])
			// Only memory is affected: chat itself stays as configured.
			require.True(t, featureFlags["chat"])
			require.True(t, featureFlags["dashboardChat"])
		})
	}
}
