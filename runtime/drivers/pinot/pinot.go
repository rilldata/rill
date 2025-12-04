package pinot

import (
	"context"
	"fmt"
	"maps"
	"net/url"
	"strings"

	"github.com/XSAM/otelsql"
	"github.com/jmoiron/sqlx"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/pinot/sqldriver"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

func init() {
	drivers.Register("pinot", driver{})
	drivers.RegisterAsConnector("pinot", driver{})
}

var spec = drivers.Spec{
	DisplayName: "Pinot",
	Description: "Connect to Apache Pinot.",
	DocsURL:     "https://docs.rilldata.com/build/connectors/olap/pinot",
	// Important: Any edits to the below properties must be accompanied by changes to the client-side form validation schemas.
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:         "dsn",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "Connection string",
			Placeholder: "http(s)://username:password@localhost:8000?controller=localhost:9000",
			Secret:      true,
			NoPrompt:    true,
		},
		{
			Key:         "broker_host",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "Broker Host",
			Description: "Hostname or IP address of the Pinot broker server",
			Placeholder: "localhost",
		},
		{
			Key:         "broker_port",
			Type:        drivers.NumberPropertyType,
			Required:    false,
			DisplayName: "Broker Port",
			Description: "Port number of the broker Pinot broker server",
			Placeholder: "8000",
		},
		{
			Key:         "controller_host",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "Controller Host",
			Description: "Hostname or IP address of the Pinot controller server",
			Placeholder: "localhost",
		},
		{
			Key:         "controller_port",
			Type:        drivers.NumberPropertyType,
			Required:    false,
			DisplayName: "Controller Port",
			Description: "Port number of the Pinot controller server",
			Placeholder: "9000",
		},
		{
			Key:         "username",
			Type:        drivers.StringPropertyType,
			Required:    false,
			DisplayName: "Username",
			Description: "Username to connect to the Pinot server",
			Placeholder: "default",
		},
		{
			Key:         "password",
			Type:        drivers.StringPropertyType,
			Required:    false,
			DisplayName: "Password",
			Description: "Password to connect to the Pinot server",
			Placeholder: "password",
			Secret:      true,
		},
		{
			Key:         "ssl",
			Type:        drivers.BooleanPropertyType,
			Required:    true,
			DisplayName: "SSL",
			Description: "Use SSL to connect to the Pinot server",
		},
	},
	ImplementsOLAP: true,
}

type driver struct{}

type configProperties struct {
	// DSN is the connection string. Set either DSN or properties below.
	DSN            string `mapstructure:"dsn"`
	Username       string `mapstructure:"username"`
	Password       string `mapstructure:"password"`
	BrokerHost     string `mapstructure:"broker_host"`
	BrokerPort     int    `mapstructure:"broker_port"`
	ControllerHost string `mapstructure:"controller_host"`
	ControllerPort int    `mapstructure:"controller_port"`
	// SSL determines whether secured connection need to be established. To be set when setting individual fields.
	SSL bool `mapstructure:"ssl"`
	// LogQueries controls whether to log the raw SQL passed to OLAP.Execute.
	LogQueries bool `mapstructure:"log_queries"`
	// MaxOpenConns is the maximum number of open connections to the database. Set to 0 to use the default value or -1 for unlimited.
	MaxOpenConns int `mapstructure:"max_open_conns"`
	// TimeoutMS is the timeout in milliseconds for queries. Set to 0 to use the cluster default.
	TimeoutMS int64 `mapstructure:"timeout_ms"`
}

func (c *configProperties) validate() error {
	var set []string
	if c.BrokerHost != "" {
		set = append(set, "broker_host")
	}
	if c.ControllerHost != "" {
		set = append(set, "controller_host")
	}
	if c.DSN != "" && len(set) > 0 {
		return fmt.Errorf("pinot: Only one of 'dsn' or [%s] can be set", strings.Join(set, ", "))
	}
	return nil
}

