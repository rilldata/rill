package duckdb

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"

	"github.com/mitchellh/mapstructure"
)

const (
	cpuThreadRatio float64 = 0.5
	poolSizeMin    int     = 2
	poolSizeMax    int     = 5
)

// config represents the DuckDB driver config
type config struct {
	// DSN is the connection string
	DSN string `mapstructure:"dsn"`
	// PoolSize is the number of concurrent connections and queries allowed
	PoolSize int `mapstructure:"pool_size"`
	// AllowHostAccess denotes whether to limit access to the local environment and file system
	AllowHostAccess bool `mapstructure:"allow_host_access"`
	// ErrorOnIncompatibleVersion controls whether to return error or delete DBFile created with older duckdb version.
	ErrorOnIncompatibleVersion bool `mapstructure:"error_on_incompatible_version"`
	// ExtTableStorage controls if every table is stored in a different db file
	ExtTableStorage bool `mapstructure:"external_table_storage"`
	// CPU cores available for the DB
	CPU int `mapstructure:"cpu"`
	// MemoryLimitGB is the amount of memory available for the DB
	MemoryLimitGB int `mapstructure:"memory_limit_gb"`
	// StorageLimitBytes is the amount of disk storage available for the DB
	StorageLimitBytes int64 `mapstructure:"storage_limit_bytes"`
	// MaxMemoryOverride sets a hard override for the "max_memory" DuckDB setting
	MaxMemoryGBOverride int `mapstructure:"max_memory_gb_override"`
	// ThreadsOverride sets a hard override for the "threads" DuckDB setting. Set to -1 for unlimited threads.
	ThreadsOverride int `mapstructure:"threads_override"`
	// BootQueries is queries to run on boot. Use ; to separate multiple queries. Common use case is to provide project specific memory and threads ratios.
	BootQueries string `mapstructure:"boot_queries"`
	// DBFilePath is the path where the database is stored. It is inferred from the DSN (can't be provided by user).
	DBFilePath string `mapstructure:"-"`
	// ExtStoragePath is the path where the database files are stored in case external_table_storage is true. It is inferred from the DSN (can't be provided by user).
	ExtStoragePath string `mapstructure:"-"`
}

func newConfig(cfgMap map[string]any) (*config, error) {
	cfg := &config{}
	err := mapstructure.WeakDecode(cfgMap, cfg)
	if err != nil {
		return nil, fmt.Errorf("could not decode config: %w", err)
	}

	// Parse DSN as URL
	uri, err := url.Parse(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("could not parse dsn: %w", err)
	}
	qry, err := url.ParseQuery(uri.RawQuery)
	if err != nil {
		return nil, fmt.Errorf("could not parse dsn: %w", err)
	}

	// Infer DBFilePath
	cfg.DBFilePath = uri.Path
	if cfg.ExtTableStorage {
		cfg.ExtStoragePath = filepath.Dir(cfg.DBFilePath)
	}

	// Set memory limit
	maxMemory := cfg.MemoryLimitGB
	if cfg.MaxMemoryGBOverride != 0 {
		maxMemory = cfg.MaxMemoryGBOverride
	}
	if maxMemory > 0 {
		qry.Add("max_memory", fmt.Sprintf("%dGB", maxMemory))
	}

	// Set threads limit
	var threads int
	if cfg.ThreadsOverride != 0 {
		threads = cfg.ThreadsOverride
	} else if cfg.CPU > 0 {
		threads = int(cpuThreadRatio * float64(cfg.CPU))
		if threads <= 0 {
			threads = 1
		}
	}
	if threads > 0 { // NOTE: threads=0 or threads=-1 means no limit
		qry.Add("threads", strconv.Itoa(threads))
	}

	// Set pool size
	poolSize := cfg.PoolSize
	if qry.Has("rill_pool_size") {
		// For backwards compatibility, we also support overriding the pool size via the DSN when "rill_pool_size" is a query argument.

		// Remove from query string (so not passed into DuckDB config)
		val := qry.Get("rill_pool_size")
		qry.Del("rill_pool_size")

		// Parse as integer
		poolSize, err = strconv.Atoi(val)
		if err != nil {
			return nil, fmt.Errorf("could not parse dsn: 'rill_pool_size' is not an integer")
		}
	}
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
	if !qry.Has("custom_user_agent") {
		qry.Add("custom_user_agent", "rill")
	}
	// Rebuild DuckDB DSN (which should be "path?key=val&...")
	// this is required since spaces and other special characters are valid in db file path but invalid and hence encoded in URL
	cfg.DSN = generateDSN(uri.Path, qry.Encode())

	return cfg, nil
}

func generateDSN(path, encodedQuery string) string {
	if encodedQuery == "" {
		return path
	}
	return path + "?" + encodedQuery
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
