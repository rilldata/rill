package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/mark3labs/mcp-go/mcp"
	"go.uber.org/zap"
)

// Custom error types
type (
	AdminAPIError struct {
		StatusCode int
		Message    string
		Err        error
	}

	OpenAIError struct {
		Message string
		Err     error
	}

	ChartGenerationError struct {
		Message string
		Err     error
	}

	ValidationError struct {
		Field   string
		Message string
	}
)

func (e *AdminAPIError) Error() string {
	return fmt.Sprintf("admin API error (status %d): %s: %v", e.StatusCode, e.Message, e.Err)
}

func (e *OpenAIError) Error() string {
	return fmt.Sprintf("OpenAI error: %s: %v", e.Message, e.Err)
}

func (e *ChartGenerationError) Error() string {
	return fmt.Sprintf("chart generation error: %s: %v", e.Message, e.Err)
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error in field %s: %s", e.Field, e.Message)
}

// Config holds the server configuration
type Config struct {
	// AdminBaseURL is the base URL for the admin API
	AdminBaseURL string `json:"admin_base_url"`
	// OrganizationName is the name of the organization
	OrganizationName string `json:"organization_name"`
	// ProjectName is the name of the project
	ProjectName string `json:"project_name"`
	// ServiceToken is the token for authenticating with the admin API
	ServiceToken string `json:"service_token"`
	// OpenAIAPIKey is the API key for OpenAI
	OpenAIAPIKey string `json:"openai_api_key"`
	// EnableVisualization enables the visualization server
	EnableVisualization bool `json:"enable_visualization"`
	// LogLevel is the logging level (debug, info, warn, error)
	LogLevel string `json:"log_level"`
	// RequestTimeout is the timeout for API requests
	RequestTimeout time.Duration `json:"request_timeout"`
	// MaxRetries is the maximum number of retries for failed requests
	MaxRetries int `json:"max_retries"`
	// RetryBackoff is the backoff duration between retries
	RetryBackoff time.Duration `json:"retry_backoff"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		AdminBaseURL:        "https://admin.rilldata.com",
		LogLevel:            "info",
		RequestTimeout:      30 * time.Second,
		MaxRetries:          3,
		RetryBackoff:        time.Second,
		EnableVisualization: true,
	}
}

// loadConfig loads configuration from environment variables with fallbacks
func loadConfig() (*Config, error) {
	config := DefaultConfig()

	// Load from environment variables
	if value := os.Getenv("RILL_ADMIN_BASE_URL"); value != "" {
		config.AdminBaseURL = value
	}
	if value := os.Getenv("RILL_ORGANIZATION_NAME"); value != "" {
		config.OrganizationName = value
	}
	if value := os.Getenv("RILL_PROJECT_NAME"); value != "" {
		config.ProjectName = value
	}
	if value := os.Getenv("RILL_SERVICE_TOKEN"); value != "" {
		config.ServiceToken = value
	}
	if value := os.Getenv("OPENAI_API_KEY"); value != "" {
		config.OpenAIAPIKey = value
	}
	if value := os.Getenv("RILL_LOG_LEVEL"); value != "" {
		config.LogLevel = value
	}
	if value := os.Getenv("RILL_REQUEST_TIMEOUT"); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			config.RequestTimeout = duration
		}
	}
	if value := os.Getenv("RILL_MAX_RETRIES"); value != "" {
		if retries, err := strconv.Atoi(value); err == nil {
			config.MaxRetries = retries
		}
	}
	if value := os.Getenv("RILL_RETRY_BACKOFF"); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			config.RetryBackoff = duration
		}
	}
	if value := os.Getenv("RILL_ENABLE_VISUALIZATION"); value != "" {
		config.EnableVisualization = value != "false"
	}

	// Validate required fields
	if config.OrganizationName == "" {
		return nil, fmt.Errorf("RILL_ORGANIZATION_NAME is required")
	}
	if config.ProjectName == "" {
		return nil, fmt.Errorf("RILL_PROJECT_NAME is required")
	}
	if config.ServiceToken == "" {
		return nil, fmt.Errorf("RILL_SERVICE_TOKEN is required")
	}
	if config.EnableVisualization && config.OpenAIAPIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is required when visualization is enabled")
	}

	return config, nil
}

// MCPServer represents the MCP server
type MCPServer struct {
	config       *Config
	adminClient  *http.Client
	openaiClient *http.Client
	logger       *zap.Logger
}

// NewMCPServer creates a new MCP server instance
func NewMCPServer() (*MCPServer, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Configure logger
	var logger *zap.Logger
	switch config.LogLevel {
	case "debug":
		logger, err = zap.NewDevelopment()
	case "info":
		logger, err = zap.NewProduction()
	case "warn":
		config := zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
		logger, err = config.Build()
	case "error":
		config := zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
		logger, err = config.Build()
	default:
		logger, err = zap.NewProduction()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Configure HTTP clients
	adminClient := &http.Client{
		Timeout: config.RequestTimeout,
	}
	openaiClient := &http.Client{
		Timeout: config.RequestTimeout,
	}

	return &MCPServer{
		config:       config,
		adminClient:  adminClient,
		openaiClient: openaiClient,
		logger:       logger,
	}, nil
}

// getEnvOrDefault gets an environment variable with a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// callAdminAPI makes a request to the admin API with retries
func (s *MCPServer) callAdminAPI(ctx context.Context, method, endpoint string, body io.Reader) (map[string]interface{}, error) {
	url := s.config.AdminBaseURL + endpoint
	s.logger.Info("calling admin API",
		zap.String("method", method),
		zap.String("endpoint", endpoint))

	var lastErr error
	for i := 0; i <= s.config.MaxRetries; i++ {
		if i > 0 {
			s.logger.Info("retrying admin API call",
				zap.Int("attempt", i+1),
				zap.Int("max_retries", s.config.MaxRetries),
				zap.Error(lastErr))
			time.Sleep(s.config.RetryBackoff)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, body)
		if err != nil {
			lastErr = &AdminAPIError{
				Message: "failed to create request",
				Err:     err,
			}
			continue
		}
		req.Header.Set("Authorization", "Bearer "+s.config.ServiceToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.adminClient.Do(req)
		if err != nil {
			lastErr = &AdminAPIError{
				Message: "failed to send request",
				Err:     err,
			}
			continue
		}

		// Read response body
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = &AdminAPIError{
				Message: "failed to read response body",
				Err:     err,
			}
			continue
		}

		if resp.StatusCode != http.StatusOK {
			s.logger.Error("admin API request failed",
				zap.Int("status_code", resp.StatusCode),
				zap.String("response", string(body)))
			lastErr = &AdminAPIError{
				StatusCode: resp.StatusCode,
				Message:    fmt.Sprintf("unexpected status code: %s", string(body)),
			}
			continue
		}

		var response map[string]interface{}
		if err := json.Unmarshal(body, &response); err != nil {
			lastErr = &AdminAPIError{
				Message: "failed to decode response",
				Err:     err,
			}
			continue
		}

		return response, nil
	}

	return nil, fmt.Errorf("failed after %d retries: %w", s.config.MaxRetries, lastErr)
}

// callOpenAI makes a request to the OpenAI API with retries
func (s *MCPServer) callOpenAI(ctx context.Context, prompt string) (map[string]interface{}, error) {
	s.logger.Info("calling OpenAI API")

	url := "https://api.openai.com/v1/chat/completions"
	requestBody := map[string]interface{}{
		"model": "gpt-4",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a data visualization expert. Generate a Vega-Lite chart specification based on the provided data and user prompt.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"response_format": map[string]string{
			"type": "json_object",
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, &OpenAIError{
			Message: "failed to marshal request body",
			Err:     err,
		}
	}

	var lastErr error
	for i := 0; i <= s.config.MaxRetries; i++ {
		if i > 0 {
			s.logger.Info("retrying OpenAI API call",
				zap.Int("attempt", i+1),
				zap.Int("max_retries", s.config.MaxRetries),
				zap.Error(lastErr))
			time.Sleep(s.config.RetryBackoff)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(jsonBody)))
		if err != nil {
			lastErr = &OpenAIError{
				Message: "failed to create request",
				Err:     err,
			}
			continue
		}
		req.Header.Set("Authorization", "Bearer "+s.config.OpenAIAPIKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.openaiClient.Do(req)
		if err != nil {
			lastErr = &OpenAIError{
				Message: "failed to send request",
				Err:     err,
			}
			continue
		}

		// Read response body
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = &OpenAIError{
				Message: "failed to read response body",
				Err:     err,
			}
			continue
		}

		if resp.StatusCode != http.StatusOK {
			s.logger.Error("OpenAI API request failed",
				zap.Int("status_code", resp.StatusCode),
				zap.String("response", string(body)))
			lastErr = &OpenAIError{
				Message: fmt.Sprintf("unexpected status code: %s", string(body)),
			}
			continue
		}

		var response map[string]interface{}
		if err := json.Unmarshal(body, &response); err != nil {
			lastErr = &OpenAIError{
				Message: "failed to decode response",
				Err:     err,
			}
			continue
		}

		choices, ok := response["choices"].([]interface{})
		if !ok || len(choices) == 0 {
			lastErr = &OpenAIError{
				Message: "invalid response format: missing choices",
			}
			continue
		}

		choice, ok := choices[0].(map[string]interface{})
		if !ok {
			lastErr = &OpenAIError{
				Message: "invalid response format: invalid choice",
			}
			continue
		}

		message, ok := choice["message"].(map[string]interface{})
		if !ok {
			lastErr = &OpenAIError{
				Message: "invalid response format: missing message",
			}
			continue
		}

		content, ok := message["content"].(string)
		if !ok {
			lastErr = &OpenAIError{
				Message: "invalid response format: missing content",
			}
			continue
		}

		var vegaSpec map[string]interface{}
		if err := json.Unmarshal([]byte(content), &vegaSpec); err != nil {
			lastErr = &OpenAIError{
				Message: "failed to parse Vega-Lite specification",
				Err:     err,
			}
			continue
		}

		return vegaSpec, nil
	}

	return nil, fmt.Errorf("failed after %d retries: %w", s.config.MaxRetries, lastErr)
}

// renderVegaLiteToPNG renders a Vega-Lite specification to a PNG image
func (s *MCPServer) renderVegaLiteToPNG(ctx context.Context, vegaSpec map[string]interface{}) ([]byte, error) {
	s.logger.Info("rendering Vega-Lite specification to PNG")

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	taskCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	taskCtx, cancel = context.WithTimeout(taskCtx, 30*time.Second)
	defer cancel()

	vegaSpecJSON, err := json.Marshal(vegaSpec)
	if err != nil {
		return nil, &ChartGenerationError{
			Message: "failed to marshal Vega-Lite specification",
			Err:     err,
		}
	}

	html := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<script src="https://cdn.jsdelivr.net/npm/vega@5"></script>
			<script src="https://cdn.jsdelivr.net/npm/vega-lite@5"></script>
			<script src="https://cdn.jsdelivr.net/npm/vega-embed@6"></script>
		</head>
		<body>
			<div id="vis"></div>
			<script>
				const spec = %s;
				vegaEmbed('#vis', spec, {renderer: 'svg'}).then(function(result) {
					setTimeout(function() {}, 1000);
				});
			</script>
		</body>
		</html>
	`, string(vegaSpecJSON))

	tmpFile, err := os.CreateTemp("", "vega-lite-*.html")
	if err != nil {
		return nil, &ChartGenerationError{
			Message: "failed to create temporary file",
			Err:     err,
		}
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(html); err != nil {
		return nil, &ChartGenerationError{
			Message: "failed to write HTML to temporary file",
			Err:     err,
		}
	}
	if err := tmpFile.Close(); err != nil {
		return nil, &ChartGenerationError{
			Message: "failed to close temporary file",
			Err:     err,
		}
	}

	var buf []byte
	if err := chromedp.Run(taskCtx,
		chromedp.Navigate("file://"+tmpFile.Name()),
		chromedp.WaitVisible("#vis svg"),
		chromedp.Screenshot("#vis", &buf, chromedp.NodeVisible),
	); err != nil {
		return nil, &ChartGenerationError{
			Message: "failed to render chart",
			Err:     err,
		}
	}

	return buf, nil
}

