package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

// Client can be used for such use cases as autoscaling, health checks, breached quotas, and usage calculations for billing.
type Client struct {
	RuntimeHost string
	InstanceID  string
	AccessToken string
}

// NewClient creates a new metrics project client.
// It must be re-created after the provided access token expires.
// TODO: Move minting and refresh of the access token to the client.
func NewClient(runtimeHost, instanceID, accessToken string) *Client {
	return &Client{
		RuntimeHost: runtimeHost,
		InstanceID:  instanceID,
		AccessToken: accessToken,
	}
}

// AutoscalerSlotsRecommendation represents a recommendation for the number of slots to use for a project.
type AutoscalerSlotsRecommendation struct {
	ProjectID        string    `json:"project_id"`
	RecommendedSlots int       `json:"recommended_slots"`
	UpdatedOn        time.Time `json:"updated_on"`
}

// AutoscalerSlotsRecommendations invokes the "autoscaler-slots-recommendations" API endpoint to get a list of recommendations for the number of slots to use for projects.
func (c *Client) AutoscalerSlotsRecommendations(ctx context.Context, limit, offset int) ([]AutoscalerSlotsRecommendation, error) {
	// Create the URL for the request
	var runtimeHost string

	// In production, the REST and gRPC endpoints are the same, but in development, they're served on different ports.
	// TODO: move to http and grpc to the same c.RuntimeHost for local development.
	// Until we make that change, this is a convenient hack for local development (assumes REST on port 8081).
	if strings.Contains(c.RuntimeHost, "localhost") {
		runtimeHost = "http://localhost:8081"
	} else {
		runtimeHost = c.RuntimeHost
	}

	uri, err := url.Parse(runtimeHost)
	if err != nil {
		return nil, err
	}
	uri.Path = path.Join("/v1/instances", c.InstanceID, "/api/autoscaler-slots-recommendations")

	// Add URL query parameters
	qry := uri.Query()
	if limit > 0 {
		qry.Set("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		qry.Set("offset", fmt.Sprintf("%d", offset))
	}
	uri.RawQuery = qry.Encode()
	apiURL := uri.String()

	// Create a new HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, http.NoBody)
	if err != nil {
		return nil, err
	}

	// Set the access token in the request headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken))

	// Send the request
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}

	// Decode the JSON response into AutoscalerSlotsRecommendation structs
	var recommendations []AutoscalerSlotsRecommendation
	err = json.NewDecoder(resp.Body).Decode(&recommendations)
	if err != nil {
		return nil, err
	}

	return recommendations, nil
}

type Usage struct {
	OrgID      string    `json:"org_id"`
	ProjectID  string    `json:"project_id"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	MetricName string    `json:"metric_name"`
	Amount     float64   `json:"amount"`
}

func (c *Client) GetUsageMetrics(ctx context.Context, startTime, endTime time.Time, grain string) ([]*Usage, error) {
	// Create the URL for the request
	var runtimeHost string

	// In production, the REST and gRPC endpoints are the same, but in development, they're served on different ports.
	if strings.Contains(c.RuntimeHost, "localhost") {
		runtimeHost = "http://localhost:8081"
	} else {
		runtimeHost = c.RuntimeHost
	}

	uri, err := url.Parse(runtimeHost)
	if err != nil {
		return nil, err
	}

	uri.Path = path.Join("/v1/instances", c.InstanceID, "/api/billing-usage")
	/*  sql api will be like -
	  (SELECT
		org_id,
		project_id,
		date_trunc('{{ .args.grain }}', time) as start_time,
		date_trunc('{{ .args.grain }}', time) + INTERVAL 1 {{ .args.grain }} as end_time,
		event_name AS metric_name,
		MAX(value) AS amount
	  FROM rill_metrics
	  WHERE
		time >= '{{ .args.start_time }}' AND time < '{{ .args.end_time }}' and event_name='data_dir_size_bytes'
	  GROUP BY ALL
	  ORDER BY 1,2,3)
	UNION ALL
	  (SELECT
		org_id,
		project_id,
		date_trunc('{{ .args.grain }}', time) as start_time,
		date_trunc('{{ .args.grain }}', time) + INTERVAL 1 {{ .args.grain }} as end_time,
		event_name AS metric_name,
		sum(value) AS amount
	  FROM rill_metrics
	  WHERE
	    time >= '2024-06-11T00:00:00Z' AND time < '{{ .args.start_time }}' and event_name='{{ .args.end_time }}'
	  GROUP BY ALL
	  ORDER BY 1,2,3)
	...
	// if we move to syncing raw events then we may not use aggregation function and UNION and use date_trunc('{{ .args.grain }}', insert_time) as event_time instead of start_time and end_time
	*/

	// Add URL query parameters
	qry := uri.Query()
	qry.Add("start_time", startTime.Format(time.RFC3339))
	qry.Add("end_time", endTime.Format(time.RFC3339))
	qry.Add("grain", grain)

	uri.RawQuery = qry.Encode()
	apiURL := uri.String()

	// Create a new HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, http.NoBody)
	if err != nil {
		return nil, err
	}

	// Set the access token in the request headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken))

	// Send the request
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}

	// Decode the JSON response into UsageMetric struct
	var usage []*Usage
	err = json.NewDecoder(resp.Body).Decode(&usage)
	if err != nil {
		return nil, err
	}

	return usage, nil
}
