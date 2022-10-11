package connectors

import (
	"context"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/runtime/connectors/misc"
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

// Connector interface will deal with connecting to the remote source if there is any
// Otherwise it will return a singleton ConnectorInstance
type Connector interface {
	Open(name string, options misc.ConnectorOpenOptions) (ConnectorInstance, error)
}

// ConnectorInstance will have the implementation to load the data into the olap store
type ConnectorInstance interface {
	Ingest(ctx context.Context, options misc.ConnectorIngestOptions, olap drivers.OLAPStore) (*sqlx.Rows, error)
}
