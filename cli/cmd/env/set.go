package env

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

// SetCmd is sub command for env. Sets the variable for a project
func SetCmd(cfg *config.Config) *cobra.Command {
	var projectName string
	setCmd := &cobra.Command{
		Use:   "set <key> <value>",
		Args:  cobra.ExactArgs(2),
		Short: "Set variable",
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]
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

			if val, ok := resp.Variables[key]; ok && val == value {
				return nil
			}

			if resp.Variables == nil {
				resp.Variables = make(map[string]string)
			}
			resp.Variables[key] = value
			_, err = client.UpdateProjectVariables(ctx, &adminv1.UpdateProjectVariablesRequest{
				OrganizationName: cfg.Org,
				Name:             projectName,
				Variables:        resp.Variables,
			})
			if err != nil {
				return err
			}

			cmdutil.SuccessPrinter("Updated project variables")
			return nil
		},
	}

	setCmd.Flags().StringVar(&projectName, "project", "", "")
	_ = setCmd.MarkFlagRequired("project")
	return setCmd
}
