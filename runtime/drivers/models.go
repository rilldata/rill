package drivers

import "context"

type ModelExecutor interface {
	Execute(ctx context.Context, opts *ModelExecuteOptions) (*ModelResult, error)
	Concurrency(desired int) (int, bool)
}

type ModelManager interface {
	Rename(ctx context.Context, res *ModelResult, newName string, env *ModelEnv) (*ModelResult, error)
	Exists(ctx context.Context, res *ModelResult) (bool, error)
	Delete(ctx context.Context, res *ModelResult) error
	MergeSplitResults(a, b *ModelResult) (*ModelResult, error)
}

type ModelExecutorOptions struct {
	Env                         *ModelEnv
	ModelName                   string
	InputHandle                 Handle
	InputConnector              string
	PreliminaryInputProperties  map[string]any
	OutputHandle                Handle
	OutputConnector             string
	PreliminaryOutputProperties map[string]any
}

type ModelExecuteOptions struct {
	*ModelExecutorOptions
	InputProperties  map[string]any
	OutputProperties map[string]any
	Priority         int
	Incremental      bool
	IncrementalRun   bool
	SplitRun         bool
	PreviousResult   *ModelResult
}

type ModelEnv struct {
	AllowHostAccess    bool
	RepoRoot           string
	StageChanges       bool
	DefaultMaterialize bool
	AcquireConnector   func(ctx context.Context, name string) (Handle, func(), error)
}

type ModelResult struct {
	Connector  string
	Properties map[string]any
	Table      string
}

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
