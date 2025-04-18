package duckdb

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	rillmysql "github.com/rilldata/rill/runtime/drivers/mysql"
	"github.com/rilldata/rill/runtime/drivers/postgres"
)

type sqlStoreToSelfInputProps struct {
	SQL         string `mapstructure:"sql"`
	DSN         string `mapstructure:"dsn"`
	DatabaseURL string `mapstructure:"database_url"`
}

func (p *sqlStoreToSelfInputProps) resolveDSN() string {
	if p.DSN != "" {
		return p.DSN
	}
	return p.DatabaseURL
}

func (p *sqlStoreToSelfInputProps) Validate() error {
	if p.SQL == "" {
		return fmt.Errorf("missing property 'sql'")
	}
	if p.DSN != "" && p.DatabaseURL != "" {
		return fmt.Errorf("cannot set both 'dsn' and 'database_url'")
	}
	return nil
}

type sqlStoreToSelfExecutor struct {
	c *connection
}

var _ drivers.ModelExecutor = &sqlStoreToSelfExecutor{}

func (e *sqlStoreToSelfExecutor) Concurrency(desired int) (int, bool) {
	if desired > 1 {
		return 0, false
	}
	return 1, true
}

func (e *sqlStoreToSelfExecutor) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
	inputProps := &sqlStoreToSelfInputProps{}
	if err := mapstructure.WeakDecode(opts.InputProperties, inputProps); err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}
	if err := inputProps.Validate(); err != nil {
		return nil, fmt.Errorf("invalid input properties: %w", err)
	}

	// Build the model executor options with updated input properties
	clone := *opts
	newInputProps, err := e.modelInputProperties(opts.ModelName, opts.InputConnector, opts.InputHandle, inputProps)
	if err != nil {
		return nil, err
	}
	clone.InputProperties = newInputProps
	newOpts := &clone

	// execute
	executor := &selfToSelfExecutor{c: e.c}
	return executor.Execute(ctx, newOpts)
}

func (e *sqlStoreToSelfExecutor) modelInputProperties(modelName, inputConnector string, inputHandle drivers.Handle, inputProps *sqlStoreToSelfInputProps) (map[string]any, error) {
	m := &ModelInputProperties{}
	dbName := fmt.Sprintf("%s__%s", modelName, inputConnector)
	safeDBName := safeName(dbName)
	userQuery, _ := strings.CutSuffix(inputProps.SQL, ";") // trim trailing semi colon
	switch inputHandle.Driver() {
	case "mysql":
		dsn := inputProps.resolveDSN()
		if dsn == "" {
			// may be configured via a connector
			var config *rillmysql.ConfigProperties
			if err := mapstructure.Decode(inputHandle.Config(), &config); err != nil {
				return nil, err
			}
			dsn = rewriteMySQLDSN(config.DSN)
		}
		if dsn == "" {
			return nil, fmt.Errorf("must set `dsn` for models that transfer data from `mysql` to `duckdb`")
		}
		m.PreExec = fmt.Sprintf("INSTALL 'MYSQL'; LOAD 'MYSQL'; ATTACH %s AS %s (TYPE mysql, READ_ONLY)", safeSQLString(dsn), safeDBName)
		m.SQL = fmt.Sprintf("SELECT * FROM mysql_query(%s, %s)", safeSQLString(dbName), safeSQLString(userQuery))
	case "postgres":
		dsn := inputProps.resolveDSN()
		if dsn == "" {
			// may be configured via a connector
			var config *postgres.ConfigProperties
			if err := mapstructure.Decode(inputHandle.Config(), &config); err != nil {
				return nil, err
			}
			dsn = config.ResolveDSN()
		}
		if dsn == "" {
			return nil, fmt.Errorf("must set `database_url` or `dsn` for models that transfer data from `postgres` to `duckdb`")
		}
		m.PreExec = fmt.Sprintf("INSTALL 'POSTGRES'; LOAD 'POSTGRES'; ATTACH %s AS %s (TYPE postgres, READ_ONLY)", safeSQLString(dsn), safeDBName)
		m.SQL = fmt.Sprintf("SELECT * FROM postgres_query(%s, %s)", safeSQLString(dbName), safeSQLString(userQuery))
	default:
		return nil, fmt.Errorf("internal error: unsupported external database: %s", inputHandle.Driver())
	}
	m.PostExec = fmt.Sprintf("DETACH %s", safeDBName)
	propsMap := make(map[string]any)
	if err := mapstructure.Decode(m, &propsMap); err != nil {
		return nil, err
	}
	return propsMap, nil
}

// rewriteMySQLDSN rewrites a MySQL DSN to a format that DuckDB expects.
// DuckDB does not support the URI based DSN format yet. It expects the DSN to be in the form of key=value pairs.
// This function parses the MySQL URI based DSN and converts it to the key=value format. It only converts the common parameters.
// For more advanced parameters like SSL configs, the user should manually convert the DSN to the key=value format.
// If there is an error parsing the DSN, it returns the DSN as is.
func rewriteMySQLDSN(dsn string) string {
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		// If we can't parse the DSN, just return it as is. May be it is already in the form duckdb expects.
		return dsn
	}

	var sb strings.Builder

	if cfg.User != "" {
		sb.WriteString(fmt.Sprintf("user=%s ", cfg.User))
	}
	if cfg.Passwd != "" {
		sb.WriteString(fmt.Sprintf("password=%s ", cfg.Passwd))
	}
	if cfg.DBName != "" {
		sb.WriteString(fmt.Sprintf("database=%s ", cfg.DBName))
	}
	switch cfg.Net {
	case "unix":
		sb.WriteString(fmt.Sprintf("socket=%s ", cfg.Addr))
	case "tcp", "tcp6":
		host, port, err := net.SplitHostPort(cfg.Addr)
		if err != nil {
			return dsn
		}
		sb.WriteString(fmt.Sprintf("host=%s ", host))
		if port != "" {
			sb.WriteString(fmt.Sprintf("port=%s ", port))
		}
	default:
		return dsn
	}
	return sb.String()
}
