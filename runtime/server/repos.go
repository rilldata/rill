package server

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ListFiles implements RuntimeService.
func (s *Server) ListFiles(ctx context.Context, req *runtimev1.ListFilesRequest) (*runtimev1.ListFilesResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.glob", req.Glob),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadRepo) {
		return nil, ErrForbidden
	}

	glob := req.Glob
	if glob == "" {
		glob = "**"
	}

	files, err := s.runtime.ListFiles(ctx, req.InstanceId, glob)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	var entries []*runtimev1.DirEntry
	for _, file := range files {
		entries = append(entries, &runtimev1.DirEntry{
			Path:  file.Path,
			IsDir: file.IsDir,
		})
	}

	return &runtimev1.ListFilesResponse{Files: entries}, nil
}

// WatchFiles implements RuntimeService.
func (s *Server) WatchFiles(req *runtimev1.WatchFilesRequest, ss runtimev1.RuntimeService_WatchFilesServer) error {
	observability.AddRequestAttributes(ss.Context(),
		attribute.String("args.instance_id", req.InstanceId),
		attribute.Bool("args.replay", req.Replay),
	)

	if !auth.GetClaims(ss.Context()).CanInstance(req.InstanceId, auth.ReadRepo) {
		return ErrForbidden
	}

	repo, release, err := s.runtime.Repo(ss.Context(), req.InstanceId)
	if err != nil {
		return err
	}
	defer release()

	if req.Replay {
		files, err := repo.ListRecursive(ss.Context(), "**", false)
		if err != nil {
			return err
		}
		for _, f := range files {
			err = ss.Send(&runtimev1.WatchFilesResponse{
				Event: runtimev1.FileEvent_FILE_EVENT_WRITE,
				Path:  f.Path,
				IsDir: f.IsDir,
			})
			if err != nil {
				return err
			}
		}
	}

	return repo.Watch(ss.Context(), func(events []drivers.WatchEvent) {
		for _, event := range events {
			err := ss.Send(&runtimev1.WatchFilesResponse{
				Event: event.Type,
				Path:  event.Path,
				IsDir: event.Dir,
			})
			if err != nil {
				s.logger.Info("failed to send watch event", zap.Error(err))
			}
		}
	})
}

// GetFile implements RuntimeService.
func (s *Server) GetFile(ctx context.Context, req *runtimev1.GetFileRequest) (*runtimev1.GetFileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.path", req.Path),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadRepo) {
		return nil, ErrForbidden
	}

	unescapedPath, err := url.PathUnescape(req.Path)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	blob, lastUpdated, err := s.runtime.GetFile(ctx, req.InstanceId, unescapedPath)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.GetFileResponse{Blob: blob, UpdatedOn: timestamppb.New(lastUpdated)}, nil
}

// PutFile implements RuntimeService.
func (s *Server) PutFile(ctx context.Context, req *runtimev1.PutFileRequest) (*runtimev1.PutFileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.path", req.Path),
		attribute.Bool("args.create", req.Create),
		attribute.Bool("args.create_only", req.CreateOnly),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.EditRepo) {
		return nil, ErrForbidden
	}

	unescapedPath, err := url.PathUnescape(req.Path)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	err = s.runtime.PutFile(ctx, req.InstanceId, unescapedPath, strings.NewReader(req.Blob), req.Create, req.CreateOnly)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.PutFileResponse{}, nil
}

// CreateDirectory implements RuntimeService.
func (s *Server) CreateDirectory(ctx context.Context, req *runtimev1.CreateDirectoryRequest) (*runtimev1.CreateDirectoryResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.path", req.Path),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.EditRepo) {
		return nil, ErrForbidden
	}

	unescapedPath, err := url.PathUnescape(req.Path)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	err = s.runtime.MakeDir(ctx, req.InstanceId, unescapedPath)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.CreateDirectoryResponse{}, nil
}

// DeleteFile implements RuntimeService.
func (s *Server) DeleteFile(ctx context.Context, req *runtimev1.DeleteFileRequest) (*runtimev1.DeleteFileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.path", req.Path),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.EditRepo) {
		return nil, ErrForbidden
	}

	unescapedPath, err := url.PathUnescape(req.Path)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	err = s.runtime.DeleteFile(ctx, req.InstanceId, unescapedPath, req.Force)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.DeleteFileResponse{}, nil
}

// RenameFile implements RuntimeService.
func (s *Server) RenameFile(ctx context.Context, req *runtimev1.RenameFileRequest) (*runtimev1.RenameFileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.from_path", req.FromPath),
		attribute.String("args.to_path", req.ToPath),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.EditRepo) {
		return nil, ErrForbidden
	}

	err := s.runtime.RenameFile(ctx, req.InstanceId, req.FromPath, req.ToPath)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.RenameFileResponse{}, nil
}

// UploadMultipartFile implements the same functionality as PutFile, but for multipart HTTP upload.
// It's mounted only on as a REST API and enables upload of large files (such as data files).
func (s *Server) UploadMultipartFile(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
	if !auth.GetClaims(req.Context()).CanInstance(pathParams["instance_id"], auth.EditRepo) {
		http.Error(w, "action not allowed", http.StatusUnauthorized)
		return
	}

	ctx := context.Background()
	if err := req.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse request: %s", err), http.StatusBadRequest)
		return
	}

	f, _, err := req.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse file in request: %s", err), http.StatusBadRequest)
		return
	}

	if pathParams["path"] == "" {
		http.Error(w, "must have a path to file", http.StatusBadRequest)
		return
	}

	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", pathParams["instance_id"]),
		attribute.String("args.path", pathParams["path"]),
	)

	s.addInstanceRequestAttributes(ctx, pathParams["instance_id"])

	err = s.runtime.PutFile(ctx, pathParams["instance_id"], pathParams["path"], f, true, false)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to write file: %s", err), http.StatusBadRequest)
		return
	}

	res, err := protojson.Marshal(&runtimev1.PutFileResponse{})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to serialize response: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(res)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to write response data: %s", err), http.StatusInternalServerError)
		return
	}
}
