package cmdutil

import (
	"context"
	"os"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"go.uber.org/zap"

	// Ensure file driver is loaded
	_ "github.com/rilldata/rill/runtime/drivers/file"
)

// RepoForProjectPath creates an ad-hoc drivers.RepoStore for a local project file path
func RepoForProjectPath(path string) (drivers.RepoStore, string, error) {
	instanceID := "default"
	repoHandle, err := drivers.Open("file", instanceID, map[string]any{"dsn": path}, storage.MustNew(os.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	if err != nil {
		return nil, "", err
	}
	repo, _ := repoHandle.AsRepoStore(instanceID)
	return repo, instanceID, nil
}

func HasRillProject(dir string) bool {
	repo, _, err := RepoForProjectPath(dir)
	if err != nil {
		return false
	}
	_, err = repo.Get(context.Background(), "rill.yaml")
	return err == nil
}
