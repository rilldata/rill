package user

import (
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func AssumeCmd(cfg *config.Config) *cobra.Command {
	assumeCmd := &cobra.Command{
		Use:   "assume <email>",
		Args:  cobra.ExactArgs(1),
		Short: "Assume users by email",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return assumeCmd
}
