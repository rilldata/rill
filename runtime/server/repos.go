package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/bufbuild/connect-go"
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
func (s *Server) ListFiles(ctx context.Context, req *connect.Request[runtimev1.ListFilesRequest]) (*connect.Response[runtimev1.ListFilesResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.Msg.InstanceId),
		attribute.String("args.glob", req.Msg.Glob),
	)

	s.addInstanceRequestAttributes(ctx, req.Msg.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.Msg.InstanceId, auth.ReadRepo) {
		return nil, ErrForbidden
	}

	glob := req.Msg.Glob
	if glob == "" {
		glob = "**"
	}

	paths, err := s.runtime.ListFiles(ctx, req.Msg.InstanceId, glob)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	return connect.NewResponse(&runtimev1.ListFilesResponse{Paths: paths}), nil
}

// WatchFiles implements RuntimeService.
func (s *Server) WatchFiles(ctx context.Context, req *connect.Request[runtimev1.WatchFilesRequest], ss *connect.ServerStream[runtimev1.WatchFilesResponse]) error {	
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.Msg.InstanceId),
		attribute.Bool("args.replay", req.Msg.Replay),
	)

	if !auth.GetClaims(ctx).CanInstance(req.Msg.InstanceId, auth.ReadRepo) {
		return ErrForbidden
	}

	repo, err := s.runtime.Repo(ctx, req.Msg.InstanceId)
	if err != nil {
		return err
	}

	if req.Msg.Replay {
		paths, err := repo.ListRecursive(ctx, req.Msg.InstanceId, "**")
		if err != nil {
			return err
		}
		for _, p := range paths {
			err = ss.Conn().Send(&runtimev1.WatchFilesResponse{
				Event: runtimev1.FileEvent_FILE_EVENT_WRITE,
				Path:  p,
			})
			if err != nil {
				return err
			}
		}
	}

	return repo.Watch(ctx, "", func(events []drivers.WatchEvent) {
		for _, event := range events {
			if !event.Dir {
				err := ss.Conn().Send(&runtimev1.WatchFilesResponse{
					Event: event.Type,
					Path:  event.Path,
				})
				if err != nil {
					s.logger.Info("failed to send watch event", zap.Error(err))
				}
			}
		}
	})
}

// GetFile implements RuntimeService.
func (s *Server) GetFile(ctx context.Context, req *connect.Request[runtimev1.GetFileRequest]) (*connect.Response[runtimev1.GetFileResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.Msg.InstanceId),
		attribute.String("args.path", req.Msg.Path),
	)

	s.addInstanceRequestAttributes(ctx, req.Msg.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.Msg.InstanceId, auth.ReadRepo) {
		return nil, ErrForbidden
	}

	blob, lastUpdated, err := s.runtime.GetFile(ctx, req.Msg.InstanceId, req.Msg.Path)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return connect.NewResponse(&runtimev1.GetFileResponse{Blob: blob, UpdatedOn: timestamppb.New(lastUpdated)}), nil
}

// PutFile implements RuntimeService.
func (s *Server) PutFile(ctx context.Context, req *connect.Request[runtimev1.PutFileRequest]) (*connect.Response[runtimev1.PutFileResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.Msg.InstanceId),
		attribute.String("args.path", req.Msg.Path),
		attribute.Bool("args.create", req.Msg.Create),
		attribute.Bool("args.create_only", req.Msg.CreateOnly),
	)

	s.addInstanceRequestAttributes(ctx, req.Msg.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.Msg.InstanceId, auth.EditRepo) {
		return nil, ErrForbidden
	}

	err := s.runtime.PutFile(ctx, req.Msg.InstanceId, req.Msg.Path, strings.NewReader(req.Msg.Blob), req.Msg.Create, req.Msg.CreateOnly)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return connect.NewResponse(&runtimev1.PutFileResponse{}), nil
}

// DeleteFile implements RuntimeService.
func (s *Server) DeleteFile(ctx context.Context, req *connect.Request[runtimev1.DeleteFileRequest]) (*connect.Response[runtimev1.DeleteFileResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.Msg.InstanceId),
		attribute.String("args.path", req.Msg.Path),
	)

	s.addInstanceRequestAttributes(ctx, req.Msg.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.Msg.InstanceId, auth.EditRepo) {
		return nil, ErrForbidden
	}

	err := s.runtime.DeleteFile(ctx, req.Msg.InstanceId, req.Msg.Path)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return connect.NewResponse(&runtimev1.DeleteFileResponse{}), nil
}

// RenameFile implements RuntimeService.
func (s *Server) RenameFile(ctx context.Context, req *connect.Request[runtimev1.RenameFileRequest]) (*connect.Response[runtimev1.RenameFileResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.Msg.InstanceId),
		attribute.String("args.from_path", req.Msg.FromPath),
		attribute.String("args.to_path", req.Msg.ToPath),
	)

	s.addInstanceRequestAttributes(ctx, req.Msg.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.Msg.InstanceId, auth.EditRepo) {
		return nil, ErrForbidden
	}

	err := s.runtime.RenameFile(ctx, req.Msg.InstanceId, req.Msg.FromPath, req.Msg.ToPath)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return connect.NewResponse(&runtimev1.RenameFileResponse{}), nil
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
