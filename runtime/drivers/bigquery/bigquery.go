package bigquery

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"go.uber.org/zap"
	"gocloud.dev/gcp"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

var errNoCredentials = errors.New("empty credentials: set `google_application_credentials` env variable")

func init() {
	drivers.Register("bigquery", driver{})
	drivers.RegisterAsConnector("bigquery", driver{})
}

// spec for bigquery connector
var spec = drivers.Spec{
	DisplayName:        "BigQuery",
	Description:        "Import data from BigQuery.",
	ServiceAccountDocs: "https://docs.rilldata.com/deploy/credentials/gcs",
	SourceProperties: []drivers.PropertySchema{
		{
			Key:         "query",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "Query",
			Description: "Query to extract data from BigQuery.",
			Placeholder: "select * from my_db.my_table;",
		},
		{
			Key:         "projectid",
			Type:        drivers.StringPropertyType,
			DisplayName: "Project ID",
			Description: "Google project ID.",
			Required:    true,
			Placeholder: "projectID",
		},
		{
			Key:         "google_application_credentials",
			DisplayName: "GCP credentials",
			Description: "GCP credentials inferred from your local environment.",
			Type:        drivers.InformationalPropertyType,
			Hint:        "Set your local credentials: <code>gcloud auth application-default login</code> Click to learn more.",
			Href:        "https://docs.rilldata.com/develop/import-data#configure-credentials-for-gcs",
		},
	},
	ConfigProperties: []drivers.PropertySchema{
		{
			Key:  "google_application_credentials",
			Hint: "Enter path of file to load from.",
			ValidateFunc: func(any interface{}) error {
				val := any.(string)
				if val == "" {
					// user can chhose to leave empty for public sources
					return nil
				}

				path, err := fileutil.ExpandHome(strings.TrimSpace(val))
				if err != nil {
					return err
				}

				_, err = os.Stat(path)
				return err
			},
			TransformFunc: func(any interface{}) interface{} {
				val := any.(string)
				if val == "" {
					return ""
				}

				path, err := fileutil.ExpandHome(strings.TrimSpace(val))
				if err != nil {
					return err
				}
				// ignoring error since PathError is already validated
				content, _ := os.ReadFile(path)
				return string(content)
			},
		},
	},
}

type driver struct{}

type configProperties struct {
	SecretJSON      string `mapstructure:"google_application_credentials"`
	AllowHostAccess bool   `mapstructure:"allow_host_access"`
}

func (d driver) Open(config map[string]any, logger *zap.Logger) (drivers.Connection, error) {
	conf := &configProperties{}
	err := mapstructure.Decode(config, conf)
	if err != nil {
		return nil, err
	}

	conn := &Connection{
		config: conf,
		logger: logger,
	}
	return conn, nil
}

func (d driver) Drop(config map[string]any, logger *zap.Logger) error {
	return drivers.ErrDropNotSupported
}

func (d driver) Spec() drivers.Spec {
	return spec
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, src drivers.Source, logger *zap.Logger) (bool, error) {
	c, err := d.Open(nil, logger)
	if err != nil {
		return false, err
	}

	dbsrc, ok := src.DatabaseSource()
	if !ok {
		return false, fmt.Errorf("require database source")
	}
	iter, err := c.(*Connection).Exec(ctx, dbsrc)
	if err != nil {
		return false, err
	}
	defer iter.Close()

	_, err = iter.ResultSchema(ctx)
	return err == nil, err
}

type Connection struct {
	config *configProperties
	logger *zap.Logger
}

var _ drivers.Connection = &Connection{}

var _ drivers.SQLStore = &Connection{}

// Driver implements drivers.Connection.
func (c *Connection) Driver() string {
	return "gcs"
}

// Config implements drivers.Connection.
func (c *Connection) Config() map[string]any {
	m := make(map[string]any, 0)
	_ = mapstructure.Decode(c.config, m)
	return m
}

// Close implements drivers.Connection.
func (c *Connection) Close() error {
	return nil
}

