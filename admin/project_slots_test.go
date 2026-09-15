package admin

import (
	"context"
	"testing"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/jobs"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestUpdateProjectDevSlots(t *testing.T) {
	for _, slots := range []int{2, 4} {
		t.Run(map[int]string{2: "unchanged", 4: "resized"}[slots], func(t *testing.T) {
			project := &database.Project{ID: "project", PrimaryBranch: "main", ProdSlots: 2, DevSlots: 2}
			db := &slotUpdateDB{project: project}
			jobClient := &slotUpdateJobs{}
			svc := &Service{DB: db, Jobs: jobClient, Logger: zap.NewNop()}

			updated, err := svc.UpdateProject(t.Context(), project, &database.UpdateProjectOptions{
				PrimaryBranch: "main", ProdSlots: 2, DevSlots: slots,
			})
			require.NoError(t, err)
			require.Equal(t, slots, updated.DevSlots)
			require.Equal(t, 2, updated.ProdSlots)
			if slots == 2 {
				require.Empty(t, jobClient.deploymentIDs)
			} else {
				require.Equal(t, []string{"dev-deployment"}, jobClient.deploymentIDs)
				require.Equal(t, database.DeploymentStatusRunning, db.desiredStatus)
			}
		})
	}
}

// Only the database and queue operations used by UpdateProject are implemented.
type slotUpdateDB struct {
	database.DB
	project       *database.Project
	desiredStatus database.DeploymentStatus
}

func (db *slotUpdateDB) FindDeploymentsForProject(_ context.Context, _, _, branch string) ([]*database.Deployment, error) {
	if branch != "" {
		return nil, nil
	}
	return []*database.Deployment{{ID: "dev-deployment", Environment: "dev", Branch: "feature", DesiredStatus: database.DeploymentStatusRunning}}, nil
}

func (db *slotUpdateDB) UpdateProject(_ context.Context, _ string, opts *database.UpdateProjectOptions) (*database.Project, error) {
	updated := *db.project
	updated.ProdSlots = opts.ProdSlots
	updated.DevSlots = opts.DevSlots
	return &updated, nil
}

func (db *slotUpdateDB) UpdateDeploymentDesiredStatus(_ context.Context, id string, desired database.DeploymentStatus) (*database.Deployment, error) {
	db.desiredStatus = desired
	return &database.Deployment{ID: id, DesiredStatus: desired}, nil
}

type slotUpdateJobs struct {
	jobs.Client
	deploymentIDs []string
}

func (j *slotUpdateJobs) ReconcileDeployment(_ context.Context, id string) (*jobs.InsertResult, error) {
	j.deploymentIDs = append(j.deploymentIDs, id)
	return &jobs.InsertResult{}, nil
}
