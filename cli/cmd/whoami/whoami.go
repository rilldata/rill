package whoami

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

// VersionCmd represents the version command
func WhoamiCmd(ch *cmdutil.Helper) *cobra.Command {
	whoamiCmd := &cobra.Command{
		Use:               "whoami",
		Short:             "Show current user",
		PersistentPreRunE: cmdutil.CheckAuth(ch),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			res, err := client.GetCurrentUser(cmd.Context(), &adminv1.GetCurrentUserRequest{})
			if err != nil {
				return err
			}

			fmt.Printf("Email: %s\n", res.User.Email)
			fmt.Printf("Name: %s\n", res.User.DisplayName)

			return nil
		},
	}

	return whoamiCmd
}