// Registry implements drivers.Connection.
func (c *Connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// Catalog implements drivers.Connection.
func (c *Connection) AsCatalogStore() (drivers.CatalogStore, bool) {
	return nil, false
}

// Repo implements drivers.Connection.
func (c *Connection) AsRepoStore() (drivers.RepoStore, bool) {
	return nil, false
}

// OLAP implements drivers.Connection.
func (c *Connection) AsOLAP() (drivers.OLAPStore, bool) {
	return nil, false
}

// Migrate implements drivers.Connection.
func (c *Connection) Migrate(ctx context.Context) (err error) {
	return nil
}

// MigrationStatus implements drivers.Connection.
func (c *Connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// AsObjectStore implements drivers.Connection.
func (c *Connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsSQLStore implements drivers.Connection.
func (c *Connection) AsSQLStore() (drivers.SQLStore, bool) {
	return c, true
}

// AsTransporter implements drivers.Connection.
func (c *Connection) AsTransporter(from, to drivers.Connection) (drivers.Transporter, bool) {
	return nil, false
}

func (c *Connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

type sourceProperties struct {
	ProjectID        string `mapstructure:"project_id"`
	EnableStorageAPI bool   `mapstructure:"enable_storage_api"`
}

func parseSourceProperties(props map[string]any) (*sourceProperties, error) {
	conf := &sourceProperties{}
	err := mapstructure.Decode(props, conf)
	return conf, err
}

func (c *Connection) Exec(ctx context.Context, src *drivers.DatabaseSource) (drivers.RowIterator, error) {
	props, err := parseSourceProperties(src.Props)
	if err != nil {
		return nil, err
	}

	client, err := c.createClient(ctx, props)
	if err != nil {
		return nil, fmt.Errorf("failed to create bigquery client: %w", err)
	}

	if props.EnableStorageAPI {
		if err := client.EnableStorageReadClient(ctx); err != nil {
			return nil, err
		}
	}

	q := client.Query(src.Query)
	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to run query: %w", err)
	}

	return &rowIterator{
		client: client,
		bqIter: it,
	}, nil
}

func (c *Connection) createClient(ctx context.Context, props *sourceProperties) (*bigquery.Client, error) {
	creds, err := c.resolvedCredentials(ctx)
	if err != nil {
		if !errors.Is(err, errNoCredentials) {
			return nil, err
		}

		return bigquery.NewClient(ctx, props.ProjectID)
	}
	return bigquery.NewClient(ctx, props.ProjectID, option.WithCredentials(creds))
}

func (c *Connection) resolvedCredentials(ctx context.Context) (*google.Credentials, error) {
	if c.config.SecretJSON != "" {
		// google_application_credentials is set, use credentials from json string provided by user
		return google.CredentialsFromJSON(ctx, []byte(c.config.SecretJSON), "https://www.googleapis.com/auth/cloud-platform")
	}
	// google_application_credentials is not set
	if c.config.AllowHostAccess {
		// use host credentials
		creds, err := gcp.DefaultCredentials(ctx)
		if err != nil {
			if strings.Contains(err.Error(), "google: could not find default credentials") {
				return nil, errNoCredentials
			}

			return nil, err
		}
		return creds, nil
	}
	return nil, errNoCredentials
}

type rowIterator struct {
	client  *bigquery.Client
	next    []any
	nexterr error
	schema  drivers.Schema
	bqIter  *bigquery.RowIterator
}

var _ drivers.RowIterator = &rowIterator{}

func (r *rowIterator) ResultSchema(ctx context.Context) (drivers.Schema, error) {
	if r.schema != nil {
		return r.schema, nil
	}

	// schema is only available after first next call
	r.next, r.nexterr = r.Next(ctx)
	if r.nexterr != nil {
		return nil, r.nexterr
	}

	r.schema = make([]drivers.Field, len(r.bqIter.Schema))
	for i, s := range r.bqIter.Schema {
		dbt, err := bqToDuckDB(string(s.Type))
		if err != nil {
			return nil, err
		}

		r.schema[i] = drivers.Field{Name: s.Name, Type: dbt}
	}
	return r.schema, nil
}

func (r *rowIterator) Next(ctx context.Context) ([]any, error) {
	if r.next != nil || r.nexterr != nil {
		next, err := r.next, r.nexterr
		r.next = nil
		r.nexterr = nil
		return next, err
	}

	var row row = make([]any, 0)
	if err := r.bqIter.Next(&row); err != nil {
		return nil, err
	}

	return row, nil
}

func (r *rowIterator) Close() error {
	return r.client.Close()
}

type row []any

var _ bigquery.ValueLoader = &row{}

func (r *row) Load(v []bigquery.Value, s bigquery.Schema) error {
	m := make([]any, len(v))
	for i := 0; i < len(v); i++ {
		if s[i].Type == bigquery.RecordFieldType {
			return fmt.Errorf("repeated or nested data is not supported")
		}

		m[i] = convert(v[i])
	}
	*r = m
	return nil
}

func bqToDuckDB(dbt string) (string, error) {
	switch dbt {
	case "STRING":
		return "VARCHAR", nil
	case "JSON":
		return "VARCHAR", nil
	case "INTERVAL":
		return "VARCHAR", nil
	case "GEOGRAPHY":
		return "VARCHAR", nil
	case "FLOAT":
		return "DOUBLE", nil
	// TODO :: NUMERIC and BIGNUMERIC are represented as *big.Rat type.
	// There is no support for these types in go-duckdb driver.
	// Users can cast these to duckdb types in model.
	case "NUMERIC": // TODO :: fix this to correct duckdb type
		return "VARCHAR", nil
	case "BIGNUMERIC": // TODO :: fix this to correct duckdb type
		return "VARCHAR", nil
	case "TIMESTAMP":
		return "TIMESTAMP", nil
	// TODO :: DATETIME, TIME, DATE doesn't have equivalent constructs in go and not supported in go-duckdb driver.
	// Users can cast these to duckdb types in model.
	case "DATETIME":
		return "VARCHAR", nil
	case "TIME":
		return "VARCHAR", nil
	case "DATE":
		return "VARCHAR", nil
	case "BOOLEAN":
		return "BOOLEAN", nil
	case "INTEGER":
		return "INTEGER", nil
	case "BYTES":
		return "BLOB", nil
	case "RECORD":
		return "", fmt.Errorf("record type not supported")
	default:
		// TODO :: may be just use VARCHAR ?
		return "", fmt.Errorf("type %s not supported", dbt)
	}
}

func convert(v any) any {
	if v == nil {
		return nil
	}
	// refer to documentation on bigquery.RowIterator.Next for the superset of all go types possible
	switch val := v.(type) {
	case civil.Date:
		return val.String()
	case civil.Time:
		return val.String()
	case civil.DateTime:
		return val.String()
	case *big.Rat:
		return val.String()
	default:
		return val
	}
}
