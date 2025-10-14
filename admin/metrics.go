package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/rilldata/rill/admin/metrics"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/server/auth"
)

const metricsProjectClientTTL = 30 * time.Minute

// OpenMetricsProject opens a client for accessing the metrics project.
// If a metrics project is not configured, it returns false for the second return value.
// The returned client has a TTL of 30 minutes.
// TODO: Encapsulate token refresh logic in the metrics client.
func (s *Service) OpenMetricsProject(ctx context.Context) (*metrics.Client, bool, error) {
	// Check if a metrics project is configured
	if s.MetricsProjectID == "" {
		return nil, false, nil
	}

	// Find the production deployment for the metrics project
	proj, err := s.DB.FindProject(ctx, s.MetricsProjectID)
	if err != nil {
		return nil, false, err
	}
	if proj.ProdDeploymentID == nil {
		return nil, false, fmt.Errorf("project does not have a production deployment")
	}
	depl, err := s.DB.FindDeployment(ctx, *proj.ProdDeploymentID)
	if err != nil {
		return nil, false, err
	}
	s.Used.Deployment(depl.ID)

	// Mint a JWT for the metrics project
	jwt, err := s.issuer.NewToken(auth.TokenOptions{
		AudienceURL: depl.RuntimeAudience,
		Subject:     "admin-service",
		TTL:         metricsProjectClientTTL,
		InstancePermissions: map[string][]runtime.Permission{
			depl.RuntimeInstanceID: {
				runtime.ReadAPI,
				runtime.ReadMetrics,
				runtime.ReadObjects,
			},
		},
		Attributes: map[string]any{"admin": true},
	})
	if err != nil {
		return nil, false, fmt.Errorf("could not issue jwt: %w", err)
	}

	// Create the metrics project client
	client := metrics.NewClient(depl.RuntimeHost, depl.RuntimeInstanceID, jwt)
	return client, true, nil
}
