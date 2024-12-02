package duckdb

import (
	"context"
	"fmt"
	"os"

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

	return t.to.db.CreateTableAsSelect(ctx, sinkCfg.Table, srcConfig.SQL, &rduckdb.CreateTableOptions{
		// InitSQL: fmt.Sprintf("INSTALL 'motherduck'; LOAD 'motherduck'; SET motherduck_token='%s'; ATTACH '%s'", token, srcConfig.DSN),
	})
}
