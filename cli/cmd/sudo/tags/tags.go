package tags

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func TagsCmd(ch *cmdutil.Helper) *cobra.Command {
	tagsCmd := &cobra.Command{
		Use:   "tags",
		Short: "Manage Tags for project in an organization",
	}

	tagsCmd.AddCommand(GetCmd(ch))
	tagsCmd.AddCommand(SetCmd(ch))

	return tagsCmd
}
