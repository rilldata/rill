package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ListFiles implements RuntimeService
func (s *Server) ListFiles(ctx context.Context, req *runtimev1.ListFilesRequest) (*runtimev1.ListFilesResponse, error) {
	glob := req.Glob
	if glob == "" {
		glob = "**"
	}

	paths, err := s.runtime.ListFiles(ctx, req.InstanceId, glob)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	return &runtimev1.ListFilesResponse{Paths: paths}, nil
}

// GetFile implements RuntimeService
func (s *Server) GetFile(ctx context.Context, req *runtimev1.GetFileRequest) (*runtimev1.GetFileResponse, error) {
	blob, lastUpdated, err := s.runtime.GetFile(ctx, req.InstanceId, req.Path)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.GetFileResponse{Blob: blob, UpdatedOn: timestamppb.New(lastUpdated)}, nil
}

// PutFile implements RuntimeService
func (s *Server) PutFile(ctx context.Context, req *runtimev1.PutFileRequest) (*runtimev1.PutFileResponse, error) {
	err := s.runtime.PutFile(ctx, req.InstanceId, req.Path, strings.NewReader(req.Blob), req.Create, req.CreateOnly)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.PutFileResponse{}, nil
}

// DeleteFile implements RuntimeService
func (s *Server) DeleteFile(ctx context.Context, req *runtimev1.DeleteFileRequest) (*runtimev1.DeleteFileResponse, error) {
	err := s.runtime.DeleteFile(ctx, req.InstanceId, req.Path)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.DeleteFileResponse{}, nil
}

// RenameFile implements RuntimeService
func (s *Server) RenameFile(ctx context.Context, req *runtimev1.RenameFileRequest) (*runtimev1.RenameFileResponse, error) {
	err := s.runtime.RenameFile(ctx, req.InstanceId, req.FromPath, req.ToPath)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.RenameFileResponse{}, nil
}

func (s *Server) ExportTable(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
	var exportString string
	switch pathParams["format"] {
	case "csv":
		exportString = "FORMAT CSV, HEADER"
	case "parquet":
		exportString = "FORMAT PARQUET"
	default:
		http.Error(w, fmt.Sprintf("unknown format: %s", pathParams), http.StatusBadRequest)
	}

	if pathParams["instance_id"] == "" || pathParams["table_name"] == "" {
		http.Error(w, "missing params", http.StatusBadRequest)
	}

	fileName := fmt.Sprintf("%s.%s", pathParams["table_name"], pathParams["format"])
	filePath := path.Join(os.TempDir(), fileName)

	// select * from the table and write to the temp file
	// using duckdb for this. TODO: druid
	_, err := s.query(req.Context(), pathParams["instance_id"], &drivers.Statement{
		Query:    fmt.Sprintf("COPY (SELECT * FROM %s) TO '%s' (%s)", pathParams["table_name"], filePath, exportString),
		DryRun:   false,
		Priority: 0,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	defer os.Remove(filePath)

	// set the header to trigger download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	w.Header().Set("Content-Type", req.Header.Get("Content-Type"))

	// read and stream the file
	file, err := os.Open(filePath)
	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

// UploadMultipartFile implements the same functionality as PutFile, but for multipart HTTP upload.
// It's mounted only on as a REST API and enables upload of large files (such as data files).
func (s *Server) UploadMultipartFile(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
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
	w.Write(res)
}
