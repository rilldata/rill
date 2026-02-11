package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJSONSchemaForRillYAML(t *testing.T) {
	schema, err := JSONSchemaForRillYAML()
	require.NoError(t, err)
	require.NotNil(t, schema)
	require.Equal(t, "Project YAML", schema.Title)
	require.NotEmpty(t, schema.AllOf, "schema should have properties")
}

func TestJSONSchemaForResourceType(t *testing.T) {
	tests := []struct {
		name         string
		resourceType ResourceKind
		wantTitle    string
		wantAllOf    bool
		wantErr      bool
	}{
		{
			name:         "model schema",
			resourceType: ResourceKindModel,
			wantTitle:    "Models YAML",
			wantAllOf:    true,
		},
		{
			name:         "metrics view schema",
			resourceType: ResourceKindMetricsView,
			wantTitle:    "Metrics View YAML",
			wantAllOf:    true,
		},
		{
			name:         "connector schema",
			resourceType: ResourceKindConnector,
			wantTitle:    "Connector YAML",
			wantAllOf:    true,
		},
		{
			name:         "explore schema",
			resourceType: ResourceKindExplore,
			wantTitle:    "Explore Dashboard YAML",
			wantAllOf:    true,
		},
		{
			name:         "canvas schema",
			resourceType: ResourceKindCanvas,
			wantTitle:    "Canvas Dashboard YAML",
			wantAllOf:    true,
		},
		{
			name:         "alert schema",
			resourceType: ResourceKindAlert,
			wantTitle:    "Alert YAML",
			wantAllOf:    true,
		},
		{
			name:         "theme schema",
			resourceType: ResourceKindTheme,
			wantTitle:    "Theme YAML",
			wantAllOf:    true,
		},
		{
			name:         "api schema",
			resourceType: ResourceKindAPI,
			wantTitle:    "API YAML",
			wantAllOf:    true,
		},
		{
			name:         "component schema",
			resourceType: ResourceKindComponent,
			wantTitle:    "Component YAML",
			wantAllOf:    true,
		},
		{
			name:         "unsupported resource type",
			resourceType: ResourceKindMigration,
			wantErr:      true,
		},
		{
			name:         "unspecified resource type",
			resourceType: ResourceKindUnspecified,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := JSONSchemaForResourceType(tt.resourceType)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, schema)
			require.Equal(t, tt.wantTitle, schema.Title)

			if tt.wantAllOf {
				require.NotEmpty(t, schema.AllOf, "schema should have allOf")
			}
		})
	}
}
