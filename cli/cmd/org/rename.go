package org

import (
	"context"
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
	renameCmd := &cobra.Command{
		Use:   "rename <org-name>",
		Args:  cobra.MaximumNArgs(1),
		Short: "Rename",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			var orgName string

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if len(args) > 0 {
				orgName = args[0]
			} else {
				// Get the new org name from user if not provided in the args
				questions := []*survey.Question{
					{
						Name: "name",
						Prompt: &survey.Input{
							Message: "Enter the new org name",
						},
						Validate: func(any interface{}) error {
							name := any.(string)
							if name == "" {
								return fmt.Errorf("empty name")
							}

							exist, err := cmdutil.OrgExists(ctx, client, name)
							if err != nil {
								return fmt.Errorf("org name %q is already taken", name)
							}

							if exist {
								// this should always be true but adding this check from completeness POV
								return fmt.Errorf("org with name %q already exists", name)
							}

							if name == cfg.Org {
								return fmt.Errorf("org with name %v is same as current/default org name", name)
							}
							return nil
						},
					},
				}

				if err := survey.Ask(questions, &orgName); err != nil {
					return err
				}
			}

			confirm := false
			prompt := &survey.Confirm{
				Message: fmt.Sprintf("Do you want to rename %q to %q?", color.YellowString(cfg.Org), color.YellowString(orgName)),
			}

			err = survey.AskOne(prompt, &confirm)
			if err != nil {
				return err
			}

			if !confirm {
				return nil
			}

			resp, err := client.GetOrganization(context.Background(), &adminv1.GetOrganizationRequest{Name: cfg.Org})
			if err != nil {
				return err
			}

			org := resp.Organization
			updatedOrg, err := client.UpdateOrganization(context.Background(), &adminv1.UpdateOrganizationRequest{
				Id:          org.Id,
				Name:        orgName,
				Description: org.Description,
			})
			if err != nil {
				return err
			}

			err = dotrill.SetDefaultOrg(orgName)
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
