package dbt_cloud

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultBaseURL    = "https://cloud.getdbt.com"
	defaultAPIVersion = "v2"
	maxManifestSize   = 100 << 20 // 100MB
)

// Client communicates with the dbt Cloud REST API.
type Client struct {
	apiToken   string
	accountID  string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new dbt Cloud API client.
func NewClient(apiToken, accountID, baseURL string) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Client{
		apiToken:  apiToken,
		accountID: accountID,
		baseURL:   baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Run represents a dbt Cloud job run.
type Run struct {
	ID             int    `json:"id"`
	Status         int    `json:"status"`
	StatusText     string `json:"status_humanized"`
	FinishedAt     string `json:"finished_at"`
	JobID          int    `json:"job_id"`
	ArtifactsSaved bool   `json:"artifacts_saved"`
	Environment    struct {
		ID int `json:"id"`
	} `json:"environment"`
}

// runsResponse wraps the API response for listing runs.
type runsResponse struct {
	Data []Run `json:"data"`
}

// Manifest represents a parsed dbt manifest.json artifact.
type Manifest struct {
	Metadata       ManifestMetadata                  `json:"metadata"`
	Nodes          map[string]*ManifestNode          `json:"nodes"`
	Metrics        map[string]*ManifestMetric        `json:"metrics"`
	SemanticModels map[string]*ManifestSemanticModel `json:"semantic_models"`
}

// ManifestMetadata contains metadata about the manifest.
type ManifestMetadata struct {
	GeneratedAt  string `json:"generated_at"`
	InvocationID string `json:"invocation_id"`
	AdapterType  string `json:"adapter_type"` // "snowflake", "bigquery", "postgres", etc.
}

// ManifestNode represents a dbt model/seed/snapshot node in the manifest.
type ManifestNode struct {
	UniqueID       string                     `json:"unique_id"`
	Name           string                     `json:"name"`
	ResourceType   string                     `json:"resource_type"` // "model", "seed", "snapshot", etc.
	Schema         string                     `json:"schema"`
	Database       string                     `json:"database"`
	RelationName   string                     `json:"relation_name"` // fully qualified table name
	Columns        map[string]*ManifestColumn `json:"columns"`
	DependsOn      ManifestDependsOn          `json:"depends_on"`
	Description    string                     `json:"description"`
	Materialized   string                     `json:"config.materialized"`
	MaterializedAs string                     // derived from config
}

// ManifestDependsOn tracks node dependencies.
type ManifestDependsOn struct {
	Nodes []string `json:"nodes"`
}

// ManifestColumn describes a column in a dbt node.
type ManifestColumn struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	DataType    string `json:"data_type"`
}

// ManifestMetric represents a dbt metric definition.
type ManifestMetric struct {
	UniqueID    string                  `json:"unique_id"`
	Name        string                  `json:"name"`
	Label       string                  `json:"label"`
	Description string                  `json:"description"`
	Type        string                  `json:"type"` // "simple", "derived", "cumulative", "ratio"
	TypeParams  ManifestMetricTypeParam `json:"type_params"`
	DependsOn   ManifestDependsOn       `json:"depends_on"`
}

// ManifestMetricTypeParam contains type-specific parameters for a metric.
type ManifestMetricTypeParam struct {
	Measure *struct {
		Name string `json:"name"`
	} `json:"measure"`
	Expr string `json:"expr"`
}

// ManifestSemanticModel represents a dbt semantic model.
type ManifestSemanticModel struct {
	UniqueID   string                     `json:"unique_id"`
	Name       string                     `json:"name"`
	Model      string                     `json:"model"` // e.g. "ref('orders')"
	NodeRef    *NodeRelation              `json:"node_relation"`
	Entities   []SemanticEntity           `json:"entities"`
	Measures   []SemanticMeasure          `json:"measures"`
	Dimensions []SemanticDimension        `json:"dimensions"`
	Columns    map[string]*ManifestColumn `json:"columns"`
	DependsOn  ManifestDependsOn          `json:"depends_on"`
}

// NodeRelation describes the resolved table reference for a semantic model.
type NodeRelation struct {
	Alias        string `json:"alias"`
	SchemaName   string `json:"schema_name"`
	Database     string `json:"database"`
	RelationName string `json:"relation_name"`
}

// SemanticEntity is an entity defined on a semantic model.
type SemanticEntity struct {
	Name string `json:"name"`
	Type string `json:"type"` // "primary", "foreign", "unique", "natural"
	Expr string `json:"expr"`
}

// SemanticMeasure is a measure defined on a semantic model.
type SemanticMeasure struct {
	Name        string `json:"name"`
	Agg         string `json:"agg"` // "sum", "count", "avg", etc.
	Expr        string `json:"expr"`
	Description string `json:"description"`
}

// SemanticDimension is a dimension defined on a semantic model.
type SemanticDimension struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // "categorical", "time"
	Expr        string `json:"expr"`
	Description string `json:"description"`
	TypeParams  *struct {
		TimeGranularity string `json:"time_granularity"`
	} `json:"type_params"`
}

