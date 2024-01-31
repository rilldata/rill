package project

import (
	"context"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func JwtCmd(ch *cmdutil.Helper) *cobra.Command {
	var name string
	cfg := ch.Config

	jwtCmd := &cobra.Command{
		Use:    "jwt [<project-name>]",
		Args:   cobra.MaximumNArgs(1),
		Short:  "Generate the token for connecting directly to the deployment",
		Hidden: !cfg.IsDev(),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if len(args) > 0 {
				name = args[0]
			}

			if !cmd.Flags().Changed("project") && len(args) == 0 && cfg.Interactive {
				names, err := cmdutil.ProjectNamesByOrg(ctx, client, cfg.Org)
				if err != nil {
					return err
				}

				// prompt for name from user
				name = cmdutil.SelectPrompt("Select project", names, "")
			}

			res, err := client.GetProject(context.Background(), &adminv1.GetProjectRequest{
				OrganizationName: cfg.Org,
				Name:             name,
			})
			if err != nil {
				return err
			}
			if res.ProdDeployment == nil {
				ch.Printer.PrintlnWarn("Project does not have a production deployment")
				return nil
			}

			ch.Printer.PrintlnSuccess("Runtime info")
			ch.Printer.Printf("  Host: %s\n", res.ProdDeployment.RuntimeHost)
			ch.Printer.Printf("  Instance: %s\n", res.ProdDeployment.RuntimeInstanceId)
			ch.Printer.Printf("  JWT: %s\n", res.Jwt)

			return nil
		},
	}

	jwtCmd.Flags().SortFlags = false
	jwtCmd.Flags().StringVar(&name, "project", "", "Project Name")

	return jwtCmd
}