// TimeGrain represents the supported time grain values for dimension aggregation
type TimeGrain string

const (
	TimeGrainUnspecified TimeGrain = "TIME_GRAIN_UNSPECIFIED"
	TimeGrainMillisecond TimeGrain = "TIME_GRAIN_MILLISECOND"
	TimeGrainSecond      TimeGrain = "TIME_GRAIN_SECOND"
	TimeGrainMinute      TimeGrain = "TIME_GRAIN_MINUTE"
	TimeGrainHour        TimeGrain = "TIME_GRAIN_HOUR"
	TimeGrainDay         TimeGrain = "TIME_GRAIN_DAY"
	TimeGrainWeek        TimeGrain = "TIME_GRAIN_WEEK"
	TimeGrainMonth       TimeGrain = "TIME_GRAIN_MONTH"
	TimeGrainQuarter     TimeGrain = "TIME_GRAIN_QUARTER"
	TimeGrainYear        TimeGrain = "TIME_GRAIN_YEAR"
)

// Operation represents the supported operations for conditions
type Operation string

const (
	OperationUnspecified Operation = "OPERATION_UNSPECIFIED"
	OperationEQ          Operation = "OPERATION_EQ"
	OperationNEQ         Operation = "OPERATION_NEQ"
	OperationLT          Operation = "OPERATION_LT"
	OperationLTE         Operation = "OPERATION_LTE"
	OperationGT          Operation = "OPERATION_GT"
	OperationGTE         Operation = "OPERATION_GTE"
	OperationOR          Operation = "OPERATION_OR"
	OperationAND         Operation = "OPERATION_AND"
	OperationIN          Operation = "OPERATION_IN"
	OperationNIN         Operation = "OPERATION_NIN"
	OperationLIKE        Operation = "OPERATION_LIKE"
	OperationNLIKE       Operation = "OPERATION_NLIKE"
)

