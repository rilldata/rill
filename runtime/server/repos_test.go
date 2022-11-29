package server

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
)

func TestServer_Repos(t *testing.T) {
	s, instId := getTestServer(t)

	ctx := context.Background()

	_, err := s.PutFile(ctx, &runtimev1.PutFileRequest{
		InstanceId: instId,
		Path:       AdBidsRepoPath,
		Blob:       "version: 0.0.1\ntype: file",
		Create:     true,
		CreateOnly: true,
	})
	require.NoError(t, err)
	assertFiles(t, s, instId, []string{AdBidsRepoPath})

	_, err = s.PutFile(ctx, &runtimev1.PutFileRequest{
		InstanceId: instId,
		Path:       AdBidsModelRepoPath,
		Blob:       "select * from AdBids",
		Create:     true,
		CreateOnly: true,
	})
	require.NoError(t, err)
	assertFiles(t, s, instId, []string{AdBidsRepoPath, AdBidsModelRepoPath})

	// rename to existing fails
	_, err = s.RenameFile(ctx, &runtimev1.RenameFileRequest{
		InstanceId: instId,
		FromPath:   AdBidsRepoPath,
		ToPath:     AdBidsModelRepoPath,
	})
	require.Error(t, err)
	assertFiles(t, s, instId, []string{AdBidsRepoPath, AdBidsModelRepoPath})

	// rename to unique file succeeds
	_, err = s.RenameFile(ctx, &runtimev1.RenameFileRequest{
		InstanceId: instId,
		FromPath:   AdBidsRepoPath,
		ToPath:     AdBidsNewRepoPath,
	})
	require.NoError(t, err)
	assertFiles(t, s, instId, []string{AdBidsNewRepoPath, AdBidsModelRepoPath})

	// delete
	_, err = s.DeleteFile(ctx, &runtimev1.DeleteFileRequest{
		InstanceId: instId,
		Path:       AdBidsNewRepoPath,
	})
	require.NoError(t, err)
	assertFiles(t, s, instId, []string{AdBidsModelRepoPath})
}

func assertFiles(t *testing.T, s *Server, instId string, expectedPaths []string) {
	resp, err := s.ListFiles(context.Background(), &runtimev1.ListFilesRequest{
		InstanceId: instId,
	})
	require.NoError(t, err)
	require.ElementsMatch(t, resp.Paths, expectedPaths)
}
