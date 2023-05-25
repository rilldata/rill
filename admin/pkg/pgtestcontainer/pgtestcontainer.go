package pgtestcontainer

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type Container struct {
	testcontainers.Container
	DatabaseURL string
}

func New(t *testing.T) Container {
	ctx := context.Background()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:14",
			ExposedPorts: []string{"5432/tcp"},
			WaitingFor:   wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(15 * time.Second),
			Env: map[string]string{
				"POSTGRES_USER":     "postgres",
				"POSTGRES_PASSWORD": "postgres",
				"POSTGRES_DB":       "postgres",
			},
		},
	})
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)
	databaseURL := fmt.Sprintf("postgres://postgres:postgres@%s:%d/postgres", host, port.Int())

	return Container{
		Container:   container,
		DatabaseURL: databaseURL,
	}
}

func (c Container) Terminate(t *testing.T) {
	ctx := context.Background()
	err := c.Container.Terminate(ctx)
	require.NoError(t, err)
}
