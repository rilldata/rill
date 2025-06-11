package server

import (
	"context"
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ListInstances implements RuntimeService.
func (s *Server) ListInstances(ctx context.Context, req *runtimev1.ListInstancesRequest) (*runtimev1.ListInstancesResponse, error) {
	if !auth.GetClaims(ctx).Can(auth.ManageInstances) {
		return nil, ErrForbidden
	}

	instances, err := s.runtime.Instances(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	pbs := make([]*runtimev1.Instance, len(instances))
	for i, inst := range instances {
		pbs[i] = instanceToPB(inst, true)
	}

	return &runtimev1.ListInstancesResponse{Instances: pbs}, nil
}

// GetInstance implements RuntimeService.
func (s *Server) GetInstance(ctx context.Context, req *runtimev1.GetInstanceRequest) (*runtimev1.GetInstanceResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	sensitiveAccess := auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadInstance)
	if !sensitiveAccess {
		// Regular project viewers can access non-sensitive instance information.
		// NOTE: ReadObjects is not the right permission to use, but it's the closest permission that regular project viewers have.
		// TODO: We should split ReadInstance into an admin-level and viewer-level permission instead.
		if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadObjects) {
			return nil, ErrForbidden
		}
	}

	if req.Sensitive && !sensitiveAccess {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to request sensitive instance information")
	}

	inst, err := s.runtime.Instance(ctx, req.InstanceId)
	if err != nil {
		if errors.Is(err, drivers.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "instance not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.GetInstanceResponse{
		Instance: instanceToPB(inst, req.Sensitive),
	}, nil
}

// CreateInstance implements RuntimeService.
func (s *Server) CreateInstance(ctx context.Context, req *runtimev1.CreateInstanceRequest) (*runtimev1.CreateInstanceResponse, error) {
	s.addInstanceRequestAttributes(ctx, req.InstanceId)
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.olap_connector", req.OlapConnector),
		attribute.String("args.repo_connector", req.RepoConnector),
		attribute.String("args.admin_connector", req.AdminConnector),
		attribute.String("args.ai_connector", req.AiConnector),
		attribute.StringSlice("args.connectors", connectorsStrings(req.Connectors)),
	)

	if !auth.GetClaims(ctx).Can(auth.ManageInstances) {
		return nil, ErrForbidden
	}

	inst := &drivers.Instance{
		ID:             req.InstanceId,
		Environment:    req.Environment,
		OLAPConnector:  req.OlapConnector,
		RepoConnector:  req.RepoConnector,
		AdminConnector: req.AdminConnector,
		AIConnector:    req.AiConnector,
		Connectors:     req.Connectors,
		Variables:      req.Variables,
		Annotations:    req.Annotations,
		WatchRepo:      req.WatchRepo,
	}

	err := s.runtime.CreateInstance(ctx, inst)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.CreateInstanceResponse{
		Instance: instanceToPB(inst, true),
	}, nil
}

// EditInstance implements RuntimeService.
func (s *Server) EditInstance(ctx context.Context, req *runtimev1.EditInstanceRequest) (*runtimev1.EditInstanceResponse, error) {
	s.addInstanceRequestAttributes(ctx, req.InstanceId)
	observability.AddRequestAttributes(ctx, attribute.String("args.instance_id", req.InstanceId))
	if req.OlapConnector != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.olap_connector", *req.OlapConnector))
	}
	if req.RepoConnector != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.repo_connector", *req.RepoConnector))
	}
	if req.AdminConnector != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.admin_connector", *req.AdminConnector))
	}
	if req.AiConnector != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.ai_connector", *req.AiConnector))
	}
	if len(req.Connectors) > 0 {
		observability.AddRequestAttributes(ctx, attribute.StringSlice("args.connectors", connectorsStrings(req.Connectors)))
	}

	if !auth.GetClaims(ctx).Can(auth.ManageInstances) {
		return nil, ErrForbidden
	}

	oldInst, err := s.runtime.Instance(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}

	connectors := req.Connectors
	if len(connectors) == 0 { // connectors not changed
		connectors = oldInst.Connectors
	}

	variables := req.Variables
	if len(variables) == 0 { // variables not changed
		variables = oldInst.Variables
	}

	annotations := req.Annotations
	if len(annotations) == 0 { // annotations not changed
		annotations = oldInst.Annotations
	}

	inst := &drivers.Instance{
		ID:                   req.InstanceId,
		Environment:          valOrDefault(req.Environment, oldInst.Environment),
		OLAPConnector:        valOrDefault(req.OlapConnector, oldInst.OLAPConnector),
		ProjectOLAPConnector: oldInst.ProjectOLAPConnector,
		RepoConnector:        valOrDefault(req.RepoConnector, oldInst.RepoConnector),
		AdminConnector:       valOrDefault(req.AdminConnector, oldInst.AdminConnector),
		AIConnector:          valOrDefault(req.AiConnector, oldInst.AIConnector),
		Connectors:           connectors,
		ProjectConnectors:    oldInst.ProjectConnectors,
		Variables:            variables,
		ProjectVariables:     oldInst.ProjectVariables,
		FeatureFlags:         oldInst.FeatureFlags,
		Annotations:          annotations,
		WatchRepo:            valOrDefault(req.WatchRepo, oldInst.WatchRepo),
		AIInstructions:       oldInst.AIInstructions,
	}

	err = s.runtime.EditInstance(ctx, inst, true)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.EditInstanceResponse{
		Instance: instanceToPB(inst, true),
	}, nil
}

