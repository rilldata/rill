package drivers

import "context"

type ModelExecutor interface {
	Execute(ctx context.Context) (*ModelResult, error)
}

type ModelManager interface {
	Rename(ctx context.Context, res *ModelResult, newName string, env *ModelEnv) (*ModelResult, error)
	Exists(ctx context.Context, res *ModelResult) (bool, error)
	Delete(ctx context.Context, res *ModelResult) error
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

type ModelExecutorOptions struct {
	Env              *ModelEnv
	ModelName        string
	InputHandle      Handle
	InputConnector   string
	InputProperties  map[string]any
	StageConnector   string
	StageProperties  map[string]any
	OutputHandle     Handle
	OutputConnector  string
	OutputProperties map[string]any
	Priority         int
	Incremental      bool
	IncrementalRun   bool
	PreviousResult   *ModelResult
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