// Request validation structs
type (
	// MetricsViewAggregationRequest represents a request to aggregate metrics view data.
	// Example:
	// {
	//   "metrics_view": "sales",
	//   "measures": [{"name": "revenue", "aggregation": "sum"}],
	//   "dimensions": [{"name": "product", "time_grain": "MONTH"}],
	//   "time_range": {"start": "2023-01-01", "end": "2023-12-31"},
	//   "comparison_time_range": {"start": "2022-01-01", "end": "2022-12-31"},
	//   "where": {"cond": {"op": "OPERATION_EQ", "exprs": [{"ident": "region"}, {"val": "North"}]}},
	//   "having": {"cond": {"op": "OPERATION_GT", "exprs": [{"ident": "revenue"}, {"val": 1000}]}},
	//   "sort": [{"name": "revenue", "desc": true}],
	//   "limit": "100",
	//   "offset": "0",
	//   "exact": true,
	//   "fill_missing": true,
	//   "rows": false,
	//   "pivot_on": ["product", "region"]
	// }
	MetricsViewAggregationRequest struct {
		// MetricsView is the name of the metrics view to query
		MetricsView string `json:"metrics_view"`
		// Measures are the metrics to aggregate (e.g., sum, avg, count)
		Measures []Measure `json:"measures"`
		// Dimensions are the columns to group by
		Dimensions []Dimension `json:"dimensions"`
		// TimeRange specifies the time period for the query
		TimeRange *TimeRange `json:"time_range,omitempty"`
		// ComparisonTimeRange specifies the time period for comparison
		ComparisonTimeRange *TimeRange `json:"comparison_time_range,omitempty"`
		// Where contains filter conditions
		Where *Expression `json:"where,omitempty"`
		// Having contains post-aggregation filter conditions
		Having *Expression `json:"having,omitempty"`
		// Sort specifies the ordering of results
		Sort []SortClause `json:"sort,omitempty"`
		// Limit restricts the number of results
		Limit string `json:"limit,omitempty"`
		// Offset specifies the number of results to skip
		Offset string `json:"offset,omitempty"`
		// Exact specifies whether to use exact calculations
		Exact *bool `json:"exact,omitempty"`
		// FillMissing specifies whether to fill missing values
		FillMissing *bool `json:"fill_missing,omitempty"`
		// Rows specifies whether to return rows instead of pivoted data
		Rows *bool `json:"rows,omitempty"`
		// PivotOn specifies dimensions to pivot on
		PivotOn []string `json:"pivot_on,omitempty"`
	}

	// Measure represents a metric to aggregate
	Measure struct {
		Name        string `json:"name"`
		Aggregation string `json:"aggregation,omitempty"`
	}

	// Dimension represents a column to group by
	Dimension struct {
		Name      string    `json:"name"`
		TimeGrain TimeGrain `json:"time_grain,omitempty"`
	}

	// SortClause represents an ordering specification
	SortClause struct {
		Name string `json:"name"`
		Desc *bool  `json:"desc,omitempty"`
	}

	// TimeRange represents a time period
	TimeRange struct {
		Start time.Time `json:"start"`
		End   time.Time `json:"end"`
	}

	// Expression represents a condition expression
	Expression struct {
		Ident    *string     `json:"ident,omitempty"`
		Val      interface{} `json:"val,omitempty"`
		Cond     *Condition  `json:"cond,omitempty"`
		Subquery *Subquery   `json:"subquery,omitempty"`
	}

	// Condition represents a condition with operation and expressions
	Condition struct {
		Op    Operation    `json:"op"`
		Exprs []Expression `json:"exprs"`
	}

	// Subquery represents a subquery expression
	Subquery struct {
		Dimension *string     `json:"dimension,omitempty"`
		Measures  []string    `json:"measures,omitempty"`
		Where     *Expression `json:"where,omitempty"`
		Having    *Expression `json:"having,omitempty"`
	}

	// MetricsViewResourceRequest represents a request for a specific metrics view
	MetricsViewResourceRequest struct {
		Name string `json:"name"`
	}

	// MetricsViewTimeRangeSummaryRequest represents a request for time range information
	MetricsViewTimeRangeSummaryRequest struct {
		MetricsView string `json:"metrics_view"`
	}
)

