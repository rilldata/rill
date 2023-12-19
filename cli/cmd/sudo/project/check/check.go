package check

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func CheckCmd(ch *cmdutil.Helper) *cobra.Command {
	checkCmd := &cobra.Command{
		Use:   "project",
		Short: "Project search for support users",
	}

	checkCmd.AddCommand(HealthCmd(ch))

	return checkCmd
}
