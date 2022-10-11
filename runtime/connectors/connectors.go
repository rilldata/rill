package connectors

import (
	"context"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/runtime/connectors/sources"
	"github.com/rilldata/rill/runtime/drivers"
	"reflect"
)

// ErrNotFound indicates the resource wasn't found
var ErrNotFound = errors.New("connector: not found")

const propertyTagName = "key"

var Connectors = make(map[string]Connector)

func Register(name string, connector Connector) {
	if Connectors[name] != nil {
		panic(fmt.Errorf("already registered connector with name '%s'", name))
	}
	Connectors[name] = connector
}

func Create(name string) (Connector, error) {
	if Connectors[name] == nil {
		return nil, ErrNotFound
	}
	return Connectors[name], nil
}

// Connector interface abstract all interactions with a remote/local connection
type Connector interface {
	Ingest(ctx context.Context, source sources.Source, olap drivers.OLAPStore) (*sqlx.Rows, error)
	Validate(source sources.Source) error
	Spec() []sources.Property
}

func ValidatePropertiesAndExtract(source sources.Source, spec []sources.Property, targetStruct interface{}) error {
	structType := reflect.TypeOf(targetStruct).Elem()
	structValue := reflect.ValueOf(targetStruct).Elem()

	// TODO: assumes order of fields in spec and struct is same. match them by iterating fields in structType
	for propIdx, prop := range spec {
		if source.Properties[prop.Key] == nil {
			if prop.Required {
				// TODO: better error object
				return errors.New(fmt.Sprintf("missing key: %s in properties", prop.Key))
			}
			continue
		}

		structField := structType.Field(propIdx)
		structFieldValue := structValue.FieldByName(structField.Name)

		propertyValue := reflect.ValueOf(source.Properties[prop.Key])

		if structFieldValue.Type() != propertyValue.Type() {
			return errors.New(fmt.Sprintf("mismatch type for %s in properties", prop.Key))
		}

		structFieldValue.Set(propertyValue)
	}

	return nil
}
