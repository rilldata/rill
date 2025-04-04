package local

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/runtime/parser"
)

func ParseDotenv(ctx context.Context, projectPath string) (map[string]string, error) {
	repo, instanceID, err := cmdutil.RepoForProjectPath(projectPath)
	if err != nil {
		return nil, err
	}
	p, err := parser.Parse(ctx, repo, instanceID, "prod", "duckdb")
	if err != nil {
		return nil, err
	}
	if p.RillYAML == nil {
		return nil, fmt.Errorf("not a valid Rill project (missing a rill.yaml file)")
	}

	return p.GetDotEnv(), nil
}
