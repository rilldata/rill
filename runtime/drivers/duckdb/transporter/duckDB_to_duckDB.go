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

func (t *duckDBToDuckDB) Transfer(ctx context.Context, source drivers.Source, sink drivers.Sink, opts *drivers.TransferOpts, p drivers.Progress) error {
	src, ok := source.DatabaseSource()
	if !ok {
		return fmt.Errorf("type of source should `drivers.DatabaseSource`")
	}
	fSink, ok := sink.DatabaseSink()
	if !ok {
		return fmt.Errorf("type of source should `drivers.DatabaseSink`")
	}

	qry := fmt.Sprintf("CREATE OR REPLACE TABLE %q AS (%s)", fSink.Table, src.SQL)
	return t.to.Exec(ctx, &drivers.Statement{Query: qry, Priority: 1})
}
