package connectors

import (
	"context"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

var ErrIngestionLimitExceeded = fmt.Errorf("connectors: source ingestion exceeds limit")

// Connectors tracks all registered connector drivers.
var Connectors = make(map[string]Connector)

// Register tracks a connector driver.
func Register(name string, connector Connector) {
	if Connectors[name] != nil {
		panic(fmt.Errorf("already registered connector with name '%s'", name))
	}
	Connectors[name] = connector
}

// Connector is a driver for ingesting data from an external system.
type Connector interface {
	Spec() Spec

	// TODO: Add method that extracts a source and outputs a schema and buffered
	// iterator for data in it. For consumption by a drivers.OLAPStore. Also consider
	// how to communicate splits and long-running/streaming data (e.g. for Kafka).
	// Consume(ctx context.Context, source Source) error

	ConsumeAsIterator(ctx context.Context, env *Env, source *Source) (FileIterator, error)
}

// Spec provides metadata about a connector and the properties it supports.
type Spec struct {
	DisplayName        string
	Description        string
	Properties         []PropertySchema
	ConnectorVariables []VariableSchema
	Help               string
}

// PropertySchema provides the schema for a property supported by a connector.
type PropertySchema struct {
	Key         string
	Type        PropertySchemaType
	Required    bool
	DisplayName string
	Description string
	Placeholder string
	Hint        string
	Href        string
}

type VariableSchema struct {
	Key           string
	Default       string
	Help          string
	Secret        bool
	ValidateFunc  func(any interface{}) error
	TransformFunc func(any interface{}) interface{}
}

// PropertySchemaType is an enum of types supported for connector properties.
type PropertySchemaType int

const (
	UnspecifiedPropertyType PropertySchemaType = iota
	StringPropertyType
	NumberPropertyType
	BooleanPropertyType
	InformationalPropertyType
)

// ValidateType checks that val has the correct type.
func (ps PropertySchema) ValidateType(val any) bool {
	switch val.(type) {
	case string:
		return ps.Type == StringPropertyType
	case bool:
		return ps.Type == BooleanPropertyType
	case int, int8, int16, int32, int64, float32, float64:
		return ps.Type == NumberPropertyType
	default:
		return false
	}
}

// Env contains contextual information for a source, such as the repo it came from
// and (in the future) secrets configured by the user.
type Env struct {
	RepoDriver string
	RepoDSN    string
	// user provided env variables kept with keys converted to uppercase
	Variables            map[string]string
	AllowHostCredentials bool
	StorageLimitInBytes  int64
}

// Source represents a dataset to ingest using a specific connector (like a connector instance).
type Source struct {
	Name          string
	Connector     string
	ExtractPolicy *runtimev1.Source_ExtractPolicy
	Properties    map[string]any
	Timeout       int32
}

// SamplePolicy tells the connector to only ingest a sample of data from the source.
// Support for it is currently not implemented.
type SamplePolicy struct {
	Strategy string
	Sample   float32
	Limit    int
}

// FileIterator provides ways to iteratively ingest files downloaded from external sources
// Clients should call close once they are done with iterator to release any resources
type FileIterator interface {
	// Close do cleanup and release resources
	Close() error
	// NextBatch returns a list of file downloaded from external sources
	// NextBatch cleanups file created in previous batch
	NextBatch(limit int) ([]string, error)
	// HasNext can be utlisied to check if iterator has more elements left
	HasNext() bool
}

// Validate checks the source's properties against its connector's spec.
func (s *Source) Validate() error {
	connector, ok := Connectors[s.Connector]
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

func ConsumeAsIterator(ctx context.Context, env *Env, source *Source) (FileIterator, error) {
	connector, ok := Connectors[source.Connector]
	if !ok {
		return nil, fmt.Errorf("connector: not found")
	}
	return connector.ConsumeAsIterator(ctx, env, source)
}

func (s *Source) PropertiesEquals(o *Source) bool {
	if len(s.Properties) != len(o.Properties) {
		return false
	}

	for k1, v1 := range s.Properties {
		v2, ok := o.Properties[k1]
		if !ok || v1 != v2 {
			return false
		}
	}

	return true
}
