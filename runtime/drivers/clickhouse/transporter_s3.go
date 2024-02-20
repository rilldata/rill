package clickhouse

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
)

type s3transporter struct {
	from   drivers.Handle
	to     drivers.OLAPStore
	logger *zap.Logger
}

var _ drivers.Transporter = &s3transporter{}

type sourceProperties struct {
	URI string `mapstructure:"uri"`
}

func NewS3Transporter(from drivers.Handle, olap drivers.OLAPStore, logger *zap.Logger) drivers.Transporter {
	return &s3transporter{
		from:   from,
		to:     olap,
		logger: logger,
	}
}

func (t *s3transporter) Transfer(ctx context.Context, srcProps, sinkProps map[string]any, opts *drivers.TransferOptions) error {
	sinkCfg, err := parseSinkProperties(sinkProps)
	if err != nil {
		return err
	}

	t.logger = t.logger.With(zap.String("source", sinkCfg.Table))

	conf := &sourceProperties{}
	if err := mapstructure.WeakDecode(srcProps, conf); err != nil {
		return err
	}

	// credentials are expected to be added to clickhouse server
	tableName := fmt.Sprintf("s3_engine_%s_table", safeSQLName(sinkCfg.Table))
	if err := t.to.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("CREATE TABLE %s ENGINE=S3('%s')", safeSQLName(tableName), conf.URI)}); err != nil {
		return fmt.Errorf("failed to create table %q with http engine: %w", tableName, err)
	}

	defer func() {
		if err := t.to.DropTable(ctx, tableName, false); err != nil {
			t.logger.Error("failed to drop table", zap.String("name", tableName), zap.Error(err))
		}
	}()

	return t.to.CreateTableAsSelect(ctx, sinkCfg.Table, false, fmt.Sprintf("SELECT * FROM %s", safeSQLName(tableName)))
}
