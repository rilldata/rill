package clickhouse

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const cloudAPIBaseURL = "https://api.clickhouse.cloud/v1"

// CloudAPIClient provides access to the ClickHouse Cloud REST API.
type CloudAPIClient struct {
	keyID      string
	keySecret  string
	httpClient *http.Client
}

// CloudServiceInfo holds metadata about a ClickHouse Cloud service.
type CloudServiceInfo struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Status        string  `json:"status"` // "running", "idle", "stopped", "provisioning"
	CloudProvider string  `json:"cloudProvider"`
	Region        string  `json:"region"`
	Tier          string  `json:"tier"` // "development", "production"
	IdleScaling   bool    `json:"idleScaling"`
	MinMemoryGB   float64 `json:"minTotalMemoryGb"`
	MaxMemoryGB   float64 `json:"maxTotalMemoryGb"`
	NumReplicas   int     `json:"numReplicas"`
}

// NewCloudAPIClient creates a client for the ClickHouse Cloud API.
// Returns nil if either credential is empty.
func NewCloudAPIClient(keyID, keySecret string) *CloudAPIClient {
	if keyID == "" || keySecret == "" {
		return nil
	}
	return &CloudAPIClient{
		keyID:     keyID,
		keySecret: keySecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FindServiceByHost looks up the CHC service whose endpoint matches the given host.
// It iterates all organizations the API key has access to and all services in each.
func (c *CloudAPIClient) FindServiceByHost(ctx context.Context, host string) (*CloudServiceInfo, error) {
	orgs, err := c.listOrganizations(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing organizations: %w", err)
	}

	for _, orgID := range orgs {
		svc, err := c.findServiceInOrg(ctx, orgID, host)
		if err != nil {
			continue // try next org
		}
		if svc != nil {
			return svc, nil
		}
	}

	return nil, fmt.Errorf("no ClickHouse Cloud service found matching host %q", host)
}

// GetServiceInfo fetches info for a known org + service ID pair.
func (c *CloudAPIClient) GetServiceInfo(ctx context.Context, orgID, serviceID string) (*CloudServiceInfo, error) {
	path := fmt.Sprintf("/organizations/%s/services/%s", orgID, serviceID)
	var resp struct {
		Result serviceResponse `json:"result"`
	}
	if err := c.doGet(ctx, path, &resp); err != nil {
		return nil, err
	}
	return mapServiceResponse(&resp.Result), nil
}

func (c *CloudAPIClient) listOrganizations(ctx context.Context) ([]string, error) {
	var resp struct {
		Result []orgResponse `json:"result"`
	}
	if err := c.doGet(ctx, "/organizations", &resp); err != nil {
		return nil, err
	}
	ids := make([]string, len(resp.Result))
	for i, o := range resp.Result {
		ids[i] = o.ID
	}
	return ids, nil
}

func (c *CloudAPIClient) findServiceInOrg(ctx context.Context, orgID, host string) (*CloudServiceInfo, error) {
	path := fmt.Sprintf("/organizations/%s/services", orgID)
	var resp struct {
		Result []serviceResponse `json:"result"`
	}
	if err := c.doGet(ctx, path, &resp); err != nil {
		return nil, err
	}

	// Normalize the target host: strip port if present
	targetHost := strings.Split(host, ":")[0]
	targetHost = strings.ToLower(targetHost)

	for _, svc := range resp.Result {
		for _, ep := range svc.Endpoints {
			epHost := strings.ToLower(ep.Host)
			if epHost == targetHost || strings.HasPrefix(targetHost, strings.Split(epHost, ".")[0]) {
				return mapServiceResponse(&svc), nil
			}
		}
	}
	return nil, nil
}

func (c *CloudAPIClient) doGet(ctx context.Context, path string, result any) error {
	url := cloudAPIBaseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.SetBasicAuth(c.keyID, c.keySecret)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ClickHouse Cloud API returned %d: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}
	return nil
}

// API response types

type orgResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type serviceResponse struct {
	ID                  string             `json:"id"`
	Name                string             `json:"name"`
	State               string             `json:"state"`
	CloudProvider       string             `json:"provider"`
	Region              string             `json:"region"`
	Tier                string             `json:"tier"`
	Endpoints           []endpointResponse `json:"endpoints"`
	IdleScaling         bool               `json:"idleScaling"`
	MinReplicaMemoryGB  float64            `json:"minReplicaMemoryGb"`
	MaxReplicaMemoryGB  float64            `json:"maxReplicaMemoryGb"`
	MinTotalMemoryGB    float64            `json:"minTotalMemoryGb"`
	MaxTotalMemoryGB    float64            `json:"maxTotalMemoryGb"`
	NumReplicas         int                `json:"numReplicas"`
}

type endpointResponse struct {
	Protocol string `json:"protocol"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
}

func mapServiceResponse(svc *serviceResponse) *CloudServiceInfo {
	return &CloudServiceInfo{
		ID:            svc.ID,
		Name:          svc.Name,
		Status:        svc.State,
		CloudProvider: svc.CloudProvider,
		Region:        svc.Region,
		Tier:          svc.Tier,
		IdleScaling:   svc.IdleScaling,
		MinMemoryGB:   svc.MinReplicaMemoryGB,
		MaxMemoryGB:   svc.MaxReplicaMemoryGB,
		NumReplicas:   svc.NumReplicas,
	}
}
