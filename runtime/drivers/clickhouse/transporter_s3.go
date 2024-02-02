package clickhouse

import (
	"context"
	"fmt"
	"strconv"
	"time"

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
	URI       string `mapstructure:"uri"`
	AWSRegion string `mapstructure:"region"`
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

	config := t.from.Config()
	var useEnvCreds string
	if v, ok := config["allow_host_access"].(bool); ok {
		useEnvCreds = strconv.FormatBool(v)
	}
	settings := fmt.Sprintf("url='%s', use_environment_credentials=%s", conf.URI, useEnvCreds)
	if conf.AWSRegion != "" {
		settings += fmt.Sprintf(", region='%s'", conf.AWSRegion)
	}
	if v, ok := config["aws_access_key_id"].(string); ok && v != "" {
		settings += fmt.Sprintf(", access_key_id='%s'", v)
	}
	if v, ok := config["aws_secret_access_key"].(string); ok && v != "" {
		settings += fmt.Sprintf(", secret_access_key='%s'", v)
	}
	if v, ok := config["aws_access_token"].(string); ok && v != "" {
		settings += fmt.Sprintf(", session_token='%s'", v)
	}

	collectionName := fmt.Sprintf("s3_%v", time.Now().UnixNano())
	if err := t.to.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("CREATE NAMED COLLECTION %v AS %v", collectionName, settings)}); err != nil {
		return fmt.Errorf("failed to create named collection %q: %w", collectionName, err)
	}

	defer func() {
		if err := t.to.Exec(context.Background(), &drivers.Statement{Query: fmt.Sprintf("DROP NAMED COLLECTION %v", collectionName)}); err != nil {
			t.logger.Info("failed to drop named collection", zap.String("collection", collectionName), zap.Error(err))
		}
	}()

	return t.to.CreateTableAsSelect(ctx, sinkCfg.Table, false, fmt.Sprintf("SELECT * FROM s3(%v)", collectionName))
}
