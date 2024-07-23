package snowflake

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/s3"
)

type selfToObjectStoreExecutor struct {
	c     *connection
	store drivers.ObjectStore
	opts  *drivers.ModelExecutorOptions
}

var _ drivers.ModelExecutor = &selfToObjectStoreExecutor{}

func (e *selfToObjectStoreExecutor) Execute(ctx context.Context) (*drivers.ModelResult, error) {
	props := &modelOutputProperties{}
	if err := mapstructure.Decode(e.opts.OutputProperties, props); err != nil {
		return nil, err
	}
	outputGlob, err := e.export(ctx, e.opts.InputProperties, props.Path)
	if err != nil {
		return nil, err
	}
	resProps := &modelResultProperties{Path: outputGlob}
	res := make(map[string]any)
	err = mapstructure.Decode(resProps, &res)
	if err != nil {
		return nil, err
	}

	return &drivers.ModelResult{
		Connector:  e.opts.OutputConnector,
		Properties: res,
	}, nil
}

func (e *selfToObjectStoreExecutor) export(ctx context.Context, props map[string]any, outputLocation string) (string, error) {
	conf := &sourceProperties{}
	err := mapstructure.Decode(props, conf)
	if err != nil {
		return "", err
	}
	if conf.SQL == "" {
		return "", fmt.Errorf("property 'sql' is mandatory for connector \"snowflake\"")
	}

	var dsn string
	if conf.DSN != "" { // get from src properties
		dsn = conf.DSN
	} else if e.c.configProperties.DSN != "" { // get from driver configs
		dsn = e.c.configProperties.DSN
	} else {
		return "", fmt.Errorf("the property 'dsn' is required for Snowflake. Provide 'dsn' in the YAML properties or pass '--var connector.snowflake.dsn=...' to 'rill start'")
	}

	db, err := sql.Open("snowflake", dsn)
	if err != nil {
		return "", err
	}

	outputLocation, err = url.JoinPath(outputLocation, "rill-tmp-"+uuid.New().String(), "/")
	if err != nil {
		return "", err
	}

	creds, err := creds(e.store)
	if err != nil {
		return "", err
	}

	//nolint:gosec // can't pass as query args
	query := fmt.Sprintf(`
	COPY INTO '%s'
		FROM (%s) 
		CREDENTIALS = %s
		HEADER = TRUE
		MAX_FILE_SIZE = 536870912 
		FILE_FORMAT = (TYPE='PARQUET' COMPRESSION = 'SNAPPY')`, outputLocation, conf.SQL, creds)
	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return "", err
	}
	return url.JoinPath(outputLocation, "*."+string(drivers.FileFormatParquet))
}

func creds(store drivers.ObjectStore) (string, error) {
	h := store.(drivers.Handle)
	switch h.Driver() {
	case "s3":
		conf := &s3.ConfigProperties{}
		if err := mapstructure.Decode(h.Config(), conf); err != nil {
			return "", err
		}
		return fmt.Sprintf("(AWS_KEY_ID='%s' AWS_SECRET_KEY='%s' AWS_TOKEN='%s')", conf.AccessKeyID, conf.SecretAccessKey, conf.SessionToken), nil
	case "gcs":
		return "", fmt.Errorf("snowflake connector can't export to connector 'gcs'. Use s3 compatibility.")
	default:
		return "", fmt.Errorf("snowflake connector can't export to connector %q", h.Driver())
	}
}

type modelOutputProperties struct {
	Path string `mapstructure:"path"`
}

func (p *modelOutputProperties) Validate(opts *drivers.ModelExecutorOptions) error {
	if p.Path == "" {
		return fmt.Errorf("missing property 'path'")
	}
	return nil
}

type modelResultProperties struct {
	Path string `mapstructure:"path"`
}
