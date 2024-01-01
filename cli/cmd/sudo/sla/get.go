package sla

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func GetCmd(ch *cmdutil.Helper) *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get",
		Args:  cobra.ExactArgs(2),
		Short: "Get SLA for project in an organization",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cfg := ch.Config

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()
			res, err := client.GetProject(ctx, &adminv1.GetProjectRequest{
				OrganizationName: args[0],
				Name:             args[1],
			})
			if err != nil {
				return err
			}

			sla := res.Project.ProdSla
			fmt.Printf("Project: %s\n", res.Project.Name)
			fmt.Printf("Organization: %s\n", res.Project.OrgName)
			fmt.Printf("SLA: %v\n", sla)

			return nil
		},
	}

	return getCmd
}
