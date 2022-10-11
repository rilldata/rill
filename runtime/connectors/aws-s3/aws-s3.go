package aws_s3

import (
	"context"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/connectors/misc"
	"github.com/rilldata/rill/runtime/drivers"
)

func init() {
	connectors.Register(misc.AWSS3ConnectorName, connector{})
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
	rows, err := olap.Ingest(ctx, misc.AWSS3ConnectorName, options)
	if err != nil || rows != nil {
		return rows, err
	}
	// TODO: download from s3 and ingest as local file
	return nil, errors.New("OLAP doesnt support s3 files")
}
