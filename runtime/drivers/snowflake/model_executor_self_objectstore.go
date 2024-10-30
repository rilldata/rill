package snowflake

import (
	"context"
	"fmt"
	"net/url"

	"github.com/XSAM/otelsql"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/s3"
)

type selfToObjectStoreExecutor struct {
	c     *connection
	store drivers.ObjectStore
}

var _ drivers.ModelExecutor = &selfToObjectStoreExecutor{}

func (e *selfToObjectStoreExecutor) Concurrency(desired int) (int, bool) {
	if desired > 0 {
		return desired, true
	}
	return 10, true // Default
}

func (e *selfToObjectStoreExecutor) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
	props := &drivers.ObjectStoreModelOutputProperties{}
	if err := mapstructure.Decode(opts.OutputProperties, props); err != nil {
		return nil, err
	}
	var format drivers.FileFormat
	if props.Format != "" {
		format = props.Format
	} else {
		format = drivers.FileFormatParquet
	}
	outputLocation, err := e.export(ctx, opts.InputProperties, props.Path, format)
	if err != nil {
		return nil, err
	}
	resProps := &drivers.ObjectStoreModelResultProperties{Path: outputLocation, Format: string(drivers.FileFormatParquet)}
	res := make(map[string]any)
	err = mapstructure.Decode(resProps, &res)
	if err != nil {
		return nil, err
	}

	return &drivers.ModelResult{
		Connector:  opts.OutputConnector,
		Properties: res,
	}, nil
}

func (e *selfToObjectStoreExecutor) export(ctx context.Context, props map[string]any, outputLocation string, format drivers.FileFormat) (string, error) {
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
		return "", fmt.Errorf("the property 'dsn' is required for Snowflake. Provide 'dsn' in the YAML properties or pass '--env connector.snowflake.dsn=...' to 'rill start'")
	}

	db, err := otelsql.Open("snowflake", dsn)
	if err != nil {
		return "", err
	}
	defer db.Close()

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
		FILE_FORMAT = (TYPE='%s' COMPRESSION = 'SNAPPY')`, outputLocation, conf.SQL, creds, string(format))
	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return "", err
	}
	return outputLocation, nil
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