// ValidateExpression validates that an Expression has exactly one of ident, val, cond, or subquery set
func (e *Expression) ValidateExpression() error {
	fields := []string{"ident", "val", "cond", "subquery"}
	setFields := 0
	if e.Ident != nil {
		setFields++
	}
	if e.Val != nil {
		setFields++
	}
	if e.Cond != nil {
		setFields++
	}
	if e.Subquery != nil {
		setFields++
	}
	if setFields > 1 {
		return &ValidationError{
			Field:   "expression",
			Message: fmt.Sprintf("only one of %v can be set, but got %d fields set", fields, setFields),
		}
	}
	if setFields == 0 {
		return &ValidationError{
			Field:   "expression",
			Message: fmt.Sprintf("one of %v must be set", fields),
		}
	}
	return nil
}

// Response types for structured responses
type (
	// MetricsViewSpec represents the specification of a metrics view
	MetricsViewSpec struct {
		Connector     string            `json:"connector"`
		Database      string            `json:"database"`
		Table         string            `json:"table"`
		DisplayName   string            `json:"display_name"`
		Description   string            `json:"description"`
		TimeDimension string            `json:"time_dimension"`
		Dimensions    []DimensionSpec   `json:"dimensions"`
		Measures      []MeasureSpec     `json:"measures"`
		Properties    map[string]string `json:"properties,omitempty"`
	}

	// DimensionSpec represents a dimension in the metrics view
	DimensionSpec struct {
		Name        string `json:"name"`
		Label       string `json:"label"`
		Description string `json:"description,omitempty"`
		Type        string `json:"type"`
	}

	// MeasureSpec represents a measure in the metrics view
	MeasureSpec struct {
		Name        string `json:"name"`
		Label       string `json:"label"`
		Description string `json:"description,omitempty"`
		Expression  string `json:"expression"`
		Format      string `json:"format,omitempty"`
	}

	// TimeRangeSummary represents the available time range for a metrics view
	TimeRangeSummary struct {
		Min      time.Time `json:"min"`
		Max      time.Time `json:"max"`
		Interval string    `json:"interval"`
	}

	// AggregationResult represents the result of a metrics view aggregation
	AggregationResult struct {
		Data       []map[string]interface{} `json:"data"`
		TotalCount int64                    `json:"total_count"`
		Meta       map[string]interface{}   `json:"meta,omitempty"`
	}
)

