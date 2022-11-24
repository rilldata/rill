package server

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/services/catalog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Reconcile implements RuntimeService
func (s *Server) Reconcile(ctx context.Context, req *runtimev1.ReconcileRequest) (*runtimev1.ReconcileResponse, error) {
	service, err := s.serviceCache.createCatalogService(ctx, s, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp, err := service.Reconcile(ctx, catalog.ReconcileConfig{
		DryRun:       req.Dry,
		Strict:       req.Strict,
		ChangedPaths: req.ChangedPaths,
	})
	if err != nil {
		return nil, err
	}

	return &runtimev1.ReconcileResponse{
		Errors:        resp.Errors,
		AffectedPaths: resp.AffectedPaths,
	}, nil
}

// PutFileAndReconcile implements RuntimeService
func (s *Server) PutFileAndReconcile(ctx context.Context, req *runtimev1.PutFileAndReconcileRequest) (*runtimev1.PutFileAndReconcileResponse, error) {
	_, err := s.PutFile(ctx, &runtimev1.PutFileRequest{
		InstanceId: req.InstanceId,
		Path:       req.Path,
		Blob:       req.Blob,
		Create:     req.Create,
		CreateOnly: req.CreateOnly,
	})
	if err != nil {
		return nil, err
	}
	res, err := s.Reconcile(ctx, &runtimev1.ReconcileRequest{
		InstanceId:   req.InstanceId,
		ChangedPaths: []string{req.Path},
		Dry:          false,
		Strict:       false,
	})
	if err != nil {
		return nil, err
	}
	return &runtimev1.PutFileAndReconcileResponse{
		Errors:        res.Errors,
		AffectedPaths: res.AffectedPaths,
	}, nil
}

func (s *Server) RenameFileAndReconcile(ctx context.Context, req *runtimev1.RenameFileAndReconcileRequest) (*runtimev1.RenameFileAndReconcileResponse, error) {
	_, err := s.RenameFile(ctx, &runtimev1.RenameFileRequest{
		InstanceId: req.InstanceId,
		FromPath:   req.FromPath,
		ToPath:     req.ToPath,
	})
	if err != nil {
		return nil, err
	}
	res, err := s.Reconcile(ctx, &runtimev1.ReconcileRequest{
		InstanceId:   req.InstanceId,
		ChangedPaths: []string{req.FromPath, req.ToPath},
		Dry:          false,
		Strict:       false,
	})
	if err != nil {
		return nil, err
	}
	return &runtimev1.RenameFileAndReconcileResponse{
		Errors:        res.Errors,
		AffectedPaths: res.AffectedPaths,
	}, nil
}

func (s *Server) DeleteFileAndReconcile(ctx context.Context, req *runtimev1.DeleteFileAndReconcileRequest) (*runtimev1.DeleteFileAndReconcileResponse, error) {
	_, err := s.DeleteFile(ctx, &runtimev1.DeleteFileRequest{
		InstanceId: req.InstanceId,
		Path:       req.Path,
	})
	if err != nil {
		return nil, err
	}
	res, err := s.Reconcile(ctx, &runtimev1.ReconcileRequest{
		InstanceId:   req.InstanceId,
		ChangedPaths: []string{req.Path},
		Dry:          false,
		Strict:       false,
	})
	if err != nil {
		return nil, err
	}
	return &runtimev1.DeleteFileAndReconcileResponse{
		Errors:        res.Errors,
		AffectedPaths: res.AffectedPaths,
	}, nil
}
