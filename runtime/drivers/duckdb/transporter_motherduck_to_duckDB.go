package duckdb

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
)

type motherduckToDuckDB struct {
	to     drivers.OLAPStore
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

func NewMotherduckToDuckDB(from drivers.Handle, to drivers.OLAPStore, logger *zap.Logger) drivers.Transporter {
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

	// we first ingest data in a temporary table in the main db
	// and then copy it to the final table to ensure that the final table is always created using CRUD APIs which takes care
	// whether table goes in main db or in separate table specific db
	tmpTable := fmt.Sprintf("__%s_tmp_motherduck", sinkCfg.Table)
	defer func() {
		// ensure temporary table is cleaned
		err := t.to.Exec(context.Background(), &drivers.Statement{
			Query:       fmt.Sprintf("DROP TABLE IF EXISTS %s", tmpTable),
			Priority:    100,
			LongRunning: true,
		})
		if err != nil {
			t.logger.Error("failed to drop temp table", zap.String("table", tmpTable), zap.Error(err))
		}
	}()

	err = t.to.WithConnection(ctx, 1, true, false, func(ctx, ensuredCtx context.Context, _ *sql.Conn) error {
		// load motherduck extension; connect to motherduck service
		err = t.to.Exec(ctx, &drivers.Statement{Query: "INSTALL 'motherduck'; LOAD 'motherduck';"})
		if err != nil {
			return fmt.Errorf("failed to load motherduck extension %w", err)
		}

		if err = t.to.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("SET motherduck_token='%s'", token)}); err != nil {
			if !strings.Contains(err.Error(), "can only be set during initialization") {
				return fmt.Errorf("failed to set motherduck token %w", err)
			}
		}

		// ignore attach error since it might be already attached
		_ = t.to.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("ATTACH '%s'", srcConfig.DSN)})
		userQuery := strings.TrimSpace(srcConfig.SQL)
		userQuery, _ = strings.CutSuffix(userQuery, ";") // trim trailing semi colon
		query := fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (%s\n);", safeName(tmpTable), userQuery)
		return t.to.Exec(ctx, &drivers.Statement{Query: query})
	})
	if err != nil {
		return err
	}

	// copy data from temp table to target table
	return t.to.CreateTableAsSelect(ctx, sinkCfg.Table, false, fmt.Sprintf("SELECT * FROM %s", tmpTable), nil)
}
