package server

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var modelNameRegexp = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// overwriteAllowedV1 controls whether overwriting existing models is allowed.
// Per PRD, this is false for V1.
const overwriteAllowedV1 = false

func (s *Server) PublishModel(ctx context.Context, req *runtimev1.PublishModelRequest) (*runtimev1.PublishModelResponse, error) {
	// Validate required fields
	if req.InstanceId == "" {
		return nil, status.Error(codes.InvalidArgument, "instance_id is required")
	}
	if req.Sql == "" {
		return nil, status.Error(codes.InvalidArgument, "sql is required")
	}
	if req.ModelName == "" {
		return nil, status.Error(codes.InvalidArgument, "model_name is required")
	}

	// Validate model name format
	if err := validateModelNameFormat(req.ModelName); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid model name: %s", err.Error())
	}

	// Get the instance
	inst, err := s.runtime.Instance(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "instance not found: %s", err.Error())
	}

	// Check for duplicate names in the catalog
	if err := validateModelName(ctx, s.runtime, req.InstanceId, req.ModelName); err != nil {
		s.emitPublishTelemetry(ctx, req.InstanceId, req.ModelName, "", "model_publish_failed_duplicate_name", err.Error())
		return nil, err
	}

	// Classify the model type based on SQL analysis
	modelType, err := classifyModelFromSQL(ctx, s.runtime, req.InstanceId, req.Sql)
	if err != nil {
		// Classification failure is non-fatal; default to source_model
		modelType = runtimev1.ModelType_MODEL_TYPE_SOURCE
	}

	// Generate the model YAML content
	yamlContent := generateModelYAML(req.ModelName, req.Sql, modelType)

	// Determine the file path for the new model
	filePath := filepath.Join("models", req.ModelName+".sql")
	yamlFilePath := filepath.Join("models", req.ModelName+".yaml")

	// Write the SQL file
	if err := s.runtime.PutFile(ctx, req.InstanceId, filePath, strings.NewReader(req.Sql), false, false); err != nil {
		s.emitPublishTelemetry(ctx, req.InstanceId, req.ModelName, modelType.String(), "model_publish_failed", err.Error())
		return nil, status.Errorf(codes.Internal, "failed to write model SQL file: %s", err.Error())
	}

	// Write the YAML metadata file
	if err := s.runtime.PutFile(ctx, req.InstanceId, yamlFilePath, strings.NewReader(yamlContent), false, false); err != nil {
		// Best-effort cleanup of the SQL file
		_ = s.runtime.DeleteFile(ctx, req.InstanceId, filePath, false)
		s.emitPublishTelemetry(ctx, req.InstanceId, req.ModelName, modelType.String(), "model_publish_failed", err.Error())
		return nil, status.Errorf(codes.Internal, "failed to write model YAML file: %s", err.Error())
	}

	// Reconcile to pick up the new resource
	if err := s.runtime.Reconcile(ctx, req.InstanceId); err != nil {
		// Non-fatal: the reconciler will eventually pick it up
		s.logger.Warn("reconcile after model publish failed", logArgs(inst, err)...)
	}

	now := timestamppb.Now()

	// Emit success telemetry
	s.emitPublishTelemetry(ctx, req.InstanceId, req.ModelName, modelType.String(), "model_published", "")

	return &runtimev1.PublishModelResponse{
		ModelName: req.ModelName,
		ModelType: modelType,
		FilePath:  yamlFilePath,
		CreatedOn: now,
	}, nil
}

// validateModelNameFormat checks that the model name follows naming conventions.
func validateModelNameFormat(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("model name cannot be empty")
	}
	if len(name) > 128 {
		return fmt.Errorf("model name cannot exceed 128 characters")
	}
	if !modelNameRegexp.MatchString(name) {
		return fmt.Errorf("model name must start with a letter or underscore and contain only alphanumeric characters and underscores")
	}
	return nil
}

// validateModelName checks that no existing resource in the catalog already uses the
// given name. It checks models, sources, and metrics views. In V1, overwriting is not
// allowed, so any collision results in an error.
func validateModelName(ctx context.Context, rt *runtime.Runtime, instanceID, name string) error {
	// Normalize name to lower-case for case-insensitive comparison, since
	// Rill resource names are case-insensitive.
	normalized := strings.ToLower(name)

	// List all resources in the instance catalog and check for name collisions.
	resources, err := rt.ListResources(ctx, instanceID)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to list catalog resources: %s", err.Error())
	}

	for _, r := range resources {
		if r.Meta == nil || r.Meta.Name == nil {
			continue
		}

		rKind := r.Meta.Name.Kind
		rName := strings.ToLower(r.Meta.Name.Name)

		// Check against resource kinds that would conflict with a new model.
		switch rKind {
		case runtime.ResourceKindModel,
			runtime.ResourceKindSource,
			runtime.ResourceKindMetricsView,
			runtime.ResourceKindExplore:
			if rName == normalized {
				if !overwriteAllowedV1 {
					return status.Errorf(
						codes.AlreadyExists,
						"Model already exists. Choose a different name.",
					)
				}
			}
		}
	}

	// Additionally, check if a file already exists at the target path.
	// This catches cases where a file exists but hasn't been reconciled yet.
	for _, ext := range []string{".sql", ".yaml", ".yml"} {
		filePath := filepath.Join("models", name+ext)
		exists, _ := rt.FileExists(ctx, instanceID, filePath)
		if exists {
			if !overwriteAllowedV1 {
				return status.Errorf(
					codes.AlreadyExists,
					"Model already exists. Choose a different name.",
				)
			}
		}
	}

	return nil
}

