package dbt_cloud

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFetchLatestRunWithArtifacts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Token test-token", r.Header.Get("Authorization"))
		require.Contains(t, r.URL.Path, "/api/v2/accounts/12345/runs/")

		resp := runsResponse{
			Data: []Run{
				{ID: 100, Status: 20, StatusText: "Error", ArtifactsSaved: true, FinishedAt: "2024-01-01T00:00:00Z"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := NewClient("test-token", "12345", srv.URL)
	run, err := client.FetchLatestRunWithArtifacts(context.Background(), "67890")
	require.NoError(t, err)
	require.Equal(t, 100, run.ID)
	require.True(t, run.ArtifactsSaved)
}

func TestFetchManifest(t *testing.T) {
	manifest := &Manifest{
		Metadata: ManifestMetadata{
			AdapterType: "snowflake",
		},
		Nodes: map[string]*ManifestNode{
			"model.project.orders": {
				UniqueID: "model.project.orders",
				Name:     "orders",
				Schema:   "analytics",
				Database: "raw",
			},
		},
		Metrics: map[string]*ManifestMetric{
			"metric.project.revenue": {
				UniqueID: "metric.project.revenue",
				Name:     "revenue",
				Label:    "Total Revenue",
				Type:     "simple",
				DependsOn: ManifestDependsOn{
					Nodes: []string{"semantic_model.project.orders"},
				},
			},
		},
		SemanticModels: map[string]*ManifestSemanticModel{
			"semantic_model.project.orders": {
				UniqueID: "semantic_model.project.orders",
				Name:     "orders",
				DependsOn: ManifestDependsOn{
					Nodes: []string{"model.project.orders"},
				},
			},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Token test-token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(manifest)
	}))
	defer srv.Close()

	client := NewClient("test-token", "12345", srv.URL)
	result, err := client.FetchManifest(context.Background(), 100)
	require.NoError(t, err)
	require.Equal(t, "snowflake", result.Metadata.AdapterType)
	require.Len(t, result.Nodes, 1)
	require.Len(t, result.Metrics, 1)
	require.Len(t, result.SemanticModels, 1)
}

func TestGetOutputTable(t *testing.T) {
	manifest := &Manifest{
		Nodes: map[string]*ManifestNode{
			"model.project.orders": {
				UniqueID: "model.project.orders",
				Name:     "orders",
				Schema:   "analytics",
				Database: "raw",
			},
		},
		Metrics: map[string]*ManifestMetric{
			"metric.project.revenue": {
				UniqueID: "metric.project.revenue",
				Name:     "revenue",
				DependsOn: ManifestDependsOn{
					Nodes: []string{"semantic_model.project.orders"},
				},
			},
		},
		SemanticModels: map[string]*ManifestSemanticModel{
			"semantic_model.project.orders": {
				UniqueID: "semantic_model.project.orders",
				Name:     "orders",
				DependsOn: ManifestDependsOn{
					Nodes: []string{"model.project.orders"},
				},
			},
		},
	}

	// Test by name
	db, schema, table, err := GetOutputTable(manifest, "revenue")
	require.NoError(t, err)
	require.Equal(t, "raw", db)
	require.Equal(t, "analytics", schema)
	require.Equal(t, "orders", table)

	// Test by unique ID
	db, schema, table, err = GetOutputTable(manifest, "metric.project.revenue")
	require.NoError(t, err)
	require.Equal(t, "raw", db)
	require.Equal(t, "analytics", schema)
	require.Equal(t, "orders", table)

	// Test not found
	_, _, _, err = GetOutputTable(manifest, "nonexistent")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestListMetrics(t *testing.T) {
	manifest := &Manifest{
		Metrics: map[string]*ManifestMetric{
			"metric.project.revenue": {Name: "revenue"},
			"metric.project.orders":  {Name: "orders"},
		},
	}

	metrics := ListMetrics(manifest)
	require.Len(t, metrics, 2)
}
