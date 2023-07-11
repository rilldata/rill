package connectors

import (
	"fmt"
	"reflect"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

// Source represents a dataset to ingest using a specific connector (like a connector instance).
type Source struct {
	Name          string
	Connector     string
	ExtractPolicy *runtimev1.Source_ExtractPolicy
	Properties    map[string]any
	Timeout       int32
}

// Validate checks the source's properties against its connector's spec.
func (s *Source) Validate() error {
	connector, ok := drivers.Connectors[s.Connector]
	if !ok {
		return fmt.Errorf("connector: not found %q", s.Connector)
	}

	for _, propSchema := range connector.Spec().Properties {
		val, ok := s.Properties[propSchema.Key]
		if !ok {
			if propSchema.Required {
				return fmt.Errorf("missing required property '%s'", propSchema.Key)
			}
			continue
		}

		if !propSchema.ValidateType(val) {
			return fmt.Errorf("unexpected type '%T' for property '%s'", val, propSchema.Key)
		}
	}

	return nil
}

func (s *Source) PropertiesEquals(o *Source) bool {
	return reflect.DeepEqual(s.Properties, o.Properties)
}
