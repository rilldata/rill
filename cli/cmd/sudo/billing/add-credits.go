package billing

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func AddCreditsCmd(ch *cmdutil.Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-credits",
		Short: "Credits are now managed by Orb",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Credits are now managed by Orb. Use the Orb dashboard to add credits.")
			return nil
		},
	}
	return cmd
}
