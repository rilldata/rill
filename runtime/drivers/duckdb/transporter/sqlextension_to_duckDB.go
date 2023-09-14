package transporter

import (
	"context"
	"database/sql"
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

func (t *sqlextensionToDuckDB) Transfer(ctx context.Context, srcProps, sinkProps map[string]any, opts *drivers.TransferOpts, p drivers.Progress) error {
	srcCfg, err := parseDBSourceProperties(srcProps)
	if err != nil {
		return err
	}

	sinkCfg, err := parseSinkProperties(sinkProps)
	if err != nil {
		return err
	}

	extensionName := t.from.Driver()
	return t.to.WithConnection(ctx, 1, true, false, func(ctx, ensuredCtx context.Context, _ *sql.Conn) error {
		if err := t.to.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("INSTALL '%s'; LOAD '%s';", extensionName, extensionName)}); err != nil {
			return fmt.Errorf("failed to load %s extension %w", extensionName, err)
		}

		userQuery := strings.TrimSpace(srcCfg.SQL)
		userQuery, _ = strings.CutSuffix(userQuery, ";") // trim trailing semi colon
		query := fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (%s);", safeName(sinkCfg.Table), userQuery)

		return t.to.Exec(ctx, &drivers.Statement{Query: query, Priority: 1})
	})
}