// Package mcp implements the Rill MCP Server.
// This server exposes APIs for querying metrics views (Rill's analytical units).
//
// Workflow Overview:
// 1. List Metrics Views: Use list_metrics_views to discover available metrics views in a project.
// 2. Get Metrics View Spec: Use get_metrics_view_spec to fetch a metrics view's spec. This is important to understand all the dimensions and measures in the metrics view.
// 3. Get Time Range: Use get_metrics_view_time_range_summary to obtain the available time range for a metrics view. This is important to understand what time range the data spans.
// 4. Query Aggregations: Use get_metrics_view_aggregation to run queries.
//
// In the workflow, do not proceed with the next step until the previous step has been completed.
// If the information from the previous step is already known (let's say for subsequent queries), you can skip it.

// handleListMetricsViews lists all metrics views in the current project.
// This is the first step in the workflow to discover available metrics views.
func (s *MCPServer) handleListMetricsViews(ctx context.Context, name string, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	s.logger.Info("handling list metrics views request")

	endpoint := fmt.Sprintf("/runtime/resources?organization=%s&project=%s&kind=rill.runtime.v1.MetricsView", s.config.OrganizationName, s.config.ProjectName)
	response, err := s.callAdminAPI(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		s.logger.Error("failed to list metrics views",
			zap.Error(err))
		return nil, fmt.Errorf("failed to list metrics views: %w", err)
	}

	resources, ok := response["resources"].([]interface{})
	if !ok {
		return nil, &ValidationError{
			Field:   "resources",
			Message: "invalid response format: missing resources",
		}
	}

	var names []string
	for _, resource := range resources {
		resourceMap, ok := resource.(map[string]interface{})
		if !ok {
			continue
		}
		meta, ok := resourceMap["meta"].(map[string]interface{})
		if !ok {
			continue
		}
		name, ok := meta["name"].(map[string]interface{})
		if !ok {
			continue
		}
		nameStr, ok := name["name"].(string)
		if !ok {
			continue
		}
		names = append(names, nameStr)
	}

	jsonNames, err := json.Marshal(names)
	if err != nil {
		return nil, &ValidationError{
			Field:   "names",
			Message: "failed to marshal names",
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(jsonNames),
			},
		},
	}, nil
}

