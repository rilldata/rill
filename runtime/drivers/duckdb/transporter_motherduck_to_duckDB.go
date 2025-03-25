package duckdb

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/rduckdb"
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

	beforeCreateFn := func(ctx context.Context, conn *sqlx.Conn) error {
		_, err = conn.ExecContext(ctx, fmt.Sprintf("ATTACH '%s'", srcConfig.DSN))
		if err != nil {
			return fmt.Errorf("failed to attach motherduck DSN: %w", err)
		}
		return err
	}
	userQuery := strings.TrimSpace(srcConfig.SQL)
	userQuery, _ = strings.CutSuffix(userQuery, ";") // trim trailing semi colon
	db, release, err := t.to.acquireDB()
	if err != nil {
		return err
	}
	defer func() {
		_ = release()
	}()
	_, err = db.CreateTableAsSelect(ctx, sinkCfg.Table, userQuery, &rduckdb.CreateTableOptions{
		BeforeCreateFn: beforeCreateFn,
		InitQueries: []string{
			"INSTALL 'motherduck'; LOAD 'motherduck';",
			fmt.Sprintf("SET motherduck_token='%s'", token),
		},
	})
	return err
}
