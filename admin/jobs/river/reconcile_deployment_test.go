package river

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/rilldata/rill/admin/database"
	"github.com/riverqueue/river"
	"github.com/stretchr/testify/require"
)

func TestReconcileDeploymentStateMatrix(t *testing.T) {
	// Exercise the complete desired/current status matrix so adding a status
	// cannot silently turn a destructive transition into the default no-op path.
	statuses := []database.DeploymentStatus{
		database.DeploymentStatusUnspecified,
		database.DeploymentStatusPending,
		database.DeploymentStatusRunning,
		database.DeploymentStatusErrored,
		database.DeploymentStatusStopped,
		database.DeploymentStatusUpdating,
		database.DeploymentStatusStopping,
		database.DeploymentStatusDeleting,
		database.DeploymentStatusDeleted,
	}
	for _, desired := range statuses {
		for _, current := range statuses {
			name := fmt.Sprintf("desired_%s/current_%s", desired, current)
			t.Run(name, func(t *testing.T) {
				deployment := &database.Deployment{
					ID:                     "deployment",
					Status:                 current,
					DesiredStatus:          desired,
					DesiredStatusUpdatedOn: time.Unix(100, 0),
				}
				var actions []string
				var statusWrites []database.DeploymentStatus
				worker := newStateMatrixWorker(deployment, &actions, &statusWrites)

				err := worker.Work(t.Context(), reconcileDeploymentJob())

				require.NoError(t, err)
				wantActions, wantStatuses := expectedReconcileTransition(desired, current)
				require.Equal(t, wantActions, actions)
				require.Equal(t, wantStatuses, statusWrites)
			})
		}
	}
}

func TestReconcileDeploymentFailureBoundaries(t *testing.T) {
	// Every database boundary and external lifecycle call must stop subsequent
	// work, preserving an intermediate status that a retry can safely resume.
	injected := errors.New("injected failure")
	tests := []struct {
		name            string
		desired         database.DeploymentStatus
		current         database.DeploymentStatus
		failStatusCall  int
		failAction      string
		wantStatusCalls int
		wantActionCalls int
	}{
		{name: "start pre-status", desired: database.DeploymentStatusRunning, current: database.DeploymentStatusStopped, failStatusCall: 1, wantStatusCalls: 1},
		{name: "start operation", desired: database.DeploymentStatusRunning, current: database.DeploymentStatusStopped, failAction: "start", wantStatusCalls: 1, wantActionCalls: 1},
		{name: "start final status", desired: database.DeploymentStatusRunning, current: database.DeploymentStatusStopped, failStatusCall: 2, wantStatusCalls: 2, wantActionCalls: 1},
		{name: "running update", desired: database.DeploymentStatusRunning, current: database.DeploymentStatusRunning, failAction: "update", wantActionCalls: 1},
		{name: "running final status", desired: database.DeploymentStatusRunning, current: database.DeploymentStatusRunning, failStatusCall: 1, wantStatusCalls: 1, wantActionCalls: 1},
		{name: "stop pre-status", desired: database.DeploymentStatusStopped, current: database.DeploymentStatusRunning, failStatusCall: 1, wantStatusCalls: 1},
		{name: "stop operation", desired: database.DeploymentStatusStopped, current: database.DeploymentStatusRunning, failAction: "stop", wantStatusCalls: 1, wantActionCalls: 1},
		{name: "stop final status", desired: database.DeploymentStatusStopped, current: database.DeploymentStatusRunning, failStatusCall: 2, wantStatusCalls: 2, wantActionCalls: 1},
		{name: "delete pre-status", desired: database.DeploymentStatusDeleted, current: database.DeploymentStatusRunning, failStatusCall: 1, wantStatusCalls: 1},
		{name: "delete operation", desired: database.DeploymentStatusDeleted, current: database.DeploymentStatusRunning, failAction: "delete", wantStatusCalls: 1, wantActionCalls: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deployment := &database.Deployment{ID: "deployment", Status: tt.current, DesiredStatus: tt.desired}
			statusCalls := 0
			actionCalls := 0
			worker := &ReconcileDeploymentWorker{
				findDeployment: func(context.Context, string) (*database.Deployment, error) {
					clone := *deployment
					return &clone, nil
				},
				updateDeploymentStatus: func(_ context.Context, _ string, status database.DeploymentStatus, _ string) (*database.Deployment, error) {
					statusCalls++
					if statusCalls == tt.failStatusCall {
						return nil, injected
					}
					clone := *deployment
					clone.Status = status
					deployment = &clone
					return &clone, nil
				},
			}
			setLifecycleAction := func(name string) func(context.Context, *database.Deployment) error {
				return func(context.Context, *database.Deployment) error {
					actionCalls++
					if tt.failAction == name {
						return injected
					}
					return nil
				}
			}
			worker.startDeployment = setLifecycleAction("start")
			worker.updateDeployment = setLifecycleAction("update")
			worker.stopDeployment = setLifecycleAction("stop")
			worker.deleteDeployment = setLifecycleAction("delete")

			err := worker.Work(t.Context(), reconcileDeploymentJob())

			require.ErrorIs(t, err, injected)
			require.Equal(t, tt.wantStatusCalls, statusCalls)
			require.Equal(t, tt.wantActionCalls, actionCalls)
		})
	}
}

