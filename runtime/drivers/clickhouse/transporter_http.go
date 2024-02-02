package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
)

type httpTransporter struct {
	from   drivers.Handle
	to     drivers.OLAPStore
	logger *zap.Logger
}

var _ drivers.Transporter = &s3transporter{}

type httpSourceProperties struct {
	URI string `mapstructure:"uri"`
}

func NewHTTPTransporter(from drivers.Handle, olap drivers.OLAPStore, logger *zap.Logger) drivers.Transporter {
	return &httpTransporter{
		from:   from,
		to:     olap,
		logger: logger,
	}
}

func (t *httpTransporter) Transfer(ctx context.Context, srcProps, sinkProps map[string]any, opts *drivers.TransferOptions) error {
	sinkCfg, err := parseSinkProperties(sinkProps)
	if err != nil {
		return err
	}

	t.logger = t.logger.With(zap.String("source", sinkCfg.Table))

	conf := &httpSourceProperties{}
	if err := mapstructure.WeakDecode(srcProps, conf); err != nil {
		return err
	}

	tableName := fmt.Sprintf("http_%v", time.Now().UnixNano())
	if err := t.to.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("CREATE TABLE %s ENGINE=URL('%s')", safeSQLName(tableName), conf.URI)}); err != nil {
		return fmt.Errorf("failed to create table %q with http engine: %w", tableName, err)
	}

	defer func() {
		if err := t.to.DropTable(ctx, tableName, false); err != nil {
			t.logger.Error("failed to drop table", zap.String("name", tableName), zap.Error(err))
		}
	}()

	return t.to.CreateTableAsSelect(ctx, sinkCfg.Table, false, fmt.Sprintf("SELECT * FROM %s", safeSQLName(tableName)))
}
