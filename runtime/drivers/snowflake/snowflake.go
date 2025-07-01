package snowflake

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/snowflakedb/gosnowflake"
	"go.uber.org/zap"
)

func init() {
	drivers.Register("snowflake", driver{})
	drivers.RegisterAsConnector("snowflake", driver{})
}

var spec = drivers.Spec{
	DisplayName: "Snowflake",
	Description: "Connect to Snowflake.",
	DocsURL:     "https://docs.rilldata.com/reference/connectors/snowflake",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:    "dsn",
			Type:   drivers.StringPropertyType,
			Secret: true,
		},
	},
	// Important: Any edits to the below properties must be accompanied by changes to the client-side form validation schemas.
	SourceProperties: []*drivers.PropertySpec{
		{
			Key:         "sql",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "SQL",
			Description: "Query to extract data from Snowflake.",
			Placeholder: "select * from table",
		},
		{
			Key:         "dsn",
			Type:        drivers.StringPropertyType,
			DisplayName: "Snowflake Connection String",
			Required:    false,
			DocsURL:     "https://docs.rilldata.com/reference/connectors/snowflake",
			Placeholder: "<username>@<account_identifier>/<database>/<schema>?warehouse=<warehouse>&role=<role>&authenticator=SNOWFLAKE_JWT&privateKey=<privateKey_base64_url_encoded>",
			Hint:        "Can be configured here or by setting the 'connector.snowflake.dsn' environment variable (using '.env' or '--env')",
			Secret:      true,
		},
		{
			Key:         "name",
			Type:        drivers.StringPropertyType,
			DisplayName: "Source name",
			Description: "The name of the source",
			Placeholder: "my_new_source",
			Required:    true,
		},
	},
	ImplementsWarehouse: true,
}

type driver struct{}

type configProperties struct {
	DSN                string         `mapstructure:"dsn"`
	Account            string         `mapstructure:"account"`
	User               string         `mapstructure:"user"`
	Password           string         `mapstructure:"password"`
	Database           string         `mapstructure:"database"`
	Schema             string         `mapstructure:"schema"`
	Warehouse          string         `mapstructure:"warehouse"`
	Role               string         `mapstructure:"role"`
	Authenticator      string         `mapstructure:"authenticator"`
	PrivateKey         string         `mapstructure:"privateKey"`
	ParallelFetchLimit int            `mapstructure:"parallel_fetch_limit"`
	Extras             map[string]any `mapstructure:",remain"`
}

func (cp configProperties) resolveDSN() (string, error) {
	if cp.DSN != "" {
		return cp.DSN, nil
	}

	if cp.Account == "" || cp.User == "" || cp.Database == "" {
		return "", errors.New("missing required fields: account, user, or database")
	}

	if cp.Password == "" && cp.PrivateKey == "" {
		return "", errors.New("either password or privateKey must be provided")
	}

	cfg := &gosnowflake.Config{
		Account:   cp.Account,
		User:      cp.User,
		Password:  cp.Password,
		Database:  cp.Database,
		Schema:    cp.Schema,
		Warehouse: cp.Warehouse,
		Role:      cp.Role,
		Params:    map[string]*string{},
	}

	if cp.PrivateKey != "" {
		privateKey, err := parseRSAPrivateKey(cp.PrivateKey)
		if err != nil {
			return "", err
		}
		cfg.PrivateKey = privateKey
		cfg.Authenticator = gosnowflake.AuthTypeJwt
	} else if cp.Authenticator != "" {
		cfg.Params["authenticator"] = &cp.Authenticator
	}

	// Apply extra params
	for k, v := range cp.Extras {
		switch val := v.(type) {
		case string:
			cfg.Params[k] = &val
		default:
			strVal := fmt.Sprintf("%v", val)
			cfg.Params[k] = &strVal
		}
	}

	return gosnowflake.DSN(cfg)
}

func (d driver) Open(instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("snowflake driver can't be shared")
	}

	conf := &configProperties{}
	err := mapstructure.WeakDecode(config, conf)
	if err != nil {
		return nil, err
	}

	return &connection{
		configProperties: conf,
		storage:          st,
		logger:           logger,
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
	configProperties *configProperties
	storage          *storage.Client
	logger           *zap.Logger
}

// Ping implements drivers.Handle.
func (c *connection) Ping(ctx context.Context) error {
	dsn, err := c.configProperties.resolveDSN()
	if err != nil {
		return err
	}

	db, err := sqlx.Open("snowflake", dsn)
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}
	defer db.Close()
	return db.PingContext(ctx)
}

// Migrate implements drivers.Connection.
func (c *connection) Migrate(ctx context.Context) (err error) {
	return nil
}

// MigrationStatus implements drivers.Handle.
func (c *connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// Driver implements drivers.Connection.
func (c *connection) Driver() string {
	return "snowflake"
}

// Config implements drivers.Connection.
func (c *connection) Config() map[string]any {
	m := make(map[string]any, 0)
	_ = mapstructure.Decode(c.configProperties, &m)
	return m
}

// Close implements drivers.Connection.
func (c *connection) Close() error {
	return nil
}

// AsRegistry implements drivers.Connection.
func (c *connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// AsCatalogStore implements drivers.Connection.
func (c *connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// AsRepoStore implements drivers.Connection.
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

// AsOLAP implements drivers.Connection.
func (c *connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return nil, false
}

// AsObjectStore implements drivers.Connection.
func (c *connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsModelExecutor implements drivers.Handle.
func (c *connection) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, bool) {
	if opts.InputHandle == c {
		if store, ok := opts.OutputHandle.AsObjectStore(); ok {
			return &selfToObjectStoreExecutor{
				c:     c,
				store: store,
			}, true
		}
	}
	return nil, false
}

// AsModelManager implements drivers.Handle.
func (c *connection) AsModelManager(instanceID string) (drivers.ModelManager, bool) {
	return nil, false
}

// AsFileStore implements drivers.Connection.
func (c *connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsWarehouse implements drivers.Handle.
func (c *connection) AsWarehouse() (drivers.Warehouse, bool) {
	return c, true
}

// AsNotifier implements drivers.Connection.
func (c *connection) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}

// parseRSAPrivateKey parses a private key string
func parseRSAPrivateKey(keyStr string) (*rsa.PrivateKey, error) {
	var keyBytes []byte

	// 1. Try standard Base64 decoding (common in env vars or configs)
	if decoded, err := base64.StdEncoding.DecodeString(keyStr); err == nil {
		if block, _ := pem.Decode(decoded); block != nil {
			keyBytes = block.Bytes // decoded base64 was PEM
		} else {
			keyBytes = decoded // decoded base64 was raw DER
		}
	} else if decoded, err := base64.URLEncoding.DecodeString(keyStr); err == nil {
		// 2. Try URL-safe Base64 (used by Snowflake SDK)
		if block, _ := pem.Decode(decoded); block != nil {
			keyBytes = block.Bytes
		} else {
			keyBytes = decoded
		}
	} else {
		// 3. Fallback: maybe it's a raw PEM string (with BEGIN/END)
		if block, _ := pem.Decode([]byte(keyStr)); block != nil {
			keyBytes = block.Bytes
		} else {
			return nil, errors.New("invalid private key: not valid base64 or PEM")
		}
	}

	// Try PKCS#8
	if key, err := x509.ParsePKCS8PrivateKey(keyBytes); err == nil {
		if rsaKey, ok := key.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
		return nil, errors.New("unsupported key type: not RSA (PKCS#8)")
	}

	return nil, errors.New("failed to parse RSA private key not PKCS#8)")
}
