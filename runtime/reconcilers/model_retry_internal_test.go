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
	outcome, err := r.executeWithRetry(context.Background(), &runtimev1.Resource{Meta: &runtimev1.ResourceMeta{Name: &runtimev1.ResourceName{Name: "m"}}}, model, func(context.Context) (*drivers.ModelResult, error) {
		attempts++
		if attempts < 3 {
			return nil, errors.New("retryable: temporary error")
		}
		return &drivers.ModelResult{Table: "ok"}, nil
	})

	require.NoError(t, err)
	require.NotNil(t, outcome)
	require.NotNil(t, outcome.Result)
	require.Equal(t, "ok", outcome.Result.Table)
	require.Equal(t, 3, attempts)
	require.Equal(t, uint32(3), outcome.Retry.Used)
	require.Equal(t, uint32(3), outcome.Retry.Max)
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
	outcome, err := r.executeWithRetry(context.Background(), &runtimev1.Resource{Meta: &runtimev1.ResourceMeta{Name: &runtimev1.ResourceName{Name: "m"}}}, model, func(context.Context) (*drivers.ModelResult, error) {
		attempts++
		return nil, errors.New("hard error")
	})

	require.Error(t, err)
	require.NotNil(t, outcome)
	require.Equal(t, 1, attempts)
	require.Equal(t, uint32(1), outcome.Retry.Used)
	require.Equal(t, uint32(3), outcome.Retry.Max)
}

func uint32Ptr(v uint32) *uint32 {
	return &v
}
