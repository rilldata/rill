package server

import (
	"context"
	"fmt"
	"net/http"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ListFiles implements RuntimeService
func (s *Server) ListFiles(ctx context.Context, req *runtimev1.ListFilesRequest) (*runtimev1.ListFilesResponse, error) {
	registry, _ := s.metastore.RegistryStore()
	inst, found, err := registry.FindInstance(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, status.Error(codes.NotFound, "instance not found")
	}

	conn, err := drivers.Open(inst.RepoDriver, inst.RepoDSN)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	glob := req.Glob
	if glob == "" {
		glob = "**"
	}

	repoStore, _ := conn.RepoStore()
	paths, err := repoStore.ListRecursive(ctx, inst.ID, glob)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	return &runtimev1.ListFilesResponse{Paths: paths}, nil
}

// GetFile implements RuntimeService
func (s *Server) GetFile(ctx context.Context, req *runtimev1.GetFileRequest) (*runtimev1.GetFileResponse, error) {
	registry, _ := s.metastore.RegistryStore()
	inst, found, err := registry.FindInstance(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, status.Error(codes.NotFound, "instance not found")
	}

	conn, err := drivers.Open(inst.RepoDriver, inst.RepoDSN)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	repoStore, _ := conn.RepoStore()
	blob, err := repoStore.Get(ctx, inst.ID, req.Path)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// TODO: Could we return Stat as part of Get?
	stat, err := repoStore.Stat(ctx, inst.ID, req.Path)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.GetFileResponse{Blob: blob, UpdatedOn: timestamppb.New(stat.LastUpdated)}, nil
}

// PutFile implements RuntimeService
func (s *Server) PutFile(ctx context.Context, req *runtimev1.PutFileRequest) (*runtimev1.PutFileResponse, error) {
	registry, _ := s.metastore.RegistryStore()
	inst, found, err := registry.FindInstance(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, status.Error(codes.NotFound, "instance not found")
	}

	conn, err := drivers.Open(inst.RepoDriver, inst.RepoDSN)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// TODO: Handle req.Create, req.CreateOnly
	repoStore, _ := conn.RepoStore()
	err = repoStore.PutBlob(ctx, inst.ID, req.Path, req.Blob)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.PutFileResponse{}, nil
}

// DeleteFile implements RuntimeService
func (s *Server) DeleteFile(ctx context.Context, req *runtimev1.DeleteFileRequest) (*runtimev1.DeleteFileResponse, error) {
	registry, _ := s.metastore.RegistryStore()
	inst, found, err := registry.FindInstance(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, status.Error(codes.NotFound, "instance not found")
	}

	conn, err := drivers.Open(inst.RepoDriver, inst.RepoDSN)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	repoStore, _ := conn.RepoStore()
	err = repoStore.Delete(ctx, inst.ID, req.Path)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.DeleteFileResponse{}, nil
}

// RenameFile implements RuntimeService
func (s *Server) RenameFile(ctx context.Context, req *runtimev1.RenameFileRequest) (*runtimev1.RenameFileResponse, error) {
	registry, _ := s.metastore.RegistryStore()
	inst, found, err := registry.FindInstance(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, status.Error(codes.NotFound, "instance not found")
	}

	conn, err := drivers.Open(inst.RepoDriver, inst.RepoDSN)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	repoStore, _ := conn.RepoStore()
	err = repoStore.Rename(ctx, inst.ID, req.FromPath, req.ToPath)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.RenameFileResponse{}, nil
}

// UploadMultipartFile implements the same functionality as PutFile, but for multipart HTTP upload.
// It's mounted only on as a REST API and enables upload of large files (such as data files).
func (s *Server) UploadMultipartFile(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
	ctx := context.Background()
	if err := req.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse request: %s", err), http.StatusBadRequest)
		return
	}

	registry, _ := s.metastore.RegistryStore()
	inst, found, err := registry.FindInstance(ctx, pathParams["instance_id"])
	if err != nil {
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}
	if !found {
		http.Error(w, "instance not found", http.StatusBadRequest)
		return
	}

	conn, err := drivers.Open(inst.RepoDriver, inst.RepoDSN)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to open driver: %s", err), http.StatusBadRequest)
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

	repoStore, _ := conn.RepoStore()
	filePath, err := repoStore.PutReader(ctx, inst.ID, pathParams["path"], f)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to write file: %s", err), http.StatusBadRequest)
		return
	}

	res, err := protojson.Marshal(&runtimev1.PutFileResponse{
		FilePath: filePath,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to serialize response: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}
