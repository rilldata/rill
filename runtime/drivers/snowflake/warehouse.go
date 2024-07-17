package snowflake

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
)

var _ drivers.Warehouse = &connection{}

func (c *connection) Export(ctx context.Context, props map[string]any, store drivers.ObjectStore, outputLocation string) (*drivers.ExportResult, error) {
	conf := &sourceProperties{}
	err := mapstructure.Decode(props, conf)
	if err != nil {
		return nil, err
	}
	if conf.SQL == "" {
		return nil, fmt.Errorf("property 'sql' is mandatory for connector \"snowflake\"")
	}

	var dsn string
	if conf.DSN != "" { // get from src properties
		dsn = conf.DSN
	} else if c.configProperties.DSN != "" { // get from driver configs
		dsn = c.configProperties.DSN
	} else {
		return nil, fmt.Errorf("the property 'dsn' is required for Snowflake. Provide 'dsn' in the YAML properties or pass '--var connector.snowflake.dsn=...' to 'rill start'")
	}

	db, err := sql.Open("snowflake", dsn)
	if err != nil {
		return nil, err
	}

	outputLocation, err = url.JoinPath(outputLocation, "rill-tmp-"+uuid.New().String(), "/")
	if err != nil {
		return nil, err
	}

	creds, err := creds(store)
	if err != nil {
		return nil, err
	}

	//nolint:g201
	query := fmt.Sprintf(`
	COPY INTO '%s'
		FROM (%s) 
		CREDENTIALS = %s
		HEADER = TRUE
		MAX_FILE_SIZE = 536870912 
		FILE_FORMAT = (TYPE='PARQUET' COMPRESSION = 'SNAPPY')`, outputLocation, conf.SQL, creds)
	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return &drivers.ExportResult{Path: outputLocation, Format: drivers.FileFormatParquet}, nil
}

func creds(store drivers.ObjectStore) (string, error) {
	h := store.(drivers.Handle)
	switch h.Driver() {
	case "s3":
		conf := &s3ConfigProperties{}
		fmt.Println(h.Config())
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

type s3ConfigProperties struct {
	AccessKeyID     string `mapstructure:"aws_access_key_id"`
	SecretAccessKey string `mapstructure:"aws_secret_access_key"`
	SessionToken    string `mapstructure:"aws_access_token"`
}
