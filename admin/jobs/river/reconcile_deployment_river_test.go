package river

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rilldata/rill/admin/pkg/pgtestcontainer"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/require"
)

func TestReconcileDeploymentUniquenessIncludesRetryableJobs(t *testing.T) {
	// River enforces uniqueness in Postgres, so this test uses the real schema
	// and concurrent inserts instead of approximating the unique-key behavior.
	pg := pgtestcontainer.New(t)
	t.Cleanup(func() { pg.Terminate(t) })
	pool, err := pgxpool.New(t.Context(), pg.DatabaseURL)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	driver := riverpgxv5.New(pool)
	migrator, err := rivermigrate.New(driver, nil)
	require.NoError(t, err)
	_, err = migrator.Migrate(t.Context(), rivermigrate.DirectionUp, nil)
	require.NoError(t, err)

	riverClient, err := river.NewClient(driver, &river.Config{TestOnly: true})
	require.NoError(t, err)
	client := &Client{riverClient: riverClient, dbPool: pool}

	states := []rivertype.JobState{
		rivertype.JobStateAvailable,
		rivertype.JobStateRunning,
		rivertype.JobStateScheduled,
		rivertype.JobStateRetryable,
	}
	for _, state := range states {
		t.Run(string(state), func(t *testing.T) {
			// Each state represents an operative job. Sixteen simultaneous callers
			// must all resolve to the original job, including during retry backoff.
			deploymentID := fmt.Sprintf("deployment-%s", state)
			original, err := client.ReconcileDeployment(t.Context(), deploymentID)
			require.NoError(t, err)
			require.False(t, original.Duplicate)

			if state != rivertype.JobStateAvailable {
				_, err = pool.Exec(t.Context(), "UPDATE river_job SET state = $1 WHERE id = $2", state, original.ID)
				require.NoError(t, err)
			}

			const callers = 16
			start := make(chan struct{})
			results := make([]int64, callers)
			duplicates := make([]bool, callers)
			errs := make([]error, callers)
			var wg sync.WaitGroup
			for i := 0; i < callers; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					<-start
					result, insertErr := client.ReconcileDeployment(context.Background(), deploymentID)
					errs[index] = insertErr
					if result != nil {
						results[index] = result.ID
						duplicates[index] = result.Duplicate
					}
				}(i)
			}
			close(start)
			wg.Wait()

			for i := range callers {
				require.NoError(t, errs[i])
				require.True(t, duplicates[i], "caller %d inserted a conflicting job", i)
				require.Equal(t, original.ID, results[i])
			}
			var count int
			err = pool.QueryRow(t.Context(), `SELECT COUNT(*) FROM river_job WHERE kind = $1 AND args->>'DeploymentID' = $2`, ReconcileDeploymentArgs{}.Kind(), deploymentID).Scan(&count)
			require.NoError(t, err)
			require.Equal(t, 1, count)
		})
	}
}
