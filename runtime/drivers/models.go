package drivers

import (
	"context"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// ModelExecutor executes models.
// A ModelExecutor may either be the a model's input or output connector.
type ModelExecutor interface {
	// Execute runs the model. The execution may be a full, incremental, or partition run.
	// For partition runs, Execute may be called concurrently by multiple workers.
	Execute(ctx context.Context, opts *ModelExecuteOptions) (*ModelResult, error)

	// Concurrency returns the number of concurrent calls that may be made to Execute given a user-provided desired concurrency.
	// If the desired concurrency is 0, it should return the recommended default concurrency.
	// If the desired concurrency is too high, it should return false.
	Concurrency(desired int) (int, bool)
}

// ModelManager manages model results returned by ModelExecutor.
// Unlike ModelExecutor, the result connector will always be used as the ModelManager.
type ModelManager interface {
	// Rename is called when a model is renamed, giving the ModelManager a chance to update state derived from the model's name (such as a table name).
	Rename(ctx context.Context, res *ModelResult, newName string, env *ModelEnv) (*ModelResult, error)

	// Exists returns whether the result still exists in the connector (for integrity checks).
	Exists(ctx context.Context, res *ModelResult) (bool, error)

	// Delete removes the result from the connector.
	Delete(ctx context.Context, res *ModelResult) error

	// MergePartitionResults merges two results produced by concurrent incremental partition runs.
	MergePartitionResults(a, b *ModelResult) (*ModelResult, error)
}

// ModelExecutorOptions are options passed when acquiring a ModelExecutor.
type ModelExecutorOptions struct {
	// Env contains contextual info about the model's instance.
	Env *ModelEnv
	// ModelName is the name of the model.
	ModelName string
	// InputHandle is the handle of the model's input connector.
	InputHandle Handle
	// InputConnector is the name of the model's input connector.
	InputConnector string
	// PreliminaryInputProperties are the preliminary properties of the model's input connector.
	// It may not always be available and may contain templating variables that have not yet been resolved.
	PreliminaryInputProperties map[string]any
	// OutputHandle is the handle of the model's output connector.
	OutputHandle Handle
	// OutputConnector is the name of the model's output connector.
	OutputConnector string
	// PreliminaryOutputProperties are the preliminary properties of the model's output connector.
	// It may not always be available and may contain templating variables that have not yet been resolved.
	PreliminaryOutputProperties map[string]any
}

// ModelExecuteOptions are options passed to a model executor's Execute function.
// They embed the ModelExecutorOptions that were used to initialize the ModelExecutor, plus additional options for the current execution step.
type ModelExecuteOptions struct {
	*ModelExecutorOptions
	// InputProperties are the resolved properties of the model's input connector.
	InputProperties map[string]any
	// OutputProperties are the resolved properties of the model's output connector.
	OutputProperties map[string]any
	// Priority is the priority of the model execution.
	Priority int
	// Incremental is true if the model is an incremental model.
	Incremental bool
	// IncrementalRun is true if the execution is an incremental run.
	IncrementalRun bool
	// PartitionRun is true if the execution is a partition run.
	PartitionRun bool
	// PartitionKey is the unique key for the partition currently being run.
	// It is empty when PartitionRun is false.
	PartitionKey string
	// PreviousResult is the result of a previous execution.
	// For concurrent partition execution, it may not be the most recent previous result.
	PreviousResult *ModelResult
	// TempDir is a temporary directory for storing intermediate data.
	TempDir string
}

// ModelEnv contains contextual info about the model's instance.
type ModelEnv struct {
	AllowHostAccess    bool
	RepoRoot           string
	StageChanges       bool
	DefaultMaterialize bool
	Connectors         []*runtimev1.Connector
	AcquireConnector   func(ctx context.Context, name string) (Handle, func(), error)
}

// ModelResult contains metadata about the result of a model execution.
type ModelResult struct {
	Connector    string
	Properties   map[string]any
	Table        string
	ExecDuration time.Duration
}

// IncrementalStrategy is a strategy to use for incrementally inserting data into a SQL table.
type IncrementalStrategy string

const (
	IncrementalStrategyUnspecified        IncrementalStrategy = ""
	IncrementalStrategyAppend             IncrementalStrategy = "append"
	IncrementalStrategyMerge              IncrementalStrategy = "merge"
	IncrementalStrategyPartitionOverwrite IncrementalStrategy = "partition_overwrite"
)

// FileFormat is a file format for importing or exporting data.
type FileFormat string

const (
	FileFormatUnspecified FileFormat = ""
	FileFormatParquet     FileFormat = "parquet"
	FileFormatCSV         FileFormat = "csv"
	FileFormatJSON        FileFormat = "json"
	FileFormatXLSX        FileFormat = "xlsx"
)

func (f FileFormat) Filename(stem string) string {
	return stem + "." + string(f)
}

func (f FileFormat) Valid() bool {
	switch f {
	case FileFormatParquet, FileFormatCSV, FileFormatJSON, FileFormatXLSX:
		return true
	}
	return false
}
