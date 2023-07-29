package server

import (
	"context"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/bigquery"
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
	s3Conn, err := s.getS3Conn(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer s3Conn.Close()

	buckets, err := s3Conn.ListBuckets(ctx)
	if err != nil {
		return nil, err
	}

	return &runtimev1.S3ListBucketsResponse{
		Buckets: buckets,
	}, nil
}

func (s *Server) S3ListObjects(ctx context.Context, req *runtimev1.S3ListObjectsRequest) (*runtimev1.S3ListObjectsResponse, error) {
	s3Conn, err := s.getS3Conn(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer s3Conn.Close()

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
	s3Conn, err := s.getS3Conn(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer s3Conn.Close()

	region, err := s3Conn.GetBucketMetadata(ctx, req)
	if err != nil {
		return nil, err
	}

	return &runtimev1.S3GetBucketMetadataResponse{
		Region: region,
	}, nil
}

func (s *Server) S3GetCredentialsInfo(ctx context.Context, req *runtimev1.S3GetCredentialsInfoRequest) (*runtimev1.S3GetCredentialsInfoResponse, error) {
	s3Conn, err := s.getS3Conn(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer s3Conn.Close()

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
	gcsConn, err := s.getGCSConn(ctx, req.InstanceId)
	if err != nil {
		return nil, err
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
	gcsConn, err := s.getGCSConn(ctx, req.InstanceId)
	if err != nil {
		return nil, err
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
	gcsConn, err := s.getGCSConn(ctx, req.InstanceId)
	if err != nil {
		return nil, err
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

func (s *Server) OLAPListTables(ctx context.Context, req *runtimev1.OLAPListTablesRequest) (*runtimev1.OLAPListTablesResponse, error) {
	instance, err := s.runtime.FindInstance(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}

	env := convertLower(instance.ResolveVariables())
	vars := connectorVariables(req.Connector, env)
	conn, err := drivers.Open(req.Connector, vars, s.logger)
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
	return &runtimev1.OLAPListTablesResponse{
		Tables: res,
	}, nil
}

func (s *Server) BigQueryListDatasets(ctx context.Context, req *runtimev1.BigQueryListDatasetsRequest) (*runtimev1.BigQueryListDatasetsResponse, error) {
	bq, err := s.getBigQueryConn(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}

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
	bq, err := s.getBigQueryConn(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}

	names, nextToken, err := bq.ListTables(ctx, req)
	if err != nil {
		return nil, err
	}

	return &runtimev1.BigQueryListTablesResponse{
		Names:         names,
		NextPageToken: nextToken,
	}, nil
}

func (s *Server) getGCSConn(ctx context.Context, instanceID string) (*gcs.Connection, error) {
	instance, err := s.runtime.FindInstance(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	env := convertLower(instance.ResolveVariables())
	vars := connectorVariables("gcs", env)
	conn, err := drivers.Open("gcs", vars, s.logger)
	if err != nil {
		return nil, fmt.Errorf("can't open connection to gcs %w", err)
	}

	gcsConn, ok := conn.(*gcs.Connection)
	if !ok {
		panic("conn is not gcs connection")
	}
	return gcsConn, nil
}

func (s *Server) getS3Conn(ctx context.Context, instanceID string) (*s3.Connection, error) {
	instance, err := s.runtime.FindInstance(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	env := convertLower(instance.ResolveVariables())
	vars := connectorVariables("s3", env)
	conn, err := drivers.Open("s3", vars, s.logger)
	if err != nil {
		return nil, fmt.Errorf("can't open connection to s3 %w", err)
	}

	s3Conn, ok := conn.(*s3.Connection)
	if !ok {
		panic("conn is not s3 connection")
	}
	return s3Conn, nil
}

func (s *Server) getBigQueryConn(ctx context.Context, instanceID string) (*bigquery.Connection, error) {
	instance, err := s.runtime.FindInstance(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	env := convertLower(instance.ResolveVariables())
	vars := connectorVariables("bigquery", env)
	conn, err := drivers.Open("bigquery", vars, s.logger)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	bq, ok := conn.(*bigquery.Connection)
	if !ok {
		panic("conn is not bigquery connection")
	}
	return bq, nil
}

func convertLower(in map[string]string) map[string]string {
	m := make(map[string]string, len(in))
	for key, value := range in {
		m[strings.ToLower(key)] = value
	}
	return m
}

func connectorVariables(connector string, env map[string]string) map[string]any {
	vars := map[string]any{
		"allow_host_access": strings.EqualFold(env["allow_host_access"], "true"),
	}
	switch connector {
	case "s3":
		vars["aws_access_key_id"] = env["aws_access_key_id"]
		vars["aws_secret_access_key"] = env["aws_secret_access_key"]
		vars["aws_session_token"] = env["aws_session_token"]
	case "gcs":
		vars["google_application_credentials"] = env["google_application_credentials"]
	case "motherduck":
		vars["token"] = env["token"]
		vars["dsn"] = ""
	case "bigquery":
		vars["google_application_credentials"] = env["google_application_credentials"]
	}
	return vars
}
