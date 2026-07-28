package server

import (
	"context"
	"errors"
	"os"

	"github.com/rilldata/rill/admin/database"
	"go.uber.org/zap"
)

// userFileContent returns the effective content of a user file.
// For staged (dirty) rows that is the staged bytes.
// For synced rows the file in Git is authoritative, so the current content on the served branch is
// fetched (users may have edited it there since the sync). If the file was deleted from Git, an error
// wrapping os.ErrNotExist is returned: the deletion is authoritative and the row's stale copy must not
// resurrect the file. If Git is unreachable for any other reason, it falls back to the last-synced copy.
func (s *Server) userFileContent(ctx context.Context, proj *database.Project, vf *database.VirtualFile) ([]byte, error) {
	if vf.SyncState != database.VirtualFileSyncStateSynced {
		return vf.Data, nil
	}
	data, err := s.admin.ReadUserFileFromGit(ctx, proj, vf.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		s.logger.Warn("failed to read user file from git, falling back to last-synced copy",
			zap.String("project_id", proj.ID), zap.String("path", vf.Path), zap.Error(err))
		return vf.Data, nil
	}
	return data, nil
}
