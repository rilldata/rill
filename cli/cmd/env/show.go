package env

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/variable"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ShowEnvCmd(ch *cmdutil.Helper) *cobra.Command {
	var projectName string
	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Show credentials and other variables",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			resp, err := client.GetProjectVariables(cmd.Context(), &adminv1.GetProjectVariablesRequest{
				OrganizationName: ch.Org,
				Name:             projectName,
			})
			if err != nil {
				return err
			}

			vals := variable.Serialize(resp.Variables)
			for _, v := range vals {
				fmt.Println(v)
			}

			return nil
		},
	}

	showCmd.Flags().StringVar(&projectName, "project", "", "")
	_ = showCmd.MarkFlagRequired("project")

	return showCmd
}
