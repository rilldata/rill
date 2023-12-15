package user

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func UserCmd(ch *cmdutil.Helper) *cobra.Command {
	userCmd := &cobra.Command{
		Use:   "user",
		Short: "Manage users",
	}

	userCmd.AddCommand(ListCmd(ch))
	userCmd.AddCommand(SearchCmd(ch))
	userCmd.AddCommand(AssumeCmd(ch))
	userCmd.AddCommand(UnassumeCmd(ch))
	userCmd.AddCommand(OpenCmd(ch))

	return userCmd
}
