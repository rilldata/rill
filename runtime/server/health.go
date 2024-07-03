package server

import (
	"context"
	"net"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// Healthz implements RuntimeService
func (s *Server) Healthz(ctx context.Context, req *runtimev1.HealthzRequest) (*runtimev1.HealthzResponse, error) {
	resp := &runtimev1.HealthzResponse{}

	// limiter
	if err := s.limiter.Ping(ctx); err != nil {
		resp.Limiter = err.Error()
	}

	// internet access
	if err := pingCloudfareDNS(ctx); err != nil {
		resp.Network = err.Error()
	}

	// get runtime health
	status := s.runtime.Health(ctx)
	// hung connections
	if status.HungConn {
		resp.ConnCache = "found hung connections"
	}
	// metastore
	resp.Metastore = status.Registry.Error()

	// instance specific health
	resp.Instances = make(map[string]*runtimev1.InstanceHealth)
	for k, v := range status.Instances {
		resp.Instances[k] = v.To()
	}
	return resp, nil
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