func TestReconcileDeploymentLookupFailures(t *testing.T) {
	// A deleted deployment is a successful idempotent no-op, while an actual
	// database outage must remain retryable and must not invoke lifecycle work.
	t.Run("not found", func(t *testing.T) {
		worker := &ReconcileDeploymentWorker{findDeployment: func(context.Context, string) (*database.Deployment, error) {
			return nil, database.ErrNotFound
		}}
		require.NoError(t, worker.Work(t.Context(), reconcileDeploymentJob()))
	})

	t.Run("database failure", func(t *testing.T) {
		injected := errors.New("database unavailable")
		worker := &ReconcileDeploymentWorker{findDeployment: func(context.Context, string) (*database.Deployment, error) {
			return nil, injected
		}}
		require.ErrorIs(t, worker.Work(t.Context(), reconcileDeploymentJob()), injected)
	})
}

func TestReconcileDeploymentSchedulesOneFollowUpForConcurrentDesiredChange(t *testing.T) {
	// A desired-state write during an external start is detected from its
	// timestamp after the final status write and results in exactly one follow-up.
	originalUpdate := time.Unix(100, 0)
	changedUpdate := originalUpdate.Add(time.Second)
	deployment := &database.Deployment{
		ID:                     "deployment",
		Status:                 database.DeploymentStatusStopped,
		DesiredStatus:          database.DeploymentStatusRunning,
		DesiredStatusUpdatedOn: originalUpdate,
	}
	statusCalls := 0
	enqueueCalls := 0
	worker := &ReconcileDeploymentWorker{
		findDeployment: func(context.Context, string) (*database.Deployment, error) {
			clone := *deployment
			return &clone, nil
		},
		updateDeploymentStatus: func(_ context.Context, _ string, status database.DeploymentStatus, _ string) (*database.Deployment, error) {
			statusCalls++
			clone := *deployment
			clone.Status = status
			if statusCalls == 2 {
				clone.DesiredStatus = database.DeploymentStatusStopped
				clone.DesiredStatusUpdatedOn = changedUpdate
			}
			deployment = &clone
			return &clone, nil
		},
		startDeployment: func(context.Context, *database.Deployment) error { return nil },
		enqueueReconcile: func(_ context.Context, deploymentID string) (int64, error) {
			enqueueCalls++
			require.Equal(t, "deployment", deploymentID)
			return 42, nil
		},
	}

	err := worker.Work(t.Context(), reconcileDeploymentJob())

	require.NoError(t, err)
	require.Equal(t, 1, enqueueCalls)
}

