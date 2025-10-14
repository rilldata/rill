package duckdb

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/azure"
	"github.com/rilldata/rill/runtime/drivers/gcs"
	"github.com/rilldata/rill/runtime/drivers/s3"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/pkg/globutil"
)

var errGCSUsesNativeCreds = errors.New("GCS uses native credentials")

type objectStoreToSelfExecutor struct {
	c *connection
}

var _ drivers.ModelExecutor = &objectStoreToSelfExecutor{}

func (e *objectStoreToSelfExecutor) Concurrency(desired int) (int, bool) {
	if desired > 1 {
		return 0, false
	}
	return 1, true
}

func (e *objectStoreToSelfExecutor) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
	// Build the model executor options with updated input properties
	clone := *opts
	newInputProps, err := e.modelInputProperties(ctx, opts)
	if err != nil {
		if errors.Is(err, errGCSUsesNativeCreds) {
			e := &objectStoreToSelfExecutorNonNative{c: e.c}
			return e.Execute(ctx, opts)
		}
		return nil, err
	}
	clone.InputProperties = newInputProps
	newOpts := &clone

	// execute
	executor := &selfToSelfExecutor{c: e.c}
	return executor.Execute(ctx, newOpts)
}

func (e *objectStoreToSelfExecutor) modelInputProperties(ctx context.Context, opts *drivers.ModelExecuteOptions) (map[string]any, error) {
	parsed := &drivers.ObjectStoreModelInputProperties{}
	if err := parsed.Decode(opts.InputProperties); err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}

	m := &ModelInputProperties{}
	var format string
	if parsed.Format != "" {
		format = fmt.Sprintf(".%s", parsed.Format)
	} else {
		format = fileutil.FullExt(parsed.Path)
	}

	// Generate secret SQL to access the service and set as pre_exec_query
	var err error
	m.PreExec, err = objectStoreSecretSQL(ctx, opts, opts.InputConnector, parsed.Path, opts.InputProperties)
	if err != nil {
		return nil, err
	}

	// Set SQL to read from the external source
	from, err := sourceReader([]string{parsed.Path}, format, parsed.DuckDB)
	if err != nil {
		return nil, err
	}
	m.SQL = "SELECT * FROM " + from

	propsMap := make(map[string]any)
	if err := mapstructure.Decode(m, &propsMap); err != nil {
		return nil, err
	}
	return propsMap, nil
}

