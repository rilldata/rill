package python

import (
	"context"
	"fmt"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
)

// connectorEnvMapping maps connector config keys to environment variable names.
// Only keys that have a well-known env var equivalent are listed.
var connectorEnvMapping = map[string]map[string]string{
	"gcs": {
		"google_application_credentials": "GOOGLE_APPLICATION_CREDENTIALS",
	},
	"bigquery": {
		"google_application_credentials": "GOOGLE_APPLICATION_CREDENTIALS",
	},
	"s3": {
		"aws_access_key_id":     "AWS_ACCESS_KEY_ID",
		"aws_secret_access_key": "AWS_SECRET_ACCESS_KEY",
		"aws_session_token":     "AWS_SESSION_TOKEN",
		"aws_role_arn":          "AWS_ROLE_ARN",
	},
	"athena": {
		"aws_access_key_id":     "AWS_ACCESS_KEY_ID",
		"aws_secret_access_key": "AWS_SECRET_ACCESS_KEY",
		"aws_session_token":     "AWS_SESSION_TOKEN",
		"aws_region":            "AWS_REGION",
	},
	"azure": {
		"azure_storage_connection_string": "AZURE_STORAGE_CONNECTION_STRING",
		"azure_storage_key":               "AZURE_STORAGE_KEY",
		"azure_storage_sas_token":         "AZURE_STORAGE_SAS_TOKEN",
		"azure_storage_account":           "AZURE_STORAGE_ACCOUNT",
	},
	"motherduck": {
		"token": "MOTHERDUCK_TOKEN",
	},
	"snowflake": {
		"dsn":      "SNOWFLAKE_DSN",
		"password": "SNOWFLAKE_PASSWORD",
	},
	"postgres": {
		"database_url": "POSTGRES_DSN",
		"password":     "POSTGRES_PASSWORD",
	},
}

// ResolveConnectorEnvVars resolves env vars from the specified connectors.
// It acquires each connector's config and maps known keys to env var names.
func ResolveConnectorEnvVars(ctx context.Context, connectorNames []string, acquireConnector func(ctx context.Context, name string) (drivers.Handle, func(), error)) (map[string]string, error) {
	if len(connectorNames) == 0 {
		return nil, nil
	}

	envVars := make(map[string]string)
	for _, connName := range connectorNames {
		handle, release, err := acquireConnector(ctx, connName)
		if err != nil {
			return nil, fmt.Errorf("python: failed to acquire connector %q for create_secrets_from_connectors: %w", connName, err)
		}

		config := handle.Config()
		driverName := handle.Driver()
		release()

		// Look up known mappings for this driver
		mapping, ok := connectorEnvMapping[driverName]
		if !ok {
			// No known mapping; pass all non-empty config values as uppercase env vars prefixed with connector name
			for k, v := range config {
				if s, ok := v.(string); ok && s != "" {
					envName := strings.ToUpper(connName + "_" + k)
					envVars[envName] = s
				}
			}
			continue
		}

		// Apply known mappings
		for configKey, envName := range mapping {
			val, ok := config[configKey]
			if !ok {
				continue
			}
			s, ok := val.(string)
			if !ok || s == "" {
				continue
			}
			envVars[envName] = s
		}
	}

	return envVars, nil
}
