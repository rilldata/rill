package reconcilers

import (
	"context"
	"errors"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestExecuteWithRetry_TracksRetryStatsOnSuccessAfterRetries(t *testing.T) {
	r := &ModelReconciler{C: &runtime.Controller{Logger: zap.NewNop()}}
	model := &runtimev1.Model{
		Spec: &runtimev1.ModelSpec{
			RetryAttempts:       uint32Ptr(3),
			RetryDelaySeconds:   uint32Ptr(0),
			RetryIfErrorMatches: []string{"retryable"},
		},
	}

	attempts := 0
	updates := make([]modelRetryStats, 0, 4)
	res, err := r.executeWithRetry(context.Background(), &runtimev1.Resource{Meta: &runtimev1.ResourceMeta{Name: &runtimev1.ResourceName{Name: "m"}}}, model, func(context.Context) (*drivers.ModelResult, error) {
		attempts++
		if attempts < 3 {
			return nil, errors.New("retryable: temporary error")
		}
		return &drivers.ModelResult{Table: "ok"}, nil
	}, func(stats modelRetryStats, _ error) error {
		updates = append(updates, stats)
		return nil
	})

	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, "ok", res.Table)
	require.Equal(t, 3, attempts)
	require.Equal(t, []modelRetryStats{{Used: 0, Max: 3}, {Used: 1, Max: 3}, {Used: 2, Max: 3}, {Used: 3, Max: 3}}, updates)
}

func TestExecuteWithRetry_TracksRetryStatsOnNonRetryableFailure(t *testing.T) {
	r := &ModelReconciler{C: &runtime.Controller{Logger: zap.NewNop()}}
	model := &runtimev1.Model{
		Spec: &runtimev1.ModelSpec{
			RetryAttempts:       uint32Ptr(3),
			RetryDelaySeconds:   uint32Ptr(0),
			RetryIfErrorMatches: []string{"retryable"},
		},
	}

	attempts := 0
	updates := make([]modelRetryStats, 0, 2)
	res, err := r.executeWithRetry(context.Background(), &runtimev1.Resource{Meta: &runtimev1.ResourceMeta{Name: &runtimev1.ResourceName{Name: "m"}}}, model, func(context.Context) (*drivers.ModelResult, error) {
		attempts++
		return nil, errors.New("hard error")
	}, func(stats modelRetryStats, _ error) error {
		updates = append(updates, stats)
		return nil
	})

	require.Error(t, err)
	require.Nil(t, res)
	require.Equal(t, 1, attempts)
	require.Equal(t, []modelRetryStats{{Used: 0, Max: 3}, {Used: 1, Max: 3}}, updates)
}

func TestExecuteWithRetry_TracksRetryStatsOnExhaustedRetries(t *testing.T) {
	r := &ModelReconciler{C: &runtime.Controller{Logger: zap.NewNop()}}
	model := &runtimev1.Model{
		Spec: &runtimev1.ModelSpec{
			RetryAttempts:       uint32Ptr(2),
			RetryDelaySeconds:   uint32Ptr(0),
			RetryIfErrorMatches: []string{"retryable"},
		},
	}

	attempts := 0
	updates := make([]modelRetryStats, 0, 3)
	res, err := r.executeWithRetry(context.Background(), &runtimev1.Resource{Meta: &runtimev1.ResourceMeta{Name: &runtimev1.ResourceName{Name: "m"}}}, model, func(context.Context) (*drivers.ModelResult, error) {
		attempts++
		return nil, errors.New("retryable: temporary error")
	}, func(stats modelRetryStats, _ error) error {
		updates = append(updates, stats)
		return nil
	})

	require.Error(t, err)
	require.Nil(t, res)
	require.Equal(t, 2, attempts)
	require.Equal(t, []modelRetryStats{{Used: 0, Max: 2}, {Used: 1, Max: 2}, {Used: 2, Max: 2}}, updates)
}

func uint32Ptr(v uint32) *uint32 {
	return &v
}
