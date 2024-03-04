package env

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

// RmCmd is sub command for env. Removes the variable for a project
func RmCmd(ch *cmdutil.Helper) *cobra.Command {
	var projectName string
	rmCmd := &cobra.Command{
		Use:   "rm <key>",
		Args:  cobra.ExactArgs(1),
		Short: "Remove variable",
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			client, err := ch.Client()
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			resp, err := client.GetProjectVariables(ctx, &adminv1.GetProjectVariablesRequest{
				OrganizationName: ch.Org,
				Name:             projectName,
			})
			if err != nil {
				return err
			}

			if _, ok := resp.Variables[key]; !ok {
				return nil
			}

			delete(resp.Variables, key)
			_, err = client.UpdateProjectVariables(ctx, &adminv1.UpdateProjectVariablesRequest{
				OrganizationName: ch.Org,
				Name:             projectName,
				Variables:        resp.Variables,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Updated project\n")
			return nil
		},
	}
	rmCmd.Flags().StringVar(&projectName, "project", "", "")
	_ = rmCmd.MarkFlagRequired("project")
	return rmCmd
}
