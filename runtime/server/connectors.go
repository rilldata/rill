package server

import (
	"context"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/connectors/gcs"
	"github.com/rilldata/rill/runtime/connectors/s3"
)

// ListConnectors implements RuntimeService.
func (s *Server) ListConnectors(ctx context.Context, req *runtimev1.ListConnectorsRequest) (*runtimev1.ListConnectorsResponse, error) {
	var pbs []*runtimev1.Connector
	for name, connector := range connectors.Connectors {
		// Build protobufs for properties
		propPBs := make([]*runtimev1.Connector_Property, len(connector.Spec().Properties))
		for j, propSchema := range connector.Spec().Properties {
			// Get type
			var t runtimev1.Connector_Property_Type
			switch propSchema.Type {
			case connectors.StringPropertyType:
				t = runtimev1.Connector_Property_TYPE_STRING
			case connectors.NumberPropertyType:
				t = runtimev1.Connector_Property_TYPE_NUMBER
			case connectors.BooleanPropertyType:
				t = runtimev1.Connector_Property_TYPE_BOOLEAN
			case connectors.InformationalPropertyType:
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
	connector, ok := connectors.Connectors["s3"]
	if !ok {
		panic("s3 connector not found")
	}

	s3Conn := connector.(s3.Connector)
	buckets, err := s3Conn.ListBuckets(ctx, &connectors.Env{AllowHostAccess: s.runtime.AllowHostAccess()})
	if err != nil {
		return nil, err
	}

	return &runtimev1.S3ListBucketsResponse{
		Buckets: buckets,
	}, nil
}

func (s *Server) S3ListObjects(ctx context.Context, req *runtimev1.S3ListObjectsRequest) (*runtimev1.S3ListObjectsResponse, error) {
	connector, ok := connectors.Connectors["s3"]
	if !ok {
		panic("s3 connector not found")
	}

	s3Conn := connector.(s3.Connector)
	objects, nextToken, err := s3Conn.ListObjects(ctx, req, &connectors.Env{AllowHostAccess: s.runtime.AllowHostAccess()})
	if err != nil {
		return nil, err
	}

	return &runtimev1.S3ListObjectsResponse{
		Objects:       objects,
		NextPageToken: nextToken,
	}, nil
}

func (s *Server) S3GetBucketMetadata(ctx context.Context, req *runtimev1.S3GetBucketMetadataRequest) (*runtimev1.S3GetBucketMetadataResponse, error) {
	connector, ok := connectors.Connectors["s3"]
	if !ok {
		panic("s3 connector not found")
	}

	s3Conn := connector.(s3.Connector)
	region, err := s3Conn.GetBucketMetadata(ctx, req, &connectors.Env{AllowHostAccess: s.runtime.AllowHostAccess()})
	if err != nil {
		return nil, err
	}

	return &runtimev1.S3GetBucketMetadataResponse{
		Region: region,
	}, nil
}

func (s *Server) GCSListBuckets(ctx context.Context, req *runtimev1.GCSListBucketsRequest) (*runtimev1.GCSListBucketsResponse, error) {
	connector, ok := connectors.Connectors["gcs"]
	if !ok {
		panic("gcs connector not found")
	}

	gcsConn := connector.(gcs.Connector)
	buckets, next, err := gcsConn.ListBuckets(ctx, req, &connectors.Env{AllowHostAccess: s.runtime.AllowHostAccess()})
	if err != nil {
		return nil, err
	}

	return &runtimev1.GCSListBucketsResponse{
		Buckets:       buckets,
		NextPageToken: next,
	}, nil
}

func (s *Server) GCSListObjects(ctx context.Context, req *runtimev1.GCSListObjectsRequest) (*runtimev1.GCSListObjectsResponse, error) {
	connector, ok := connectors.Connectors["gcs"]
	if !ok {
		panic("gcs connector not found")
	}

	gcsConn := connector.(gcs.Connector)
	objects, nextToken, err := gcsConn.ListObjects(ctx, req, &connectors.Env{AllowHostAccess: s.runtime.AllowHostAccess()})
	if err != nil {
		return nil, err
	}

	return &runtimev1.GCSListObjectsResponse{
		Objects:       objects,
		NextPageToken: nextToken,
	}, nil
}
