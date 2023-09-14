package transporter

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
)

type duckDBToDuckDB struct {
	to     drivers.OLAPStore
	logger *zap.Logger
}

func NewDuckDBToDuckDB(to drivers.OLAPStore, logger *zap.Logger) drivers.Transporter {
	return &duckDBToDuckDB{
		to:     to,
		logger: logger,
	}
}

var _ drivers.Transporter = &duckDBToDuckDB{}

func (t *duckDBToDuckDB) Transfer(ctx context.Context, srcProps, sinkProps map[string]any, opts *drivers.TransferOpts, p drivers.Progress) error {
	srcCfg, err := parseSourceProperties(srcProps)
	if err != nil {
		return err
	}

	sinkCfg, err := parseSinkProperties(sinkProps)
	if err != nil {
		return err
	}

	qry := fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (%s)", safeName(sinkCfg.Table), srcCfg.SQL)
	return t.to.Exec(ctx, &drivers.Statement{Query: qry, Priority: 1, LongRunning: true})
}
