package server

import (
	"context"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GenerateMetricsViewFile generates a metrics view YAML file from a table in an OLAP database
func (s *Server) GenerateMetricsViewFile(ctx context.Context, req *runtimev1.GenerateMetricsViewFileRequest) (*runtimev1.GenerateMetricsViewFileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.connector", req.Connector),
		attribute.String("args.table", req.Table),
		attribute.String("args.path", req.Path),
	)
	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	// Must have edit permissions on the repo
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.EditRepo) {
		return nil, ErrForbidden
	}

	// Get instance
	inst, err := s.runtime.Instance(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}

	// If a connector is not provided, default to the instance's OLAP connector
	if req.Connector == "" {
		req.Connector = inst.OLAPConnector
	}

	// Connect to connector and check it's an OLAP db
	handle, release, err := s.runtime.AcquireHandle(ctx, req.InstanceId, req.Connector)
	if err != nil {
		return nil, err
	}
	defer release()

	olap, ok := handle.AsOLAP(req.InstanceId)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "connector is not an OLAP connector")
	}

	// Get table info
	tbl, err := olap.InformationSchema().Lookup(ctx, req.Table)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "table not found")
	}

	// Generate the YAML
	var yaml string
	if req.UseAi {
		// Connect to the AI service configured for the instance
		ai, release, err := s.runtime.AI(ctx, req.InstanceId)
		if err != nil {
			return nil, err
		}
		defer release()

		// Call AI service to infer a metrics view YAML
		yaml, err = ai.GenerateMetricsViewYAML(ctx, tbl.Name, olap.Dialect().String(), tbl.Schema)
		if err != nil {
			s.logger.Error("failed to generate metrics view YAML using AI", zap.Error(err))
		}

		// TODO: Validate the YAML using the parser
		// TODO: Remove invalid dimensions and measures (use validation logic from the reconciler)
		// TODO: Fallback to current metrics view generator if something fails OR after a timeout
		// TODO: Add a comment in the output for whether it was generated with AI or static analysis
	}

	// TODO: If yaml == nil, fall back to the basic generator

	// Write the file to the repo
	repo, release, err := s.runtime.Repo(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()
	err = repo.Put(ctx, req.Path, strings.NewReader(yaml))
	if err != nil {
		return nil, err
	}

	return &runtimev1.GenerateMetricsViewFileResponse{}, nil
}
