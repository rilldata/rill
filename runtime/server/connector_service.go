package server

import (
	"context"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) ListBuckets(ctx context.Context, req *runtimev1.ListBucketsRequest) (*runtimev1.ListBucketsResponse, error) {
	handle, release, err := s.runtime.AcquireHandle(ctx, req.InstanceId, req.Connector)
	if err != nil {
		return nil, err
	}
	defer release()

	os, ok := handle.AsObjectStore()
	if !ok {
		return nil, fmt.Errorf("connector %q does not implement object store", req.Connector)
	}

	buckets, nextPageToken, err := os.ListBuckets(ctx, req.PageSize, req.PageToken)
	if err != nil {
		return nil, err
	}

	return &runtimev1.ListBucketsResponse{
		Buckets:       buckets,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *Server) ListObjects(ctx context.Context, req *runtimev1.ListObjectsRequest) (*runtimev1.ListObjectsResponse, error) {
	handle, release, err := s.runtime.AcquireHandle(ctx, req.InstanceId, req.Connector)
	if err != nil {
		return nil, err
	}
	defer release()
	os, ok := handle.AsObjectStore()
	if !ok {
		return nil, fmt.Errorf("connector %q does not implement object store", req.Connector)
	}
	objects, nextToken, err := os.ListObjects(ctx, req.Bucket, req.Path, req.Delimiter, req.PageSize, req.PageToken)
	if err != nil {
		return nil, err
	}
	pbObjects := make([]*runtimev1.Object, len(objects))
	for i, obj := range objects {
		pbObjects[i] = &runtimev1.Object{
			Name:       obj.Path,
			Size:       obj.Size,
			IsDir:      obj.IsDir,
			ModifiedOn: timestamppb.New(obj.UpdatedOn),
		}
	}
	return &runtimev1.ListObjectsResponse{
		Objects:       pbObjects,
		NextPageToken: nextToken,
	}, nil
}

func (s *Server) OLAPListTables(ctx context.Context, req *runtimev1.OLAPListTablesRequest) (*runtimev1.OLAPListTablesResponse, error) {
	olap, release, err := s.runtime.OLAP(ctx, req.InstanceId, req.Connector)
	if err != nil {
		return nil, err
	}
	defer release()

	i := olap.InformationSchema()
	tables, next, err := i.All(ctx, req.SearchPattern, req.PageSize, req.PageToken)
	if err != nil {
		return nil, err
	}
	_ = i.LoadPhysicalSize(ctx, tables)

	res := make([]*runtimev1.OlapTableInfo, len(tables))
	for i, table := range tables {
		res[i] = &runtimev1.OlapTableInfo{
			Database:                table.Database,
			DatabaseSchema:          table.DatabaseSchema,
			IsDefaultDatabase:       table.IsDefaultDatabase,
			IsDefaultDatabaseSchema: table.IsDefaultDatabaseSchema,
			Name:                    table.Name,
			HasUnsupportedDataTypes: len(table.UnsupportedCols) != 0,
			PhysicalSizeBytes:       table.PhysicalSizeBytes,
		}
	}
	return &runtimev1.OLAPListTablesResponse{
		Tables:        res,
		NextPageToken: next,
	}, nil
}

func (s *Server) OLAPGetTable(ctx context.Context, req *runtimev1.OLAPGetTableRequest) (*runtimev1.OLAPGetTableResponse, error) {
	olap, release, err := s.runtime.OLAP(ctx, req.InstanceId, req.Connector)
	if err != nil {
		return nil, err
	}
	defer release()

	table, err := olap.InformationSchema().Lookup(ctx, req.Database, req.DatabaseSchema, req.Table)
	if err != nil {
		return nil, err
	}
	_ = olap.InformationSchema().LoadPhysicalSize(ctx, []*drivers.OlapTable{table})

	return &runtimev1.OLAPGetTableResponse{
		Schema:             table.Schema,
		UnsupportedColumns: table.UnsupportedCols,
		View:               table.View,
		PhysicalSizeBytes:  table.PhysicalSizeBytes,
	}, nil
}

func (s *Server) ListDatabaseSchemas(ctx context.Context, req *runtimev1.ListDatabaseSchemasRequest) (*runtimev1.ListDatabaseSchemasResponse, error) {
	handle, release, err := s.runtime.AcquireHandle(ctx, req.InstanceId, req.Connector)
	if err != nil {
		return nil, err
	}
	defer release()

	is, ok := handle.AsInformationSchema()
	if !ok {
		return nil, fmt.Errorf("connector %q does not implement information schema", req.Connector)
	}

	items, next, err := is.ListDatabaseSchemas(ctx, req.PageSize, req.PageToken)
	if err != nil {
		return nil, err
	}
	res := make([]*runtimev1.DatabaseSchemaInfo, len(items))
	for i, schema := range items {
		res[i] = &runtimev1.DatabaseSchemaInfo{
			Database:       schema.Database,
			DatabaseSchema: schema.DatabaseSchema,
		}
	}
	return &runtimev1.ListDatabaseSchemasResponse{
		NextPageToken:   next,
		DatabaseSchemas: res,
	}, nil
}

func (s *Server) ListTables(ctx context.Context, req *runtimev1.ListTablesRequest) (*runtimev1.ListTablesResponse, error) {
	handle, release, err := s.runtime.AcquireHandle(ctx, req.InstanceId, req.Connector)
	if err != nil {
		return nil, err
	}
	defer release()

	is, ok := handle.AsInformationSchema()
	if !ok {
		return nil, fmt.Errorf("connector %q does not implement information schema", req.Connector)
	}

	items, next, err := is.ListTables(ctx, req.Database, req.DatabaseSchema, req.PageSize, req.PageToken)
	if err != nil {
		return nil, err
	}
	res := make([]*runtimev1.TableInfo, len(items))
	for i, table := range items {
		res[i] = &runtimev1.TableInfo{
			Name: table.Name,
			View: table.View,
		}
	}
	return &runtimev1.ListTablesResponse{
		NextPageToken: next,
		Tables:        res,
	}, nil
}

func (s *Server) GetTable(ctx context.Context, req *runtimev1.GetTableRequest) (*runtimev1.GetTableResponse, error) {
	handle, release, err := s.runtime.AcquireHandle(ctx, req.InstanceId, req.Connector)
	if err != nil {
		return nil, err
	}
	defer release()

	is, ok := handle.AsInformationSchema()
	if !ok {
		return nil, fmt.Errorf("connector %q does not implement information schema", req.Connector)
	}

	tableMetadata, err := is.GetTable(ctx, req.Database, req.DatabaseSchema, req.Table)
	if err != nil {
		return nil, err
	}

	return &runtimev1.GetTableResponse{
		Schema: tableMetadata.Schema,
	}, nil
}
