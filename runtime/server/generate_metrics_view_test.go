package server_test

import (
	"context"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGenerateMetricsView(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(testCtx(), 25*time.Second)
	defer cancel()

	_, err = server.GenerateMetricsViewFile(ctx, &runtimev1.GenerateMetricsViewFileRequest{
		InstanceId: instanceID,
		Table:      "ad_bids",
		Path:       "/dashboards/ad_bids_metrics_view.yaml",
		UseAi:      false,
	})
	require.NoError(t, err)

	repo, release, err := rt.Repo(ctx, instanceID)
	require.NoError(t, err)
	defer release()

	data, err := repo.Get(ctx, "/dashboards/ad_bids_metrics_view.yaml")
	require.NoError(t, err)

	require.Contains(t, data, "model: ad_bids")
	require.Contains(t, data, "valid_percent_of_total:")
	require.Contains(t, data, "measures:")
}
