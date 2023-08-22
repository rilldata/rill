package server

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/bigquery"
	"github.com/rilldata/rill/runtime/drivers/gcs"
	"github.com/rilldata/rill/runtime/drivers/s3"
)

// ListConnectors implements RuntimeService.
func (s *Server) ListConnectors(ctx context.Context, req *connect.Request[runtimev1.ListConnectorsRequest]) (*connect.Response[runtimev1.ListConnectorsResponse], error) {
	var pbs []*runtimev1.ConnectorSpec
	for name, connector := range drivers.Connectors {
		// Build protobufs for properties
		srcProps := connector.Spec().SourceProperties
		propPBs := make([]*runtimev1.ConnectorSpec_Property, len(srcProps))
		for j := range connector.Spec().SourceProperties {
			propSchema := srcProps[j]
			// Get type
			var t runtimev1.ConnectorSpec_Property_Type
			switch propSchema.Type {
			case drivers.StringPropertyType:
				t = runtimev1.ConnectorSpec_Property_TYPE_STRING
			case drivers.NumberPropertyType:
				t = runtimev1.ConnectorSpec_Property_TYPE_NUMBER
			case drivers.BooleanPropertyType:
				t = runtimev1.ConnectorSpec_Property_TYPE_BOOLEAN
			case drivers.InformationalPropertyType:
				t = runtimev1.ConnectorSpec_Property_TYPE_INFORMATIONAL
			default:
				panic(fmt.Errorf("property type '%v' not handled", propSchema.Type))
			}

			// Add protobuf for property
			propPBs[j] = &runtimev1.ConnectorSpec_Property{
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
		pbs = append(pbs, &runtimev1.ConnectorSpec{
			Name:        name,
			DisplayName: connector.Spec().DisplayName,
			Description: connector.Spec().Description,
			Properties:  propPBs,
		})
	}

	return connect.NewResponse(&runtimev1.ListConnectorsResponse{Connectors: pbs}), nil
}

func (s *Server) S3ListBuckets(ctx context.Context, req *connect.Request[runtimev1.S3ListBucketsRequest]) (*connect.Response[runtimev1.S3ListBucketsResponse], error) {
	s3Conn, release, err := s.getS3Conn(ctx, req.Msg.Connector, req.Msg.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	buckets, err := s3Conn.ListBuckets(ctx)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&runtimev1.S3ListBucketsResponse{
		Buckets: buckets,
	}), nil
}

func (s *Server) S3ListObjects(ctx context.Context, req *connect.Request[runtimev1.S3ListObjectsRequest]) (*connect.Response[runtimev1.S3ListObjectsResponse], error) {
	s3Conn, release, err := s.getS3Conn(ctx, req.Msg.Connector, req.Msg.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	objects, nextToken, err := s3Conn.ListObjects(ctx, req)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&runtimev1.S3ListObjectsResponse{
		Objects:       objects,
		NextPageToken: nextToken,
	}), nil
}

func (s *Server) S3GetBucketMetadata(ctx context.Context, req *connect.Request[runtimev1.S3GetBucketMetadataRequest]) (*connect.Response[runtimev1.S3GetBucketMetadataResponse], error) {
	s3Conn, release, err := s.getS3Conn(ctx, req.Msg.Connector, req.Msg.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	region, err := s3Conn.GetBucketMetadata(ctx, req)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&runtimev1.S3GetBucketMetadataResponse{
		Region: region,
	}), nil
}

func (s *Server) S3GetCredentialsInfo(ctx context.Context, req *connect.Request[runtimev1.S3GetCredentialsInfoRequest]) (*connect.Response[runtimev1.S3GetCredentialsInfoResponse], error) {
	s3Conn, release, err := s.getS3Conn(ctx, req.Msg.Connector, req.Msg.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	provider, exist, err := s3Conn.GetCredentialsInfo(ctx)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&runtimev1.S3GetCredentialsInfoResponse{
		Exist:    exist,
		Provider: provider,
	}), nil
}

func (s *Server) GCSListBuckets(ctx context.Context, req *connect.Request[runtimev1.GCSListBucketsRequest]) (*connect.Response[runtimev1.GCSListBucketsResponse], error) {
	gcsConn, release, err := s.getGCSConn(ctx, req.Msg.Connector, req.Msg.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	buckets, next, err := gcsConn.ListBuckets(ctx, req)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&runtimev1.GCSListBucketsResponse{
		Buckets:       buckets,
		NextPageToken: next,
	}), nil
}

func (s *Server) GCSListObjects(ctx context.Context, req *connect.Request[runtimev1.GCSListObjectsRequest]) (*connect.Response[runtimev1.GCSListObjectsResponse], error) {
	gcsConn, release, err := s.getGCSConn(ctx, req.Msg.Connector, req.Msg.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	objects, nextToken, err := gcsConn.ListObjects(ctx, req)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&runtimev1.GCSListObjectsResponse{
		Objects:       objects,
		NextPageToken: nextToken,
	}), nil
}

func (s *Server) GCSGetCredentialsInfo(ctx context.Context, req *connect.Request[runtimev1.GCSGetCredentialsInfoRequest]) (*connect.Response[runtimev1.GCSGetCredentialsInfoResponse], error) {
	gcsConn, release, err := s.getGCSConn(ctx, req.Msg.Connector, req.Msg.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	projectID, exist, err := gcsConn.GetCredentialsInfo(ctx)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&runtimev1.GCSGetCredentialsInfoResponse{
		ProjectId: projectID,
		Exist:     exist,
	}), nil
}

func (s *Server) MotherduckListTables(ctx context.Context, req *connect.Request[runtimev1.OLAPListTablesRequest]) (*connect.Response[runtimev1.OLAPListTablesResponse], error) {
	conn, release, err := s.runtime.AcquireHandle(ctx, req.Msg.InstanceId, req.Msg.Connector)
	if err != nil {
		return nil, err
	}
	defer release()

	olap, _ := conn.AsOLAP(req.Msg.InstanceId)
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
	return connect.NewResponse(&runtimev1.OLAPListTablesResponse{
		Tables: res,
	}), nil
}

func (s *Server) BigQueryListDatasets(ctx context.Context, req *runtimev1.BigQueryListDatasetsRequest) (*runtimev1.BigQueryListDatasetsResponse, error) {
	bq, release, err := s.getBigQueryConn(ctx, req.Connector, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	names, nextToken, err := bq.ListDatasets(ctx, req)
	if err != nil {
		return nil, err
	}

	return &runtimev1.BigQueryListDatasetsResponse{
		Names:         names,
		NextPageToken: nextToken,
	}, nil
}

func (s *Server) BigQueryListTables(ctx context.Context, req *runtimev1.BigQueryListTablesRequest) (*runtimev1.BigQueryListTablesResponse, error) {
	bq, release, err := s.getBigQueryConn(ctx, req.Connector, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	names, nextToken, err := bq.ListTables(ctx, req)
	if err != nil {
		return nil, err
	}

	return &runtimev1.BigQueryListTablesResponse{
		Names:         names,
		NextPageToken: nextToken,
	}, nil
}

func (s *Server) getGCSConn(ctx context.Context, connector, instanceID string) (*gcs.Connection, func(), error) {
	conn, release, err := s.runtime.AcquireHandle(ctx, instanceID, connector)
	if err != nil {
		return nil, nil, fmt.Errorf("can't open connection to gcs %w", err)
	}

	gcsConn, ok := conn.(*gcs.Connection)
	if !ok {
		panic("conn is not gcs connection")
	}
	return gcsConn, release, nil
}

func (s *Server) getS3Conn(ctx context.Context, connector, instanceID string) (*s3.Connection, func(), error) {
	conn, release, err := s.runtime.AcquireHandle(ctx, instanceID, connector)
	if err != nil {
		return nil, nil, fmt.Errorf("can't open connection to s3 %w", err)
	}

	s3Conn, ok := conn.(*s3.Connection)
	if !ok {
		panic("conn is not s3 connection")
	}
	return s3Conn, release, nil
}

func (s *Server) getBigQueryConn(ctx context.Context, connector, instanceID string) (*bigquery.Connection, func(), error) {
	conn, release, err := s.runtime.AcquireHandle(ctx, instanceID, connector)
	if err != nil {
		return nil, nil, err
	}

	bq, ok := conn.(*bigquery.Connection)
	if !ok {
		panic("conn is not bigquery connection")
	}
	return bq, release, nil
}
