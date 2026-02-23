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
	ImplementsNotifier    bool
	ImplementsWarehouse   bool
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
	// EnvVarName is the conventional env var name for this property (e.g. AWS_ACCESS_KEY_ID, GOOGLE_APPLICATION_CREDENTIALS).
	// It must be specified explicitly because the mapping doesn't follow a mechanical pattern;
	// some keys use well-known names shared across drivers (AWS_*), others add infixes (AZURE_STORAGE_*),
	// and others diverge entirely from the key name (GCS key_id -> GCP_ACCESS_KEY_ID).
	EnvVarName string
	NoPrompt   bool
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
