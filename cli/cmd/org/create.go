package org

import (
	"fmt"
	"strings"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func CreateCmd(ch *cmdutil.Helper) *cobra.Command {
	var name, displayName, description string

	createCmd := &cobra.Command{
		Use:   "create <org-name>",
		Short: "Create organization",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			res, err := client.CreateOrganization(cmd.Context(), &adminv1.CreateOrganizationRequest{
				Name:        name,
				DisplayName: displayName,
				Description: description,
			})
			if err != nil {
				if !isNameExistsErr(err) {
					return err
				}
				return fmt.Errorf("an org with name %q already exists", name)
			}

			// Switching to the created org
			err = ch.DotRill.SetDefaultOrg(res.Organization.Name)
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Created org %q\n", res.Organization.Name)
			return nil
		},
	}
	createCmd.Flags().SortFlags = false
	createCmd.Flags().StringVar(&displayName, "display-name", "", "Display name")
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
