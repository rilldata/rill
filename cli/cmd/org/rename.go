package org

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RenameCmd(cfg *config.Config) *cobra.Command {
	var org, newName string

	renameCmd := &cobra.Command{
		Use:   "rename",
		Args:  cobra.NoArgs,
		Short: "Rename",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if len(args) == 1 {
				return fmt.Errorf("Invalid args provided, required 0 or 2 args")
			}

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if !cmd.Flags().Changed("org") {
				// Get the new org name from user if not provided in the flag
				err := cmdutil.PromptIfUnset(&org, "Enter org to rename", org)
				if err != nil {
					return err
				}
			}

			if !cmd.Flags().Changed("new_name") {
				// Get the new org name from user if not provided in the flag
				err := cmdutil.PromptIfUnset(&newName, "Rename to", newName)
				if err != nil {
					return err
				}
			}

			exist, err := cmdutil.OrgExists(ctx, client, newName)
			if err != nil {
				return err
			}

			if exist {
				return fmt.Errorf("Org name %q already exists", newName)
			}

			fmt.Println("Warn: Renaming an org would invalidate dashboard URLs")

			confirm := false
			prompt := &survey.Confirm{
				Message: fmt.Sprintf("Do you want to rename org \"%s\" to \"%s\"?", color.YellowString(org), color.YellowString(newName)),
			}

			err = survey.AskOne(prompt, &confirm)
			if err != nil {
				return err
			}

			if !confirm {
				return nil
			}

			resp, err := client.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Name: org})
			if err != nil {
				return err
			}

			org := resp.Organization
			updatedOrg, err := client.UpdateOrganization(ctx, &adminv1.UpdateOrganizationRequest{
				Id:          org.Id,
				Name:        newName,
				Description: org.Description,
			})
			if err != nil {
				return err
			}

			err = dotrill.SetDefaultOrg(newName)
			if err != nil {
				return err
			}

			cmdutil.SuccessPrinter("Renamed organization\n")
			cmdutil.TablePrinter(toRow(updatedOrg.Organization))
			return nil
		},
	}
	renameCmd.Flags().SortFlags = false
	renameCmd.Flags().StringVar(&org, "org", cfg.Org, "Name")
	renameCmd.Flags().StringVar(&newName, "new_name", cfg.Org, "Description")

	return renameCmd
}
