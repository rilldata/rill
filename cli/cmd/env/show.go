package env

import (
	"os"

	"github.com/lensesio/tableprinter"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/variable"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ShowEnvCmd(cfg *config.Config) *cobra.Command {
	var projectName string
	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Show credentials and other variables",
		RunE: func(cmd *cobra.Command, args []string) error {
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
	showCmd.Flags().StringVar(&projectName, "project", "", "")
	_ = showCmd.MarkFlagRequired("project")
	return showCmd
}
