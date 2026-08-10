package duckdb

import (
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func TestModelOutputPropertiesOnSchemaChangeValidation(t *testing.T) {
	mode := OnSchemaChangeAppendNewColumns
	tests := []struct {
		name      string
		opts      *drivers.ModelExecuteOptions
		props     ModelOutputProperties
		wantErr   string
		wantMode  OnSchemaChange
		wantEmpty bool
	}{
		{
			name: "valid merge",
			opts: schemaChangeModelExecuteOptions(true, true),
			props: ModelOutputProperties{
				IncrementalStrategy: drivers.IncrementalStrategyMerge,
				UniqueKey:           []string{"id"},
				OnSchemaChange:      &mode,
			},
			wantMode: mode,
		},
		{
			name: "eligible omitted defaults to fail",
			opts: schemaChangeModelExecuteOptions(true, true),
			props: ModelOutputProperties{
				IncrementalStrategy: drivers.IncrementalStrategyMerge,
				UniqueKey:           []string{"id"},
			},
			wantMode: OnSchemaChangeFail,
		},
		{
			name: "non incremental",
			opts: schemaChangeModelExecuteOptions(false, true),
			props: ModelOutputProperties{
				IncrementalStrategy: drivers.IncrementalStrategyMerge,
				UniqueKey:           []string{"id"},
				OnSchemaChange:      &mode,
			},
			wantErr: "incremental models with partitions",
		},
		{
			name: "without partitions",
			opts: schemaChangeModelExecuteOptions(true, false),
			props: ModelOutputProperties{
				IncrementalStrategy: drivers.IncrementalStrategyMerge,
				UniqueKey:           []string{"id"},
				OnSchemaChange:      &mode,
			},
			wantErr: "incremental models with partitions",
		},
		{
			name: "append strategy",
			opts: schemaChangeModelExecuteOptions(true, true),
			props: ModelOutputProperties{
				IncrementalStrategy: drivers.IncrementalStrategyAppend,
				OnSchemaChange:      &mode,
			},
			wantErr: "merge",
		},
		{
			name: "unknown mode",
			opts: schemaChangeModelExecuteOptions(true, true),
			props: ModelOutputProperties{
				IncrementalStrategy: drivers.IncrementalStrategyMerge,
				UniqueKey:           []string{"id"},
				OnSchemaChange:      schemaChangeModePtr("unknown"),
			},
			wantErr: "invalid on_schema_change mode",
		},
		{
			name: "ineligible omitted remains unspecified",
			opts: schemaChangeModelExecuteOptions(false, false),
			props: ModelOutputProperties{
				IncrementalStrategy: drivers.IncrementalStrategyAppend,
			},
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &ModelInputProperties{SQL: "SELECT 1"}
			err := tt.props.validateAndApplyDefaults(tt.opts, input, &tt.props)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			if tt.wantEmpty {
				require.Nil(t, tt.props.OnSchemaChange)
			} else {
				require.NotNil(t, tt.props.OnSchemaChange)
				require.Equal(t, tt.wantMode, *tt.props.OnSchemaChange)
			}
		})
	}
}

func schemaChangeModelExecuteOptions(incremental, partitionRun bool) *drivers.ModelExecuteOptions {
	return &drivers.ModelExecuteOptions{
		ModelExecutorOptions: &drivers.ModelExecutorOptions{
			InputConnector:  "duckdb",
			OutputConnector: "duckdb",
		},
		Incremental:  incremental,
		PartitionRun: partitionRun,
		PartitionKey: "partition",
	}
}

func schemaChangeModePtr(mode OnSchemaChange) *OnSchemaChange {
	return &mode
}