// handleGetMetricsViewSpec retrieves the specification for a given metrics view, including available measures and dimensions.
// This is the second step in the workflow to understand the structure of a metrics view.
func (s *MCPServer) handleGetMetricsViewSpec(ctx context.Context, name string, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	s.logger.Info("handling get metrics view spec request")

	var req MetricsViewResourceRequest
	if err := validateAndDecode(arguments, &req); err != nil {
		s.logger.Error("failed to validate request",
			zap.Error(err))
		return nil, err
	}

	endpoint := fmt.Sprintf("/runtime/resource?name.kind=rill.runtime.v1.MetricsView&name.name=%s", req.Name)
	response, err := s.callAdminAPI(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		s.logger.Error("failed to get metrics view spec",
			zap.Error(err),
			zap.String("name", req.Name))
		return nil, fmt.Errorf("failed to get metrics view spec: %w", err)
	}

	resource, ok := response["resource"].(map[string]interface{})
	if !ok {
		return nil, &ValidationError{
			Field:   "resource",
			Message: "invalid response format: missing resource",
		}
	}

	metricsViewData, ok := resource["metricsView"].(map[string]interface{})
	if !ok {
		return nil, &ValidationError{
			Field:   "metricsView",
			Message: "invalid response format: missing metricsView",
		}
	}

	state, ok := metricsViewData["state"].(map[string]interface{})
	if !ok {
		return nil, &ValidationError{
			Field:   "state",
			Message: "invalid response format: missing state",
		}
	}

	validSpec, ok := state["validSpec"].(map[string]interface{})
	if !ok {
		validSpec = map[string]interface{}{}
	}

	// Prune empty values from the spec
	prunedSpec := prune(validSpec)

	jsonSpec, err := json.Marshal(prunedSpec)
	if err != nil {
		return nil, &ValidationError{
			Field:   "spec",
			Message: "failed to marshal spec",
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(jsonSpec),
			},
		},
	}, nil
}

// handleGetMetricsViewTimeRangeSummary retrieves the total time range available for a given metrics view.
// This is the third step in the workflow to understand the time range constraints.
// All subsequent queries of the metrics view should be constrained to this time range to ensure accurate results.
func (s *MCPServer) handleGetMetricsViewTimeRangeSummary(ctx context.Context, name string, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	s.logger.Info("handling get metrics view time range summary request")

	var req MetricsViewTimeRangeSummaryRequest
	if err := validateAndDecode(arguments, &req); err != nil {
		s.logger.Error("failed to validate request",
			zap.Error(err))
		return nil, err
	}

	endpoint := fmt.Sprintf("/runtime/queries/metrics-views/%s/time-range-summary", req.MetricsView)
	response, err := s.callAdminAPI(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		s.logger.Error("failed to get time range summary",
			zap.Error(err),
			zap.String("metrics_view", req.MetricsView))
		return nil, fmt.Errorf("failed to get time range summary: %w", err)
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return nil, &ValidationError{
			Field:   "response",
			Message: "failed to marshal response",
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(jsonResponse),
			},
		},
	}, nil
}

