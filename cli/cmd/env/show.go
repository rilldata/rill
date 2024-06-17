package env

import (
	"github.com/joho/godotenv"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ShowCmd(ch *cmdutil.Helper) *cobra.Command {
	var projectPath, projectName string

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
				OrganizationName: ch.Org,
				Name:             projectName,
			})
			if err != nil {
				return err
			}

			res, err := godotenv.Marshal(resp.Variables)
			if err != nil {
				return err
			}

			ch.Println(res)

			return nil
		},
	}

	showCmd.Flags().StringVar(&projectName, "project", "", "Cloud project name (will attempt to infer from Git remote if not provided)")
	showCmd.Flags().StringVar(&projectPath, "path", ".", "Project directory")

	return showCmd
}
