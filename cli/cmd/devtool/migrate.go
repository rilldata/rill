package devtool

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func MigrateCmd(ch *cmdutil.Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Apply migrations to the database",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("MigrateCmd")
			return nil
		},
	}
	return cmd
}
