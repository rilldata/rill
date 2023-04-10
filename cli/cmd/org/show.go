package org

import (
	"context"
	"fmt"
	"os"

	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ShowCmd(cfg *config.Config) *cobra.Command {
	showCmd := &cobra.Command{
		Use:   "show [<org-name>]",
		Short: "Show",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var name string
			if len(args) == 0 {
				name = cfg.Org
				if name == "" {
					fmt.Printf("No organization is set. Run 'rill org create org-name' to create one.")
					os.Exit(1)
				}
			} else {
				name = args[0]
			}

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			org, err := client.GetOrganization(context.Background(), &adminv1.GetOrganizationRequest{
				Name: name,
			})
			if err != nil {
				return err
			}

			cmdutil.SuccessPrinter("Found organization \n")
			cmdutil.TablePrinter(toRow(org.Organization))
			return nil
		},
	}

	return showCmd
}
