package user

import (
	"errors"

	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RemoveCmd(cfg *config.Config) *cobra.Command {
	var orgName string
	var projectName string
	var email string

	removeCmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove",
		RunE: func(cmd *cobra.Command, args []string) error {
			if orgName == "" && projectName == "" {
				return errors.New("either organization or project has to be specified")
			}
			if orgName != "" && projectName != "" {
				return errors.New("only one of organization or project has to be specified")
			}

			if orgName == "" {
				orgName = cfg.Org
			}

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			_, err = client.RemoveProjectMember(cmd.Context(), &adminv1.RemoveProjectMemberRequest{
				Organization: orgName,
				Project:      projectName,
				Email:        email,
			})
			if err != nil {
				return err
			}

			cmdutil.SuccessPrinter("Removed")
			return nil
		},
	}

	removeCmd.Flags().StringVar(&orgName, "org", "", "Organization")
	removeCmd.Flags().StringVar(&projectName, "project", "", "Project")
	removeCmd.Flags().StringVar(&email, "email", "", "Email of the user")

	return removeCmd
}
