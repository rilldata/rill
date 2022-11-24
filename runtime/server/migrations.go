package server

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/services/catalog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Migrate implements RuntimeService
func (s *Server) Migrate(ctx context.Context, req *runtimev1.MigrateRequest) (*runtimev1.MigrateResponse, error) {
	service, err := s.serviceCache.createCatalogService(ctx, s, req.InstanceId, req.RepoId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp, err := service.Migrate(ctx, catalog.MigrationConfig{
		DryRun:       req.Dry,
		Strict:       req.Strict,
		ChangedPaths: req.ChangedPaths,
	})
	if err != nil {
		return nil, err
	}

	return &runtimev1.MigrateResponse{
		Errors:        resp.Errors,
		AffectedPaths: resp.AffectedPaths,
	}, nil
}

// PutFileAndMigrate implements RuntimeService
func (s *Server) PutFileAndMigrate(ctx context.Context, req *runtimev1.PutFileAndMigrateRequest) (*runtimev1.PutFileAndMigrateResponse, error) {
	_, err := s.PutFile(ctx, &runtimev1.PutFileRequest{
		RepoId:     req.RepoId,
		Path:       req.Path,
		Blob:       req.Blob,
		Create:     req.Create,
		CreateOnly: req.CreateOnly,
	})
	if err != nil {
		return nil, err
	}
	migrateResp, err := s.Migrate(ctx, &runtimev1.MigrateRequest{
		InstanceId:   req.InstanceId,
		RepoId:       req.RepoId,
		ChangedPaths: []string{req.Path},
		Dry:          false,
		Strict:       false,
	})
	if err != nil {
		return nil, err
	}
	return &runtimev1.PutFileAndMigrateResponse{
		Errors:        migrateResp.Errors,
		AffectedPaths: migrateResp.AffectedPaths,
	}, nil
}

func (s *Server) RenameFileAndMigrate(ctx context.Context, req *runtimev1.RenameFileAndMigrateRequest) (*runtimev1.RenameFileAndMigrateResponse, error) {
	_, err := s.RenameFile(ctx, &runtimev1.RenameFileRequest{
		RepoId:   req.RepoId,
		FromPath: req.FromPath,
		ToPath:   req.ToPath,
	})
	if err != nil {
		return nil, err
	}
	migrateResp, err := s.Migrate(ctx, &runtimev1.MigrateRequest{
		InstanceId:   req.InstanceId,
		RepoId:       req.RepoId,
		ChangedPaths: []string{req.FromPath, req.ToPath},
		Dry:          false,
		Strict:       false,
	})
	if err != nil {
		return nil, err
	}
	return &runtimev1.RenameFileAndMigrateResponse{
		Errors:        migrateResp.Errors,
		AffectedPaths: migrateResp.AffectedPaths,
	}, nil
}

func (s *Server) DeleteFileAndMigrate(ctx context.Context, req *runtimev1.DeleteFileAndMigrateRequest) (*runtimev1.DeleteFileAndMigrateResponse, error) {
	_, err := s.DeleteFile(ctx, &runtimev1.DeleteFileRequest{
		RepoId: req.RepoId,
		Path:   req.Path,
	})
	if err != nil {
		return nil, err
	}
	migrateResp, err := s.Migrate(ctx, &runtimev1.MigrateRequest{
		InstanceId:   req.InstanceId,
		RepoId:       req.RepoId,
		ChangedPaths: []string{req.Path},
		Dry:          false,
		Strict:       false,
	})
	if err != nil {
		return nil, err
	}
	return &runtimev1.DeleteFileAndMigrateResponse{
		Errors:        migrateResp.Errors,
		AffectedPaths: migrateResp.AffectedPaths,
	}, nil
}
