package superuser

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func SuperuserCmd(ch *cmdutil.Helper) *cobra.Command {
	superuserCmd := &cobra.Command{
		Use:   "superuser",
		Short: "Manage superusers",
	}

	superuserCmd.AddCommand(ListCmd(ch))
	superuserCmd.AddCommand(AddCmd(ch))
	superuserCmd.AddCommand(RemoveCmd(ch))

	return superuserCmd
}
