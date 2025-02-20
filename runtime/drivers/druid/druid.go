package druid

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/XSAM/otelsql"
	"github.com/jmoiron/sqlx"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	// Load Druid database/sql driver
	_ "github.com/rilldata/rill/runtime/drivers/druid/druidsqldriver"
)

func init() {
	drivers.Register("druid", &driver{})
	drivers.RegisterAsConnector("druid", &driver{})
}

var spec = drivers.Spec{
	DisplayName: "Druid",
	Description: "Connect to Apache Druid.",
	DocsURL:     "https://docs.rilldata.com/reference/olap-engines/druid",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:         "dsn",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "Connection string",
			Placeholder: "https://example.com/druid/v2/sql/avatica-protobuf?authentication=BASIC&avaticaUser=username&avaticaPassword=password",
			Secret:      true,
			NoPrompt:    true,
		},
		{
			Key:         "host",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "Host",
			Description: "Hostname or IP address of the Druid server",
			Placeholder: "localhost",
		},
		{
			Key:         "port",
			Type:        drivers.NumberPropertyType,
			Required:    false,
			DisplayName: "Port",
			Description: "Port number of the Druid server",
			Placeholder: "8888",
		},
		{
			Key:         "username",
			Type:        drivers.StringPropertyType,
			Required:    false,
			DisplayName: "Username",
			Description: "Username to connect to the Druid server",
			Placeholder: "default",
		},
		{
			Key:         "password",
			Type:        drivers.StringPropertyType,
			Required:    false,
			DisplayName: "Password",
			Description: "Password to connect to the Druid server",
			Placeholder: "password",
			Secret:      true,
		},
		{
			Key:         "ssl",
			Type:        drivers.BooleanPropertyType,
			Required:    true,
			DisplayName: "SSL",
			Description: "Use SSL to connect to the Druid server",
		},
	},
	ImplementsOLAP: true,
}

type driver struct{}

var _ drivers.Driver = &driver{}

