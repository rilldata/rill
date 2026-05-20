package drivers_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func testRegistry(t *testing.T, reg drivers.RegistryStore) {
	ctx := context.Background()
	inst := &drivers.Instance{
		Environment:      "test",
		OLAPConnector:    "duckdb",
		RepoConnector:    "repo",
		CatalogConnector: "catalog",
		Connectors: []*runtimev1.Connector{
			{
				Type:   "file",
				Name:   "repo",
				Config: testruntime.Must(structpb.NewStruct(map[string]any{"dsn": "."})),
			},
			{
				Type:   "duckdb",
				Name:   "duckdb",
				Config: testruntime.Must(structpb.NewStruct(map[string]any{"dsn": ":memory:"})),
			},
			{
				Type:   "sqlite",
				Name:   "catalog",
				Config: testruntime.Must(structpb.NewStruct(map[string]any{"dsn": "file:rill?mode=memory&cache=shared"})),
			},
		},
	}

	err := reg.CreateInstance(ctx, inst)
	require.NoError(t, err)
	_, err = uuid.Parse(inst.ID)
	require.NoError(t, err)
	require.Equal(t, "test", inst.Environment)
	require.Equal(t, "duckdb", inst.OLAPConnector)
	require.Equal(t, "repo", inst.RepoConnector)
	require.Equal(t, "catalog", inst.CatalogConnector)
	require.Greater(t, time.Minute, time.Since(inst.CreatedOn))
	require.Greater(t, time.Minute, time.Since(inst.UpdatedOn))

	// edit instance
	inst.ProjectDisplayName = "My Project"
	err = reg.EditInstance(ctx, inst)
	require.NoError(t, err)

	res, err := reg.FindInstance(ctx, inst.ID)
	require.NoError(t, err)
	require.Equal(t, inst.OLAPConnector, res.OLAPConnector)
	require.Equal(t, inst.RepoConnector, res.RepoConnector)
	require.Equal(t, inst.CatalogConnector, res.CatalogConnector)
	require.Equal(t, "My Project", res.ProjectDisplayName)
	require.ElementsMatch(t, inst.Connectors, res.Connectors)

	err = reg.CreateInstance(ctx, &drivers.Instance{OLAPConnector: "druid"})
	require.NoError(t, err)

	insts, err := reg.FindInstances(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, len(insts))

	err = reg.DeleteInstance(ctx, inst.ID)
	require.NoError(t, err)

	_, err = reg.FindInstance(ctx, inst.ID)
	require.EqualError(t, err, drivers.ErrNotFound.Error())

	insts, err = reg.FindInstances(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, len(insts))
}

func TestInstanceConfigAITimeouts(t *testing.T) {
	inst := &drivers.Instance{
		Environment: "prod",
		Variables: map[string]string{
			"rill.ai.llm_request_timeout_seconds": "600",
			"rill.ai.chat_timeout_seconds":        "600",
		},
	}

	cfg, err := inst.Config()
	require.NoError(t, err)
	require.Equal(t, uint32(600), cfg.AILLMRequestTimeoutSeconds)
	require.Equal(t, uint32(600), cfg.AIChatTimeoutSeconds)
}

func TestInstanceConfigAITimeoutDefaults(t *testing.T) {
	inst := &drivers.Instance{Environment: "prod"}

	cfg, err := inst.Config()
	require.NoError(t, err)
	require.Equal(t, uint32(180), cfg.AILLMRequestTimeoutSeconds)
	require.Equal(t, uint32(300), cfg.AIChatTimeoutSeconds)
}

func TestInstanceConfigAITimeoutsZeroFallsBackToDefault(t *testing.T) {
	inst := &drivers.Instance{
		Environment: "prod",
		Variables: map[string]string{
			"rill.ai.llm_request_timeout_seconds": "0",
			"rill.ai.chat_timeout_seconds":        "0",
		},
	}

	cfg, err := inst.Config()
	require.NoError(t, err)
	require.Equal(t, uint32(180), cfg.AILLMRequestTimeoutSeconds)
	require.Equal(t, uint32(300), cfg.AIChatTimeoutSeconds)
}

func TestInstanceConfigAITimeoutsClampedToMax(t *testing.T) {
	inst := &drivers.Instance{
		Environment: "prod",
		Variables: map[string]string{
			"rill.ai.llm_request_timeout_seconds": "86400",
			"rill.ai.chat_timeout_seconds":        "86400",
		},
	}

	cfg, err := inst.Config()
	require.NoError(t, err)
	require.Equal(t, uint32(1800), cfg.AILLMRequestTimeoutSeconds)
	require.Equal(t, uint32(1800), cfg.AIChatTimeoutSeconds)
}
