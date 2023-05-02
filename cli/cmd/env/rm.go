package env

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

// RmCmd is sub command for env. Removes the variable for a project
func RmCmd(cfg *config.Config) *cobra.Command {
	var projectName string
	rmCmd := &cobra.Command{
		Use:   "rm <key>",
		Args:  cobra.ExactArgs(1),
		Short: "Remove variable",
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := cmd.Context()
			resp, err := client.GetProjectVariables(ctx, &adminv1.GetProjectVariablesRequest{
				OrganizationName: cfg.Org,
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
				OrganizationName: cfg.Org,
				Name:             projectName,
				Variables:        resp.Variables,
			})
			if err != nil {
				return err
			}

			cmdutil.SuccessPrinter("Updated project")
			return nil
		},
	}
	rmCmd.Flags().StringVar(&projectName, "project", "", "")
	_ = rmCmd.MarkFlagRequired("project")
	return rmCmd
}
