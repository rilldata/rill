package testclickhouse

import (
	"context"
	"fmt"
	"path/filepath"
	goruntime "runtime"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/clickhouse"
	tc "github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestingT satisfies both *testing.T and *testing.B.
type TestingT interface {
	Name() string
	TempDir() string
	FailNow()
	Errorf(format string, args ...interface{})
	Cleanup(f func())
}

// Start starts a ClickHouse container for testing.
// It returns the DSN for connecting to the container.
// The container is automatically terminated when the test ends.
func Start(t TestingT) string {
	_, currentFile, _, _ := goruntime.Caller(0)
	testdataPath := filepath.Join(currentFile, "..", "testdata")

	ctx := context.Background()
	clickHouseContainer, err := clickhouse.Run(
		ctx,
		"clickhouse/clickhouse-server:25.5.1.2782",
		clickhouse.WithConfigFile(filepath.Join(testdataPath, "clickhouse-config.xml")),
		testcontainers.CustomizeRequestOption(func(req *testcontainers.GenericContainerRequest) error {
			cf := testcontainers.ContainerFile{
				HostFilePath:      filepath.Join(testdataPath, "users.xml"),
				ContainerFilePath: "/etc/clickhouse-server/users.d/default.xml",
				FileMode:          0o755,
			}
			req.Files = append(req.Files, cf)
			return nil
		}),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		err := clickHouseContainer.Terminate(ctx)
		require.NoError(t, err)
	})

	host, err := clickHouseContainer.Host(ctx)
	require.NoError(t, err)
	port, err := clickHouseContainer.MappedPort(ctx, "9000/tcp")
	require.NoError(t, err)

	dsn := fmt.Sprintf("clickhouse://default:default@%v:%v", host, port.Port())
	return dsn
}

// StartCluster starts a ClickHouse cluster for testing.
// It returns the DSN for connecting to the cluster and the cluster name.
// The cluster is automatically terminated when the test ends.
func StartCluster(t TestingT) (string, string) {
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
