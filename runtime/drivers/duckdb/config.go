package duckdb

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

const (
	poolSizeMin int = 2
	poolSizeMax int = 5
)

// config represents the DuckDB driver config
type config struct {
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
	// BootQueries is SQL to execute when initializing a new connection. It runs before any extensions are loaded or default settings are set.
	BootQueries string `mapstructure:"boot_queries"`
	// InitSQL is SQL to execute when initializing a new connection. It runs after extensions are loaded and and default settings are set.
	InitSQL string `mapstructure:"init_sql"`
	// LogQueries controls whether to log the raw SQL passed to OLAP.Execute. (Internal queries will not be logged.)
	LogQueries bool `mapstructure:"log_queries"`
}

func newConfig(cfgMap map[string]any) (*config, error) {
	cfg := &config{
		ReadWriteRatio: 0.5,
	}
	err := mapstructure.WeakDecode(cfgMap, cfg)
	if err != nil {
		return nil, fmt.Errorf("could not decode config: %w", err)
	}

	// Set pool size
	poolSize := cfg.PoolSize
	if poolSize == 0 && cfg.CPU != 0 {
		poolSize = min(poolSizeMax, cfg.CPU) // Only enforce max pool size when inferred from CPU
	}
	poolSize = max(poolSizeMin, poolSize) // Always enforce min pool size
	cfg.PoolSize = poolSize
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
