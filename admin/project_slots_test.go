package admin

import (
	"context"
	"sync"
	"testing"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/jobs"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestUpdateProjectSlots(t *testing.T) {
	tests := []struct {
		name        string
		prodSlots   int
		devSlots    int
		subpath     string
		expectedIDs []string
	}{
		{name: "unchanged", prodSlots: 2, devSlots: 2},
		{name: "production only", prodSlots: 4, devSlots: 2, expectedIDs: []string{"prod-deployment", "prod-branch-deployment"}},
		{name: "development only", prodSlots: 2, devSlots: 4, expectedIDs: []string{"dev-deployment", "dev-branch-deployment"}},
		{name: "both environments", prodSlots: 4, devSlots: 4, expectedIDs: []string{"prod-deployment", "prod-branch-deployment", "dev-deployment", "dev-branch-deployment"}},
		{name: "shared configuration", prodSlots: 2, devSlots: 2, subpath: "project", expectedIDs: []string{"prod-deployment", "prod-branch-deployment", "dev-deployment", "dev-branch-deployment"}},
		{name: "shared configuration with production slots", prodSlots: 4, devSlots: 2, subpath: "project", expectedIDs: []string{"prod-deployment", "prod-branch-deployment", "dev-deployment", "dev-branch-deployment"}},
		{name: "shared configuration with development slots", prodSlots: 2, devSlots: 4, subpath: "project", expectedIDs: []string{"prod-deployment", "prod-branch-deployment", "dev-deployment", "dev-branch-deployment"}},
	}
	for _, tt := range tests {
		for _, status := range []database.DeploymentStatus{database.DeploymentStatusRunning, database.DeploymentStatusStopped} {
			t.Run(tt.name+"/"+status.String(), func(t *testing.T) {
				project := &database.Project{ID: "project", PrimaryBranch: "main", ProdSlots: 2, DevSlots: 2}
				db := &slotUpdateDB{
					project: project,
					deployments: []*database.Deployment{
						{ID: "prod-deployment", Environment: "prod", Branch: "main", DesiredStatus: status},
						{ID: "prod-branch-deployment", Environment: "prod", Branch: "release", DesiredStatus: status},
						{ID: "dev-deployment", Environment: "dev", Branch: "feature", DesiredStatus: status},
						{ID: "dev-branch-deployment", Environment: "dev", Branch: "another-feature", DesiredStatus: status},
					},
					desiredStatuses: make(map[string]database.DeploymentStatus),
				}
				jobClient := &slotUpdateJobs{}
				svc := &Service{DB: db, Jobs: jobClient, Logger: zap.NewNop()}

				updated, err := svc.UpdateProject(t.Context(), project, &database.UpdateProjectOptions{
					PrimaryBranch: "main", ProdSlots: tt.prodSlots, DevSlots: tt.devSlots, Subpath: tt.subpath,
				})
				require.NoError(t, err)
				require.Equal(t, tt.devSlots, updated.DevSlots)
				require.Equal(t, tt.prodSlots, updated.ProdSlots)
				require.Equal(t, tt.subpath, updated.Subpath)
				require.ElementsMatch(t, tt.expectedIDs, jobClient.deploymentIDs)
				expectedStatuses := make(map[string]database.DeploymentStatus)
				for _, id := range tt.expectedIDs {
					expectedStatuses[id] = status
				}
				// Unaffected deployments must not be touched, and hibernated deployments must stay stopped.
				require.Equal(t, expectedStatuses, db.desiredStatuses)
			})
		}
	}
}

// Only the database and queue operations used by UpdateProject are implemented.
type slotUpdateDB struct {
	database.DB
	mu              sync.Mutex
	project         *database.Project
	deployments     []*database.Deployment
	desiredStatuses map[string]database.DeploymentStatus
}

func (db *slotUpdateDB) FindDeploymentsForProject(_ context.Context, _, environment, branch string) ([]*database.Deployment, error) {
	var deployments []*database.Deployment
	for _, d := range db.deployments {
		if (environment == "" || d.Environment == environment) && (branch == "" || d.Branch == branch) {
			deployments = append(deployments, d)
		}
	}
	return deployments, nil
}

func (db *slotUpdateDB) UpdateProject(_ context.Context, _ string, opts *database.UpdateProjectOptions) (*database.Project, error) {
	updated := *db.project
	updated.ProdSlots = opts.ProdSlots
	updated.DevSlots = opts.DevSlots
	updated.Subpath = opts.Subpath
	return &updated, nil
}

func (db *slotUpdateDB) UpdateDeploymentDesiredStatus(_ context.Context, id string, desired database.DeploymentStatus) (*database.Deployment, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.desiredStatuses[id] = desired
	return &database.Deployment{ID: id, DesiredStatus: desired}, nil
}

type slotUpdateJobs struct {
	jobs.Client
	mu            sync.Mutex
	deploymentIDs []string
}

func (j *slotUpdateJobs) ReconcileDeployment(_ context.Context, id string) (*jobs.InsertResult, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.deploymentIDs = append(j.deploymentIDs, id)
	return &jobs.InsertResult{}, nil
}
