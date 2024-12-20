package duckdb

import (
	"context"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/azure"
	"github.com/rilldata/rill/runtime/drivers/s3"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
)

type s3InputProps struct {
	Path   string             `mapstructure:"path"`
	Format drivers.FileFormat `mapstructure:"format"`
	DuckDB map[string]any     `mapstructure:"duckdb"`
}

func (p *s3InputProps) Validate() error {
	if p.Path == "" {
		return fmt.Errorf("missing property `path`")
	}
	return nil
}

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
	inputProps := &s3InputProps{}
	if err := mapstructure.WeakDecode(opts.InputProperties, inputProps); err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}
	if err := inputProps.Validate(); err != nil {
		return nil, fmt.Errorf("invalid input properties: %w", err)
	}

	// Build the model executor options with updated input and output properties
	clone := *opts
	newInputProps, err := e.modelInputProperties(opts.InputHandle, inputProps, opts.ModelName)
	if err != nil {
		return nil, err
	}
	clone.InputProperties = newInputProps
	newOpts := &clone

	// execute
	executor := &selfToSelfExecutor{c: e.c}
	return executor.Execute(ctx, newOpts)
}

func (e *objectStoreToSelfExecutor) modelInputProperties(inputHandle drivers.Handle, inputProps *s3InputProps, model string) (map[string]any, error) {
	m := &ModelInputProperties{}
	var format string
	if inputProps.Format != "" {
		format = fmt.Sprintf(".%s", inputProps.Format)
	} else {
		format = fileutil.FullExt(inputProps.Path)
	}

	// Generate secret SQL to access the service and set as pre_exec_query
	switch inputHandle.Driver() {
	case "s3":
		safeSecretName := safeName(model + "_s3_secret_")
		s3Config := &s3.ConfigProperties{}
		err := mapstructure.WeakDecode(inputHandle.Config(), s3Config)
		if err != nil {
			return nil, fmt.Errorf("failed to parse s3 config properties: %w", err)
		}
		var sb strings.Builder
		sb.WriteString("CREATE OR REPLACE TEMPORARY SECRET ")
		sb.WriteString(safeSecretName)
		sb.WriteString(" (TYPE S3,")
		if s3Config.AllowHostAccess {
			sb.WriteString(" PROVIDER CREDENTIAL_CHAIN")
		}
		if s3Config.AccessKeyID != "" {
			fmt.Fprintf(&sb, ", KEY_ID %s, SECRET %s, SESSION_TOKEN %s", safeSQLString(s3Config.AccessKeyID), safeSQLString(s3Config.SecretAccessKey), safeSQLString(s3Config.SessionToken))
		}
		if s3Config.Endpoint != "" {
			sb.WriteString(", ENDPOINT ")
			sb.WriteString(safeSQLString(s3Config.Endpoint))
		}
		if s3Config.Region != "" {
			sb.WriteString(", REGION ")
			sb.WriteString(safeSQLString(s3Config.Region))
		}
		sb.WriteRune(')')
		m.PreExec = sb.String()
	case "gcs":
		safeSecretName := safeName(model + "_gcs_secret_")
		// GCS works via S3 compatibility mode
		s3Config := &s3.ConfigProperties{}
		err := mapstructure.WeakDecode(inputHandle.Config(), s3Config)
		if err != nil {
			return nil, fmt.Errorf("failed to parse s3 config properties: %w", err)
		}
		var sb strings.Builder
		sb.WriteString("CREATE OR REPLACE TEMPORARY SECRET ")
		sb.WriteString(safeSecretName)
		sb.WriteString(" (TYPE GCS,")
		if s3Config.AllowHostAccess {
			sb.WriteString(" PROVIDER CREDENTIAL_CHAIN")
		}
		if s3Config.AccessKeyID != "" {
			fmt.Fprintf(&sb, ", KEY_ID %s, SECRET %s", safeSQLString(s3Config.AccessKeyID), safeSQLString(s3Config.SecretAccessKey))
		}
		sb.WriteRune(')')
		m.PreExec = sb.String()
	case "azure":
		safeSecretName := safeName(model + "_azure_secret_")
		azureConfig := &azure.ConfigProperties{}
		err := mapstructure.WeakDecode(inputHandle.Config(), azureConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to parse s3 config properties: %w", err)
		}
		var sb strings.Builder
		sb.WriteString("CREATE OR REPLACE TEMPORARY SECRET ")
		sb.WriteString(safeSecretName)
		sb.WriteString(" (TYPE AZURE,")
		if azureConfig.AllowHostAccess {
			sb.WriteString(" PROVIDER CREDENTIAL_CHAIN")
		}
		if azureConfig.ConnectionString != "" {
			fmt.Fprintf(&sb, ", CONNECTION_STRING %s", safeSQLString(azureConfig.ConnectionString))
		}
		if azureConfig.Account != "" {
			fmt.Fprintf(&sb, ", ACCOUNT_NAME %s", safeSQLString(azureConfig.Account))
		}
		sb.WriteRune(')')
		m.PreExec = sb.String()
	default:
		return nil, fmt.Errorf("internal error: unsupported object store: %s", inputHandle.Driver())
	}

	// Set SQL to read from the external source
	from, err := sourceReader([]string{inputProps.Path}, format, inputProps.DuckDB)
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
