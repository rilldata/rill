package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ListRepos implements RuntimeService
func (s *Server) ListRepos(ctx context.Context, req *api.ListReposRequest) (*api.ListReposResponse, error) {
	registry, _ := s.metastore.RegistryStore()
	repos := registry.FindRepos(ctx)

	pbs := make([]*api.Repo, len(repos))
	for i, repo := range repos {
		pbs[i] = repoToPB(repo)
	}

	return &api.ListReposResponse{Repos: pbs}, nil
}

// GetRepo implements RuntimeService
func (s *Server) GetRepo(ctx context.Context, req *api.GetRepoRequest) (*api.GetRepoResponse, error) {
	registry, _ := s.metastore.RegistryStore()
	repo, found := registry.FindRepo(ctx, req.RepoId)
	if !found {
		return nil, status.Error(codes.NotFound, "repo not found")
	}

	return &api.GetRepoResponse{
		Repo: repoToPB(repo),
	}, nil
}

// CreateRepo implements RuntimeService
func (s *Server) CreateRepo(ctx context.Context, req *api.CreateRepoRequest) (*api.CreateRepoResponse, error) {
	repo := &drivers.Repo{
		Driver: req.Driver,
		DSN:    req.Dsn,
	}

	// Check that it's a valid repo
	conn, err := drivers.Open(repo.Driver, repo.DSN)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	_, ok := conn.RepoStore()
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "not a valid repo driver")
	}

	registry, _ := s.metastore.RegistryStore()
	err = registry.CreateRepo(ctx, repo)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &api.CreateRepoResponse{
		Repo: repoToPB(repo),
	}, nil
}

// DeleteRepo implements RuntimeService
func (s *Server) DeleteRepo(ctx context.Context, req *api.DeleteRepoRequest) (*api.DeleteRepoResponse, error) {
	registry, _ := s.metastore.RegistryStore()
	err := registry.DeleteRepo(ctx, req.RepoId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &api.DeleteRepoResponse{}, nil
}

// ListFiles implements RuntimeService
func (s *Server) ListFiles(ctx context.Context, req *api.ListFilesRequest) (*api.ListFilesResponse, error) {
	registry, _ := s.metastore.RegistryStore()
	repo, found := registry.FindRepo(ctx, req.RepoId)
	if !found {
		return nil, status.Error(codes.NotFound, "repo not found")
	}

	conn, err := drivers.Open(repo.Driver, repo.DSN)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	repoStore, _ := conn.RepoStore()
	paths, err := repoStore.ListRecursive(ctx, repo.ID) // TODO: use req.Glob
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	return &api.ListFilesResponse{Paths: paths}, nil
}

// GetFile implements RuntimeService
func (s *Server) GetFile(ctx context.Context, req *api.GetFileRequest) (*api.GetFileResponse, error) {
	registry, _ := s.metastore.RegistryStore()
	repo, found := registry.FindRepo(ctx, req.RepoId)
	if !found {
		return nil, status.Error(codes.NotFound, "repo not found")
	}

	conn, err := drivers.Open(repo.Driver, repo.DSN)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	repoStore, _ := conn.RepoStore()
	blob, err := repoStore.Get(ctx, repo.ID, req.Path)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// TODO: Could we return Stat as part of Get?
	stat, err := repoStore.Stat(ctx, repo.ID, req.Path)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &api.GetFileResponse{Blob: blob, UpdatedOn: timestamppb.New(stat.LastUpdated)}, nil
}

// PutFile implements RuntimeService
func (s *Server) PutFile(ctx context.Context, req *api.PutFileRequest) (*api.PutFileResponse, error) {
	registry, _ := s.metastore.RegistryStore()
	repo, found := registry.FindRepo(ctx, req.RepoId)
	if !found {
		return nil, status.Error(codes.NotFound, "repo not found")
	}

	conn, err := drivers.Open(repo.Driver, repo.DSN)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// TODO: Handle req.Create, req.CreateOnly, req.Delete
	repoStore, _ := conn.RepoStore()
	err = repoStore.PutBlob(ctx, repo.ID, req.Path, req.Blob)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &api.PutFileResponse{}, nil
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
	repo, found := registry.FindRepo(ctx, pathParams["repo_id"])
	if !found {
		http.Error(w, "repo not found", http.StatusBadRequest)
		return
	}

	conn, err := drivers.Open(repo.Driver, repo.DSN)
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
	filePath, err := repoStore.PutReader(ctx, repo.ID, pathParams["path"], f)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to write file: %s", err), http.StatusBadRequest)
		return
	}

	res, err := protojson.Marshal(&api.PutFileResponse{
		FilePath: filePath,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to serialize response: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}

func repoToPB(repo *drivers.Repo) *api.Repo {
	return &api.Repo{
		RepoId: repo.ID,
		Driver: repo.Driver,
		Dsn:    repo.DSN,
	}
}
