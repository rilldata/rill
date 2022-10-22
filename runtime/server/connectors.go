package server

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/connectors"
)

// ListConnectors implements RuntimeService
func (s *Server) ListConnectors(ctx context.Context, req *api.ListConnectorsRequest) (*api.ListConnectorsResponse, error) {
	var pbs []*api.Connector
	for name, connector := range connectors.Connectors {
		// Build protobufs for properties
		propPBs := make([]*api.Connector_Property, len(connector.Spec().Properties))
		for j, propSchema := range connector.Spec().Properties {
			// Get type
			var t api.Connector_Property_Type
			switch propSchema.Type {
			case connectors.StringPropertyType:
				t = api.Connector_Property_TYPE_STRING
			case connectors.NumberPropertyType:
				t = api.Connector_Property_TYPE_NUMBER
			case connectors.BooleanPropertyType:
				t = api.Connector_Property_TYPE_BOOLEAN
			default:
				panic(fmt.Errorf("property type '%v' not handled", propSchema.Type))
			}

			// Add protobuf for property
			propPBs[j] = &api.Connector_Property{
				Key:         propSchema.Key,
				DisplayName: propSchema.DisplayName,
				Description: propSchema.Description,
				Placeholder: propSchema.Placeholder,
				Type:        t,
				Nullable:    !propSchema.Required,
			}
		}

		// Add connector
		pbs = append(pbs, &api.Connector{
			Name:        name,
			DisplayName: connector.Spec().DisplayName,
			Description: connector.Spec().Description,
			Properties:  propPBs,
		})
	}

	return &api.ListConnectorsResponse{Connectors: pbs}, nil
}
