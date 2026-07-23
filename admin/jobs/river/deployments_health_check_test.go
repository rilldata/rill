package river

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestDeploymentsHealthCheckPaginationBoundaries(t *testing.T) {
	// Exact page boundaries are easy to get wrong: a full page must trigger one
	// more fetch, while the cursor must still let the 10,001st row be checked.
	tests := []struct {
		count       int
		wantCursors []string
	}{
		{count: 99, wantCursors: []string{""}},
		{count: 100, wantCursors: []string{"", "deployment-099"}},
		{count: 101, wantCursors: []string{"", "deployment-099"}},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d deployments", tt.count), func(t *testing.T) {
			deployments := make([]*database.Deployment, tt.count)
			for i := range deployments {
				deployments[i] = runningDeployment(i, fmt.Sprintf("host-%03d", i))
			}

			var cursors []string
			var healthCalls atomic.Int32
			worker := &DeploymentsHealthCheckWorker{
				logger: zap.NewNop(),
				findDeployments: func(_ context.Context, afterID string, limit int) ([]*database.Deployment, error) {
					cursors = append(cursors, afterID)
					start := 0
					if afterID != "" {
						for i, deployment := range deployments {
							if deployment.ID == afterID {
								start = i + 1
								break
							}
						}
					}
					end := min(start+limit, len(deployments))
					return deployments[start:end], nil
				},
				healthCheck: func(_ context.Context, deployment *database.Deployment) ([]string, bool) {
					healthCalls.Add(1)
					return []string{deployment.RuntimeInstanceID}, true
				},
				findDeploymentByInstanceID: func(context.Context, string) (*database.Deployment, error) {
					t.Fatal("a healthy instance must not be reloaded from the database")
					return nil, nil
				},
			}

			err := worker.Work(t.Context(), nil)

			require.NoError(t, err)
			require.Equal(t, tt.wantCursors, cursors)
			require.Equal(t, int32(tt.count), healthCalls.Load())
		})
	}
}

func TestDeploymentsHealthCheckDeduplicatesHostsAndHandlesDeletionRace(t *testing.T) {
	// One runtime health response describes every instance on its host. A
	// deployment deleted between listing and comparison is not a missing-instance alert.
	deployments := []*database.Deployment{
		runningDeployment(0, "shared-host"),
		runningDeployment(1, "shared-host"),
		runningDeployment(2, "shared-host"),
	}
	core, logs := observer.New(zap.ErrorLevel)
	var healthCalls atomic.Int32
	worker := &DeploymentsHealthCheckWorker{
		logger: zap.New(core),
		findDeployments: func(_ context.Context, afterID string, _ int) ([]*database.Deployment, error) {
			if afterID != "" {
				return nil, nil
			}
			return deployments, nil
		},
		healthCheck: func(context.Context, *database.Deployment) ([]string, bool) {
			healthCalls.Add(1)
			return []string{deployments[0].RuntimeInstanceID, deployments[1].RuntimeInstanceID}, true
		},
		findDeploymentByInstanceID: func(_ context.Context, instanceID string) (*database.Deployment, error) {
			require.Equal(t, deployments[2].RuntimeInstanceID, instanceID)
			return nil, database.ErrNotFound
		},
	}

	err := worker.Work(t.Context(), nil)

	require.NoError(t, err)
	require.Equal(t, int32(1), healthCalls.Load())
	require.Equal(t, 0, logs.FilterMessage("deployment health check: missing instance on runtime").Len())
}

func TestDeploymentsHealthCheckReportsOnlyConfirmedMissingInstances(t *testing.T) {
	// A successful runtime response without an expected instance is actionable
	// only after the database confirms that the deployment still exists.
	deployment := runningDeployment(0, "host")
	core, logs := observer.New(zap.ErrorLevel)
	worker := &DeploymentsHealthCheckWorker{
		logger: zap.New(core),
		findDeployments: func(context.Context, string, int) ([]*database.Deployment, error) {
			return []*database.Deployment{deployment}, nil
		},
		healthCheck: func(context.Context, *database.Deployment) ([]string, bool) {
			return []string{}, true
		},
		findDeploymentByInstanceID: func(context.Context, string) (*database.Deployment, error) {
			return deployment, nil
		},
		deploymentAnnotations: func(context.Context, *database.Deployment) (*admin.DeploymentAnnotations, error) {
			return &admin.DeploymentAnnotations{}, nil
		},
	}

	err := worker.Work(t.Context(), nil)

	require.NoError(t, err)
	require.Equal(t, 1, logs.FilterMessage("deployment health check: missing instance on runtime").Len())
}

func TestDeploymentsHealthCheckSkipsMissingComparisonWhenRuntimeUnavailable(t *testing.T) {
	// An unavailable runtime yields no trustworthy instance inventory, so it
	// must not fan out misleading database lookups and missing-instance alerts.
	deployment := runningDeployment(0, "host")
	var lookups atomic.Int32
	worker := &DeploymentsHealthCheckWorker{
		logger: zap.NewNop(),
		findDeployments: func(context.Context, string, int) ([]*database.Deployment, error) {
			return []*database.Deployment{deployment}, nil
		},
		healthCheck: func(context.Context, *database.Deployment) ([]string, bool) {
			return nil, false
		},
		findDeploymentByInstanceID: func(context.Context, string) (*database.Deployment, error) {
			lookups.Add(1)
			return nil, database.ErrNotFound
		},
	}

	err := worker.Work(t.Context(), nil)

	require.NoError(t, err)
	require.Zero(t, lookups.Load())
}

