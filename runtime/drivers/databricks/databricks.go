package databricks

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"

	dbsqllog "github.com/databricks/databricks-sql-go/logger"
	"github.com/jmoiron/sqlx"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"

	// Load Databricks SQL driver
	_ "github.com/databricks/databricks-sql-go"
)

func init() {
	drivers.Register("databricks", driver{})
	drivers.RegisterAsConnector("databricks", driver{})

	_ = dbsqllog.SetLogLevel("disabled")
}

var spec = drivers.Spec{
	DisplayName: "Databricks",
	Description: "Connect to Databricks.",
	DocsURL:     "https://docs.rilldata.com/developers/build/connectors/data-source/databricks",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:         "dsn",
			Type:        drivers.StringPropertyType,
			DisplayName: "Databricks Connection String",
			Placeholder: "token:<token>@<host>:443/<http_path>?catalog=<catalog>&schema=<schema>",
			Hint:        "Can be configured here or by setting the 'connector.databricks.dsn' environment variable (using '.env' or '--env').",
			Secret:      true,
		},
		{
			Key:         "host",
			Type:        drivers.StringPropertyType,
			DisplayName: "Host",
			Required:    true,
			Placeholder: "dbc-xxxxxxxx-xxxx.cloud.databricks.com",
			Hint:        "Databricks SQL warehouse hostname",
		},
		{
			Key:         "http_path",
			Type:        drivers.StringPropertyType,
			DisplayName: "HTTP Path",
			Required:    true,
			Placeholder: "/sql/1.0/warehouses/xxxxxxxxxxxxxxxx",
			Hint:        "HTTP path for the SQL warehouse",
		},
		{
			Key:         "token",
			Type:        drivers.StringPropertyType,
			DisplayName: "Access Token",
			Required:    true,
			Placeholder: "dapi...",
			Hint:        "Databricks personal access token",
			Secret:      true,
		},
		{
			Key:         "catalog",
			Type:        drivers.StringPropertyType,
			DisplayName: "Catalog",
			Placeholder: "main",
			Hint:        "Unity Catalog name (optional; defaults to the workspace default)",
		},
		{
			Key:         "schema",
			Type:        drivers.StringPropertyType,
			DisplayName: "Schema",
			Placeholder: "default",
			Hint:        "Schema within the catalog (optional; defaults to the workspace default)",
		},
		{
			Key:         "use_kernel",
			Type:        drivers.BooleanPropertyType,
			DisplayName: "Use SEA (Statement Execution API)",
			Hint: "Force the SEA backend instead of Thrift. Lakehouse//RT warehouses are " +
				"auto-detected and switched to SEA automatically, so this is usually unnecessary. " +
				"Requires a Rill build with the databricks_kernel backend (CGO); defaults to false.",
		},
	},
	ImplementsOLAP:      true,
	ImplementsWarehouse: true,
}

type driver struct{}

type configProperties struct {
	DSN        string `mapstructure:"dsn"`
	Host       string `mapstructure:"host"`
	HTTPPath   string `mapstructure:"http_path"`
	Token      string `mapstructure:"token"`
	Catalog    string `mapstructure:"catalog"`
	Schema     string `mapstructure:"schema"`
	LogQueries bool   `mapstructure:"log_queries"`
	// UseKernel forces the driver's SEA backend. Lakehouse//RT is auto-detected and
	// switched to SEA even when this is false (see getDB), so it's an override.
	// Needs a build with the databricks_kernel backend; default false = Thrift.
	UseKernel bool `mapstructure:"use_kernel"`
}

func (c *configProperties) validate() error {
	var set []string
	if c.Host != "" {
		set = append(set, "host")
	}
	if c.HTTPPath != "" {
		set = append(set, "http_path")
	}
	if c.Token != "" {
		set = append(set, "token")
	}
	if c.Catalog != "" {
		set = append(set, "catalog")
	}
	if c.Schema != "" {
		set = append(set, "schema")
	}
	if c.DSN != "" && len(set) > 0 {
		return fmt.Errorf("databricks: only one of 'dsn' or [%s] can be set", strings.Join(set, ", "))
	}
	if c.DSN == "" && (c.Host == "" || c.HTTPPath == "" || c.Token == "") {
		return errors.New("databricks: either 'dsn' or 'host', 'http_path', and 'token' are required")
	}
	return nil
}

