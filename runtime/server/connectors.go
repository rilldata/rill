package server

import (
	"context"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/gcs"
	"github.com/rilldata/rill/runtime/drivers/s3"
)

// ListConnectors implements RuntimeService.
func (s *Server) ListConnectors(ctx context.Context, req *runtimev1.ListConnectorsRequest) (*runtimev1.ListConnectorsResponse, error) {
	var pbs []*runtimev1.Connector
	for name, connector := range drivers.Connectors {
		// Build protobufs for properties
		srcProps := connector.Spec().SourceProperties
		propPBs := make([]*runtimev1.Connector_Property, len(srcProps))
		for j := range connector.Spec().SourceProperties {
			propSchema := srcProps[j]
			// Get type
			var t runtimev1.Connector_Property_Type
			switch propSchema.Type {
			case drivers.StringPropertyType:
				t = runtimev1.Connector_Property_TYPE_STRING
			case drivers.NumberPropertyType:
				t = runtimev1.Connector_Property_TYPE_NUMBER
			case drivers.BooleanPropertyType:
				t = runtimev1.Connector_Property_TYPE_BOOLEAN
			case drivers.InformationalPropertyType:
				t = runtimev1.Connector_Property_TYPE_INFORMATIONAL
			default:
				panic(fmt.Errorf("property type '%v' not handled", propSchema.Type))
			}

			// Add protobuf for property
			propPBs[j] = &runtimev1.Connector_Property{
				Key:         propSchema.Key,
				DisplayName: propSchema.DisplayName,
				Description: propSchema.Description,
				Placeholder: propSchema.Placeholder,
				Type:        t,
				Nullable:    !propSchema.Required,
				Hint:        propSchema.Hint,
				Href:        propSchema.Href,
			}
		}

		// Add connector
		pbs = append(pbs, &runtimev1.Connector{
			Name:        name,
			DisplayName: connector.Spec().DisplayName,
			Description: connector.Spec().Description,
			Properties:  propPBs,
		})
	}

	return &runtimev1.ListConnectorsResponse{Connectors: pbs}, nil
}

func (s *Server) S3ListBuckets(ctx context.Context, req *runtimev1.S3ListBucketsRequest) (*runtimev1.S3ListBucketsResponse, error) {
	conn, err := drivers.Open("s3", map[string]any{"allow_host_access": s.runtime.AllowHostAccess()}, s.logger)
	if err != nil {
		return nil, fmt.Errorf("can't open connection to s3 %w", err)
	}
	defer conn.Close()

	s3Conn, ok := conn.(*s3.Connection)
	if !ok {
		panic("s3 connector is not an object store")
	}

	buckets, err := s3Conn.ListBuckets(ctx)
	if err != nil {
		return nil, err
	}

	return &runtimev1.S3ListBucketsResponse{
		Buckets: buckets,
	}, nil
}

func (s *Server) S3ListObjects(ctx context.Context, req *runtimev1.S3ListObjectsRequest) (*runtimev1.S3ListObjectsResponse, error) {
	conn, err := drivers.Open("s3", map[string]any{"allow_host_access": s.runtime.AllowHostAccess()}, s.logger)
	if err != nil {
		return nil, fmt.Errorf("can't open connection to s3 %w", err)
	}
	defer conn.Close()

	s3Conn, ok := conn.(*s3.Connection)
	if !ok {
		panic("s3 connector is not an object store")
	}

	objects, nextToken, err := s3Conn.ListObjects(ctx, req)
	if err != nil {
		return nil, err
	}

	return &runtimev1.S3ListObjectsResponse{
		Objects:       objects,
		NextPageToken: nextToken,
	}, nil
}

func (s *Server) S3GetBucketMetadata(ctx context.Context, req *runtimev1.S3GetBucketMetadataRequest) (*runtimev1.S3GetBucketMetadataResponse, error) {
	conn, err := drivers.Open("s3", map[string]any{"allow_host_access": s.runtime.AllowHostAccess()}, s.logger)
	if err != nil {
		return nil, fmt.Errorf("can't open connection to s3 %w", err)
	}
	defer conn.Close()

	s3Conn, ok := conn.(*s3.Connection)
	if !ok {
		panic("s3 connector is not an object store")
	}

	region, err := s3Conn.GetBucketMetadata(ctx, req)
	if err != nil {
		return nil, err
	}

	return &runtimev1.S3GetBucketMetadataResponse{
		Region: region,
	}, nil
}

