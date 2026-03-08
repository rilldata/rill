package server

import (
	"context"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/dbt_cloud"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ImportDbtMetrics implements runtimev1.RuntimeServiceServer.
// It imports dbt metrics from a dbt Cloud connector as model and metrics_view YAML files.
func (s *Server) ImportDbtMetrics(ctx context.Context, req *runtimev1.ImportDbtMetricsRequest) (*runtimev1.ImportDbtMetricsResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.connector", req.Connector),
		attribute.Int("args.metric_refs_count", len(req.MetricRefs)),
	)
	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	// Must have edit permissions on the repo
	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.EditRepo) {
		return nil, ErrForbidden
	}

	if req.Connector == "" {
		return nil, status.Error(codes.InvalidArgument, "connector is required")
	}

	// Acquire the dbt_cloud connector handle
	handle, release, err := s.runtime.AcquireHandle(ctx, req.InstanceId, req.Connector)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "connector %q not found: %v", req.Connector, err)
	}
	defer release()

	if handle.Driver() != "dbt_cloud" {
		return nil, status.Errorf(codes.InvalidArgument, "connector %q is not a dbt_cloud connector", req.Connector)
	}

	// Fetch manifest
	provider, ok := handle.(dbt_cloud.ManifestProvider)
	if !ok {
		return nil, status.Error(codes.Internal, "connector does not support manifest fetching")
	}
	manifest, err := provider.GetManifest(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch dbt manifest: %v", err)
	}

	// Get adapter type from manifest
	adapterType := manifest.Metadata.AdapterType

	// Get all available metrics
	metrics := dbt_cloud.ListMetrics(manifest)

	// List-only mode: return metric info, adapter type, and matching connectors
	if req.ListOnly {
		var available []*runtimev1.DbtMetricInfo
		for _, m := range metrics {
			available = append(available, &runtimev1.DbtMetricInfo{
				Name:        m.Name,
				Label:       m.Label,
				Description: m.Description,
				Type:        m.Type,
			})
		}

		// Find connectors matching the adapter type
		matching, err := s.findConnectorsByAdapterType(ctx, req.InstanceId, adapterType)
		if err != nil {
			s.logger.Warn("failed to find matching connectors", zap.Error(err))
		}

		return &runtimev1.ImportDbtMetricsResponse{
			AvailableMetrics:  available,
			AdapterType:       adapterType,
			MatchingConnectors: matching,
		}, nil
	}

	// Resolve the warehouse connector for import
	warehouseConnector, err := s.resolveWarehouseConnector(ctx, req, handle, adapterType)
	if err != nil {
		return nil, err
	}

	// Filter metrics if specific refs were requested
	if len(req.MetricRefs) > 0 {
		requested := make(map[string]bool, len(req.MetricRefs))
		for _, ref := range req.MetricRefs {
			requested[ref] = true
		}
		var filtered []*dbt_cloud.ManifestMetric
		for _, m := range metrics {
			if requested[m.Name] || requested[m.UniqueID] {
				filtered = append(filtered, m)
			}
		}
		if len(filtered) == 0 {
			return nil, status.Errorf(codes.NotFound, "none of the requested metrics were found in the manifest")
		}
		metrics = filtered
	}

	// Get the repo to write files
	repo, release, err := s.runtime.Repo(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get repo: %v", err)
	}
	defer release()

	// Ensure rill.yaml has metrics_compiler set to dbt_cloud
	if err := s.ensureMetricsCompiler(ctx, repo, "dbt_cloud"); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update rill.yaml: %v", err)
	}

	var generatedFiles []string
	for _, metric := range metrics {
		modelPath := fmt.Sprintf("/models/dbt_%s.yaml", metric.Name)
		mvPath := fmt.Sprintf("/metrics/dbt_%s.yaml", metric.Name)

		// Idempotency: skip if the model file already exists
		if _, err := repo.Get(ctx, modelPath); err == nil {
			continue
		}

		// Resolve the output table for display purposes
		displayName := metric.Label
		if displayName == "" {
			displayName = identifierToDisplayName(metric.Name)
		}

		// Generate model YAML
		modelYAML := generateDbtModelYAML(warehouseConnector, metric.Name)
		if err := repo.Put(ctx, modelPath, strings.NewReader(modelYAML)); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to write model file %q: %v", modelPath, err)
		}
		generatedFiles = append(generatedFiles, modelPath)

		// Generate metrics_view YAML
		mvYAML := generateDbtMetricsViewYAML(metric.Name, displayName)
		if err := repo.Put(ctx, mvPath, strings.NewReader(mvYAML)); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to write metrics view file %q: %v", mvPath, err)
		}
		generatedFiles = append(generatedFiles, mvPath)
	}

	return &runtimev1.ImportDbtMetricsResponse{
		GeneratedFiles: generatedFiles,
		AdapterType:    adapterType,
	}, nil
}

