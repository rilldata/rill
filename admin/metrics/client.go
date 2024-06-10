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

const TimeBoundFormat = "2006-01-02T15:04:050Z"

// It can be used for such use cases as autoscaling, health checks, breached quotas, and usage calculations for billing.
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
	ProjectName string    `json:"project_name"`
	MinTime     time.Time `json:"min_time"`
	MaxTime     time.Time `json:"max_time"`
}

type Usage struct {
	MetricName string  `json:"metric_name"`
	Amount     float64 `json:"amount"`
}

func (c *Client) GetProjectUsageAvailability(ctx context.Context, orgName string, projectNames []string, lowerTimeBound, upperTimeBound time.Time) ([]ProjectUsageAvailability, error) {
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
	/*  api query -
	select project_id, min(__time) as min_time, max(__time) as max_time
	from druid-metrics where org_name='' and project_name IN ('','',...) and __time >= '<lower-bound>' and __time < '<upper-bound>' group by 1
	*/

	// Add URL query parameters
	qry := uri.Query()
	qry.Add("org_name", orgName)
	qry.Add("lower_time_bound", lowerTimeBound.Format(TimeBoundFormat))
	qry.Add("upper_time_bound", upperTimeBound.Format(TimeBoundFormat))
	for _, id := range projectNames {
		qry.Add("project_name", id)
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

	// Decode the JSON response into ProjectUsageAvailability structs
	var availability []ProjectUsageAvailability
	err = json.NewDecoder(resp.Body).Decode(&availability)
	if err != nil {
		return nil, err
	}

	return availability, nil
}

func (c *Client) GetProjectUsageMetrics(ctx context.Context, orgName, projectName string, lowerTimeBound, upperTimeBound time.Time, metricNames []string) ([]Usage, error) {
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
		uri.Path = path.Join("/v1/instances", c.InstanceID, fmt.Sprintf("/api/%s", metric))
		/*  api query -
		select metric, AGG_FUNC(metric_value) as usage
		from druid-metrics where org_name='' and project_name ='' and __time >= '<lower-bound>' and __time < '<upper-bound>' and metric = 'metric' group by 1
		*/

		// Add URL query parameters
		qry := uri.Query()
		qry.Add("org_name", orgName)
		qry.Add("project_name", projectName)
		qry.Add("lower_time_bound", lowerTimeBound.Format(TimeBoundFormat))
		qry.Add("upper_time_bound", upperTimeBound.Format(TimeBoundFormat))

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
