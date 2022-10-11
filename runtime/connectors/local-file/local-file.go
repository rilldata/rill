package local_file

import (
	"context"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/connectors/misc"
	"github.com/rilldata/rill/runtime/drivers"
)

func init() {
	connectors.Register(misc.LocalFileConnectorName, connector{})
}

type connector struct{}

func (c connector) Open(name string, options misc.ConnectorOpenOptions) (connectors.ConnectorInstance, error) {
	return new(connectorInstance), nil
}

type connectorInstance struct{}

func (i *connectorInstance) Ingest(
	ctx context.Context,
	options misc.ConnectorIngestOptions,
	olap drivers.OLAPStore,
) (*sqlx.Rows, error) {
	rows, err := olap.Ingest(ctx, misc.LocalFileConnectorName, options)
	if err != nil || rows != nil {
		return rows, err
	}
	return nil, errors.New("OLAP doesnt support local file")
}
