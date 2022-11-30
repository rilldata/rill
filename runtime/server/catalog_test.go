package server

import (
	"context"
	"path/filepath"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/services/catalog/testutils"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestServer_PutFileAndReconcile(t *testing.T) {
	ctx := context.Background()
	rt, instanceID := testruntime.NewInstance(t)
	srv, err := NewServer(&Options{}, rt, nil)
	require.NoError(t, err)

	cat, err := rt.Catalog(ctx, instanceID)
	require.NoError(t, err)

	sourcePath := "/sources/ad_bids_source.yaml"
	csvPath := filepath.Join("../testruntime/testdata/ad_bids/data", "AdBids.csv.gz")
	pqPath := filepath.Join("../testruntime/testdata/ad_bids/data", "AdImpressions.parquet")

	artifact := testutils.CreateSource(t, cat, "ad_bids_source", csvPath, sourcePath)
	resp, err := srv.PutFileAndReconcile(ctx, &runtimev1.PutFileAndReconcileRequest{
		InstanceId: instanceID,
		Path:       sourcePath,
		Blob:       artifact,
	})
	require.NoError(t, err)
	require.Len(t, resp.Errors, 0)
	testutils.AssertTable(t, cat, "ad_bids_source", sourcePath)

	// replace with same name different file
	artifact = testutils.CreateSource(t, cat, "ad_bids_source", pqPath, sourcePath)
	resp, err = srv.PutFileAndReconcile(ctx, &runtimev1.PutFileAndReconcileRequest{
		InstanceId: instanceID,
		Path:       sourcePath,
		Blob:       artifact,
	})
	require.NoError(t, err)
	require.Len(t, resp.Errors, 0)
	testutils.AssertTable(t, cat, "ad_bids_source", sourcePath)

	// rename
	testutils.CreateSource(t, cat, "ad_bids_new", csvPath, sourcePath)
	renameResp, err := srv.RenameFileAndReconcile(ctx, &runtimev1.RenameFileAndReconcileRequest{
		InstanceId: instanceID,
		FromPath:   sourcePath,
		ToPath:     "/sources/ad_bids_new.yaml",
	})
	require.NoError(t, err)
	require.Len(t, renameResp.Errors, 0)
	testutils.AssertTableAbsence(t, cat, "ad_bids_source")
	testutils.AssertTable(t, cat, "ad_bids_new", "/sources/ad_bids_new.yaml")

	// delete
	delResp, err := srv.DeleteFileAndReconcile(ctx, &runtimev1.DeleteFileAndReconcileRequest{
		InstanceId: instanceID,
		Path:       "/sources/ad_bids_new.yaml",
	})
	require.NoError(t, err)
	require.Len(t, delResp.Errors, 0)
	testutils.AssertTableAbsence(t, cat, "ad_bids_source")
	testutils.AssertTableAbsence(t, cat, "ad_bids_new")
}
