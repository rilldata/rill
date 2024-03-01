package env

import (
	"path/filepath"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/spf13/cobra"
)

func EnvCmd(ch *cmdutil.Helper) *cobra.Command {
	envCmd := &cobra.Command{
		Use:               "env",
		Short:             "Manage variables for a project",
		PersistentPreRunE: cmdutil.CheckChain(cmdutil.CheckAuth(ch), cmdutil.CheckOrganization(ch)),
	}
	envCmd.AddCommand(ConfigureCmd(ch))
	envCmd.AddCommand(PullCmd(ch))
	envCmd.AddCommand(PushCmd(ch))
	envCmd.AddCommand(SetCmd(ch))
	envCmd.AddCommand(RmCmd(ch))
	envCmd.AddCommand(ShowCmd(ch))
	return envCmd
}

// normalizeProjectPath normalizes a project path provided by the user
func normalizeProjectPath(projectPath string) (string, error) {
	res, err := fileutil.ExpandHome(projectPath)
	if err != nil {
		return "", err
	}
	res, err = filepath.Abs(res)
	if err != nil {
		return "", err
	}
	return res, nil
}
