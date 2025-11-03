package druid

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

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
	DocsURL:     "https://docs.rilldata.com/build/connectors/olap/druid",
	// Important: Any edits to the below properties must be accompanied by changes to the client-side form validation schemas.
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
	// MaxOpenConns is the maximum number of open connections to the database. Set to 0 to use the default value or -1 for unlimited.
	MaxOpenConns int `mapstructure:"max_open_conns"`
	// SkipVersionCheck skips the version check.
	SkipVersionCheck bool `mapstructure:"skip_version_check"`
	// SkipQueryPriority indicates whether to skip passing query priority to Druid.
	SkipQueryPriority bool `mapstructure:"skip_query_priority"`
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

	maxOpenConns := conf.MaxOpenConns
	if maxOpenConns == 0 {
		maxOpenConns = 20 // default value
	}
	db.SetMaxOpenConns(maxOpenConns)

	err = otelsql.RegisterDBStatsMetrics(db, otelsql.WithAttributes(attribute.String("instance_id", instanceID)))
	if err != nil {
		return nil, fmt.Errorf("druid: failed to register db stats metrics: %w", err)
	}

	dbx := sqlx.NewDb(db, "druid")
	err = dbx.Ping()
	if err != nil {
		return nil, fmt.Errorf("druid: %w", err)
	}

	if !conf.SkipVersionCheck {
		err = d.checkVersion(dsn)
		if err != nil {
			return nil, fmt.Errorf("druid: %w", err)
		}
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

func (d driver) checkVersion(dsn string) error {
	parsedURL, err := url.Parse(dsn)
	if err != nil {
		return err
	}
	parsedURL.Path = "/status"
	statusURL := parsedURL.String()

	req, err := http.NewRequest(http.MethodGet, statusURL, http.NoBody)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("druid version check failed with status code: %d", resp.StatusCode)
	}

	var statusResponse struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&statusResponse); err != nil {
		return fmt.Errorf("failed to decode Druid status response: %w", err)
	}

	if statusResponse.Version != "" {
		majorVersion := strings.Split(statusResponse.Version, ".")[0]
		if ver, err := strconv.Atoi(majorVersion); err == nil {
			if ver < 28 {
				return fmt.Errorf("druid version %s is not supported, please use 28.0.0 or higher", statusResponse.Version)
			}
		} else {
			return fmt.Errorf("failed to parse Druid version: %w", err)
		}
	} else {
		return fmt.Errorf("druid version information not found in the response")
	}

	return nil
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

// Driver implements drivers.Handle.
func (c *connection) Driver() string {
	return "druid"
}

// Config used to open the Connection
func (c *connection) Config() map[string]any {
	m := make(map[string]any, 0)
	_ = mapstructure.Decode(c.config, &m)
	return m
}

// Close implements drivers.Handle.
func (c *connection) Close() error {
	return c.db.Close()
}

// Registry implements drivers.Handle.
func (c *connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// Catalog implements drivers.Handle.
func (c *connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// Repo implements drivers.Handle.
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

// OLAP implements drivers.Handle.
func (c *connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return c, true
}

// AsInformationSchema implements drivers.Handle.
func (c *connection) AsInformationSchema() (drivers.InformationSchema, bool) {
	return c, true
}

// Migrate implements drivers.Handle.
func (c *connection) Migrate(ctx context.Context) (err error) {
	return nil
}

// MigrationStatus implements drivers.Handle.
func (c *connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// AsObjectStore implements drivers.Handle.
func (c *connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsModelExecutor implements drivers.Handle.
func (c *connection) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, error) {
	return nil, drivers.ErrNotImplemented
}

// AsModelManager implements drivers.Handle.
func (c *connection) AsModelManager(instanceID string) (drivers.ModelManager, bool) {
	return nil, false
}

// AsFileStore implements drivers.Handle.
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
