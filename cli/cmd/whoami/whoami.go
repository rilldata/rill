package whoami

import (
	"github.com/spf13/cobra"
)

// VersionCmd represents the version command
func WhoamiCmd() *cobra.Command {
	whoamiCmd := &cobra.Command{
		Use:   "whoami",
		Short: "Show current user",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return whoamiCmd
}
