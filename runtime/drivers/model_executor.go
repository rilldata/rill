package drivers

import "context"

// ModelExecutor implements logic for building and managing a model.
type ModelExecutor interface {
	Supports(ctx context.Context, opts *ModelExecuteOptions) (bool, error)
	Run(ctx context.Context, opts *ModelExecuteOptions) (*ModelExecuteResult, error)
	Rename(ctx context.Context, opts *ModelRenameOptions) (*ModelExecuteResult, error)
	Exists(ctx context.Context, res *ModelExecuteResult) (bool, error)
	Delete(ctx context.Context, res *ModelExecuteResult) error
}

type ModelExecuteOptions struct {
	ModelName        string
	Env              *ModelExecutorEnv
	PreviousResult   *ModelExecuteResult
	Incremental      bool
	InputConnector   string
	InputProperties  map[string]any
	OutputConnector  string
	OutputProperties map[string]any
}

type ModelExecuteResult struct {
	Connector  string
	Properties map[string]any
	Table      string
}

type ModelRenameOptions struct {
	NewName        string
	PreviousName   string
	PreviousResult *ModelExecuteResult
	Env            *ModelExecutorEnv
}

type ModelExecutorEnv struct {
	AllowHostAccess    bool
	RepoRoot           string
	StageChanges       bool
	DefaultMaterialize bool
	AcquireConnector   func(ctx context.Context, name string) (Handle, func(), error)
}
