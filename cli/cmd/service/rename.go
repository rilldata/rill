package service

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func RenameCmd(cfg *config.Config) *cobra.Command {
	var name, newName string

	renameCmd := &cobra.Command{
		Use:   "rename",
		Args:  cobra.NoArgs,
		Short: "Rename service",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			fmt.Println("not implemented yet")
			return nil
		},
	}
	renameCmd.Flags().SortFlags = false
	renameCmd.Flags().StringVar(&name, "org", cfg.Org, "Current Service Name")
	renameCmd.Flags().StringVar(&newName, "new-name", "", "New Service Name")

	return renameCmd
}