// FetchLatestRunWithArtifacts fetches the latest completed run that has saved artifacts.
// dbt generates the manifest during parsing (before execution), so even failed runs
// typically have a valid manifest artifact.
func (c *Client) FetchLatestRunWithArtifacts(ctx context.Context, environmentID string) (*Run, error) {
	params := url.Values{
		"order_by":       {"-finished_at"},
		"limit":          {"5"},
		"include_related": {"[]"},
	}
	if environmentID != "" {
		params.Set("environment_id", environmentID)
	}

	u := fmt.Sprintf("%s/api/%s/accounts/%s/runs/?%s", c.baseURL, defaultAPIVersion, c.accountID, params.Encode())

	var resp runsResponse
	if err := c.doJSON(ctx, http.MethodGet, u, &resp); err != nil {
		return nil, fmt.Errorf("failed to fetch runs: %w", err)
	}

	for i := range resp.Data {
		if resp.Data[i].ArtifactsSaved {
			return &resp.Data[i], nil
		}
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no runs found for this environment")
	}
	return nil, fmt.Errorf("no runs with saved artifacts found")
}

// FetchManifest fetches the manifest.json artifact from a specific run.
func (c *Client) FetchManifest(ctx context.Context, runID int) (*Manifest, error) {
	u := fmt.Sprintf("%s/api/%s/accounts/%s/runs/%d/artifacts/manifest.json", c.baseURL, defaultAPIVersion, c.accountID, runID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Token "+c.apiToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("dbt Cloud API returned status %d: %s", resp.StatusCode, string(body))
	}

	var manifest Manifest
	reader := io.LimitReader(resp.Body, maxManifestSize)
	if err := json.NewDecoder(reader).Decode(&manifest); err != nil {
		return nil, fmt.Errorf("failed to decode manifest: %w", err)
	}

	return &manifest, nil
}

// ListMetrics returns all metric definitions from a manifest.
func ListMetrics(manifest *Manifest) []*ManifestMetric {
	var metrics []*ManifestMetric
	for _, m := range manifest.Metrics {
		metrics = append(metrics, m)
	}
	return metrics
}

// GetOutputTable resolves a metric reference to its underlying output table.
// It traverses metric → semantic model → dbt node to find the fully qualified table name.
// Returns database, schema, and table name.
func GetOutputTable(manifest *Manifest, metricRef string) (database, schema, table string, err error) {
	// Find the metric (try both bare name and fully qualified)
	var metric *ManifestMetric
	for _, m := range manifest.Metrics {
		if m.Name == metricRef || m.UniqueID == metricRef {
			metric = m
			break
		}
	}
	if metric == nil {
		return "", "", "", fmt.Errorf("metric %q not found in manifest", metricRef)
	}

	// Find the semantic model that this metric depends on
	var semanticModel *ManifestSemanticModel
	for _, dep := range metric.DependsOn.Nodes {
		if sm, ok := manifest.SemanticModels[dep]; ok {
			semanticModel = sm
			break
		}
	}
	if semanticModel == nil {
		return "", "", "", fmt.Errorf("no semantic model found for metric %q", metricRef)
	}

	// Find the dbt node that the semantic model depends on
	for _, dep := range semanticModel.DependsOn.Nodes {
		if node, ok := manifest.Nodes[dep]; ok {
			return node.Database, node.Schema, node.Name, nil
		}
	}

	return "", "", "", fmt.Errorf("no output table found for metric %q", metricRef)
}

// doJSON makes an authenticated GET/POST request and decodes JSON response.
func (c *Client) doJSON(ctx context.Context, method, url string, result any) error {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Token "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("dbt Cloud API returned status %d: %s", resp.StatusCode, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(result)
}
