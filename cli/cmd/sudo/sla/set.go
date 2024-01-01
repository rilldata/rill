package sla

import (
	"fmt"
	"strconv"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func SetCmd(ch *cmdutil.Helper) *cobra.Command {
	setCmd := &cobra.Command{
		Use:   "set <org> <project> {true,false}",
		Args:  cobra.ExactArgs(3),
		Short: "Set SLA for project in an organization",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cfg := ch.Config

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			// Convert thirdArg to a boolean value
			sla, err := strconv.ParseBool(args[2])
			if err != nil {
				ch.Printer.PrintlnError("Third argument (SLA) must be 'true' or 'false'")
				return err
			}

			res, err := client.SudoUpdateSLA(ctx, &adminv1.SudoUpdateSLARequest{
				Organization: args[0],
				Project:      args[1],
				Sla:          sla,
			})
			if err != nil {
				return err
			}

			sla = res.Project.ProdSla
			fmt.Printf("Project: %s\n", res.Project.Name)
			fmt.Printf("Organization: %s\n", res.Project.OrgName)
			fmt.Printf("SLA: %v\n", sla)

			return nil
		},
	}

	return setCmd
}
