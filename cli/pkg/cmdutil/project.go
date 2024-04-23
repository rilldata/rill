package cmdutil

import (
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"go.uber.org/zap"

	// Ensure file driver is loaded
	_ "github.com/rilldata/rill/runtime/drivers/file"
)

// RepoForProjectPath creates an ad-hoc drivers.RepoStore for a local project file path
func RepoForProjectPath(path string) (drivers.RepoStore, string, error) {
	instanceID := "default"
	repoHandle, err := drivers.Open("file", instanceID, map[string]any{"dsn": path}, activity.NewNoopClient(), zap.NewNop())
	if err != nil {
		return nil, "", err
	}
	repo, _ := repoHandle.AsRepoStore(instanceID)
	return repo, instanceID, nil
}