func (c *configProperties) resolveDSN() string {
	if c.DSN != "" {
		if c.UseKernel {
			return withUseKernel(c.DSN)
		}
		return c.DSN
	}
	params := url.Values{}
	params.Set("timezone", "UTC")
	// Opt in to the SEA backend (required for Lakehouse//RT); default is Thrift.
	if c.UseKernel {
		params.Set("useKernel", "true")
	}
	if c.Catalog != "" {
		params.Set("catalog", c.Catalog)
	}
	if c.Schema != "" {
		params.Set("schema", c.Schema)
	}
	// DSN format: https://token:<token>@<host>:443/<http_path>?catalog=<catalog>&schema=<schema>
	u := &url.URL{
		Scheme:   "https",
		User:     url.UserPassword("token", c.Token),
		Host:     c.Host + ":443",
		Path:     c.HTTPPath,
		RawQuery: params.Encode(),
	}
	return u.String()
}

// withUseKernel forces useKernel=true on a resolved DSN. It is only called when SEA is
// required (explicit use_kernel, or RT auto-detect), so any existing useKernel in the
// DSN — including useKernel=false — is dropped in favor of true. It splits at the first
// '?' (the query separator) and rewrites only the query part; it avoids url.Parse
// because the driver also accepts scheme-less DSNs (token:...@host), which url.Parse
// would misread (treating "token" as the scheme).
func withUseKernel(dsn string) string {
	base, query, hasQuery := strings.Cut(dsn, "?")
	if !hasQuery || query == "" {
		return base + "?useKernel=true"
	}
	parts := strings.Split(query, "&")
	kept := parts[:0]
	for _, p := range parts {
		if strings.HasPrefix(p, "useKernel=") {
			continue // drop any existing useKernel (true or false)
		}
		kept = append(kept, p)
	}
	kept = append(kept, "useKernel=true")
	return base + "?" + strings.Join(kept, "&")
}

// rtRequiresSEA reports whether err is a Lakehouse//RT warehouse rejecting the
// Thrift protocol (the driver surfaces the server's message verbatim).
func rtRequiresSEA(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not supported for Thrift protocol")
}

func (d driver) Open(_, instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("databricks driver can't be shared")
	}

	conf := &configProperties{}
	err := mapstructure.WeakDecode(config, conf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}
	if err := conf.validate(); err != nil {
		return nil, err
	}

	return &connection{
		config:  conf,
		storage: st,
		logger:  logger,
		dbMu:    semaphore.NewWeighted(1),
	}, nil
}

func (d driver) Spec() drivers.Spec {
	return spec
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, src map[string]any, logger *zap.Logger) (bool, error) {
	return false, nil
}

func (d driver) TertiarySourceConnectors(ctx context.Context, src map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, nil
}

type connection struct {
	config  *configProperties
	storage *storage.Client
	logger  *zap.Logger

	db    *sqlx.DB // lazily populated using getDB
	dbErr error
	dbMu  *semaphore.Weighted

	// resolvedDSN is the DSN to connect with, computed by effectiveDSN (auto-detects
	// Lakehouse//RT and upgrades to the SEA backend). Empty until a *definitive*
	// probe outcome is reached, so a transient probe failure is not cached.
	dsnMu       sync.Mutex
	resolvedDSN string
}

// Ping implements drivers.Handle.
func (c *connection) Ping(ctx context.Context) error {
	db, err := c.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to open databricks connection: %w", err)
	}
	return db.PingContext(ctx)
}

// Migrate implements drivers.Handle.
func (c *connection) Migrate(ctx context.Context) (err error) {
	return nil
}

