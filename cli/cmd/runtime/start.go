package runtime

import (
	"github.com/rilldata/rill/cli/pkg/version"
	"github.com/spf13/cobra"
)

// StartCmd starts a stand-alone runtime server. It only allows configuration using environment variables.
func StartCmd(ver version.Version) *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start stand-alone runtime server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return startCmd
}
