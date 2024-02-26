package project

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func JwtCmd(ch *cmdutil.Helper) *cobra.Command {
	var name string

	jwtCmd := &cobra.Command{
		Use:    "jwt [<project-name>]",
		Args:   cobra.MaximumNArgs(1),
		Short:  "Generate the token for connecting directly to the deployment",
		Hidden: !ch.IsDev(),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := ch.Client()
			if err != nil {
				return err
			}

			if len(args) > 0 {
				name = args[0]
			}

			if !cmd.Flags().Changed("project") && len(args) == 0 && ch.Interactive {
				names, err := projectNames(ctx, ch)
				if err != nil {
					return err
				}

				// prompt for name from user
				name = cmdutil.SelectPrompt("Select project", names, "")
			}

			res, err := client.GetProject(cmd.Context(), &adminv1.GetProjectRequest{
				OrganizationName: ch.Org,
				Name:             name,
			})
			if err != nil {
				return err
			}
			if res.ProdDeployment == nil {
				ch.PrintfWarn("Project does not have a production deployment\n")
				return nil
			}

			ch.Printf("Runtime info\n")
			ch.Printf("  Host: %s\n", res.ProdDeployment.RuntimeHost)
			ch.Printf("  Instance: %s\n", res.ProdDeployment.RuntimeInstanceId)
			ch.Printf("  JWT: %s\n", res.Jwt)

			return nil
		},
	}

	jwtCmd.Flags().SortFlags = false
	jwtCmd.Flags().StringVar(&name, "project", "", "Project Name")

	return jwtCmd
}
