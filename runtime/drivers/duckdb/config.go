package duckdb

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/mitchellh/mapstructure"
)

// config represents the DuckDB driver config
type config struct {
	// DSN is the connection string
	DSN string `mapstructure:"dsn"`
	// PoolSize is the number of concurrent connections and queries allowed
	PoolSize int `mapstructure:"pool_size"`
	// AllowHostAccess denotes whether to limit access to the local environment and file system
	AllowHostAccess bool `mapstructure:"allow_host_access"`
	// DBFilePath is the path where the database is stored. It is inferred from the DSN (can't be provided by user).
	DBFilePath string `mapstructure:"-"`
}

func newConfig(cfgMap map[string]any) (*config, error) {
	cfg := &config{
		PoolSize: 1, // Default value
	}
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

	// We also support overriding the pool size via the DSN by setting "rill_pool_size" as a query argument.
	if qry.Has("rill_pool_size") {
		// Parse as integer
		cfg.PoolSize, err = strconv.Atoi(qry.Get("rill_pool_size"))
		if err != nil {
			return nil, fmt.Errorf("could not parse dsn: 'rill_pool_size' is not an integer")
		}

		// Remove from query string (so not passed into DuckDB config)
		qry.Del("rill_pool_size")

		// Rebuild DuckDB DSN (which should be "path?key=val&...")
		uri.RawQuery = qry.Encode()
		cfg.DSN = uri.String()
	}

	// Check pool size
	if cfg.PoolSize < 1 {
		return nil, fmt.Errorf("duckdb pool size must be >= 1")
	}

	return cfg, nil
}
