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

type ProjectUsageAvailability struct {
	ProjectID string    `json:"project_id"`
	MinTime   time.Time `json:"min_time"`
	MaxTime   time.Time `json:"max_time"`
}

type Usage struct {
	MetricName string  `json:"metric_name"`
	Amount     float64 `json:"amount"`
}

func (c *Client) GetProjectUsageAvailability(ctx context.Context, projectIDs []string, start, end time.Time) ([]ProjectUsageAvailability, error) {
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

	uri.Path = path.Join("/v1/instances", c.InstanceID, "/api/project-usage-availability")
	/*  sql api  -
	  SELECT
		project_id,
		MIN(time) as min_time,
		MAX(time) as max_time
	  FROM rill-metrics
	  WHERE
		project_id IN ({{ .args.project_ids }}) and time >= '{{ .args.start }}' and time < '{{ .args.end }}'
	  GROUP BY 1
	*/

	// Add URL query parameters
	qry := uri.Query()
	qry.Add("start", start.Format(time.RFC3339))
	qry.Add("end", end.Format(time.RFC3339))
	// create a comma separated string of projectIDs with single quotes around each projectID
	qry.Add("project_ids", fmt.Sprintf("'%s'", strings.Join(projectIDs, "','")))

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

	// Decode the JSON response into ProjectUsageAvailability structs
	var availability []ProjectUsageAvailability
	err = json.NewDecoder(resp.Body).Decode(&availability)
	if err != nil {
		return nil, err
	}

	return availability, nil
}

func (c *Client) GetProjectUsageMetrics(ctx context.Context, projectID string, start, end time.Time, metricNames []string) ([]Usage, error) {
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

	var usage []Usage
	for _, metric := range metricNames {
		// metric name is the api name
		uri.Path = path.Join("/v1/instances", c.InstanceID, fmt.Sprintf("/api/%s", metric+"-usage"))
		/*  sql api -
		  SELECT
			'<event>' AS metric_name,
			MAX(value) AS usage
		  FROM rill-metrics
		  WHERE
			project_id ='{{ .args.project_id }}' AND time >= '{{ .args.start }}' AND time < '{{ .args.end }}'
		  GROUP BY 1
		*/

		// Add URL query parameters
		qry := uri.Query()
		qry.Add("project_id", projectID)
		qry.Add("start", start.Format(time.RFC3339))
		qry.Add("end", end.Format(time.RFC3339))

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

		// Check the response status code
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("request failed with status code %d", resp.StatusCode)
		}

		// Decode the JSON response into UsageMetric struct
		var usageMetric Usage
		err = json.NewDecoder(resp.Body).Decode(&usageMetric)
		if err != nil {
			return nil, err
		}
		resp.Body.Close()

		usage = append(usage, usageMetric)
	}
	return usage, nil
}