func (s *Server) S3GetCredentialsInfo(ctx context.Context, req *runtimev1.S3GetCredentialsInfoRequest) (*runtimev1.S3GetCredentialsInfoResponse, error) {
	conn, err := drivers.Open("s3", map[string]any{"allow_host_access": s.runtime.AllowHostAccess()}, s.logger)
	if err != nil {
		return nil, fmt.Errorf("can't open connection to s3 %w", err)
	}
	defer conn.Close()

	s3Conn, ok := conn.(*s3.Connection)
	if !ok {
		panic("s3 connector is not an object store")
	}

	provider, exist, err := s3Conn.GetCredentialsInfo(ctx)
	if err != nil {
		return nil, err
	}

	return &runtimev1.S3GetCredentialsInfoResponse{
		Exist:    exist,
		Provider: provider,
	}, nil
}

func (s *Server) GCSListBuckets(ctx context.Context, req *runtimev1.GCSListBucketsRequest) (*runtimev1.GCSListBucketsResponse, error) {
	conn, err := drivers.Open("gcs", map[string]any{"allow_host_access": s.runtime.AllowHostAccess()}, s.logger)
	if err != nil {
		return nil, fmt.Errorf("can't open connection to s3 %w", err)
	}
	defer conn.Close()

	gcsConn, ok := conn.(*gcs.Connection)
	if !ok {
		panic("gcs connector not found")
	}

	buckets, next, err := gcsConn.ListBuckets(ctx, req)
	if err != nil {
		return nil, err
	}

	return &runtimev1.GCSListBucketsResponse{
		Buckets:       buckets,
		NextPageToken: next,
	}, nil
}

func (s *Server) GCSListObjects(ctx context.Context, req *runtimev1.GCSListObjectsRequest) (*runtimev1.GCSListObjectsResponse, error) {
	conn, err := drivers.Open("gcs", map[string]any{"allow_host_access": s.runtime.AllowHostAccess()}, s.logger)
	if err != nil {
		return nil, fmt.Errorf("can't open connection to s3 %w", err)
	}
	defer conn.Close()

	gcsConn, ok := conn.(*gcs.Connection)
	if !ok {
		panic("gcs connector not found")
	}

	objects, nextToken, err := gcsConn.ListObjects(ctx, req)
	if err != nil {
		return nil, err
	}

	return &runtimev1.GCSListObjectsResponse{
		Objects:       objects,
		NextPageToken: nextToken,
	}, nil
}

func (s *Server) GCSGetCredentialsInfo(ctx context.Context, req *runtimev1.GCSGetCredentialsInfoRequest) (*runtimev1.GCSGetCredentialsInfoResponse, error) {
	conn, err := drivers.Open("gcs", map[string]any{"allow_host_access": s.runtime.AllowHostAccess()}, s.logger)
	if err != nil {
		return nil, fmt.Errorf("can't open connection to s3 %w", err)
	}
	defer conn.Close()

	gcsConn, ok := conn.(*gcs.Connection)
	if !ok {
		panic("gcs connector not found")
	}

	projectID, exist, err := gcsConn.GetCredentialsInfo(ctx)
	if err != nil {
		return nil, err
	}

	return &runtimev1.GCSGetCredentialsInfoResponse{
		ProjectId: projectID,
		Exist:     exist,
	}, nil
}

func (s *Server) MotherduckListTables(ctx context.Context, req *runtimev1.MotherduckListTablesRequest) (*runtimev1.MotherduckListTablesResponse, error) {
	conn, err := drivers.Open("duckdb", map[string]any{"dsn": "md:"}, s.logger)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	olap, _ := conn.AsOLAP()
	tables, err := olap.InformationSchema().All(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]*runtimev1.TableInfo, len(tables))
	for i, table := range tables {
		res[i] = &runtimev1.TableInfo{
			Database: table.Database,
			Name:     table.Name,
		}
	}
	return &runtimev1.MotherduckListTablesResponse{
		Tables: res,
	}, nil
}
