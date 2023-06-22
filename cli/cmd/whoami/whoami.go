package whoami

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

// VersionCmd represents the version command
func WhoamiCmd(cfg *config.Config) *cobra.Command {
	whoamiCmd := &cobra.Command{
		Use:   "whoami",
		Short: "Show current user",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			res, err := client.GetCurrentUser(cmd.Context(), &adminv1.GetCurrentUserRequest{})
			if err != nil {
				return err
			}

			fmt.Printf("  You Logged in as : %s\n", res.User.Email)

			return nil
		},
	}

	return whoamiCmd
}
