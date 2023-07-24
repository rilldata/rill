package drivers

import (
	"fmt"
)

// Connectors tracks all registered connector drivers.
var Connectors = make(map[string]Driver)

// RegisterAsConnector tracks a connector driver.
func RegisterAsConnector(name string, driver Driver) {
	if Connectors[name] != nil {
		panic(fmt.Errorf("already registered connector with name '%s'", name))
	}
	Connectors[name] = driver
}

// Spec provides metadata about a connector and the properties it supports.
type Spec struct {
	DisplayName        string
	Description        string
	ServiceAccountDocs string
	SourceProperties   []PropertySchema
	ConfigProperties   []PropertySchema
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
	// Default can be different from placeholder in the sense that placeholder should not be used as default value.
	// If a default is set then it should also be used as a placeholder.
	Default       string
	Hint          string
	Href          string
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
