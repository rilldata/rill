package org

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func SwitchCmd(ch *cmdutil.Helper) *cobra.Command {
	switchCmd := &cobra.Command{
		Use:   "switch [<org-name>]",
		Short: "Switch to other organization",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			var defaultOrg string
			if len(args) == 0 {
				res, err := client.ListOrganizations(cmd.Context(), &adminv1.ListOrganizationsRequest{
					PageSize: 1000,
				})
				if err != nil {
					return err
				}

				defaultOrg, err = SwitchSelectFlow(res.Organizations)
				if err != nil {
					return err
				}
			} else {
				_, err = client.GetOrganization(cmd.Context(), &adminv1.GetOrganizationRequest{
					Name: args[0],
				})
				if err != nil {
					return err
				}
				defaultOrg = args[0]
			}

			err = dotrill.SetDefaultOrg(defaultOrg)
			if err != nil {
				return err
			}
			ch.Org = defaultOrg

			ch.Printf("Set default organization to %q.\n", defaultOrg)
			return nil
		},
	}

	return switchCmd
}

func SwitchSelectFlow(orgs []*adminv1.Organization) (string, error) {
	if len(orgs) < 1 {
		fmt.Println("No organizations found, run `rill org create` first.")
		return "", nil
	}

	var orgNames []string
	for _, org := range orgs {
		orgNames = append(orgNames, org.Name)
	}

	org, err := dotrill.GetDefaultOrg()
	if err != nil {
		return "", err
	}

	return cmdutil.SelectPrompt("Select default org.", orgNames, org)
}

// SetDefaultOrg sets a default org for the user if user is part of any org.
func SetDefaultOrg(ctx context.Context, ch *cmdutil.Helper) error {
	c, err := ch.Client()
	if err != nil {
		return err
	}

	res, err := c.ListOrganizations(ctx, &adminv1.ListOrganizationsRequest{})
	if err != nil {
		return fmt.Errorf("listing orgs failed: %w", err)
	}

	if len(res.Organizations) == 1 {
		ch.Org = res.Organizations[0].Name
		if err := dotrill.SetDefaultOrg(ch.Org); err != nil {
			return err
		}
	} else if len(res.Organizations) > 1 {
		orgName, err := SwitchSelectFlow(res.Organizations)
		if err != nil {
			return fmt.Errorf("org selection failed %w", err)
		}

		ch.Org = orgName
		if err := dotrill.SetDefaultOrg(ch.Org); err != nil {
			return err
		}
	}
	return nil
}
