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
	// DataDir is the path to directory where duckdb file named `main.db` will be created. In case of external table storage all the files will also be present in DataDir's subdirectories.
	// If path is set then DataDir is ignored.
	DataDir string `mapstructure:"data_dir"`
	// PoolSize is the number of concurrent connections and queries allowed
	PoolSize int `mapstructure:"pool_size"`
	// AllowHostAccess denotes whether to limit access to the local environment and file system
	AllowHostAccess bool `mapstructure:"allow_host_access"`
	// CPU cores available for the read DB. If no CPUWrite is set and external_table_storage is enabled then this is split evenly between read and write.
	CPU int `mapstructure:"cpu"`
	// MemoryLimitGB is the amount of memory available for the read DB. If no MemoryLimitGBWrite is set and external_table_storage is enabled then this is split evenly between read and write.
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

	ReadSettings  map[string]string `mapstructure:"-"`
	WriteSettings map[string]string `mapstructure:"-"`
}

func newConfig(cfgMap map[string]any) (*config, error) {
	cfg := &config{}
	err := mapstructure.WeakDecode(cfgMap, cfg)
	if err != nil {
		return nil, fmt.Errorf("could not decode config: %w", err)
	}

	// Set memory limit
	cfg.ReadSettings = make(map[string]string)
	cfg.WriteSettings = make(map[string]string)
	if cfg.MemoryLimitGB > 0 {
		cfg.ReadSettings["max_memory"] = fmt.Sprintf("%dGB", cfg.MemoryLimitGB)
	}
	if cfg.MemoryLimitGBWrite > 0 {
		cfg.WriteSettings["max_memory"] = fmt.Sprintf("%dGB", cfg.MemoryLimitGB)
	}

	// Set threads limit
	var threads int
	if cfg.CPU > 0 {
		cfg.ReadSettings["threads"] = strconv.Itoa(cfg.CPU)
	}
	if cfg.CPUWrite > 0 {
		cfg.WriteSettings["threads"] = strconv.Itoa(cfg.CPUWrite)
	}

	// Set pool size
	poolSize := cfg.PoolSize
	if poolSize == 0 && threads != 0 {
		poolSize = threads
		if cfg.CPU != 0 && cfg.CPU < poolSize {
			poolSize = cfg.CPU
		}
		poolSize = min(poolSizeMax, poolSize) // Only enforce max pool size when inferred from threads/CPU
	}
	poolSize = max(poolSizeMin, poolSize) // Always enforce min pool size
	cfg.PoolSize = poolSize

	// useful for motherduck but safe to pass at initial connect
	cfg.WriteSettings["custom_user_agent"] = "rill"
	return cfg, nil
}

func generateDSN(path, encodedQuery string) string {
	if encodedQuery == "" {
		return path
	}
	return path + "?" + encodedQuery
}
