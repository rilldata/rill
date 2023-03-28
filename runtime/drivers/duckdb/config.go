package duckdb

import (
	"fmt"
	"net/url"
	"strconv"
)

const poolSizeKey = "rill_pool_size"

// config represents the Driver config, extracted from the DSN
type config struct {
	// DSN for DuckDB
	DSN string
	// PoolSize is the number of concurrent connections and queries allowed
	PoolSize int
	// DBFilePath is the path where database is stored
	DBFilePath string
}

func newConfig(dsn string) (*config, error) {
	// Parse DSN as URL
	uri, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("could not parse dsn: %w", err)
	}
	qry, err := url.ParseQuery(uri.RawQuery)
	if err != nil {
		return nil, fmt.Errorf("could not parse dsn: %w", err)
	}

	// If poolSizeKey is in the DSN, parse and remove it
	poolSize := 1
	if qry.Has(poolSizeKey) {
		// Parse as integer
		poolSize, err = strconv.Atoi(qry.Get(poolSizeKey))
		if err != nil {
			return nil, fmt.Errorf("duckdb Driver: %s is not an integer", poolSizeKey)
		}
		// Remove from query string (so not passed into DuckDB config)
		qry.Del(poolSizeKey)
	}
	if poolSize < 1 {
		return nil, fmt.Errorf("%s must be >= 1", poolSizeKey)
	}

	// Rebuild DuckDB DSN (which should be "path?key=val&...")
	uri.RawQuery = qry.Encode()
	dsn = uri.String()

	// Return config
	cfg := &config{
		DSN:        dsn,
		PoolSize:   poolSize,
		DBFilePath: uri.Path,
	}
	return cfg, nil
}
