package env

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ShowCmd(ch *cmdutil.Helper) *cobra.Command {
	var projectPath, projectName, environment string

	showCmd := &cobra.Command{
		Use:   "show [<project-name>]",
		Short: "Show credentials and other variables",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			if len(args) > 0 {
				projectName = args[0]
			}

			// Find the cloud project name
			if projectName == "" {
				projectName, err = ch.InferProjectName(cmd.Context(), ch.Org, projectPath)
				if err != nil {
					return fmt.Errorf("unable to infer project name (use `--project` to explicitly specify the name): %w", err)
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

			var envVars []*variable

			for _, v := range resp.Variables {
				envVars = append(envVars, &variable{
					Name:        v.Name,
					Value:       v.Value,
					Environment: v.Environment,
				})
			}

			if cmd.Flags().Changed("env") {
				printEnv(envVars)
			} else {
				ch.PrintData(envVars)
			}

			return nil
		},
	}

	showCmd.Flags().StringVar(&projectName, "project", "", "Cloud project name (will attempt to infer from Git remote if not provided)")
	showCmd.Flags().StringVar(&projectPath, "path", ".", "Project directory")
	showCmd.Flags().StringVar(&environment, "environment", "", "Optional environment to resolve for (options: dev, prod)")
	showCmd.Flags().Bool("env", false, "Print variables in shell export format")

	return showCmd
}

func formatEnvVar(name, value string) string {
	return fmt.Sprintf("%s=%s", name, value)
}

func printEnv(vars []*variable) {
	for _, v := range vars {
		fmt.Println(formatEnvVar(v.Name, v.Value))
	}
}

type variable struct {
	Name        string `header:"name"`
	Value       string `header:"value"`
	Environment string `header:"environment"`
}
