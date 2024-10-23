package env

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ShowCmd(ch *cmdutil.Helper) *cobra.Command {
	var projectPath, projectName, environment string

	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Show credentials and other variables",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			// Find the cloud project name
			if projectName == "" {
				projectName, err = ch.InferProjectName(cmd.Context(), ch.Org, projectPath)
				if err != nil {
					return err
				}
			}

			resp, err := client.GetProjectVariables(cmd.Context(), &adminv1.GetProjectVariablesRequest{
				Organization:       ch.Org,
				Project:            projectName,
				Environment:        environment,
				ForAllEnvironments: !cmd.Flags().Changed("environment"),
			})
			if err != nil {
				return err
			}

			var table []*variable

			for _, v := range resp.Variables {
				table = append(table, &variable{
					Name:        v.Name,
					Value:       v.Value,
					Environment: v.Environment,
				})
			}

			ch.PrintfSuccess("\nVariables\n\n")
			ch.PrintData(table)

			return nil
		},
	}

	showCmd.Flags().StringVar(&projectName, "project", "", "Cloud project name (will attempt to infer from Git remote if not provided)")
	showCmd.Flags().StringVar(&projectPath, "path", ".", "Project directory")
	showCmd.Flags().StringVar(&environment, "environment", "", "Optional environment to resolve for (options: dev, prod)")

	return showCmd
}

type variable struct {
	Name        string `header:"name"`
	Value       string `header:"value"`
	Environment string `header:"environment"`
}
