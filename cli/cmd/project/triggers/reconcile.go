package triggers

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func ReconcileCmd(cfg *config.Config) *cobra.Command {
	reconcileCmd := &cobra.Command{
		Use:   "reconcile",
		Short: "Reconcile",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("reconcile called")
			return nil
		},
	}

	return reconcileCmd
}