// handleGetMetricsViewAggregation performs an arbitrary aggregation on a metrics view.
// This is the final step in the workflow to query the data.
//
// Tips:
//   - Use the sort and limit parameters for best results and to avoid large, unbounded result sets.
//
// Examples:
//
//	Get the total revenue by country and product category:
//	{
//	  "metrics_view": "ecommerce_financials",
//	  "measures": [{"name": "total_revenue"}, {"name": "total_orders"}],
//	  "dimensions": [{"name": "country"}, {"name": "product_category"}],
//	  "time_range": {
//	    "start": "2024-01-01T00:00:00Z",
//	    "end": "2024-12-31T23:59:59Z"
//	  },
//	  "where": {
//	    "cond": {
//	      "op": "OPERATION_AND",
//	      "exprs": [
//	        {
//	          "cond": {
//	            "op": "OPERATION_IN",
//	            "exprs": [
//	              {"ident": "country"},
//	              {"val": ["US", "CA", "GB"]}
//	            ]
//	          }
//	        },
//	        {
//	          "cond": {
//	            "op": "OPERATION_EQ",
//	            "exprs": [
//	              {"ident": "product_category"},
//	              {"val": "Electronics"}
//	            ]
//	          }
//	        }
//	      ]
//	    },
//	  },
//	  "sort": [{"name": "total_revenue", "desc": true}],
//	  "limit": "10"
//	}
//
//	Get the total revenue by country, grouped by month:
//	{
//	  "metrics_view": "ecommerce_financials",
//	  "measures": [{"name": "total_revenue"}],
//	  "dimensions": [
//	    {"name": "transaction_timestamp", "time_grain": "TIME_GRAIN_MONTH"}
//	    {"name": "country"},
//	  ],
//	  "time_range": {
//	    "start": "2024-01-01T00:00:00Z",
//	    "end": "2024-12-31T23:59:59Z"
//	  },
//	  "sort": [
//	    {"name": "transaction_timestamp"},
//	    {"name": "total_revenue", "desc": true},
//	  ],
//	}
func (s *MCPServer) handleGetMetricsViewAggregation(ctx context.Context, name string, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	s.logger.Info("handling get metrics view aggregation request")

	var req MetricsViewAggregationRequest
	if err := validateAndDecode(arguments, &req); err != nil {
		s.logger.Error("failed to validate request",
			zap.Error(err))
		return nil, err
	}

	// Convert request to API format
	requestBody := map[string]interface{}{
		"measures":     req.Measures,
		"dimensions":   req.Dimensions,
		"time_range":   req.TimeRange,
		"where":        req.Where,
		"having":       req.Having,
		"sort":         req.Sort,
		"limit":        req.Limit,
		"offset":       req.Offset,
		"exact":        req.Exact,
		"fill_missing": req.FillMissing,
		"rows":         req.Rows,
		"pivot_on":     req.PivotOn,
	}

	// Add comparison time range if provided
	if req.ComparisonTimeRange != nil {
		requestBody["comparison_time_range"] = req.ComparisonTimeRange
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, &ValidationError{
			Field:   "request",
			Message: "failed to marshal request body",
		}
	}

	endpoint := fmt.Sprintf("/runtime/queries/metrics-views/%s/aggregation", req.MetricsView)
	response, err := s.callAdminAPI(ctx, http.MethodPost, endpoint, strings.NewReader(string(jsonBody)))
	if err != nil {
		s.logger.Error("failed to get metrics view aggregation",
			zap.Error(err),
			zap.String("metrics_view", req.MetricsView))
		return nil, fmt.Errorf("failed to get metrics view aggregation: %w", err)
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return nil, &ValidationError{
			Field:   "response",
			Message: "failed to marshal response",
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(jsonResponse),
			},
		},
	}, nil
}

// Helper function to validate and decode request arguments
func validateAndDecode(arguments map[string]interface{}, req interface{}) error {
	jsonData, err := json.Marshal(arguments)
	if err != nil {
		return &ValidationError{
			Field:   "arguments",
			Message: "failed to marshal arguments",
		}
	}

	if err := json.Unmarshal(jsonData, req); err != nil {
		return &ValidationError{
			Field:   "request",
			Message: fmt.Sprintf("failed to decode request: %v", err),
		}
	}

	// Validate specific request types
	switch r := req.(type) {
	case *MetricsViewAggregationRequest:
		if err := validateAggregationRequest(r); err != nil {
			return err
		}
	case *MetricsViewResourceRequest:
		if err := validateResourceRequest(r); err != nil {
			return err
		}
	case *MetricsViewTimeRangeSummaryRequest:
		if err := validateTimeRangeSummaryRequest(r); err != nil {
			return err
		}
	}

	return nil
}

