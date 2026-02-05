package duckdb

import (
	"context"
	"fmt"
	"os"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
)

type mdToSelfInputProps struct {
	SQL   string `mapstructure:"sql"`
	Token string `mapstructure:"motherduck_token"`
	DB    string `mapstructure:"db"`
	DSN   string `mapstructure:"dsn"`
}

type mdConfigProps struct {
	Path            string `mapstructure:"path"`
	Token           string `mapstructure:"token"`
	AllowHostAccess bool   `mapstructure:"allow_host_access"`
}

func (p *mdToSelfInputProps) Validate() error {
	if p.SQL == "" {
		return fmt.Errorf("missing property 'sql'")
	}
	if p.DSN != "" && p.DB != "" {
		return fmt.Errorf("cannot set both 'dsn' and 'db'")
	}
	return nil
}

type mdToSelfExecutor struct {
	c *connection
}

var _ drivers.ModelExecutor = &mdToSelfExecutor{}

func (e *mdToSelfExecutor) Concurrency(desired int) (int, bool) {
	if desired > 1 {
		return 0, false
	}
	return 1, true
}

func (e *mdToSelfExecutor) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
	inputProps := &mdToSelfInputProps{}
	if err := mapstructure.WeakDecode(opts.InputProperties, inputProps); err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}
	if err := inputProps.Validate(); err != nil {
		return nil, fmt.Errorf("invalid input properties: %w", err)
	}

	mdConfig := &mdConfigProps{}
	err := mapstructure.WeakDecode(opts.InputHandle.Config(), mdConfig)
	if err != nil {
		return nil, err
	}

	// get dsn
	var dsn string
	if inputProps.DSN != "" {
		dsn = inputProps.DSN
	} else if inputProps.DB != "" {
		dsn = inputProps.DB
	} else if mdConfig.Path != "" {
		dsn = mdConfig.Path
	}

	// get token
	var token string
	if inputProps.Token != "" {
		token = inputProps.Token
	} else if mdConfig.Token != "" {
		token = mdConfig.Token
	} else if mdConfig.AllowHostAccess {
		token = os.Getenv("motherduck_token")
	}
	if token == "" {
		return nil, fmt.Errorf("no motherduck token found. Refer to this documentation for instructions: https://docs.rilldata.com/build/connectors/olap/motherduck")
	}

	clone := *opts
	m := &ModelInputProperties{
		SQL:         inputProps.SQL,
		InitQueries: fmt.Sprintf("INSTALL 'motherduck'; LOAD 'motherduck'; SET motherduck_token=%s; ATTACH %s;", safeSQLString(token), safeSQLString(dsn)),
	}
	var props map[string]any
	err = mapstructure.Decode(m, &props)
	if err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}
	clone.InputProperties = props

	executor := &selfToSelfExecutor{c: e.c}
	return executor.Execute(ctx, &clone)
}
