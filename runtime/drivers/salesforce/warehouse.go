package salesforce

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

var tracer = otel.Tracer("github.com/rilldata/rill/runtime/drivers/salesforce")

const defaultClientID = "3MVG9KsVczVNcM8y6w3Kjszy.DW9gMzcYDHT97WIX3NYNYA35UvITypEhtYc6FDY8qqcDEIQc_qJgZErv6Q_d"

var _ drivers.Warehouse = &connection{}

// QueryAsFiles implements drivers.SQLStore
func (c *connection) QueryAsFiles(ctx context.Context, props map[string]any) (outIt drivers.FileIterator, outErr error) {
	ctx, span := tracer.Start(ctx, "Connection.QueryAsFiles")
	defer func() {
		if outErr != nil {
			span.SetStatus(codes.Error, outErr.Error())
		}
		span.End()
	}()

	srcProps, err := parseSourceProperties(props)
	if err != nil {
		return nil, err
	}

	authOptions := c.authOptions()
	// Per-model source properties override the connector-level config so a
	// single connector definition can be reused across models with different
	// credentials when needed.
	srcProps.applyOverrides(&authOptions)

	if authOptions.Endpoint == "" {
		return nil, fmt.Errorf("the property 'endpoint' is required for Salesforce. Provide 'endpoint' in the YAML properties or pass '--env connector.salesforce.endpoint=...' to 'rill start'")
	}

	switch selectAuthMode(authOptions) {
	case authModeUnknown:
		return nil, fmt.Errorf("Salesforce credentials are required: provide a JWT 'key', a 'username' and 'password' (with 'client_secret'), or a 'client_secret' for the client credentials flow")
	case authModePassword:
		if authOptions.ClientSecret == "" {
			return nil, fmt.Errorf("the property 'client_secret' is required for username/password authentication. Provide 'client_secret' in the YAML properties or pass '--env connector.salesforce.client_secret=...' to 'rill start'")
		}
	case authModeJWT:
		if authOptions.Username == "" {
			return nil, fmt.Errorf("the property 'username' is required for JWT authentication. Provide 'username' in the YAML properties or pass '--env connector.salesforce.username=...' to 'rill start'")
		}
	}

	session, err := authenticate(authOptions)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	job := makeBulk2QueryJob(session, c.logger)
	if err := job.startJob(ctx, srcProps.SOQL, srcProps.QueryAll); err != nil {
		return nil, err
	}
	return job, nil
}

type sourceProperties struct {
	SOQL         string `mapstructure:"soql"`
	SQL          string `mapstructure:"sql"`
	QueryAll     bool   `mapstructure:"queryAll"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	Key          string `mapstructure:"key"`
	Endpoint     string `mapstructure:"endpoint"`
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
}

// applyOverrides copies any non-empty source-level credential fields onto the
// supplied connector-level options. Used so a model can override the connector
// for one-off credentials without changing the connector definition.
func (s *sourceProperties) applyOverrides(opts *authenticationOptions) {
	if s.Endpoint != "" {
		opts.Endpoint = s.Endpoint
	}
	if s.ClientID != "" {
		opts.ConnectedApp = s.ClientID
	}
	if s.Username != "" {
		opts.Username = s.Username
	}
	if s.Password != "" {
		opts.Password = s.Password
	}
	if s.Key != "" {
		opts.JWT = s.Key
	}
	if s.ClientSecret != "" {
		opts.ClientSecret = s.ClientSecret
	}
}

func parseSourceProperties(props map[string]any) (*sourceProperties, error) {
	conf := &sourceProperties{}
	if err := mapstructure.Decode(props, conf); err != nil {
		return nil, err
	}
	// Accept `sql:` as an alias for `soql:` so Salesforce fits the same model
	// shape as other warehouse drivers (which read `sql:` from the source).
	// Bulk API 2.0 derives the SObject from the query itself, so a separate
	// `sobject:` property is no longer required.
	if conf.SOQL == "" {
		conf.SOQL = conf.SQL
	}
	if conf.SOQL == "" {
		return nil, fmt.Errorf("property 'soql' (or 'sql') is mandatory for connector \"salesforce\"")
	}
	return conf, nil
}