// MigrationStatus implements drivers.Handle.
func (c *connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// Driver implements drivers.Handle.
func (c *connection) Driver() string {
	return "databricks"
}

// Config implements drivers.Handle.
func (c *connection) Config() map[string]any {
	var m map[string]any
	_ = mapstructure.WeakDecode(c.config, &m)
	return m
}

// Close implements drivers.Handle.
func (c *connection) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// AsRegistry implements drivers.Handle.
func (c *connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// AsCatalogStore implements drivers.Handle.
func (c *connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// AsRepoStore implements drivers.Handle.
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

// AsOLAP implements drivers.Handle.
func (c *connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return c, true
}

// AsInformationSchema implements drivers.Handle.
func (c *connection) AsInformationSchema() (drivers.InformationSchema, bool) {
	return c, true
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
func (c *connection) AsModelManager(instanceID string) (drivers.ModelManager, error) {
	return nil, drivers.ErrNotImplemented
}

// AsFileStore implements drivers.Handle.
func (c *connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsWarehouse implements drivers.Handle.
func (c *connection) AsWarehouse() (drivers.Warehouse, bool) {
	return c, true
}

// AsNotifier implements drivers.Handle.
func (c *connection) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}

func (c *connection) getDB(ctx context.Context) (*sqlx.DB, error) {
	err := c.dbMu.Acquire(ctx, 1)
	if err != nil {
		return nil, err
	}
	defer c.dbMu.Release(1)
	if c.db != nil || c.dbErr != nil {
		return c.db, c.dbErr
	}

	// Only build (and permanently cache) the shared pool once the backend is
	// *definitively* known. On a non-definitive result we surface the probe error
	// (which may be a real misconfiguration or a transient blip) without caching it,
	// so the caller can retry and we re-probe. c.db is thus never swapped or closed
	// here, so a pool already handed to concurrent callers is never closed underneath.
	dsn, definitive, probeErr := c.effectiveDSN(ctx)
	if !definitive {
		if probeErr != nil {
			return nil, probeErr
		}
		return nil, errors.New("databricks: could not determine warehouse protocol")
	}
	c.db, c.dbErr = sqlx.Open("databricks", dsn)
	if c.dbErr != nil {
		return nil, c.dbErr
	}
	return c.db, c.dbErr
}

// effectiveDSN resolves the DSN to connect with, auto-detecting Lakehouse//RT: if SEA
// wasn't explicitly requested (use_kernel) and the warehouse rejects Thrift, it upgrades
// the DSN to the SEA backend. It returns whether the result is definitive (safe to
// cache) and, when not, the probe error that prevented a decision — callers surface that
// error and re-resolve next time rather than lock onto the wrong backend or mask a real
// failure. It's the single shared RT-detection point for the OLAP (getDB) and warehouse
// ingest paths, so RT works without any configuration.
func (c *connection) effectiveDSN(ctx context.Context) (dsn string, definitive bool, probeErr error) {
	c.dsnMu.Lock()
	defer c.dsnMu.Unlock()
	if c.resolvedDSN != "" {
		return c.resolvedDSN, true, nil
	}

	base := c.config.resolveDSN()
	if c.config.UseKernel {
		c.resolvedDSN = base // already carries useKernel=true
		return base, true, nil
	}

	probe, err := sqlx.Open("databricks", base)
	if err != nil {
		return base, false, err // can't even open; don't cache
	}
	defer probe.Close() //nolint:errcheck // best-effort close of the probe connection

	// Only cache a *definitive* outcome. A clean ping means Thrift works; the RT
	// rejection means switch to SEA. Any other ping error is returned uncached (with
	// the error) — it may be transient (a blip, re-probed next call) or persistent (a
	// bad token / wrong host), and surfacing it verbatim avoids both locking onto the
	// wrong backend and masking a real misconfiguration behind auto-detection.
	switch perr := probe.PingContext(ctx); {
	case perr == nil:
		c.resolvedDSN = base
	case rtRequiresSEA(perr):
		c.logger.Info("databricks: warehouse rejected Thrift (Lakehouse//RT); switching to the SEA backend")
		c.resolvedDSN = withUseKernel(base)
	default:
		return base, false, perr // transient or persistent — surface it, don't cache
	}
	return c.resolvedDSN, true, nil
}
