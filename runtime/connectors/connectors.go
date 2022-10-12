package connectors

import (
	"context"
	"errors"
	"fmt"

	"github.com/rilldata/rill/runtime/connectors/sources"
	"github.com/rilldata/rill/runtime/drivers"
)

// ErrNotFound indicates the resource wasn't found
var ErrNotFound = errors.New("connector: not found")

var Connectors = make(map[string]Connector)

func Register(name string, connector Connector) {
	if Connectors[name] != nil {
		panic(fmt.Errorf("already registered connector with name '%s'", name))
	}
	Connectors[name] = connector
}

func Get(name string) (Connector, error) {
	if Connectors[name] == nil {
		return nil, ErrNotFound
	}
	return Connectors[name], nil
}

// Connector interface abstract all interactions with a remote/local connection
type Connector interface {
	Ingest(ctx context.Context, source sources.Source, olap drivers.OLAPStore) error
	Validate(source sources.Source) error
	Spec() []sources.Property
}

func Ingest(ctx context.Context, source sources.Source, olap drivers.OLAPStore) error {
	connector, err := Get(source.Connector)
	if err != nil {
		return err
	}

	return connector.Ingest(ctx, source, olap)
}

func Validate(source sources.Source) error {
	connector, err := Get(source.Connector)
	if err != nil {
		return err
	}

	// TODO: assumes order of fields in spec and struct is same. match them by iterating fields in structType
	for _, prop := range connector.Spec() {
		if source.Properties[prop.Key] == nil {
			if prop.Required {
				// TODO: better error object
				return errors.New(fmt.Sprintf("missing key: %s in properties", prop.Key))
			}
			continue
		}

		if !validateType(source.Properties[prop.Key], prop) {
			return errors.New(fmt.Sprintf("mismatch type for %s in properties", prop.Key))
		}
	}

	return nil
}

func validateType(value any, property sources.Property) bool {
	switch value.(type) {
	case string:
		return property.Type == sources.StringPropertyType

	case int:
		return property.Type == sources.NumberPropertyType
	case byte:
		return property.Type == sources.NumberPropertyType

	case int8:
		return property.Type == sources.NumberPropertyType
	case int16:
		return property.Type == sources.NumberPropertyType
	case int32:
		return property.Type == sources.NumberPropertyType
	case int64:
		return property.Type == sources.NumberPropertyType

	case float32:
		return property.Type == sources.NumberPropertyType
	case float64:
		return property.Type == sources.NumberPropertyType

	case bool:
		return property.Type == sources.BooleanPropertyType
	}

	return false
}
