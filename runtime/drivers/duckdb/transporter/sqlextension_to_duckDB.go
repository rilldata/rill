package transporter

import (
	"context"
	"fmt"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
)

type sqlextensionToDuckDB struct {
	to     drivers.OLAPStore
	from   drivers.Handle
	logger *zap.Logger
}

var _ drivers.Transporter = &sqlextensionToDuckDB{}

// NewSQLExtensionToDuckDB returns a transporter to transfer data to duckdb from a sql extension supported by duckdb.
// Currently only sqlite extension is supported. Postgres is not supported due to licensing issues.
func NewSQLExtensionToDuckDB(from drivers.Handle, to drivers.OLAPStore, logger *zap.Logger) drivers.Transporter {
	return &sqlextensionToDuckDB{
		to:     to,
		from:   from,
		logger: logger,
	}
}

func (t *sqlextensionToDuckDB) Transfer(ctx context.Context, source drivers.Source, sink drivers.Sink, opts *drivers.TransferOpts, p drivers.Progress) error {
	src, ok := source.DatabaseSource()
	if !ok {
		return fmt.Errorf("type of source should be `drivers.DatabaseSource`")
	}
	fSink, ok := sink.DatabaseSink()
	if !ok {
		return fmt.Errorf("type of source should be `drivers.DatabaseSink`")
	}

	extensionName := t.from.Driver()
	if err := t.to.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("INSTALL '%s'; LOAD '%s';", extensionName, extensionName)}); err != nil {
		return fmt.Errorf("failed to load %s extension %w", extensionName, err)
	}

	userQuery := strings.TrimSpace(src.SQL)
	userQuery, _ = strings.CutSuffix(userQuery, ";") // trim trailing semi colon
	query := fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (%s);", safeName(fSink.Table), userQuery)

	return t.to.Exec(ctx, &drivers.Statement{Query: query, Priority: 1})
}
