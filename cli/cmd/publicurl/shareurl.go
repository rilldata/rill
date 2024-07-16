package publicurl

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func PublicURLCmd(ch *cmdutil.Helper) *cobra.Command {
	publicURLCmd := &cobra.Command{
		Use:               "public-url",
		Short:             "Manage public URLs",
		PersistentPreRunE: cmdutil.CheckChain(cmdutil.CheckAuth(ch), cmdutil.CheckOrganization(ch)),
	}

	publicURLCmd.PersistentFlags().StringVar(&ch.Org, "org", ch.Org, "Organization Name")
	publicURLCmd.AddCommand(ListCmd(ch))
	publicURLCmd.AddCommand(CreateCmd(ch))
	publicURLCmd.AddCommand(DeleteCmd(ch))

	return publicURLCmd
}