// DeleteInstance implements RuntimeService.
func (s *Server) DeleteInstance(ctx context.Context, req *runtimev1.DeleteInstanceRequest) (*runtimev1.DeleteInstanceResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx).Can(auth.ManageInstances) {
		return nil, ErrForbidden
	}

	err := s.runtime.DeleteInstance(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.DeleteInstanceResponse{}, nil
}

// GetLogs implements runtimev1.RuntimeServiceServer
func (s *Server) GetLogs(ctx context.Context, req *runtimev1.GetLogsRequest) (*runtimev1.GetLogsResponse, error) {
	s.addInstanceRequestAttributes(ctx, req.InstanceId)
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.Bool("args.ascending", req.Ascending),
		attribute.Int("args.limit", int(req.Limit)),
		attribute.String("args.level", req.Level.String()),
	)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadObjects) {
		return nil, ErrForbidden
	}

	lvl := req.Level
	if lvl == runtimev1.LogLevel_LOG_LEVEL_UNSPECIFIED {
		lvl = runtimev1.LogLevel_LOG_LEVEL_INFO // backward compatibility
	}

	logBuffer, err := s.runtime.InstanceLogs(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &runtimev1.GetLogsResponse{Logs: logBuffer.GetLogs(req.Ascending, int(req.Limit), lvl)}, nil
}

// WatchLogs implements runtimev1.RuntimeServiceServer
func (s *Server) WatchLogs(req *runtimev1.WatchLogsRequest, srv runtimev1.RuntimeService_WatchLogsServer) error {
	ctx := srv.Context()
	s.addInstanceRequestAttributes(ctx, req.InstanceId)
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.Bool("args.replay", req.Replay),
		attribute.Int("args.replay_limit", int(req.ReplayLimit)),
		attribute.String("args.level", req.Level.String()),
	)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadObjects) {
		return ErrForbidden
	}

	lvl := req.Level
	if lvl == runtimev1.LogLevel_LOG_LEVEL_UNSPECIFIED {
		lvl = runtimev1.LogLevel_LOG_LEVEL_INFO // backward compatibility
	}

	logBuffer, err := s.runtime.InstanceLogs(ctx, req.InstanceId)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	if req.Replay {
		for _, l := range logBuffer.GetLogs(true, int(req.ReplayLimit), lvl) {
			err := srv.Send(&runtimev1.WatchLogsResponse{Log: l})
			if err != nil {
				return status.Error(codes.InvalidArgument, err.Error())
			}
		}
	}

	return logBuffer.WatchLogs(srv.Context(), func(item *runtimev1.Log) {
		err := srv.Send(&runtimev1.WatchLogsResponse{Log: item})
		if err != nil {
			s.logger.Info("failed to send log event", zap.Error(err), observability.ZapCtx(ctx))
		}
	}, lvl)
}

func instanceToPB(inst *drivers.Instance, sensitive bool) *runtimev1.Instance {
	pb := &runtimev1.Instance{
		InstanceId:     inst.ID,
		CreatedOn:      timestamppb.New(inst.CreatedOn),
		UpdatedOn:      timestamppb.New(inst.UpdatedOn),
		FeatureFlags:   inst.FeatureFlags,
		AiInstructions: inst.AIInstructions,
	}

	if sensitive {
		olapConnector := inst.OLAPConnector
		if inst.ProjectOLAPConnector != "" {
			olapConnector = inst.ProjectOLAPConnector
		}

		pb.OlapConnector = olapConnector
		pb.RepoConnector = inst.RepoConnector
		pb.AdminConnector = inst.AdminConnector
		pb.AiConnector = inst.AIConnector
		pb.Connectors = inst.Connectors
		pb.ProjectConnectors = inst.ProjectConnectors
		pb.Variables = inst.Variables
		pb.ProjectVariables = inst.ProjectVariables
		pb.Annotations = inst.Annotations
		pb.WatchRepo = inst.WatchRepo
	}

	return pb
}

func valOrDefault[T any](ptr *T, def T) T {
	if ptr != nil {
		return *ptr
	}
	return def
}

func connectorsStrings(connectors []*runtimev1.Connector) []string {
	res := make([]string, len(connectors))
	for i, c := range connectors {
		res[i] = fmt.Sprintf("%s:%s", c.Name, c.Type)
	}
	return res
}
