package billing

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func GetCreditsCmd(ch *cmdutil.Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-credits",
		Short: "Credits are now managed by Orb",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Credits are now managed by Orb. Check the Orb dashboard for credit balances.")
			return nil
		},
	}
	return cmd
}