func TestDeploymentHealthClassifiers(t *testing.T) {
	// Health severity depends on every response field, so this table protects
	// less common limiter, repository, metric, parsing, and reconcile failures.
	require.False(t, runtimeUnhealthy(&runtimev1.HealthResponse{}))
	for name, response := range map[string]*runtimev1.HealthResponse{
		"limiter":    {LimiterError: "failed"},
		"connection": {ConnCacheError: "failed"},
		"metastore":  {MetastoreError: "failed"},
		"network":    {NetworkError: "failed"},
	} {
		t.Run(name, func(t *testing.T) {
			require.True(t, runtimeUnhealthy(response))
		})
	}

	require.False(t, instanceUnhealthy(&runtimev1.InstanceHealth{}))
	for name, health := range map[string]*runtimev1.InstanceHealth{
		"olap":       {OlapError: "failed"},
		"controller": {ControllerError: "failed"},
		"repository": {RepoError: "failed"},
		"metrics":    {MetricsViewErrors: map[string]string{"orders": "failed"}},
		"parse":      {ParseErrorCount: 1},
		"reconcile":  {ReconcileErrorCount: 1},
	} {
		t.Run(name, func(t *testing.T) {
			require.True(t, instanceUnhealthy(health))
		})
	}
}

func TestValidateDeploymentsOrphanBoundary(t *testing.T) {
	// The three-hour grace period protects a deployment that may still become
	// primary; teardown is allowed only strictly after the boundary.
	now := time.Date(2026, time.July, 23, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name      string
		depl      *database.Deployment
		primaryID string
		want      bool
	}{
		{name: "exactly three hours old", depl: orphanDeployment("old", now.Add(-3*time.Hour)), primaryID: "new"},
		{name: "one nanosecond older", depl: orphanDeployment("old", now.Add(-3*time.Hour-time.Nanosecond)), primaryID: "new", want: true},
		{name: "current primary", depl: orphanDeployment("primary", now.Add(-4*time.Hour)), primaryID: "primary"},
		{name: "already stopped", depl: withDeploymentStatus(orphanDeployment("old", now.Add(-4*time.Hour)), database.DeploymentStatusStopped), primaryID: "new"},
		{name: "development environment", depl: withDeploymentEnvironment(orphanDeployment("old", now.Add(-4*time.Hour)), "dev"), primaryID: "new"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, shouldTeardownOrphanDeployment(tt.depl, tt.primaryID, now))
		})
	}
}

func TestValidateDeploymentsPrimarySwitchAndTeardownFailure(t *testing.T) {
	// After a primary switch, the old instance is removed while the new primary
	// is reconciled. A failed orphan teardown must not block later deployments.
	now := time.Date(2026, time.July, 23, 12, 0, 0, 0, time.UTC)
	primaryID := "new-primary"
	project := &database.Project{ID: "project", OrganizationID: "org", PrimaryDeploymentID: &primaryID}
	failedOrphan := orphanDeployment("failed-orphan", now.Add(-4*time.Hour))
	oldPrimary := orphanDeployment("old-primary", now.Add(-4*time.Hour))
	newPrimary := orphanDeployment(primaryID, now.Add(-time.Minute))

	var mu sync.Mutex
	var removed []string
	var reconciled []string
	worker := &ValidateDeploymentsWorker{
		admin: &admin.Service{Logger: zap.NewNop()},
		now:   func() time.Time { return now },
		findDeploymentsForProject: func(context.Context, string, string, string) ([]*database.Deployment, error) {
			return []*database.Deployment{failedOrphan, oldPrimary, newPrimary}, nil
		},
		findOrganization: func(context.Context, string) (*database.Organization, error) {
			return &database.Organization{ID: "org"}, nil
		},
		teardownDeployment: func(_ context.Context, deployment *database.Deployment) error {
			mu.Lock()
			defer mu.Unlock()
			removed = append(removed, deployment.ID)
			if deployment.ID == failedOrphan.ID {
				return fmt.Errorf("provisioner unavailable")
			}
			return nil
		},
		reconcileDeployment: func(_ context.Context, deploymentID string) error {
			mu.Lock()
			defer mu.Unlock()
			reconciled = append(reconciled, deploymentID)
			return nil
		},
	}

	err := worker.validateDeploymentsForProject(t.Context(), project)

	require.NoError(t, err)
	require.Equal(t, []string{"failed-orphan", "old-primary"}, removed)
	require.Equal(t, []string{"new-primary"}, reconciled)
}

func runningDeployment(index int, host string) *database.Deployment {
	return &database.Deployment{
		ID:                fmt.Sprintf("deployment-%03d", index),
		ProjectID:         fmt.Sprintf("project-%03d", index),
		Environment:       "prod",
		RuntimeHost:       host,
		RuntimeInstanceID: fmt.Sprintf("instance-%03d", index),
		Status:            database.DeploymentStatusRunning,
	}
}

func orphanDeployment(id string, updatedOn time.Time) *database.Deployment {
	return &database.Deployment{
		ID:          id,
		ProjectID:   "project",
		Environment: "prod",
		Status:      database.DeploymentStatusRunning,
		UpdatedOn:   updatedOn,
	}
}

func withDeploymentStatus(deployment *database.Deployment, status database.DeploymentStatus) *database.Deployment {
	clone := *deployment
	clone.Status = status
	return &clone
}

func withDeploymentEnvironment(deployment *database.Deployment, environment string) *database.Deployment {
	clone := *deployment
	clone.Environment = environment
	return &clone
}
