package org

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RenameCmd(cfg *config.Config) *cobra.Command {
	renameCmd := &cobra.Command{
		Use:   "rename <org-name>",
		Args:  cobra.MaximumNArgs(1),
		Short: "Rename",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			var currentName string
			var newName string

			if len(args) == 1 {
				return fmt.Errorf("Invalid args provided, required 0 or 2 args")
			}

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if len(args) > 1 {
				currentName = args[0]
				newName = args[1]
			} else {
				resp, err := client.ListOrganizations(ctx, &adminv1.ListOrganizationsRequest{})
				if err != nil {
					return err
				}

				var orgNames []string
				for _, org := range resp.Organizations {
					orgNames = append(orgNames, org.Name)
				}

				currentName = cmdutil.SelectPrompt("Select the org for rename", orgNames, "")

				// Get the new org name from user if not provided in the args
				questions := []*survey.Question{
					{
						Name: "name",
						Prompt: &survey.Input{
							Message: "New org name",
						},
						Validate: func(any interface{}) error {
							name := any.(string)
							if name == "" {
								return fmt.Errorf("empty name")
							}

							return nil
						},
					},
				}

				if err := survey.Ask(questions, &newName); err != nil {
					return err
				}
			}

			confirm := false
			prompt := &survey.Confirm{
				Message: fmt.Sprintf("Do you want to rename org %q to %q?", color.YellowString(currentName), color.YellowString(newName)),
			}

			err = survey.AskOne(prompt, &confirm)
			if err != nil {
				return err
			}

			if !confirm {
				return nil
			}

			resp, err := client.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Name: currentName})
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

			cmdutil.SuccessPrinter("Renamed organization \n")
			cmdutil.TablePrinter(toRow(updatedOrg.Organization))
			return nil
		},
	}

	return renameCmd
}
