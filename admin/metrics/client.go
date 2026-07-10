package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
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
	uri, err := url.Parse(c.RuntimeHost)
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
	OrgID             string    `json:"org_id"`
	ProjectID         string    `json:"project_id"`
	ProjectName       string    `json:"project_name"`
	BillingCustomerID *string   `json:"billing_customer_id"`
	StartTime         time.Time `json:"start_time"`
	EndTime           time.Time `json:"end_time"`
	EventName         string    `json:"event_name"`
	MaxValue          float64   `json:"max_value"`
	SumValue          float64   `json:"sum_value"`
	BillingService    string    `json:"billing_service"`
	InstanceID        string    `json:"instance_id"`
}

func (c *Client) GetUsageMetrics(ctx context.Context, startTime, endTime, afterTime time.Time, afterOrgID, afterProjectID, afterInstanceID, afterBillingService, afterEventName, grain string, limit int) ([]*Usage, error) {
	uri, err := url.Parse(c.RuntimeHost)
	if err != nil {
		return nil, err
	}

	uri.Path = path.Join("/v1/instances", c.InstanceID, "/api/billing-usage")
	// For the billing-usage API definition (the SQL, the event_name list, and the source-based billing filters), see the
	// billing-usage API of the rill metrics project opened by the OpenMetricsProject method in admin/jobs/river/billing_reporter.go.

	// Add URL query parameters
	qry := uri.Query()
	qry.Add("start_time", startTime.Format(time.RFC3339))
	qry.Add("end_time", endTime.Format(time.RFC3339))
	qry.Add("grain", grain)
	qry.Add("limit", strconv.Itoa(limit))
	if !afterTime.IsZero() {
		qry.Add("after_time", afterTime.Format(time.RFC3339))
		qry.Add("after_org_id", afterOrgID)
		qry.Add("after_project_id", afterProjectID)
		qry.Add("after_instance_id", afterInstanceID)
		qry.Add("after_billing_service", afterBillingService)
		qry.Add("after_event_name", afterEventName)
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

	// Decode the JSON response into UsageMetric struct
	var usage []*Usage
	err = json.NewDecoder(resp.Body).Decode(&usage)
	if err != nil {
		return nil, err
	}

	return usage, nil
}