// classifyModelFromSQL determines whether the model is a source model or a derived model
// by analyzing the SQL and checking referenced objects against the catalog.
func classifyModelFromSQL(ctx context.Context, rt *runtime.Runtime, instanceID, sql string) (runtimev1.ModelType, error) {
	// Use the classify helper from runtime package
	refs := extractTableReferences(sql)
	if len(refs) == 0 {
		// No table references (e.g., SELECT 1) — treat as derived
		return runtimev1.ModelType_MODEL_TYPE_DERIVED, nil
	}

	// Load catalog resources for lookup
	resources, err := rt.ListResources(ctx, instanceID)
	if err != nil {
		return runtimev1.ModelType_MODEL_TYPE_SOURCE, fmt.Errorf("failed to list resources: %w", err)
	}

	// Build a set of known internal model/source names
	internalNames := make(map[string]bool)
	for _, r := range resources {
		if r.Meta == nil || r.Meta.Name == nil {
			continue
		}
		switch r.Meta.Name.Kind {
		case runtime.ResourceKindModel, runtime.ResourceKindSource:
			internalNames[strings.ToLower(r.Meta.Name.Name)] = true
		}
	}

	// Check each referenced table
	allInternal := true
	for _, ref := range refs {
		normalized := strings.ToLower(ref)
		// Strip schema prefix if present (e.g., "main.table" → "table")
		parts := strings.Split(normalized, ".")
		tableName := parts[len(parts)-1]

		if !internalNames[tableName] {
			allInternal = false
			break
		}
	}

	if allInternal {
		return runtimev1.ModelType_MODEL_TYPE_DERIVED, nil
	}
	return runtimev1.ModelType_MODEL_TYPE_SOURCE, nil
}

// extractTableReferences performs a lightweight extraction of table names from SQL.
// It looks for FROM and JOIN clauses. This is a best-effort parser and does not handle
// all SQL dialects perfectly.
func extractTableReferences(sql string) []string {
	// Normalize whitespace
	normalized := strings.Join(strings.Fields(sql), " ")

	var refs []string
	seen := make(map[string]bool)

	// Pattern: FROM <table> or JOIN <table>
	// We use a simple word-boundary approach
	tokens := strings.Fields(normalized)
	for i, token := range tokens {
		upper := strings.ToUpper(token)
		if (upper == "FROM" || upper == "JOIN" ||
			upper == "INNER" || upper == "LEFT" ||
			upper == "RIGHT" || upper == "FULL" ||
			upper == "CROSS" || upper == "OUTER") && i+1 < len(tokens) {

			// For multi-word join keywords, skip to the actual table name
			if upper == "INNER" || upper == "LEFT" || upper == "RIGHT" ||
				upper == "FULL" || upper == "CROSS" || upper == "OUTER" {
				// These are join modifiers; the actual table comes after "JOIN"
				continue
			}

			nextToken := tokens[i+1]
			// Skip subqueries
			if strings.HasPrefix(nextToken, "(") {
				continue
			}
			// Clean up: remove trailing commas, parens, semicolons
			clean := strings.TrimRight(nextToken, ",;)")
			clean = strings.TrimLeft(clean, "(")
			// Remove quotes if present
			clean = strings.Trim(clean, "`\"'")

			if clean != "" && !seen[strings.ToLower(clean)] {
				seen[strings.ToLower(clean)] = true
				refs = append(refs, clean)
			}
		}
	}

	return refs
}

// generateModelYAML creates the YAML metadata content for a published model.
func generateModelYAML(name, sql string, modelType runtimev1.ModelType) string {
	typeStr := "source"
	if modelType == runtimev1.ModelType_MODEL_TYPE_DERIVED {
		typeStr = "derived"
	}

	var b strings.Builder
	b.WriteString("# Auto-generated by Rill Query Console\n")
	b.WriteString(fmt.Sprintf("# Published: %s\n", time.Now().UTC().Format(time.RFC3339)))
	b.WriteString(fmt.Sprintf("# Type: %s\n", typeStr))
	b.WriteString("\n")
	b.WriteString("type: model\n")
	b.WriteString(fmt.Sprintf("name: %s\n", name))
	b.WriteString(fmt.Sprintf("model_type: %s\n", typeStr))

	return b.String()
}

// emitPublishTelemetry emits a telemetry event for model publishing operations.
func (s *Server) emitPublishTelemetry(ctx context.Context, instanceID, modelName, modelType, eventType, errMsg string) {
	if s.activity == nil {
		return
	}

	dimensions := map[string]string{
		"instance_id": instanceID,
		"model_name":  modelName,
		"model_type":  modelType,
		"event_type":  eventType,
	}
	if errMsg != "" {
		dimensions["error"] = errMsg
	}

	s.activity.Record(ctx, activity.Event{
		Action:     eventType,
		Dimensions: dimensions,
	})
}

// logArgs is a helper that formats instance and error info for structured logging.
// This matches the pattern used elsewhere in the runtime server.
func logArgs(inst interface{}, err error) []interface{} {
	return []interface{}{"instance", inst, "error", err}
}
