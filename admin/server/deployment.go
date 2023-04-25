package server

import (
	"context"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
)

func (s *Server) TriggerReconcile(ctx context.Context, req *adminv1.TriggerReconcileRequest) (*adminv1.TriggerReconcileResponse, error) {
	proj, err := s.GetProject(ctx, &adminv1.GetProjectRequest{OrganizationName: req.OrganizationName, Name: req.Name})
	if err != nil {
		return nil, err
	}
	err = s.admin.TriggerReconcile(ctx, proj.ProdDeployment.Id)
	if err != nil {
		return nil, err
	}

	return &adminv1.TriggerReconcileResponse{}, nil
}

func (s *Server) TriggerRefreshSource(ctx context.Context, req *adminv1.TriggerRefreshSourceRequest) (*adminv1.TriggerRefreshSourceResponse, error) {
	err := s.admin.TriggerRefreshSource(ctx, req.OrganizationName, req.Name, req.SourceName)
	if err != nil {
		return nil, err
	}

	return &adminv1.TriggerRefreshSourceResponse{}, nil
}

func (s *Server) TriggerRedeploy(ctx context.Context, req *adminv1.TriggerRedeployRequest) (*adminv1.TriggerRedeployResponse, error) {
	err := s.admin.TriggerRedeploy(ctx, req.OrganizationName, req.Name)
	if err != nil {
		return nil, err
	}

	return &adminv1.TriggerRedeployResponse{}, nil
}