func TestReconcileDeploymentRetryRepeatsOnlyIdempotentLifecycleOperation(t *testing.T) {
	// If the final status write fails, the next job observes the intermediate
	// state and repeats StartDeploymentInner, whose production contract is idempotent.
	deployment := &database.Deployment{ID: "deployment", Status: database.DeploymentStatusStopped, DesiredStatus: database.DeploymentStatusRunning}
	startCalls := 0
	statusCalls := 0
	failFinalOnce := true
	worker := &ReconcileDeploymentWorker{
		findDeployment: func(context.Context, string) (*database.Deployment, error) {
			clone := *deployment
			return &clone, nil
		},
		updateDeploymentStatus: func(_ context.Context, _ string, status database.DeploymentStatus, _ string) (*database.Deployment, error) {
			statusCalls++
			if status == database.DeploymentStatusRunning && failFinalOnce {
				failFinalOnce = false
				return nil, errors.New("final status unavailable")
			}
			clone := *deployment
			clone.Status = status
			deployment = &clone
			return &clone, nil
		},
		startDeployment: func(context.Context, *database.Deployment) error {
			startCalls++
			return nil
		},
	}

	require.Error(t, worker.Work(t.Context(), reconcileDeploymentJob()))
	require.NoError(t, worker.Work(t.Context(), reconcileDeploymentJob()))
	require.Equal(t, 2, startCalls)
	require.Equal(t, database.DeploymentStatusRunning, deployment.Status)
	require.Equal(t, 4, statusCalls)
}

func newStateMatrixWorker(deployment *database.Deployment, actions *[]string, statusWrites *[]database.DeploymentStatus) *ReconcileDeploymentWorker {
	current := *deployment
	worker := &ReconcileDeploymentWorker{
		findDeployment: func(context.Context, string) (*database.Deployment, error) {
			clone := current
			return &clone, nil
		},
		updateDeploymentStatus: func(_ context.Context, _ string, status database.DeploymentStatus, _ string) (*database.Deployment, error) {
			*statusWrites = append(*statusWrites, status)
			current.Status = status
			clone := current
			return &clone, nil
		},
	}
	worker.startDeployment = recordLifecycleAction("start", actions)
	worker.updateDeployment = recordLifecycleAction("update", actions)
	worker.stopDeployment = recordLifecycleAction("stop", actions)
	worker.deleteDeployment = recordLifecycleAction("delete", actions)
	return worker
}

func recordLifecycleAction(name string, actions *[]string) func(context.Context, *database.Deployment) error {
	return func(context.Context, *database.Deployment) error {
		*actions = append(*actions, name)
		return nil
	}
}

func expectedReconcileTransition(desired, current database.DeploymentStatus) ([]string, []database.DeploymentStatus) {
	switch desired {
	case database.DeploymentStatusRunning:
		if current == database.DeploymentStatusRunning {
			return []string{"update"}, []database.DeploymentStatus{database.DeploymentStatusRunning}
		}
		return []string{"start"}, []database.DeploymentStatus{database.DeploymentStatusPending, database.DeploymentStatusRunning}
	case database.DeploymentStatusStopped:
		if current == database.DeploymentStatusStopped {
			return nil, nil
		}
		return []string{"stop"}, []database.DeploymentStatus{database.DeploymentStatusStopping, database.DeploymentStatusStopped}
	case database.DeploymentStatusDeleted:
		if current == database.DeploymentStatusDeleted {
			return nil, nil
		}
		return []string{"delete"}, []database.DeploymentStatus{database.DeploymentStatusDeleting}
	default:
		return nil, nil
	}
}

func reconcileDeploymentJob() *river.Job[ReconcileDeploymentArgs] {
	return &river.Job[ReconcileDeploymentArgs]{Args: ReconcileDeploymentArgs{DeploymentID: "deployment"}}
}