// resolveWarehouseConnector determines the warehouse connector to use for import.
// Priority: request field > connector config > auto-detect from adapter type.
func (s *Server) resolveWarehouseConnector(ctx context.Context, req *runtimev1.ImportDbtMetricsRequest, handle drivers.Handle, adapterType string) (string, error) {
	// 1. Explicit override from request
	if req.WarehouseConnector != "" {
		return req.WarehouseConnector, nil
	}

	// 2. From dbt_cloud connector config (backward compat)
	if wc, _ := handle.Config()["warehouse_connector"].(string); wc != "" {
		return wc, nil
	}

	// 3. Auto-detect from manifest adapter type
	if adapterType == "" {
		return "", status.Error(codes.FailedPrecondition, "manifest does not specify an adapter type; please provide warehouse_connector")
	}

	matching, err := s.findConnectorsByAdapterType(ctx, req.InstanceId, adapterType)
	if err != nil {
		return "", status.Errorf(codes.Internal, "failed to find matching connectors: %v", err)
	}

	switch len(matching) {
	case 0:
		return "", status.Errorf(codes.FailedPrecondition, "no Rill connector found for adapter type %q; please set up a %s connector first", adapterType, adapterType)
	case 1:
		return matching[0], nil
	default:
		return "", status.Errorf(codes.FailedPrecondition, "multiple connectors match adapter type %q: %s; please select one", adapterType, strings.Join(matching, ", "))
	}
}

// findConnectorsByAdapterType finds Rill connectors whose driver matches the dbt adapter type.
func (s *Server) findConnectorsByAdapterType(ctx context.Context, instanceID, adapterType string) ([]string, error) {
	if adapterType == "" {
		return nil, nil
	}

	driverNames := adapterTypeToDrivers(adapterType)
	if len(driverNames) == 0 {
		return nil, nil
	}

	ctrl, err := s.runtime.Controller(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	resources, err := ctrl.List(ctx, runtime.ResourceKindConnector, "", false)
	if err != nil {
		return nil, err
	}

	driverSet := make(map[string]bool, len(driverNames))
	for _, d := range driverNames {
		driverSet[d] = true
	}

	var matching []string
	for _, r := range resources {
		if r.GetConnector() == nil {
			continue
		}
		driver := r.GetConnector().GetSpec().GetDriver()
		if driverSet[driver] {
			matching = append(matching, r.Meta.Name.Name)
		}
	}

	return matching, nil
}

// adapterTypeToDrivers maps a dbt adapter type to Rill driver names.
func adapterTypeToDrivers(adapterType string) []string {
	switch strings.ToLower(adapterType) {
	case "snowflake":
		return []string{"snowflake"}
	case "bigquery":
		return []string{"bigquery"}
	case "postgres", "postgresql":
		return []string{"postgres"}
	case "redshift":
		return []string{"redshift"}
	case "mysql":
		return []string{"mysql"}
	case "duckdb":
		return []string{"duckdb", "motherduck"}
	case "athena":
		return []string{"athena"}
	default:
		return []string{adapterType}
	}
}

// generateDbtModelYAML generates a model YAML file for a dbt metric.
func generateDbtModelYAML(warehouseConnector, metricRef string) string {
	return fmt.Sprintf(`# Model for dbt metric: %s
# This file was auto-generated by Rill's dbt Cloud integration.

version: 1
type: model
connector: %s
dbt_metric_ref: %s
`, metricRef, warehouseConnector, metricRef)
}

// generateDbtMetricsViewYAML generates a metrics_view YAML file for a dbt metric.
func generateDbtMetricsViewYAML(metricName, displayName string) string {
	return fmt.Sprintf(`# Metrics view for dbt metric: %s
# This file was auto-generated by Rill's dbt Cloud integration.
# Dimensions and measures are auto-populated from the table schema by the dbt_cloud compiler.

version: 1
type: metrics_view
model: dbt_%s
compiler: dbt_cloud
display_name: "%s"
`, metricName, metricName, displayName)
}

// ensureMetricsCompiler reads rill.yaml and adds metrics_compiler if not already set.
func (s *Server) ensureMetricsCompiler(ctx context.Context, repo drivers.RepoStore, compiler string) error {
	data, err := repo.Get(ctx, "/rill.yaml")
	if err != nil {
		return fmt.Errorf("failed to read rill.yaml: %w", err)
	}

	// Check if metrics_compiler is already set
	if strings.Contains(data, "metrics_compiler:") {
		return nil
	}

	// Append metrics_compiler after olap_connector line, or at the end
	if idx := strings.Index(data, "olap_connector:"); idx != -1 {
		// Find end of the olap_connector line
		eol := strings.Index(data[idx:], "\n")
		if eol != -1 {
			insertAt := idx + eol + 1
			data = data[:insertAt] + fmt.Sprintf("metrics_compiler: %s\n", compiler) + data[insertAt:]
		} else {
			data += fmt.Sprintf("\nmetrics_compiler: %s\n", compiler)
		}
	} else {
		data += fmt.Sprintf("\nmetrics_compiler: %s\n", compiler)
	}

	return repo.Put(ctx, "/rill.yaml", strings.NewReader(data))
}
