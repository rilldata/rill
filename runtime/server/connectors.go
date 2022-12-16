package server

import (
	"context"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/connectors"
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
