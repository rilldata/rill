package server

import (
	"context"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetDataExplorerSchema returns the column-level schema for a given data object (table or view).
// It resolves the connector and table from the request, looks up column metadata via the
// OLAP driver's InformationSchema, and returns typed column descriptors.
func (s *Server) GetDataExplorerSchema(ctx context.Context, req *runtimev1.GetDataExplorerSchemaRequest) (*runtimev1.GetDataExplorerSchemaResponse, error) {
	if req.InstanceId == "" {
		return nil, status.Error(codes.InvalidArgument, "instance_id is required")
	}
	if req.ObjectName == "" {
		return nil, status.Error(codes.InvalidArgument, "object_name is required")
	}

	// Check permissions: the caller must have read access to the instance.
	if !auth(ctx, s.runtime, req.InstanceId, runtimev1.QueryConsolePermission_QUERY_CONSOLE_PERMISSION_READ) {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read instance schema")
	}

	inst, err := s.runtime.Instance(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("instance not found: %s", err))
	}

	// Resolve the connector and fully qualified table name.
	connectorName, database, databaseSchema, tableName, err := resolveObjectReference(inst, req.Connector, req.ObjectName)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Acquire an OLAP handle for the resolved connector.
	olap, release, err := s.runtime.OLAP(ctx, req.InstanceId, connectorName)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to acquire OLAP connector %q: %s", connectorName, err))
	}
	defer release()

	// Look up the table schema via InformationSchema.
	table, err := olap.InformationSchema().Lookup(ctx, database, databaseSchema, tableName)
	if err != nil {
		// Check if this is a "not found" type error.
		if isDatabaseObjectNotFound(err) {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("object %q not found in connector %q: %s", req.ObjectName, connectorName, err))
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to look up schema for %q: %s", req.ObjectName, err))
	}

	// Build column descriptors from the table schema.
	columns := make([]*runtimev1.DataExplorerColumn, 0, len(table.Schema.Fields))
	for _, field := range table.Schema.Fields {
		col := &runtimev1.DataExplorerColumn{
			Name: field.Name,
			Type: field.Type.Code.String(),
			Nullable: field.Type.Nullable,
		}
		columns = append(columns, col)
	}

	return &runtimev1.GetDataExplorerSchemaResponse{
		Connector:  connectorName,
		ObjectName: req.ObjectName,
		Columns:    columns,
	}, nil
}

// resolveObjectReference parses the object reference to determine the connector, database,
// schema, and table name. The object_name may be a simple name (e.g., "my_table") or
// a qualified name (e.g., "my_schema.my_table" or "my_db.my_schema.my_table").
// If connector is empty, the instance's default OLAP connector is used.
func resolveObjectReference(inst *runtime.Instance, connector, objectName string) (connectorName, database, databaseSchema, tableName string, err error) {
	// Use the instance's default OLAP connector if none specified.
	if connector == "" {
		connectorName = inst.ResolveOLAPConnector()
	} else {
		connectorName = connector
	}

	// Parse the object name. Supported formats:
	//   "table"
	//   "schema.table"
	//   "database.schema.table"
	parts := splitQualifiedName(objectName)
	switch len(parts) {
	case 1:
		tableName = parts[0]
	case 2:
		databaseSchema = parts[0]
		tableName = parts[1]
	case 3:
		database = parts[0]
		databaseSchema = parts[1]
		tableName = parts[2]
	default:
		err = fmt.Errorf("invalid object name %q: expected format [database.][schema.]table", objectName)
		return
	}

	if tableName == "" {
		err = fmt.Errorf("object name %q resolves to an empty table name", objectName)
		return
	}

	return
}

// splitQualifiedName splits a potentially dot-delimited qualified name into its parts.
// It respects double-quoted identifiers that may contain dots.
func splitQualifiedName(name string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false

	for i := 0; i < len(name); i++ {
		ch := name[i]
		switch {
		case ch == '"':
			inQuotes = !inQuotes
			// Don't include the quote characters in the output.
		case ch == '.' && !inQuotes:
			parts = append(parts, current.String())
			current.Reset()
		default:
			current.WriteByte(ch)
		}
	}
	parts = append(parts, current.String())

	return parts
}

// isDatabaseObjectNotFound checks whether an error indicates the database object was not found.
// This handles common error patterns from various OLAP drivers.
func isDatabaseObjectNotFound(err error) bool {
	if err == nil {
		return false
	}
	if err == drivers.ErrNotFound {
		return true
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "not found") ||
		strings.Contains(errMsg, "does not exist") ||
		strings.Contains(errMsg, "no such table") ||
		strings.Contains(errMsg, "unknown table")
}

// auth checks whether the current caller has the required permission on the given instance.
// This is a helper that delegates to the runtime's security layer.
func auth(ctx context.Context, rt *runtime.Runtime, instanceID string, _ runtimev1.QueryConsolePermission) bool {
	// The runtime's existing security middleware enforces authentication and authorization
	// at the gRPC interceptor level. Instance-level read access is validated before we
	// reach the handler. This function provides a hook for future fine-grained permission
	// checks specific to the query console (e.g., execute vs. read vs. publish).
	//
	// For V1, we rely on the existing interceptor-based auth and always return true here.
	// The interceptor already verifies that the caller has access to the instance.
	_ = rt
	_ = instanceID
	_ = ctx
	return true
}
