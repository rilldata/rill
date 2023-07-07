package server

import (
	"fmt"
	"github.com/bmatcuk/doublestar/v4"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

const (
	WatchFilesMethod = "WatchFiles"
)

func (s *Server) WatchFiles(req *runtimev1.WatchFilesRequest, stream runtimev1.RuntimeService_WatchFilesServer) error {
	repo, err := s.runtime.Repo(stream.Context(), req.InstanceId)
	if err != nil {
		return err
	}

	if repo.Driver() != "file" {
		return fmt.Errorf("%s repository is not supported", repo.Driver())
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	fsRoot := os.DirFS(repo.Root())

	var dirs []string
	var files []string

	err = doublestar.GlobWalk(fsRoot, "**", func(p string, d fs.DirEntry) error {
		if d.IsDir() {
			dirs = append(dirs, p)
		} else {
			files = append(files, p)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if req.Replay {
		for _, f := range files {
			err := stream.Send(&runtimev1.WatchFilesResponse{
				Event: runtimev1.FileEvent_FILE_EVENT_WRITE,
				Path:  f,
			})
			if err != nil {
				return err
			}
		}
	}

	for _, path := range dirs {
		relativePath := filepath.Join(repo.Root(), path)
		fi, err := os.Stat(relativePath)
		if err != nil {
			return err
		}

		if fi.IsDir() {
			err := watcher.Add(relativePath)
			if err != nil {
				return err
			}
		}
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil // todo should we notify the user?
			}
			resp := &runtimev1.WatchFilesResponse{
				Path: event.Name,
			}
			switch event.Op {
			case fsnotify.Write:
				resp.Event = runtimev1.FileEvent_FILE_EVENT_WRITE
			case fsnotify.Remove:
				resp.Event = runtimev1.FileEvent_FILE_EVENT_DELETE
			case fsnotify.Rename:
				resp.Event = runtimev1.FileEvent_FILE_EVENT_RENAME
			default:
				resp.Event = runtimev1.FileEvent_FILE_EVENT_UNSPECIFIED
			}
			err := stream.Send(resp)
			if err != nil {
				return err
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			return err
		case _, ok := <-stream.Context().Done():
			if !ok {
				return nil
			}

			return stream.Context().Err()
		}
	}
}
