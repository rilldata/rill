package project

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ResetCmd(ch *cmdutil.Helper) *cobra.Command {
	var force bool

	resetCmd := &cobra.Command{
		Use:   "reset <org> <project>",
		Args:  cobra.ExactArgs(2),
		Short: "Re-deploy the project",
		Long:  "Create a new deployment for the project (and tear down the current one)",
		RunE: func(cmd *cobra.Command, args []string) error {
			org := args[0]
			project := args[1]

			client, err := ch.Client()
			if err != nil {
				return err
			}

			if !force {
				ch.PrintfWarn("The project will be unavailable for a while as data sources are reloaded from scratch. If you just need to refresh data, use `rill project refresh`.\n")
				ok, err := cmdutil.ConfirmPrompt("Continue?", "", false)
				if err != nil {
					return err
				}
				if !ok {
					return nil
				}
			}

			_, err = client.RedeployProject(cmd.Context(), &adminv1.RedeployProjectRequest{
				Org:                  org,
				Project:              project,
				SuperuserForceAccess: true,
			})
			if err != nil {
				return err
			}

			ch.Printf("Triggered reset of %q.\n", project)

			return nil
		},
	}

	resetCmd.Flags().SortFlags = false
	resetCmd.Flags().BoolVar(&force, "force", false, "Force reset even if project is already deployed")
	return resetCmd
}
