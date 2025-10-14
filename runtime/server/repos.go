package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/r3labs/sse/v2"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
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

	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.ReadRepo) {
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

// WatchFilesHandler is a HTTP handler for watching local file changes.
// This is required as vanguard doesn't currently map streaming RPCs to SSE, so we register this handler manually override the behavior
func (s *Server) WatchFilesHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	instanceID := req.PathValue("instance_id")

	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", instanceID),
	)

	if !auth.GetClaims(ctx, instanceID).Can(runtime.ReadRepo) {
		http.Error(w, "action not allowed", http.StatusUnauthorized)
		return
	}

	replayStr := req.URL.Query().Get("replay")
	replay := replayStr == "true"

	eventServer := sse.New()
	eventServer.CreateStream("files")
	eventServer.Headers = map[string]string{
		"Content-Type":  "text/event-stream",
		"Cache-Control": "no-cache",
		"Connection":    "keep-alive",
	}

	// Create the shim that implements RuntimeService_WatchFilesServer
	shim := &watchFilesServerShim{
		r: req,
		s: eventServer,
	}

	// Create a goroutine to handle the streaming
	go func() {
		// Create the request object for WatchFiles
		watchReq := &runtimev1.WatchFilesRequest{
			InstanceId: instanceID,
			Replay:     replay,
		}

		// Call the existing WatchFiles implementation with our shim
		err := s.WatchFiles(watchReq, shim)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				s.logger.Warn("watch files error", zap.String("instance_id", instanceID), zap.Error(err))
			}

			errJSON, err := json.Marshal(map[string]string{"error": err.Error()})
			if err != nil {
				s.logger.Error("failed to marshal error as json", zap.Error(err))
			}

			eventServer.Publish("files", &sse.Event{
				Data:  errJSON,
				Event: []byte("error"),
			})
		}
		eventServer.Close()
	}()

	// Serve the SSE stream
	eventServer.ServeHTTP(w, req)
}

// WatchFiles implements RuntimeService.
func (s *Server) WatchFiles(req *runtimev1.WatchFilesRequest, ss runtimev1.RuntimeService_WatchFilesServer) error {
	observability.AddRequestAttributes(ss.Context(),
		attribute.String("args.instance_id", req.InstanceId),
		attribute.Bool("args.replay", req.Replay),
	)

	if !auth.GetClaims(ss.Context(), req.InstanceId).Can(runtime.ReadRepo) {
		return ErrForbidden
	}

	repo, release, err := s.runtime.Repo(ss.Context(), req.InstanceId)
	if err != nil {
		return err
	}
	defer release()

	if req.Replay {
		files, err := repo.ListGlob(ss.Context(), "**", false)
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

	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.ReadRepo) {
		return nil, ErrForbidden
	}

	blob, lastUpdated, err := s.runtime.GetFile(ctx, req.InstanceId, req.Path)
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

	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.EditRepo) {
		return nil, ErrForbidden
	}

	err := s.runtime.PutFile(ctx, req.InstanceId, req.Path, strings.NewReader(req.Blob), req.Create, req.CreateOnly)
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

	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.EditRepo) {
		return nil, ErrForbidden
	}

	err := s.runtime.MkdirAll(ctx, req.InstanceId, req.Path)
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

	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.EditRepo) {
		return nil, ErrForbidden
	}

	err := s.runtime.DeleteFile(ctx, req.InstanceId, req.Path, req.Force)
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

	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.EditRepo) {
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
func (s *Server) UploadMultipartFile(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	instanceID := req.PathValue("instance_id")
	path := req.PathValue("path")

	if !auth.GetClaims(req.Context(), instanceID).Can(runtime.EditRepo) {
		http.Error(w, "action not allowed", http.StatusUnauthorized)
		return
	}

	if err := req.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse request: %s", err), http.StatusBadRequest)
		return
	}

	f, _, err := req.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse file in request: %s", err), http.StatusBadRequest)
		return
	}
	defer f.Close()

	if path == "" {
		http.Error(w, "must have a path to file", http.StatusBadRequest)
		return
	}

	observability.AddRequestAttributes(ctx, attribute.String("args.instance_id", instanceID), attribute.String("args.path", path))

	s.addInstanceRequestAttributes(ctx, instanceID)

	err = s.runtime.PutFile(ctx, instanceID, path, f, true, false)
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

// A shim for runtimev1.RuntimeService_WatchFilesServer
type watchFilesServerShim struct {
	r *http.Request
	s *sse.Server
}

// Context returns the context of the request.
func (ss *watchFilesServerShim) Context() context.Context {
	return ss.r.Context()
}

// SendHeader sends a header to the client.
func (ss *watchFilesServerShim) Send(e *runtimev1.WatchFilesResponse) error {
	data, err := protojson.Marshal(e)
	if err != nil {
		return err
	}

	ss.s.Publish("files", &sse.Event{Data: data})
	return nil
}

// SetHeader sets the header for the response.
func (ss *watchFilesServerShim) SetHeader(metadata.MD) error {
	return errors.New("not implemented")
}

// SendHeader sends a header to the client.
func (ss *watchFilesServerShim) SendHeader(metadata.MD) error {
	return errors.New("not implemented")
}

// SetTrailer sets the trailer for the response.
func (ss *watchFilesServerShim) SetTrailer(metadata.MD) {}

func (ss *watchFilesServerShim) SendMsg(m any) error {
	return errors.New("not implemented")
}

// RecvMsg receives a message from the client.
func (ss *watchFilesServerShim) RecvMsg(m any) error {
	return errors.New("not implemented")
}
