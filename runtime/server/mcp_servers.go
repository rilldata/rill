package server

import (
	"context"
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ListMCPServers implements RuntimeService.
// It lists the external MCP servers the current user may use, with the tools each exposes to them.
func (s *Server) ListMCPServers(ctx context.Context, req *runtimev1.ListMCPServersRequest) (*runtimev1.ListMCPServersResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.instance_id", req.InstanceId))

	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.UseAI) {
		return nil, ErrForbidden
	}

	servers, err := s.ai.DiscoverExternalServers(ctx, req.InstanceId, claims)
	if err != nil {
		return nil, err
	}

	res := &runtimev1.ListMCPServersResponse{Servers: make([]*runtimev1.MCPServer, 0, len(servers))}
	for _, srv := range servers {
		pb := &runtimev1.MCPServer{
			Connector:   srv.Connector,
			DisplayName: srv.DisplayName,
			Oauth:       srv.OAuth,
			Authorized:  srv.Authorized,
			Connected:   srv.Connected,
			Agents:      srv.Agents,
		}
		for _, t := range srv.Tools {
			pb.Tools = append(pb.Tools, &runtimev1.MCPTool{Name: t.Name, Description: t.Description})
		}
		if srv.Err != nil {
			pb.Error = srv.Err.Error()
		}
		res.Servers = append(res.Servers, pb)
	}
	return res, nil
}

// DisconnectMCPServer implements RuntimeService.
// It removes the current user's OAuth authorization for an external MCP server:
// from the admin service's token vault in Rill Cloud, or from the connector's local state in Rill Developer.
func (s *Server) DisconnectMCPServer(ctx context.Context, req *runtimev1.DisconnectMCPServerRequest) (*runtimev1.DisconnectMCPServerResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.connector", req.Connector),
	)

	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.UseAI) {
		return nil, ErrForbidden
	}

	handle, release, err := s.runtime.AcquireHandle(ctx, req.InstanceId, req.Connector)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("connector %q not found: %s", req.Connector, err.Error()))
	}
	defer release()
	svc, ok := handle.AsMCP()
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "connector %q is not an MCP server", req.Connector)
	}
	cfg := svc.OAuthConfig()
	if cfg == nil {
		return nil, status.Errorf(codes.InvalidArgument, "connector %q does not use per-user OAuth", req.Connector)
	}

	// Rill Cloud: the admin service holds the user's token for the connector's current OAuth configuration.
	if admin, releaseAdmin, err := s.runtime.Admin(ctx, req.InstanceId); err == nil {
		err = admin.DeleteConnectorOAuthToken(ctx, req.Connector, cfg.Hash, claims.UserID)
		releaseAdmin()
		if err != nil && !errors.Is(err, drivers.ErrNotImplemented) {
			return nil, err
		}
		if err == nil {
			return &runtimev1.DisconnectMCPServerResponse{}, nil
		}
	}

	// Rill Developer: the driver holds the state.
	err = svc.ResetOAuth(ctx)
	if err != nil {
		return nil, err
	}
	return &runtimev1.DisconnectMCPServerResponse{}, nil
}
