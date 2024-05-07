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
	OutputHandle     Handle
	OutputConnector  string
	OutputProperties map[string]any
	IncrementalRun   bool
	PreviousResult   *ModelResult
}
