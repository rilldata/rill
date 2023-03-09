package org

import (
	"context"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ShowCmd(cfg *config.Config) *cobra.Command {
	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Show",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sp := cmdutil.Spinner("Switching org...")
			sp.Start()

			client, err := client.New(cfg.AdminURL, cfg.AdminToken())
			if err != nil {
				return err
			}
			defer client.Close()

			org, err := client.GetOrganization(context.Background(), &adminv1.GetOrganizationRequest{
				Name: args[0],
			})
			if err != nil {
				return err
			}

			sp.Stop()
			cmdutil.TextPrinter("Found organization \n")
			cmdutil.TablePrinter(toOrg(org.Organization))
			return nil
		},
	}

	return showCmd
}
