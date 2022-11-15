package server

import (
	"context"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/services/catalog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Migrate implements RuntimeService
func (s *Server) Migrate(ctx context.Context, req *api.MigrateRequest) (*api.MigrateResponse, error) {
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

	return &api.MigrateResponse{
		Errors:        resp.Errors,
		AffectedPaths: resp.AffectedPaths,
	}, nil
}

// PutFileAndMigrate implements RuntimeService
func (s *Server) PutFileAndMigrate(ctx context.Context, req *api.PutFileAndMigrateRequest) (*api.PutFileAndMigrateResponse, error) {
	_, err := s.PutFile(ctx, &api.PutFileRequest{
		RepoId:     req.RepoId,
		Path:       req.Path,
		Blob:       req.Blob,
		Create:     req.Create,
		CreateOnly: req.CreateOnly,
		Delete:     req.Delete,
	})
	if err != nil {
		return nil, err
	}
	migrateResp, err := s.Migrate(ctx, &api.MigrateRequest{
		InstanceId:   req.InstanceId,
		RepoId:       req.RepoId,
		ChangedPaths: []string{req.Path},
		Dry:          false,
		Strict:       false,
	})
	if err != nil {
		return nil, err
	}
	return &api.PutFileAndMigrateResponse{
		Errors:        migrateResp.Errors,
		AffectedPaths: migrateResp.AffectedPaths,
	}, nil
}
