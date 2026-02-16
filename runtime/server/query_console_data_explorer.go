package server

import (
	"context"
	"fmt"
	"sort"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/server/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ListDataExplorerObjects implements RuntimeService.ListDataExplorerObjects.
// It enumerates data objects available in the instance's catalog, grouped by connector.
// Internal resources (sources, models, metrics views) are listed alongside external connector tables.
func (s *Server) ListDataExplorerObjects(ctx context.Context, req *runtimev1.ListDataExplorerObjectsRequest) (*runtimev1.ListDataExplorerObjectsResponse, error) {
	// Validate required fields
	if req.InstanceId == "" {
		return nil, status.Error(codes.InvalidArgument, "instance_id is required")
	}

	// Check permissions — requires read access to the instance
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadObjects) {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read objects")
	}

	// Get the runtime instance
	inst, err := s.runtime.Instance(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("instance not found: %v", err))
	}

	var nodes []*runtimev1.DataExplorerTreeNode

	// 1. Enumerate internal Rill resources (sources, models, metrics views) from the catalog
	internalNodes, err := s.listInternalObjects(ctx, req.InstanceId, inst)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list internal objects: %v", err))
	}
	nodes = append(nodes, internalNodes...)

	// 2. Enumerate external connector tables/views
	externalNodes, err := s.listExternalObjects(ctx, req.InstanceId, inst)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list external objects: %v", err))
	}
	nodes = append(nodes, externalNodes...)

	// Sort top-level nodes for deterministic output
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Name < nodes[j].Name
	})

	return &runtimev1.ListDataExplorerObjectsResponse{
		Nodes: nodes,
	}, nil
}

// listInternalObjects enumerates Rill-managed resources (sources, models, metrics views)
// from the controller's catalog and groups them under an "internal" connector node.
func (s *Server) listInternalObjects(ctx context.Context, instanceID string, inst *runtime.Instance) ([]*runtimev1.DataExplorerTreeNode, error) {
	ctrl, err := s.runtime.Controller(ctx, instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get controller: %w", err)
	}

	var sourceChildren []*runtimev1.DataExplorerTreeNode
	var modelChildren []*runtimev1.DataExplorerTreeNode
	var metricsViewChildren []*runtimev1.DataExplorerTreeNode

	// List all resources from the controller
	resources, err := ctrl.List(ctx, "", false)
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}

	for _, r := range resources {
		name := r.Meta.Name
		switch r.Meta.Name.Kind {
		case runtime.ResourceKindSource:
			sourceChildren = append(sourceChildren, &runtimev1.DataExplorerTreeNode{
				Name:       name.Name,
				ObjectType: "source",
				IsLeaf:     true,
				Connector:  inst.ResolveOLAPConnector(),
			})
		case runtime.ResourceKindModel:
			modelChildren = append(modelChildren, &runtimev1.DataExplorerTreeNode{
				Name:       name.Name,
				ObjectType: "model",
				IsLeaf:     true,
				Connector:  inst.ResolveOLAPConnector(),
			})
		case runtime.ResourceKindMetricsView:
			metricsViewChildren = append(metricsViewChildren, &runtimev1.DataExplorerTreeNode{
				Name:       name.Name,
				ObjectType: "metrics_view",
				IsLeaf:     true,
				Connector:  inst.ResolveOLAPConnector(),
			})
		}
	}

	// Sort children for deterministic output
	sortTreeNodes(sourceChildren)
	sortTreeNodes(modelChildren)
	sortTreeNodes(metricsViewChildren)

	// Build the connector-level grouping node for the default OLAP connector
	olapConnector := inst.ResolveOLAPConnector()
	var children []*runtimev1.DataExplorerTreeNode

	if len(sourceChildren) > 0 {
		children = append(children, &runtimev1.DataExplorerTreeNode{
			Name:       "Sources",
			ObjectType: "group",
			IsLeaf:     false,
			Children:   sourceChildren,
			Connector:  olapConnector,
		})
	}
	if len(modelChildren) > 0 {
		children = append(children, &runtimev1.DataExplorerTreeNode{
			Name:       "Models",
			ObjectType: "group",
			IsLeaf:     false,
			Children:   modelChildren,
			Connector:  olapConnector,
		})
	}
	if len(metricsViewChildren) > 0 {
		children = append(children, &runtimev1.DataExplorerTreeNode{
			Name:       "Metrics Views",
			ObjectType: "group",
			IsLeaf:     false,
			Children:   metricsViewChildren,
			Connector:  olapConnector,
		})
	}

	if len(children) == 0 {
		return nil, nil
	}

	return []*runtimev1.DataExplorerTreeNode{
		{
			Name:       olapConnector,
			ObjectType: "connector",
			IsLeaf:     false,
			Children:   children,
			Connector:  olapConnector,
		},
	}, nil
}