// validateAggregationRequest validates a metrics view aggregation request
func validateAggregationRequest(req *MetricsViewAggregationRequest) error {
	if req.MetricsView == "" {
		return &ValidationError{
			Field:   "metrics_view",
			Message: "metrics view name is required",
		}
	}

	if len(req.Measures) == 0 {
		return &ValidationError{
			Field:   "measures",
			Message: "at least one measure is required",
		}
	}

	if len(req.Dimensions) == 0 {
		return &ValidationError{
			Field:   "dimensions",
			Message: "at least one dimension is required",
		}
	}

	// Validate time range if provided
	if req.TimeRange != nil {
		if req.TimeRange.Start.IsZero() {
			return &ValidationError{
				Field:   "time_range.start",
				Message: "start time is required",
			}
		}
		if req.TimeRange.End.IsZero() {
			return &ValidationError{
				Field:   "time_range.end",
				Message: "end time is required",
			}
		}
		if req.TimeRange.End.Before(req.TimeRange.Start) {
			return &ValidationError{
				Field:   "time_range",
				Message: "end time must be after start time",
			}
		}
	}

	// Validate comparison time range if provided
	if req.ComparisonTimeRange != nil {
		if req.ComparisonTimeRange.Start.IsZero() {
			return &ValidationError{
				Field:   "comparison_time_range.start",
				Message: "start time is required",
			}
		}
		if req.ComparisonTimeRange.End.IsZero() {
			return &ValidationError{
				Field:   "comparison_time_range.end",
				Message: "end time is required",
			}
		}
		if req.ComparisonTimeRange.End.Before(req.ComparisonTimeRange.Start) {
			return &ValidationError{
				Field:   "comparison_time_range",
				Message: "end time must be after start time",
			}
		}
	}

	// Validate where expression if provided
	if req.Where != nil {
		if err := req.Where.ValidateExpression(); err != nil {
			return err
		}
	}

	// Validate having expression if provided
	if req.Having != nil {
		if err := req.Having.ValidateExpression(); err != nil {
			return err
		}
	}

	// Validate sort clauses if provided
	for i, sort := range req.Sort {
		if sort.Name == "" {
			return &ValidationError{
				Field:   fmt.Sprintf("sort[%d].name", i),
				Message: "sort name is required",
			}
		}
	}

	return nil
}

// validateResourceRequest validates a metrics view resource request
func validateResourceRequest(req *MetricsViewResourceRequest) error {
	if req.Name == "" {
		return &ValidationError{
			Field:   "name",
			Message: "metrics view name is required",
		}
	}
	return nil
}

// validateTimeRangeSummaryRequest validates a metrics view time range summary request
func validateTimeRangeSummaryRequest(req *MetricsViewTimeRangeSummaryRequest) error {
	if req.MetricsView == "" {
		return &ValidationError{
			Field:   "metrics_view",
			Message: "metrics view name is required",
		}
	}
	return nil
}

func (s *MCPServer) handleGenerateChart(ctx context.Context, name string, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	s.logger.Info("handling generate chart request")

	if !s.config.EnableVisualization {
		return nil, &ChartGenerationError{
			Message: "visualization is disabled",
		}
	}

	data := arguments["data"].(map[string]interface{})
	prompt := arguments["prompt"].(string)

	dataJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	fullPrompt := fmt.Sprintf("Data: %s\n\nPrompt: %s\n\nGenerate a Vega-Lite specification (version 5) for a chart that addresses this prompt using the provided data. Return ONLY valid JSON for the Vega-Lite specification, nothing else.", string(dataJSON), prompt)

	vegaSpec, err := s.callOpenAI(ctx, fullPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate chart: %w", err)
	}

	// Render the Vega-Lite specification to a PNG image
	imgBytes, err := s.renderVegaLiteToPNG(ctx, vegaSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to render chart: %w", err)
	}

	// Return the image as a base64-encoded string
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.ImageContent{
				Type:     "image",
				Data:     imgBytes,
				MimeType: "image/png",
			},
		},
	}, nil
}

// prune recursively removes keys with empty, null, or non-substantial values from maps/slices
func prune(obj interface{}) interface{} {
	switch v := obj.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, val := range v {
			pruned := prune(val)
			if !isEmpty(pruned) {
				result[k] = pruned
			}
		}
		return result
	case []interface{}:
		var result []interface{}
		for _, val := range v {
			pruned := prune(val)
			if !isEmpty(pruned) {
				result = append(result, pruned)
			}
		}
		return result
	default:
		return obj
	}
}

// isEmpty checks if a value is empty, null, or non-substantial
func isEmpty(v interface{}) bool {
	if v == nil {
		return true
	}

	switch val := v.(type) {
	case string:
		return val == ""
	case map[string]interface{}:
		return len(val) == 0
	case []interface{}:
		return len(val) == 0
	default:
		return false
	}
}
