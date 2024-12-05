package duckdb

import (
	"fmt"
	"strconv"

	"github.com/mitchellh/mapstructure"
)

const (
	poolSizeMin int = 2
	poolSizeMax int = 5
)

// config represents the DuckDB driver config
type config struct {
	// DataDir is the path to directory where duckdb files will be created.
	DataDir string `mapstructure:"data_dir"`
	// PoolSize is the number of concurrent connections and queries allowed
	PoolSize int `mapstructure:"pool_size"`
	// AllowHostAccess denotes whether to limit access to the local environment and file system
	AllowHostAccess bool `mapstructure:"allow_host_access"`
	// CPU cores available for the read DB. If no CPUWrite is set then this is split evenly between read and write.
	CPU int `mapstructure:"cpu"`
	// MemoryLimitGB is the amount of memory available for the read DB. If no MemoryLimitGBWrite is set then this is split evenly between read and write.
	MemoryLimitGB int `mapstructure:"memory_limit_gb"`
	// CPUWrite is CPU available for the DB when writing data.
	CPUWrite int `mapstructure:"cpu_write"`
	// MemoryLimitGBWrite is the amount of memory available for the DB when writing data.
	MemoryLimitGBWrite int `mapstructure:"memory_limit_gb_write"`
	// BootQueries is SQL to execute when initializing a new connection. It runs before any extensions are loaded or default settings are set.
	BootQueries string `mapstructure:"boot_queries"`
	// InitSQL is SQL to execute when initializing a new connection. It runs after extensions are loaded and and default settings are set.
	InitSQL string `mapstructure:"init_sql"`
	// LogQueries controls whether to log the raw SQL passed to OLAP.Execute. (Internal queries will not be logged.)
	LogQueries bool `mapstructure:"log_queries"`
}

func newConfig(cfgMap map[string]any, dataDir string) (*config, error) {
	cfg := &config{
		DataDir: dataDir,
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
	if c.MemoryLimitGB > 0 {
		readSettings["max_memory"] = fmt.Sprintf("%dGB", c.MemoryLimitGB)
	}
	if c.CPU > 0 {
		readSettings["threads"] = strconv.Itoa(c.CPU)
	}
	return readSettings
}

func (c *config) writeSettings() map[string]string {
	writeSettings := make(map[string]string)
	if c.MemoryLimitGBWrite > 0 {
		writeSettings["max_memory"] = fmt.Sprintf("%dGB", c.MemoryLimitGBWrite)
	}
	if c.CPUWrite > 0 {
		writeSettings["threads"] = strconv.Itoa(c.CPUWrite)
	}
	// useful for motherduck but safe to pass at initial connect
	writeSettings["custom_user_agent"] = "rill"
	return writeSettings
}
