package org

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RenameCmd(ch *cmdutil.Helper) *cobra.Command {
	var name, newName string
	var force bool

	renameCmd := &cobra.Command{
		Use:   "rename",
		Args:  cobra.NoArgs,
		Short: "Rename organization",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			fmt.Println("Warn: Renaming an org would invalidate dashboard URLs")

			if !cmd.Flags().Changed("org") && ch.Interactive {
				orgNames, err := orgNames(ctx, ch)
				if err != nil {
					return err
				}

				name = cmdutil.SelectPrompt("Select org to rename", orgNames, "")
			}

			if ch.Interactive {
				err = cmdutil.SetFlagsByInputPrompts(*cmd, "new-name")
				if err != nil {
					return err
				}
			}

			if newName == "" {
				return fmt.Errorf("please provide valid org new-name, provided: %q", newName)
			}

			if !force {
				msg := fmt.Sprintf("Do you want to rename org \"%s\" to \"%s\"?", color.YellowString(name), color.YellowString(newName))
				if !cmdutil.ConfirmPrompt(msg, "", false) {
					return nil
				}
			}

			updatedOrg, err := client.UpdateOrganization(ctx, &adminv1.UpdateOrganizationRequest{
				Name:    name,
				NewName: &newName,
			})
			if err != nil {
				return err
			}

			err = dotrill.SetDefaultOrg(newName)
			if err != nil {
				return err
			}

			ch.Printer.PrintlnSuccess("Renamed organization")
			return ch.Printer.PrintResource([]*organization{toRow(updatedOrg.Organization)})
		},
	}
	renameCmd.Flags().SortFlags = false
	renameCmd.Flags().StringVar(&name, "org", ch.Org, "Current Org Name")
	renameCmd.Flags().StringVar(&newName, "new-name", "", "New Org Name")
	renameCmd.Flags().BoolVar(&force, "force", false, "Force rename org without confirmation prompt")

	return renameCmd
}
