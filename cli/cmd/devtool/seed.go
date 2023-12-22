package devtool

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func SeedCmd(ch *cmdutil.Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "seed {cloud}",
		Short: "Authenticate and deploy a seed project",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Do `rill login`

			// TODO: Deploy the seed project

			return nil
		},
	}

	return cmd
}
