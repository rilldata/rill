package admin

import (
	"github.com/rilldata/rill/cli/pkg/version"
	"github.com/spf13/cobra"
)

// StartCmd starts an admin server. It only allows configuration using environment variables.
func StartCmd(ver version.Version) *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start admin server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return startCmd
}
