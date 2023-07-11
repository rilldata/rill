package drivers

import (
	"context"
	"fmt"
)

// Connectors tracks all registered connector drivers.
var Connectors = make(map[string]Connector)

// RegisterConnector tracks a connector driver.
func RegisterConnector(name string, connector Connector) {
	if Connectors[name] != nil {
		panic(fmt.Errorf("already registered connector with name '%s'", name))
	}
	Connectors[name] = connector
}

// Connector is a driver for ingesting data from an external system.
type Connector interface {
	Spec() Spec

	// HasAnonymousAccess returns true if external system can be accessed without credentials
	HasAnonymousAccess(ctx context.Context, props map[string]any) (bool, error)
}

// Spec provides metadata about a connector and the properties it supports.
type Spec struct {
	DisplayName        string
	Description        string
	ServiceAccountDocs string
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
