package testruntime

import (
	"context"
	"fmt"
	"path/filepath"
	goruntime "runtime"

	"github.com/docker/go-connections/nat"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
)

func ClickhouseCluster(t TestingT) (string, string) {
	_, currentFile, _, _ := goruntime.Caller(0)

	compose, err := tc.NewDockerCompose(filepath.Join(currentFile, "..", "testdata", "ch_cluster_2S_2R", "docker-compose.yaml"))
	require.NoError(t, err, "NewDockerComposeAPI()")

	t.Cleanup(func() {
		require.NoError(t, compose.Down(context.Background(), tc.RemoveOrphans(true), tc.RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	compose.WaitForService("clickhouse-01", wait.ForAll(
		wait.NewHTTPStrategy("/").WithPort(nat.Port("8123/tcp")).WithStatusCodeMatcher(func(status int) bool {
			return status == 200
		}),
	))
	compose.WaitForService("clickhouse-02", wait.ForAll(
		wait.NewHTTPStrategy("/").WithPort(nat.Port("8123/tcp")).WithStatusCodeMatcher(func(status int) bool {
			return status == 200
		}),
	))
	compose.WaitForService("clickhouse-03", wait.ForAll(
		wait.NewHTTPStrategy("/").WithPort(nat.Port("8123/tcp")).WithStatusCodeMatcher(func(status int) bool {
			return status == 200
		}),
	))
	compose.WaitForService("clickhouse-04", wait.ForAll(
		wait.NewHTTPStrategy("/").WithPort(nat.Port("8123/tcp")).WithStatusCodeMatcher(func(status int) bool {
			return status == 200
		}),
	))
	compose.WaitForService("clickhouse-keeper-01", wait.ForAll(
		wait.NewHTTPStrategy("/ready").WithPort(nat.Port("9182/tcp")).WithStatusCodeMatcher(func(status int) bool {
			return status == 200
		}),
	))

	require.NoError(t, compose.Up(ctx, tc.Wait(true)), "compose.Up()")

	container, err := compose.ServiceContainer(ctx, "clickhouse-01")
	require.NoError(t, err, "compose.ServiceContainer()")

	port, err := container.MappedPort(ctx, nat.Port("9000/tcp"))
	require.NoError(t, err, "container.MappedPort()")

	host, err := container.Host(ctx)
	require.NoError(t, err, "container.Host()")
	return fmt.Sprintf("clickhouse://default@%s:%s", host, port.Port()), "cluster_2S_2R"
}

func NewInstanceWithClickhouseProject(t TestingT, withCluster bool) (*runtime.Runtime, string) {
	dsn, cluster := ClickhouseCluster(t)
	rt := New(t)
	_, currentFile, _, _ := goruntime.Caller(0)
	projectPath := filepath.Join(currentFile, "..", "testdata", "ad_bids_clickhouse")

	olapConfig := map[string]string{"dsn": dsn}
	if withCluster {
		olapConfig["cluster"] = cluster
		olapConfig["log_queries"] = "true"
	}
	inst := &drivers.Instance{
		Environment:      "test",
		OLAPConnector:    "duckdb",
		RepoConnector:    "repo",
		CatalogConnector: "catalog",
		Connectors: []*runtimev1.Connector{
			{
				Type:   "file",
				Name:   "repo",
				Config: map[string]string{"dsn": projectPath},
			},
			{
				Type:   "clickhouse",
				Name:   "clickhouse",
				Config: olapConfig,
			},
			{
				Type: "sqlite",
				Name: "catalog",
				// Setting a test-specific name ensures a unique connection when "cache=shared" is enabled.
				// "cache=shared" is needed to prevent threading problems.
				Config: map[string]string{"dsn": fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())},
			},
		},
		Variables: map[string]string{"rill.stage_changes": "false"},
	}

	err := rt.CreateInstance(context.Background(), inst)
	require.NoError(t, err)
	require.NotEmpty(t, inst.ID)

	ctrl, err := rt.Controller(context.Background(), inst.ID)
	require.NoError(t, err)

	_, err = ctrl.Get(context.Background(), runtime.GlobalProjectParserName, false)
	require.NoError(t, err)

	err = ctrl.WaitUntilIdle(context.Background(), false)
	require.NoError(t, err)

	return rt, inst.ID
}