// Open a connection to Apache Pinot using HTTP API.
func (d driver) Open(instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, fmt.Errorf("pinot driver can't be shared")
	}

	conf := &configProperties{}
	err := mapstructure.WeakDecode(config, conf)
	if err != nil {
		return nil, err
	}

	if err := conf.validate(); err != nil {
		return nil, err
	}

	var dsn string
	if conf.DSN != "" {
		dsn = conf.DSN
	} else if conf.ControllerHost != "" && conf.BrokerHost != "" {
		var controllerURL url.URL
		if conf.ControllerPort == 0 {
			controllerURL.Host = conf.ControllerHost
		} else {
			controllerURL.Host = fmt.Sprintf("%v:%v", conf.ControllerHost, conf.ControllerPort)
		}
		if conf.SSL {
			controllerURL.Scheme = "https"
		} else {
			controllerURL.Scheme = "http"
		}

		var dsnURL url.URL
		dsnURL.Host = conf.BrokerHost
		// set port
		if conf.BrokerPort != 0 {
			dsnURL.Host = fmt.Sprintf("%v:%v", conf.BrokerHost, conf.BrokerPort)
		}

		// set scheme
		if conf.SSL {
			dsnURL.Scheme = "https"
		} else {
			dsnURL.Scheme = "http"
		}

		// set username and password
		if conf.Password != "" {
			dsnURL.User = url.UserPassword(conf.Username, conf.Password)
		} else if conf.Username != "" {
			dsnURL.User = url.User(conf.Username)
		}

		dsnURL.RawQuery = "controller=" + controllerURL.String()
		dsn = dsnURL.String()
	} else {
		return nil, fmt.Errorf("pinot connection parameters not set. Set `dsn` or individual properties")
	}

	db, err := otelsql.Open("pinot", dsn)
	if err != nil {
		return nil, err
	}

	maxOpenConns := conf.MaxOpenConns
	if maxOpenConns == 0 {
		maxOpenConns = 20 // default value
	}
	db.SetMaxOpenConns(maxOpenConns)

	err = otelsql.RegisterDBStatsMetrics(db, otelsql.WithAttributes(attribute.String("instance_id", instanceID)))
	if err != nil {
		return nil, fmt.Errorf("pinot: failed to register db stats metrics: %w", err)
	}

	dbx := sqlx.NewDb(db, "pinot")
	err = dbx.Ping()
	if err != nil {
		return nil, fmt.Errorf("pinot: %w", err)
	}

	broker, controller, headers, err := sqldriver.ParseDSN(dsn)
	if err != nil {
		return nil, err
	}

	conn := &connection{
		db:         dbx,
		config:     config,
		queryURL:   broker,
		schemaURL:  controller,
		headers:    headers,
		logQueries: conf.LogQueries,
		timeoutMS:  conf.TimeoutMS,
		logger:     logger,
	}
	return conn, nil
}

func (d driver) Spec() drivers.Spec {
	return spec
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, src map[string]any, logger *zap.Logger) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (d driver) TertiarySourceConnectors(ctx context.Context, src map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

type connection struct {
	db         *sqlx.DB
	config     map[string]any
	queryURL   string
	schemaURL  string
	headers    map[string]string
	logQueries bool
	timeoutMS  int64 // timeout in milliseconds for queries, 0 means use cluster default
	logger     *zap.Logger
}

// Ping implements drivers.Handle.
func (c *connection) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
}

// Driver implements drivers.Handle.
func (c *connection) Driver() string {
	return "pinot"
}

// Config used to open the Connection
func (c *connection) Config() map[string]any {
	return maps.Clone(c.config)
}

// MayBeScaledToZero implements drivers.Handle.
func (c *connection) MayBeScaledToZero(ctx context.Context) bool {
	return false
}

// Close implements drivers.Handle.
func (c *connection) Close() error {
	return c.db.Close()
}

func (c *connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

func (c *connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

func (c *connection) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

func (c *connection) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return nil, false
}

func (c *connection) AsAI(instanceID string) (drivers.AIService, bool) {
	return nil, false
}

func (c *connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return c, true
}

// AsInformationSchema implements drivers.Handle.
func (c *connection) AsInformationSchema() (drivers.InformationSchema, bool) {
	return c, true
}

func (c *connection) Migrate(ctx context.Context) (err error) {
	return nil
}

func (c *connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

func (c *connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

func (c *connection) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, error) {
	return nil, drivers.ErrNotImplemented
}

// AsModelManager implements drivers.Handle.
func (c *connection) AsModelManager(instanceID string) (drivers.ModelManager, bool) {
	return nil, false
}

func (c *connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsWarehouse implements drivers.Handle.
func (c *connection) AsWarehouse() (drivers.Warehouse, bool) {
	return nil, false
}

// AsNotifier implements drivers.Handle.
func (c *connection) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}
