package duckdb

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
)

type motherduckToDuckDB struct {
	to     *connection
	from   drivers.Handle
	logger *zap.Logger
}

type mdSrcProps struct {
	DSN   string `mapstructure:"dsn"`
	Token string `mapstructure:"token"`
	SQL   string `mapstructure:"sql"`
}

type mdConfigProps struct {
	Token           string `mapstructure:"token"`
	AllowHostAccess bool   `mapstructure:"allow_host_access"`
}

var _ drivers.Transporter = &motherduckToDuckDB{}

func newMotherduckToDuckDB(from drivers.Handle, to *connection, logger *zap.Logger) drivers.Transporter {
	return &motherduckToDuckDB{
		to:     to,
		from:   from,
		logger: logger,
	}
}

func (t *motherduckToDuckDB) Transfer(ctx context.Context, srcProps, sinkProps map[string]any, opts *drivers.TransferOptions) error {
	srcConfig := &mdSrcProps{}
	err := mapstructure.WeakDecode(srcProps, srcConfig)
	if err != nil {
		return err
	}
	if srcConfig.SQL == "" {
		return fmt.Errorf("property \"sql\" is mandatory for connector \"motherduck\"")
	}

	sinkCfg, err := parseSinkProperties(sinkProps)
	if err != nil {
		return err
	}

	mdConfig := &mdConfigProps{}
	err = mapstructure.WeakDecode(t.from.Config(), mdConfig)
	if err != nil {
		return err
	}

	// get token
	var token string
	if srcConfig.Token != "" {
		token = srcConfig.Token
	} else if mdConfig.Token != "" {
		token = mdConfig.Token
	} else if mdConfig.AllowHostAccess {
		token = os.Getenv("motherduck_token")
	}
	if token == "" {
		return fmt.Errorf("no motherduck token found. Refer to this documentation for instructions: https://docs.rilldata.com/reference/connectors/motherduck")
	}

	t.logger = t.logger.With(zap.String("source", sinkCfg.Table))

	rwConn, release, err := t.to.acquireConn(ctx, false)
	if err != nil {
		return err
	}
	defer release()

	conn := rwConn.Connx()

	// load motherduck extension; connect to motherduck service
	_, err = conn.ExecContext(ctx, "INSTALL 'motherduck'; LOAD 'motherduck';")
	if err != nil {
		return fmt.Errorf("failed to load motherduck extension %w", err)
	}

	if _, err = conn.ExecContext(ctx, fmt.Sprintf("SET motherduck_token='%s'", token)); err != nil {
		if !strings.Contains(err.Error(), "can only be set during initialization") {
			return fmt.Errorf("failed to set motherduck token %w", err)
		}
	}

	// ignore attach error since it might be already attached
	_, _ = conn.ExecContext(ctx, fmt.Sprintf("ATTACH '%s'", srcConfig.DSN))
	userQuery := strings.TrimSpace(srcConfig.SQL)
	userQuery, _ = strings.CutSuffix(userQuery, ";") // trim trailing semi colon

	// we first ingest data in a temporary table in the main db
	// and then copy it to the final table to ensure that the final table is always created using CRUD APIs
	tmpTable := fmt.Sprintf("__%s_tmp_motherduck", sinkCfg.Table)
	defer func() {
		// ensure temporary table is cleaned
		_, err := conn.ExecContext(context.Background(), fmt.Sprintf("DROP TABLE IF EXISTS %s", tmpTable))
		if err != nil {
			t.logger.Error("failed to drop temp table", zap.String("table", tmpTable), zap.Error(err))
		}
	}()

	query := fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (%s\n);", safeName(tmpTable), userQuery)
	_, err = conn.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	// copy data from temp table to target table
	return rwConn.CreateTableAsSelect(ctx, sinkCfg.Table, fmt.Sprintf("SELECT * FROM %s", tmpTable), nil)

}
