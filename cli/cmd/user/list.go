package user

import (
	"context"
	"errors"

	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListCmd(cfg *config.Config) *cobra.Command {
	var orgName string
	var projectName string

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List",
		RunE: func(cmd *cobra.Command, args []string) error {
			if orgName == "" && projectName == "" {
				return errors.New("either organization or project has to be specified")
			}
			if orgName != "" && projectName != "" {
				return errors.New("only one of organization or project has to be specified")
			}

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if orgName != "" {
				res, err := client.ListOrganizationMembers(context.Background(), &adminv1.ListOrganizationMembersRequest{
					Organization: orgName,
				})
				if err != nil {
					return err
				}

				cmdutil.PrintMembers(res.Members)
			} else if projectName != "" {
				res, err := client.ListProjectMembers(context.Background(), &adminv1.ListProjectMembersRequest{
					Organization: cfg.Org,
					Project:      projectName,
				})
				if err != nil {
					return err
				}

				cmdutil.PrintMembers(res.Members)
				// TODO: user groups
			}

			return nil
		},
	}

	listCmd.Flags().StringVar(&orgName, "org", "", "Organization")
	listCmd.Flags().StringVar(&projectName, "project", "", "Project")

	return listCmd
}
