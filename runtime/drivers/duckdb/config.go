package duckdb

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
)

const (
	poolSizeMin int = 2
	poolSizeMax int = 5

	modeReadOnly  = "read"
	modeReadWrite = "readwrite"
)

// config represents the DuckDB driver config
type config struct {
	// Managed is set internally if the connector has `managed: true`.
	// This indicates to use an embedded DuckDB, and cannot be combined with the Path and Attach settings.
	Managed bool `mapstructure:"managed"`
	// PoolSize is the number of concurrent connections and queries allowed
	PoolSize int `mapstructure:"pool_size"`
	// AllowHostAccess denotes whether to limit access to the local environment and file system
	AllowHostAccess bool `mapstructure:"allow_host_access"`
	// CPU cores available for the DB. If no ratio is set then this is split evenly between read and write.
	CPU int `mapstructure:"cpu"`
	// MemoryLimitGB is the amount of memory available for the DB. If no ratio is set then this is split evenly between read and write.
	MemoryLimitGB int `mapstructure:"memory_limit_gb"`
	// ReadWriteRatio is the ratio of resources to allocate to the read DB. If set, CPU and MemoryLimitGB are distributed based on this ratio.
	ReadWriteRatio float64 `mapstructure:"read_write_ratio"`
	// BootQueries is deprecated. Use InitSQL instead. Retained for backward compatibility.
	BootQueries string `mapstructure:"boot_queries"`
	// InitSQL is the SQL executed during database initialization.
	InitSQL string `mapstructure:"init_sql"`
	// ConnInitSQL is the SQL executed when a new connection is initialized.
	ConnInitSQL string `mapstructure:"conn_init_sql"`
	// LogQueries controls whether to log the raw SQL passed to OLAP.Execute. (Internal queries will not be logged.)
	LogQueries bool `mapstructure:"log_queries"`
	// CreateSecretsFromConnectors is list of connector names to create temporary secrets for before executing models.
	// The secrets are not created for read queries.
	CreateSecretsFromConnectors []string `mapstructure:"create_secrets_from_connectors"`
	// Mode specifies the mode in which to open the database.
	Mode string `mapstructure:"mode"`
	// CanScaleToZero indicates if the underlying duckdb service may scale to zero when idle.
	// When set to true, we try to avoid too frequent non-user queries to the database (such as alert checks and fetching metrics).
	CanScaleToZero bool `mapstructure:"can_scale_to_zero"`
	// Path switches the implementation to use a generic rduckdb implementation backed by the db used in the Path
	Path string `mapstructure:"path"`
	// Attach allows user to pass a full ATTACH statement to attach a DuckDB database.
	// Example YAML syntax : attach: "'ducklake:metadata.ducklake' AS my_ducklake(DATA_PATH 'datafiles1')"
	Attach string `mapstructure:"attach"`
	// Token is the authentication token used for MotherDuck.
	Token string `mapstructure:"token"`
	// DatabaseName is the name of the attached DuckDB database specified in the Path.
	// This is usually not required but can be set if our auto detection of name fails.
	DatabaseName string `mapstructure:"database_name"`
	// SchemaName can be set to switch the default schema used by the DuckDB database.
	// Only applicable for the generic rduckdb implementation.
	SchemaName string `mapstructure:"schema_name"`
	// EnableBackups enables periodic backups of the DuckDB database to object storage.
	// It only takes effect if the runtime's storage config includes an object storage bucket.
	// This is an internal property that should not be documented or set by users.
	EnableBackups bool `mapstructure:"enable_backups"`
}

func newConfig(cfgMap map[string]any) (*config, error) {
	cfg := &config{
		ReadWriteRatio: 0.5,
	}
	err := mapstructure.WeakDecode(cfgMap, cfg)
	if err != nil {
		return nil, fmt.Errorf("could not decode config: %w", err)
	}

	// Validate mode if specified
	if cfg.Mode != "" && cfg.Mode != modeReadOnly && cfg.Mode != modeReadWrite {
		return nil, fmt.Errorf("invalid mode '%s': must be 'read' or 'readwrite'", cfg.Mode)
	}

	// Previously we did not require `managed: true` to use embedded DuckDB.
	// For backward compatibility, we default to managed if no external config is provided.
	hasExternalConfig := cfg.Path != "" || cfg.Attach != ""
	if !hasExternalConfig {
		cfg.Managed = true
	}

	// Validate that managed is not combined with external config.
	if cfg.Managed && hasExternalConfig {
		return nil, fmt.Errorf("'managed: true' cannot be combined with 'path' or 'attach' fields")
	}

	// Set the mode for the connection
	if cfg.Mode == "" {
		// The default mode depends on the connection type:
		// - For managed/embedded DuckDB, default to "readwrite" to maintain compatibility
		// - For external connections (Path/Attach), default to "read"
		if cfg.Managed {
			cfg.Mode = modeReadWrite
		} else {
			cfg.Mode = modeReadOnly
		}
	}

	// Set pool size
	poolSize := cfg.PoolSize
	if poolSize == 0 && cfg.CPU != 0 {
		poolSize = min(poolSizeMax, cfg.CPU) // Only enforce max pool size when inferred from CPU
	}
	poolSize = max(poolSizeMin, poolSize) // Always enforce min pool size
	cfg.PoolSize = poolSize

	// set can_scale_to_zero for motherduck by default
	if _, ok := cfgMap["can_scale_to_zero"]; !ok && cfg.isMotherduck() {
		cfg.CanScaleToZero = true
	}

	return cfg, nil
}

func (c *config) readSettings() map[string]string {
	readSettings := make(map[string]string)
	return readSettings
}

func (c *config) writeSettings() map[string]string {
	writeSettings := make(map[string]string)
	// useful for motherduck but safe to pass at initial connect
	writeSettings["custom_user_agent"] = "rill"
	return writeSettings
}

// isMotherduck returns true if the Path or Attach config options reference a Motherduck database.
func (c *config) isMotherduck() bool {
	return strings.HasPrefix(c.Path, "md:") || strings.HasPrefix(c.Attach, "'md:")
}
