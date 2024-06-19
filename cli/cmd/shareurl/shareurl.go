package shareurl

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func ShareURLCmd(ch *cmdutil.Helper) *cobra.Command {
	shareURLCmd := &cobra.Command{
		Use:               "share-url",
		Short:             "Manage shareable URLs",
		PersistentPreRunE: cmdutil.CheckChain(cmdutil.CheckAuth(ch), cmdutil.CheckOrganization(ch)),
	}

	shareURLCmd.PersistentFlags().StringVar(&ch.Org, "org", ch.Org, "Organization Name")
	shareURLCmd.AddCommand(ListCmd(ch))
	shareURLCmd.AddCommand(CreateCmd(ch))
	shareURLCmd.AddCommand(DeleteCmd(ch))

	return shareURLCmd
}
