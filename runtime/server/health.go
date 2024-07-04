package server

import (
	"context"
	"net"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// Health implements RuntimeService
func (s *Server) Health(ctx context.Context, req *runtimev1.HealthRequest) (*runtimev1.HealthResponse, error) {
	resp := &runtimev1.HealthResponse{}

	// limiter
	if err := s.limiter.Ping(ctx); err != nil {
		resp.LimiterError = err.Error()
	}

	// internet access
	if err := pingCloudfareDNS(ctx); err != nil {
		resp.NetworkError = err.Error()
	}

	// get runtime health
	status := s.runtime.Health(ctx)
	resp.ConnCacheError = status.HungConn.Error()
	resp.MetastoreError = status.Registry.Error()
	return resp, nil
}

// InstanceHealth implements RuntimeService
func (s *Server) InstanceHealth(ctx context.Context, req *runtimev1.InstanceHealthRequest) (*runtimev1.InstanceHealthResponse, error) {
	h, err := s.runtime.InstanceHealth(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	return &runtimev1.InstanceHealthResponse{
		ControllerError: h.Controller.Error(),
		RepoError:       h.Repo.Error(),
		OlapError:       h.OLAP.Error(),
	}, nil
}

func pingCloudfareDNS(ctx context.Context) error {
	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", "1.1.1.1:53")
	if err != nil {
		return err
	}
	_ = conn.Close()
	return nil
}
