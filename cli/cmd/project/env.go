package project

import (
	"os"

	"github.com/lensesio/tableprinter"
	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/variable"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

// EnvCmd sets/rm variables for a project
func EnvCmd(cfg *config.Config) *cobra.Command {
	envCmd := &cobra.Command{
		Use:   "env",
		Short: "Manage variables for a project",
	}
	envCmd.AddCommand(RmCmd(cfg))
	envCmd.AddCommand(SetCmd(cfg))
	envCmd.AddCommand(ShowEnvCmd(cfg))
	return envCmd
}

// SetCmd is sub command for env. Sets the variable for a project
func SetCmd(cfg *config.Config) *cobra.Command {
	setCmd := &cobra.Command{
		Use:   "set <project name> <key> <value>",
		Args:  cobra.ExactArgs(3),
		Short: "set variable",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]
			key := args[1]
			value := args[2]
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
			updateResp, err := client.UpdateProjectVariables(ctx, &adminv1.UpdateProjectVariablesRequest{
				OrganizationName: cfg.Org,
				Name:             projectName,
				Variables:        resp.Variables,
			})
			if err != nil {
				return err
			}

			cmdutil.SuccessPrinter("Updated project variables\n")
			tableprinter.PrintHeadList(os.Stdout, variable.Serialize(updateResp.Variables), "Project Variables")
			return nil
		},
	}
	return setCmd
}

// RmCmd is sub command for env. Removes the variable for a project
func RmCmd(cfg *config.Config) *cobra.Command {
	rmCmd := &cobra.Command{
		Use:   "rm <project name> <key>",
		Args:  cobra.ExactArgs(2),
		Short: "remove variable",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]
			key := args[1]
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
			update, err := client.UpdateProjectVariables(ctx, &adminv1.UpdateProjectVariablesRequest{
				OrganizationName: cfg.Org,
				Name:             projectName,
				Variables:        resp.Variables,
			})
			if err != nil {
				return err
			}

			cmdutil.SuccessPrinter("Updated project \n")
			tableprinter.PrintHeadList(os.Stdout, variable.Serialize(update.Variables), "Project Variables")
			return nil
		},
	}
	return rmCmd
}

func ShowEnvCmd(cfg *config.Config) *cobra.Command {
	showCmd := &cobra.Command{
		Use:   "show <project name>",
		Args:  cobra.ExactArgs(1),
		Short: "show variable for project",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			resp, err := client.GetProjectVariables(cmd.Context(), &adminv1.GetProjectVariablesRequest{
				OrganizationName: cfg.Org,
				Name:             projectName,
			})
			if err != nil {
				return err
			}

			tableprinter.PrintHeadList(os.Stdout, variable.Serialize(resp.Variables), "Project Variables")
			return nil
		},
	}
	return showCmd
}
