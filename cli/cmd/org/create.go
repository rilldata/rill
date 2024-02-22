package org

import (
	"context"
	"fmt"
	"strings"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func CreateCmd(ch *cmdutil.Helper) *cobra.Command {
	var name, description string

	createCmd := &cobra.Command{
		Use:   "create [<org-name>]",
		Short: "Create organization",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			if len(args) > 0 {
				name = args[0]
			}

			if len(args) == 0 && ch.Interactive {
				err = cmdutil.SetFlagsByInputPrompts(*cmd, "name")
				if err != nil {
					return err
				}
			}

			res, err := client.CreateOrganization(context.Background(), &adminv1.CreateOrganizationRequest{
				Name:        name,
				Description: description,
			})
			if err != nil {
				if !isNameExistsErr(err) {
					return err
				}

				fmt.Printf("Org name %q already exists\n", name)
				return nil
			}

			// Switching to the created org
			err = dotrill.SetDefaultOrg(res.Organization.Name)
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Created organization\n")
			ch.PrintOrgs([]*adminv1.Organization{res.Organization}, "")

			return nil
		},
	}
	createCmd.Flags().SortFlags = false
	createCmd.Flags().StringVar(&name, "name", "", "Organization Name")
	createCmd.Flags().StringVar(&description, "description", "", "Description")
	return createCmd
}

func isNameExistsErr(err error) bool {
	if strings.Contains(err.Error(), "already exists") {
		return true
	}
	if strings.Contains(err.Error(), "violates unique constraint") {
		return true
	}
	return false
}
