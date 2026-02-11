package server

import (
	"context"
	"errors"
	"net"
	"net/http"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/httputil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
)

// Health implements RuntimeService
func (s *Server) Health(ctx context.Context, req *runtimev1.HealthRequest) (*runtimev1.HealthResponse, error) {
	if !auth.GetClaims(ctx, "").Can(runtime.ManageInstances) {
		return nil, ErrForbidden
	}

	resp := &runtimev1.HealthResponse{}

	// limiter
	if err := s.limiter.Ping(ctx); err != nil {
		resp.LimiterError = err.Error()
	}

	// internet access
	if err := pingCloudfareDNS(ctx); err != nil {
		resp.NetworkError = err.Error()
	}

	// runtime health
	status, err := s.runtime.Health(ctx, true)
	if err != nil {
		return nil, err
	}

	if status.HangingConn != nil {
		resp.ConnCacheError = status.HangingConn.Error()
	}
	if status.Registry != nil {
		resp.MetastoreError = status.Registry.Error()
	}
	resp.InstancesHealth = make(map[string]*runtimev1.InstanceHealth, len(status.InstancesHealth))
	for id, h := range status.InstancesHealth {
		resp.InstancesHealth[id] = h.Proto()
	}

	return resp, nil
}

// InstanceHealth implements RuntimeService
func (s *Server) InstanceHealth(ctx context.Context, req *runtimev1.InstanceHealthRequest) (*runtimev1.InstanceHealthResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
	)
	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.ReadInstance) {
		return nil, ErrForbidden
	}

	h, err := s.runtime.InstanceHealth(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	return &runtimev1.InstanceHealthResponse{
		InstanceHealth: h.Proto(),
	}, nil
}

func (s *Server) healthCheckHandler(w http.ResponseWriter, req *http.Request) error {
	ctx := req.Context()
	if err := s.limiter.Ping(ctx); err != nil {
		return httputil.Error(http.StatusInternalServerError, err)
	}

	// internet access
	if err := pingCloudfareDNS(ctx); err != nil {
		return httputil.Error(http.StatusInternalServerError, err)
	}

	// runtime health
	// we don't return 5xx on hanging connections
	status, err := s.runtime.Health(ctx, false)
	if err != nil {
		return httputil.Error(http.StatusInternalServerError, err)
	}
	if status.Registry != nil {
		return httputil.Error(http.StatusInternalServerError, status.Registry)
	}
	for _, h := range status.InstancesHealth {
		if h.Controller != "" || h.Repo != "" || h.OLAP != "" {
			return httputil.Error(http.StatusInternalServerError, errors.New("controller: "+h.Controller+", repo: "+h.Repo+", olap: "+h.OLAP))
		}
	}
	w.WriteHeader(http.StatusOK)
	return nil
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
