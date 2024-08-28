package testruntime

import (
	"context"
	"fmt"
	"path/filepath"
	goruntime "runtime"

	"github.com/docker/go-connections/nat"
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
