package cmdutil

import (
	"context"
	"fmt"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/archive"
)

// UploadRepo uploads a local project files to rill managed store.
// Internally it creates an asset object on admin service and returns its id which can be supplied while creating/updating project.
func UploadRepo(ctx context.Context, repo drivers.RepoStore, ch *Helper, org, name string) (string, error) {
	// list files
	entries, err := repo.ListRecursive(ctx, "**", false)
	if err != nil {
		return "", err
	}

	adminClient, err := ch.Client()
	if err != nil {
		return "", err
	}

	// generate a upload URL
	asset, err := adminClient.CreateAsset(ctx, &adminv1.CreateAssetRequest{
		OrganizationName: org,
		Type:             "deploy",
		Name:             fmt.Sprintf("%s__%s", org, name),
		Extension:        "tar.gz",
	})
	if err != nil {
		return "", err
	}

	rootPath, err := repo.Root(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get root path: %w", err)
	}

	err = archive.Create(ctx, entries, rootPath, asset.SignedUrl, asset.SigningHeaders)
	if err != nil {
		return "", err
	}

	return asset.AssetId, nil
}
