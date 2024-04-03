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
	DisplayName           string
	Description           string
	DocsURL               string
	ConfigProperties      []*PropertySpec
	SourceProperties      []*PropertySpec
	ImplementsRegistry    bool
	ImplementsCatalog     bool
	ImplementsRepo        bool
	ImplementsAdmin       bool
	ImplementsAI          bool
	ImplementsSQLStore    bool
	ImplementsOLAP        bool
	ImplementsObjectStore bool
	ImplementsFileStore   bool
}

// PropertySpec provides metadata about a single connector property.
type PropertySpec struct {
	Key         string
	Type        PropertyType
	Required    bool
	DisplayName string
	Description string
	DocsURL     string
	Hint        string
	Default     string
	Placeholder string
	Secret      bool
}

// PropertyType is an enum of types supported for connector properties.
type PropertyType int

const (
	UnspecifiedPropertyType PropertyType = iota
	NumberPropertyType
	BooleanPropertyType
	StringPropertyType
	FilePropertyType
	InformationalPropertyType
)
