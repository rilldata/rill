package org

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func SwitchCmd(cfg *config.Config) *cobra.Command {
	switchCmd := &cobra.Command{
		Use:   "switch",
		Short: "Switch",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := client.New(cfg.AdminURL, cfg.AdminToken())
			if err != nil {
				return err
			}
			defer client.Close()

			res, err := client.ListOrganizations(context.Background(), &adminv1.ListOrganizationsRequest{})
			if err != nil {
				return err
			}

			if len(res.Organizations) < 1 {
				fmt.Println("No organizations found, run `rill org create` first.")
				return nil
			}

			var orgNames []string
			for _, org := range res.Organizations {
				orgNames = append(orgNames, org.Name)
			}

			defaultOrg := cmdutil.PromptGetSelect(orgNames, "Select default org.")

			err = dotrill.SetDefaultOrg(defaultOrg)
			if err != nil {
				return err
			}

			fmt.Printf("Set default organization to %q.\n", defaultOrg)
			return nil
		},
	}

	return switchCmd
}
