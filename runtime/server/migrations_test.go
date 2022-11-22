package server

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	_ "github.com/rilldata/rill/runtime/drivers/file"
	_ "github.com/rilldata/rill/runtime/drivers/sqlite"
	_ "github.com/rilldata/rill/runtime/services/catalog/artifacts/yaml"
	_ "github.com/rilldata/rill/runtime/services/catalog/migrator/sources"
	"github.com/rilldata/rill/runtime/services/catalog/testutils"
	"github.com/stretchr/testify/require"
)

const TestDataPath = "../../web-local/test/data"

var AdBidsCsvPath = filepath.Join(TestDataPath, "AdBids.csv")
var AdImpressionsCsvPath = filepath.Join(TestDataPath, "AdImpressions.tsv")

const AdBidsRepoPath = "/sources/AdBids.yaml"
const AdBidsNewRepoPath = "/sources/AdBidsNew.yaml"
const AdBidsModelRepoPath = "/models/AdBids_model.sql"

func TestServer_MigrateSingleSources(t *testing.T) {
	server, instanceId := getTestServer(t)

	ctx := context.Background()

	dir := t.TempDir()
	repoResp, err := server.CreateRepo(ctx, &runtimev1.CreateRepoRequest{
		Driver: "file",
		Dsn:    dir,
	})
	require.NoError(t, err)
	_, err = server.serviceCache.createCatalogService(ctx, server, instanceId, repoResp.Repo.RepoId)
	require.NoError(t, err)

	_, err = server.MigrateSingle(ctx, &runtimev1.MigrateSingleRequest{
		InstanceId: instanceId,
		Sql:        fmt.Sprintf("create source AdBids with connector = 'file', path = '%s'", AdBidsCsvPath),
		DryRun:     false,
	})
	require.NoError(t, err)
	assertTablePresence(t, server, instanceId, "AdBids", 100000)

	_, err = server.MigrateSingle(ctx, &runtimev1.MigrateSingleRequest{
		InstanceId: instanceId,
		Sql:        fmt.Sprintf("create source adbids with connector = 'file', path = '%s'", AdBidsCsvPath),
		DryRun:     false,
	})
	require.Error(t, err)

	_, err = server.MigrateSingle(ctx, &runtimev1.MigrateSingleRequest{
		InstanceId:      instanceId,
		Sql:             fmt.Sprintf("create source adbids with connector = 'file', path = '%s'", AdBidsCsvPath),
		DryRun:          false,
		CreateOrReplace: true,
	})
	require.NoError(t, err)
	assertTablePresence(t, server, instanceId, "adbids", 100000)

	_, err = server.MigrateSingle(ctx, &runtimev1.MigrateSingleRequest{
		InstanceId:      instanceId,
		Sql:             fmt.Sprintf("create source AdBids_New with connector = 'file', path = '%s'", AdBidsCsvPath),
		DryRun:          false,
		CreateOrReplace: true,
		RenameFrom:      "AdBids",
	})
	require.NoError(t, err)
	assertTablePresence(t, server, instanceId, "AdBids_New", 100000)
}

func TestServer_PutFileAndMigrate(t *testing.T) {
	server, instanceId := getTestServer(t)

	ctx := context.Background()
	dir := t.TempDir()

	repoResp, err := server.CreateRepo(ctx, &runtimev1.CreateRepoRequest{
		Driver: "file",
		Dsn:    dir,
	})
	require.NoError(t, err)
	service, err := server.serviceCache.createCatalogService(ctx, server, instanceId, repoResp.Repo.RepoId)
	require.NoError(t, err)

	artifact := testutils.CreateSource(t, service, "AdBids", AdBidsCsvPath, AdBidsRepoPath)
	resp, err := server.PutFileAndMigrate(ctx, &runtimev1.PutFileAndMigrateRequest{
		RepoId:     repoResp.Repo.RepoId,
		InstanceId: instanceId,
		Path:       AdBidsRepoPath,
		Blob:       artifact,
	})
	require.NoError(t, err)
	require.Len(t, resp.Errors, 0)
	testutils.AssertTable(t, service, "AdBids", AdBidsRepoPath)

	// replace with same name different file
	artifact = testutils.CreateSource(t, service, "AdBids", AdImpressionsCsvPath, AdBidsRepoPath)
	resp, err = server.PutFileAndMigrate(ctx, &runtimev1.PutFileAndMigrateRequest{
		RepoId:     repoResp.Repo.RepoId,
		InstanceId: instanceId,
		Path:       AdBidsRepoPath,
		Blob:       artifact,
	})
	require.NoError(t, err)
	require.Len(t, resp.Errors, 0)
	testutils.AssertTable(t, service, "AdBids", AdBidsRepoPath)

	// rename
	testutils.CreateSource(t, service, "AdBidsNew", AdBidsCsvPath, AdBidsRepoPath)
	renameResp, err := server.RenameFileAndMigrate(ctx, &runtimev1.RenameFileAndMigrateRequest{
		RepoId:     repoResp.Repo.RepoId,
		InstanceId: instanceId,
		FromPath:   AdBidsRepoPath,
		ToPath:     AdBidsNewRepoPath,
	})
	require.NoError(t, err)
	require.Len(t, renameResp.Errors, 0)
	testutils.AssertTableAbsence(t, service, "AdBids")
	testutils.AssertTable(t, service, "AdBidsNew", AdBidsNewRepoPath)

	// delete
	delResp, err := server.DeleteFileAndMigrate(ctx, &runtimev1.DeleteFileAndMigrateRequest{
		RepoId:     repoResp.Repo.RepoId,
		InstanceId: instanceId,
		Path:       AdBidsNewRepoPath,
	})
	require.NoError(t, err)
	require.Len(t, delResp.Errors, 0)
	testutils.AssertTableAbsence(t, service, "AdBids")
	testutils.AssertTableAbsence(t, service, "AdBidsNew")
}

func assertTablePresence(t *testing.T, server *Server, instanceId, tableName string, count int) {
	ctx := context.Background()

	resp, err := server.QueryDirect(ctx, &runtimev1.QueryDirectRequest{
		InstanceId: instanceId,
		Sql:        fmt.Sprintf("select count(*) as count from %s", tableName),
		Args:       nil,
		Priority:   0,
		DryRun:     false,
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.Data)
	require.Equal(t, int(resp.Data[0].Fields["count"].GetNumberValue()), count)

	catalog, _ := server.GetCatalogObject(context.Background(), &runtimev1.GetCatalogObjectRequest{
		InstanceId: instanceId,
		Name:       tableName,
	})
	require.WithinDuration(t, time.Now(), catalog.GetObject().RefreshedOn.AsTime(), time.Second)
}