func objectStoreSecretSQL(ctx context.Context, opts *drivers.ModelExecuteOptions, connector, optionalBucketURL string, optionalAdditionalConfig map[string]any) (string, error) {
	handle, release, err := opts.Env.AcquireConnector(ctx, connector)
	if err != nil {
		return "", err
	}
	defer release()

	_, ok := handle.AsObjectStore()
	if !ok {
		return "", fmt.Errorf("can only create secrets for object store connectors %q", connector)
	}

	safeSecretName := safeName(fmt.Sprintf("%s__%s__secret", opts.ModelName, connector))

	switch handle.Driver() {
	case "s3":
		conn, ok := handle.(*s3.Connection)
		if !ok {
			return "", fmt.Errorf("internal error: expected s3 connector handle")
		}
		s3Config := conn.ParsedConfig()
		err := mapstructure.WeakDecode(optionalAdditionalConfig, s3Config)
		if err != nil {
			return "", fmt.Errorf("failed to parse s3 config properties: %w", err)
		}
		var sb strings.Builder
		sb.WriteString("CREATE OR REPLACE TEMPORARY SECRET ")
		sb.WriteString(safeSecretName)
		sb.WriteString(" (TYPE S3")
		if s3Config.AllowHostAccess {
			sb.WriteString(", PROVIDER CREDENTIAL_CHAIN")
		}
		if s3Config.AccessKeyID != "" {
			fmt.Fprintf(&sb, ", KEY_ID %s, SECRET %s", safeSQLString(s3Config.AccessKeyID), safeSQLString(s3Config.SecretAccessKey))
		}
		if s3Config.SessionToken != "" {
			fmt.Fprintf(&sb, ", SESSION_TOKEN %s", safeSQLString(s3Config.SessionToken))
		}
		if s3Config.Endpoint != "" {
			uri, err := url.Parse(s3Config.Endpoint)
			if err == nil && uri.Scheme != "" { // let duckdb raise an error if the endpoint is invalid
				// for duckdb the endpoint should not have a scheme
				s3Config.Endpoint = strings.TrimPrefix(s3Config.Endpoint, uri.Scheme+"://")
				if uri.Scheme == "http" {
					sb.WriteString(", USE_SSL false")
				}
			}
			sb.WriteString(", ENDPOINT ")
			sb.WriteString(safeSQLString(s3Config.Endpoint))
			sb.WriteString(", URL_STYLE path")
		}
		if s3Config.Region != "" {
			sb.WriteString(", REGION ")
			sb.WriteString(safeSQLString(s3Config.Region))
		} else if optionalBucketURL != "" {
			// DuckDB does not automatically resolve the region as of 1.2.0 so we try to detect and set the region.
			uri, err := globutil.ParseBucketURL(optionalBucketURL)
			if err != nil {
				return "", fmt.Errorf("failed to parse path %q: %w", optionalBucketURL, err)
			}
			reg, err := conn.BucketRegion(ctx, uri.Host)
			if err != nil {
				return "", fmt.Errorf("failed to get bucket region (set `region` in s3.yaml): %w", err)
			}
			sb.WriteString(", REGION ")
			sb.WriteString(safeSQLString(reg))
		}
		sb.WriteRune(')')
		return sb.String(), nil
	case "gcs":
		// GCS works via S3 compatibility mode.
		// This means we that gcsConfig.KeyID and gcsConfig.Secret should be set instead of gcsConfig.SecretJSON.
		gcsConnectorProp := handle.Config()
		gcsConfig, err := gcs.NewConfigProperties(gcsConnectorProp)
		if err != nil {
			return "", fmt.Errorf("failed to load gcs base config: %w", err)
		}
		if err := mapstructure.WeakDecode(optionalAdditionalConfig, gcsConfig); err != nil {
			return "", fmt.Errorf("failed to parse gcs config properties: %w", err)
		}
		// If no credentials are provided we assume that the user wants to use the native credentials
		if gcsConfig.SecretJSON != "" || (gcsConfig.KeyID == "" && gcsConfig.Secret == "") {
			return "", errGCSUsesNativeCreds
		}
		var sb strings.Builder
		sb.WriteString("CREATE OR REPLACE TEMPORARY SECRET ")
		sb.WriteString(safeSecretName)
		sb.WriteString(" (TYPE GCS")
		if gcsConfig.AllowHostAccess {
			sb.WriteString(", PROVIDER CREDENTIAL_CHAIN")
		}
		if gcsConfig.KeyID != "" {
			fmt.Fprintf(&sb, ", KEY_ID %s, SECRET %s", safeSQLString(gcsConfig.KeyID), safeSQLString(gcsConfig.Secret))
		}
		sb.WriteRune(')')
		return sb.String(), nil
	case "azure":
		conn, ok := handle.(*azure.Connection)
		if !ok {
			return "", fmt.Errorf("internal error: expected azure connector handle")
		}
		azureConfig := conn.ParsedConfig()
		err := mapstructure.WeakDecode(optionalAdditionalConfig, azureConfig)
		if err != nil {
			return "", fmt.Errorf("failed to parse azure config properties: %w", err)
		}
		var sb strings.Builder
		sb.WriteString("CREATE OR REPLACE TEMPORARY SECRET ")
		sb.WriteString(safeSecretName)
		sb.WriteString(" (TYPE AZURE")
		// if connection string is set then use that and fall back to env credentials only if host access is allowed and connection string is not set
		connectionString := azureConfig.GetConnectionString()
		if connectionString != "" {
			fmt.Fprintf(&sb, ", CONNECTION_STRING %s", safeSQLString(connectionString))
		} else if azureConfig.AllowHostAccess {
			// duckdb will use default defaultazurecredential https://github.com/Azure/azure-sdk-for-cpp/blob/azure-identity_1.6.0/sdk/identity/azure-identity/README.md#defaultazurecredential
			sb.WriteString(", PROVIDER CREDENTIAL_CHAIN")
		}
		if azureConfig.GetAccount() != "" {
			fmt.Fprintf(&sb, ", ACCOUNT_NAME %s", safeSQLString(azureConfig.GetAccount()))
		}
		sb.WriteRune(')')
		return sb.String(), nil
	default:
		return "", fmt.Errorf("internal error: unsupported object store connector %q", handle.Driver())
	}
}

// objectStoreToSelfExecutorNonNative is a non-native implementation of objectStoreToSelfExecutor.
// It uses Rill's own connectors instead of duckdb's native connectors.
type objectStoreToSelfExecutorNonNative struct {
	c *connection
}

func (e *objectStoreToSelfExecutorNonNative) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
	parsed := &drivers.ObjectStoreModelInputProperties{}
	if err := parsed.Decode(opts.InputProperties); err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}

	store, ok := opts.InputHandle.AsObjectStore()
	if !ok {
		return nil, fmt.Errorf("input handle is not an object store")
	}

	iter, err := store.DownloadFiles(ctx, parsed.Path)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	// We want to batch all the files to avoid issues with schema compatibility and partition_overwrite inserts.
	// If a user encounters performance issues, we should encourage them to use `partitions:` without `incremental:` to break ingestion into smaller batches.
	iter.SetKeepFilesUntilClose()
	var files []string
	for {
		batch, err := iter.Next(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		files = append(files, batch...)
	}
	if len(files) == 0 {
		return nil, drivers.ErrNoRows
	}

	var format string
	if parsed.Format != "" {
		format = fmt.Sprintf(".%s", parsed.Format)
	} else {
		format = fileutil.FullExt(parsed.Path)
	}

	fromClause, err := sourceReader(files, format, parsed.DuckDB)
	if err != nil {
		return nil, err
	}

	m := &ModelInputProperties{SQL: "SELECT * FROM " + fromClause}
	propsMap := make(map[string]any)
	if err := mapstructure.Decode(m, &propsMap); err != nil {
		return nil, err
	}
	opts.InputProperties = propsMap

	executor := &selfToSelfExecutor{c: e.c}
	return executor.Execute(ctx, opts)
}
