package org

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func SwitchCmd(cfg *config.Config) *cobra.Command {
	switchCmd := &cobra.Command{
		Use:   "switch",
		Short: "Switch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := client.New(cfg.AdminURL, cfg.AdminToken())
			if err != nil {
				return err
			}
			defer client.Close()

			_, err = client.GetOrganization(context.Background(), &adminv1.GetOrganizationRequest{
				Name: args[0],
			})
			if err != nil {
				return err
			}

			err = dotrill.SetDefaultOrg(args[0])
			if err != nil {
				return err
			}

			fmt.Printf("Set default organization to %q", args[0])
			return nil
		},
	}

	return switchCmd
}