// listExternalObjects enumerates tables and views from external connectors
// that implement the OLAPStore interface, grouping them by database/schema/table.
func (s *Server) listExternalObjects(ctx context.Context, instanceID string, inst *runtime.Instance) ([]*runtimev1.DataExplorerTreeNode, error) {
	olapConnector := inst.ResolveOLAPConnector()
	var connectorNodes []*runtimev1.DataExplorerTreeNode

	// Iterate over all configured connectors on the instance
	for _, c := range inst.Connectors {
		// Skip the default OLAP connector — its objects are already listed as internal
		if c.Name == olapConnector {
			continue
		}

		// Try to acquire an OLAP handle for this connector
		olap, release, err := s.runtime.OLAP(ctx, instanceID, c.Name)
		if err != nil {
			// Connector doesn't support OLAP — skip it
			continue
		}
		defer release()

		// Get the information schema from the OLAP driver
		infoSchema, err := olap.InformationSchema(ctx)
		if err != nil {
			// Connector doesn't support information schema — skip
			continue
		}

		tables, err := infoSchema.All(ctx)
		if err != nil {
			// Log and skip connectors that fail enumeration
			s.logger.Warn("failed to list tables for connector",
				// Use connector name for context
			)
			continue
		}

		if len(tables) == 0 {
			continue
		}

		// Group tables by database and schema
		tableNodes := groupTablesBySchema(tables, c.Name)

		connectorNodes = append(connectorNodes, &runtimev1.DataExplorerTreeNode{
			Name:       c.Name,
			ObjectType: "connector",
			IsLeaf:     false,
			Children:   tableNodes,
			Connector:  c.Name,
		})
	}

	return connectorNodes, nil
}

// groupTablesBySchema groups a list of tables into a tree of database → schema → table nodes.
// If there's only one database and one schema, it flattens the hierarchy.
func groupTablesBySchema(tables []*drivers.Table, connectorName string) []*runtimev1.DataExplorerTreeNode {
	// Build a nested map: database -> schema -> tables
	type schemaGroup struct {
		tables []*drivers.Table
	}
	type dbGroup struct {
		schemas map[string]*schemaGroup
	}

	databases := make(map[string]*dbGroup)

	for _, t := range tables {
		dbName := t.Database
		if dbName == "" {
			dbName = "default"
		}
		schemaName := t.DatabaseSchema
		if schemaName == "" {
			schemaName = "default"
		}

		db, ok := databases[dbName]
		if !ok {
			db = &dbGroup{schemas: make(map[string]*schemaGroup)}
			databases[dbName] = db
		}

		sg, ok := db.schemas[schemaName]
		if !ok {
			sg = &schemaGroup{}
			db.schemas[schemaName] = sg
		}

		sg.tables = append(sg.tables, t)
	}

	// Check if we can flatten (single db, single schema)
	if len(databases) == 1 {
		for _, db := range databases {
			if len(db.schemas) == 1 {
				// Flatten: return table nodes directly
				for _, sg := range db.schemas {
					return tablesToLeafNodes(sg.tables, connectorName)
				}
			}
		}
	}

	// Build full hierarchy: database → schema → table
	var dbNodes []*runtimev1.DataExplorerTreeNode

	// Sort database names
	dbNames := make([]string, 0, len(databases))
	for dbName := range databases {
		dbNames = append(dbNames, dbName)
	}
	sort.Strings(dbNames)

	for _, dbName := range dbNames {
		db := databases[dbName]
		var schemaNodes []*runtimev1.DataExplorerTreeNode

		// Sort schema names
		schemaNames := make([]string, 0, len(db.schemas))
		for schemaName := range db.schemas {
			schemaNames = append(schemaNames, schemaName)
		}
		sort.Strings(schemaNames)

		for _, schemaName := range schemaNames {
			sg := db.schemas[schemaName]
			leaves := tablesToLeafNodes(sg.tables, connectorName)

			schemaNodes = append(schemaNodes, &runtimev1.DataExplorerTreeNode{
				Name:       schemaName,
				ObjectType: "schema",
				IsLeaf:     false,
				Children:   leaves,
				Connector:  connectorName,
			})
		}

		// If only one schema exists, skip the schema level
		if len(schemaNodes) == 1 {
			dbNodes = append(dbNodes, &runtimev1.DataExplorerTreeNode{
				Name:       dbName,
				ObjectType: "database",
				IsLeaf:     false,
				Children:   schemaNodes[0].Children,
				Connector:  connectorName,
			})
		} else {
			dbNodes = append(dbNodes, &runtimev1.DataExplorerTreeNode{
				Name:       dbName,
				ObjectType: "database",
				IsLeaf:     false,
				Children:   schemaNodes,
				Connector:  connectorName,
			})
		}
	}

	return dbNodes
}

// tablesToLeafNodes converts a slice of driver Table objects into leaf tree nodes.
func tablesToLeafNodes(tables []*drivers.Table, connectorName string) []*runtimev1.DataExplorerTreeNode {
	nodes := make([]*runtimev1.DataExplorerTreeNode, 0, len(tables))
	for _, t := range tables {
		objType := "table"
		if t.View {
			objType = "view"
		}

		// Build a fully qualified name for use in schema lookups
		qualifiedName := buildQualifiedName(t.Database, t.DatabaseSchema, t.Name)

		nodes = append(nodes, &runtimev1.DataExplorerTreeNode{
			Name:          t.Name,
			ObjectType:    objType,
			IsLeaf:        true,
			Connector:     connectorName,
			QualifiedName: qualifiedName,
		})
	}

	sortTreeNodes(nodes)
	return nodes
}

// buildQualifiedName constructs a dot-separated qualified name from database, schema, and table.
// Empty parts are omitted.
func buildQualifiedName(database, schema, table string) string {
	var parts []string
	if database != "" {
		parts = append(parts, database)
	}
	if schema != "" {
		parts = append(parts, schema)
	}
	parts = append(parts, table)
	return strings.Join(parts, ".")
}

// sortTreeNodes sorts tree nodes alphabetically by name.
func sortTreeNodes(nodes []*runtimev1.DataExplorerTreeNode) {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Name < nodes[j].Name
	})
}