type configProperties struct {
	// DSN is the connection string. Set either DSN or properties below.
	DSN      string `mapstructure:"dsn"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	// SSL determines whether secured connection need to be established. To be set when setting individual fields.
	SSL bool `mapstructure:"ssl"`
	// LogQueries controls whether to log the raw SQL passed to OLAP.Execute.
	LogQueries bool `mapstructure:"log_queries"`
}

// Opens a connection to Apache Druid using HTTP API.
// Note that the Druid connection string must have the form "http://user:password@host:port/druid/v2/sql".
func (d driver) Open(instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("druid driver can't be shared")
	}

	conf := &configProperties{}
	err := mapstructure.WeakDecode(config, conf)
	if err != nil {
		return nil, err
	}

	dsn, err := dsnFromConfig(conf)
	if err != nil {
		return nil, err
	}

	db, err := otelsql.Open("druid", dsn)
	if err != nil {
		return nil, err
	}
	// very roughly approximating num queries required for a typical page load
	db.SetMaxOpenConns(20)

	err = otelsql.RegisterDBStatsMetrics(db, otelsql.WithAttributes(attribute.String("instance_id", instanceID)))
	if err != nil {
		return nil, fmt.Errorf("druid: failed to register db stats metrics: %w", err)
	}

	dbx := sqlx.NewDb(db, "druid")
	err = dbx.Ping()
	if err != nil {
		return nil, fmt.Errorf("druid: %w", err)
	}

	conn := &connection{
		db:     dbx,
		config: conf,
		logger: logger,
	}
	return conn, nil
}

func (d *driver) Spec() drivers.Spec {
	return spec
}

func (d *driver) HasAnonymousSourceAccess(ctx context.Context, src map[string]any, logger *zap.Logger) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (d *driver) TertiarySourceConnectors(ctx context.Context, src map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

type connection struct {
	db     *sqlx.DB
	config *configProperties
	logger *zap.Logger
}

// Ping implements drivers.Handle.
func (c *connection) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
}

// Driver implements drivers.Connection.
func (c *connection) Driver() string {
	return "druid"
}

// Config used to open the Connection
func (c *connection) Config() map[string]any {
	m := make(map[string]any, 0)
	_ = mapstructure.Decode(c.config, &m)
	return m
}

// Close implements drivers.Connection.
func (c *connection) Close() error {
	return c.db.Close()
}

// Registry implements drivers.Connection.
func (c *connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// Catalog implements drivers.Connection.
func (c *connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// Repo implements drivers.Connection.
func (c *connection) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

// AsAdmin implements drivers.Handle.
func (c *connection) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return nil, false
}

// AsAI implements drivers.Handle.
func (c *connection) AsAI(instanceID string) (drivers.AIService, bool) {
	return nil, false
}

// OLAP implements drivers.Connection.
func (c *connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return c, true
}

// Migrate implements drivers.Connection.
func (c *connection) Migrate(ctx context.Context) (err error) {
	return nil
}

// MigrationStatus implements drivers.Connection.
func (c *connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// AsObjectStore implements drivers.Connection.
func (c *connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsModelExecutor implements drivers.Handle.
func (c *connection) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, bool) {
	return nil, false
}

// AsModelManager implements drivers.Handle.
func (c *connection) AsModelManager(instanceID string) (drivers.ModelManager, bool) {
	return nil, false
}

// AsTransporter implements drivers.Connection.
func (c *connection) AsTransporter(from, to drivers.Handle) (drivers.Transporter, bool) {
	return nil, false
}

// AsFileStore implements drivers.Connection.
func (c *connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsWarehouse implements drivers.Handle.
func (c *connection) AsWarehouse() (drivers.Warehouse, bool) {
	return nil, false
}

// AsNotifier implements drivers.Connection.
func (c *connection) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}

func (c *connection) AcquireLongRunning(ctx context.Context) (func(), error) {
	return func() {}, nil
}

func dsnFromConfig(conf *configProperties) (string, error) {
	var dsn string
	var err error
	if conf.DSN != "" {
		dsn, err = correctURL(conf.DSN)
		if err != nil {
			return "", err
		}
	} else if conf.Host != "" {
		var dsnURL url.URL
		dsnURL.Host = conf.Host
		// set port
		if conf.Port != 0 {
			dsnURL.Host = fmt.Sprintf("%v:%v", conf.Host, conf.Port)
		}

		// set scheme
		if conf.SSL {
			dsnURL.Scheme = "https"
		} else {
			dsnURL.Scheme = "http"
		}

		// set path
		dsnURL.Path = "druid/v2/sql"

		// set username and password
		if conf.Password != "" {
			dsnURL.User = url.UserPassword(conf.Username, conf.Password)
		} else if conf.Username != "" {
			dsnURL.User = url.User(conf.Username)
		}

		dsn = dsnURL.String()
	} else {
		return "", fmt.Errorf("druid connection parameters not set. Set `dsn` or individual properties")
	}
	return dsn, nil
}

func correctURL(dsn string) (string, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		if strings.Contains(err.Error(), dsn) { // avoid returning the actual DSN with the password which will be logged
			return "", fmt.Errorf("%s", strings.ReplaceAll(err.Error(), dsn, "<masked>"))
		}
		return "", err
	}

	if strings.Contains(u.Path, "avatica-protobuf") {
		avaticaUser := url.QueryEscape(u.Query().Get("avaticaUser"))
		avaticaPassword := url.QueryEscape(u.Query().Get("avaticaPassword"))

		if avaticaUser != "" {
			dsn = u.Scheme + "://" + avaticaUser + ":" + avaticaPassword + "@" + u.Host + "/druid/v2/sql"
		} else {
			dsn = u.Scheme + "://" + u.Host + "/druid/v2/sql"
		}
	}
	return dsn, nil
}
