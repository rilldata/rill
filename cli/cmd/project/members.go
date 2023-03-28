package project

import (
	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func MembersCmd(cfg *config.Config) *cobra.Command {
	membersCmd := &cobra.Command{
		Use:   "members",
		Short: "Members",
	}
	membersCmd.AddCommand(ListMembersCmd(cfg))
	membersCmd.AddCommand(AddCmd(cfg))
	membersCmd.AddCommand(RemoveCmd(cfg))
	membersCmd.AddCommand(SetRoleCmd(cfg))

	return membersCmd
}

func ListMembersCmd(cfg *config.Config) *cobra.Command {
	membersCmd := &cobra.Command{
		Use:   "list",
		Short: "List Members",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sp := cmdutil.Spinner("Listing members...")
			sp.Start()

			orgName := cfg.Org
			projectName := args[0]

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()
			resp, err := client.ListProjectMembers(cmd.Context(), &adminv1.ListProjectMembersRequest{
				Organization: orgName,
				Project:      projectName,
			})
			if err != nil {
				return err
			}

			cmdutil.PrintUsers(resp.Users)
			return nil
		},
	}

	return membersCmd
}

func AddCmd(cfg *config.Config) *cobra.Command {
	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add Member",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			orgName := cfg.Org

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()
			_, err = client.AddProjectMember(cmd.Context(), &adminv1.AddProjectMemberRequest{
				Organization: orgName,
				Project:      args[0],
				Email:        args[1],
				Role:         args[2],
			})
			if err != nil {
				return err
			}
			cmdutil.SuccessPrinter("Done")
			return nil
		},
	}
	return addCmd
}

func RemoveCmd(cfg *config.Config) *cobra.Command {
	removeCmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove Member",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			orgName := cfg.Org

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()
			_, err = client.RemoveProjectMember(cmd.Context(), &adminv1.RemoveProjectMemberRequest{
				Organization: orgName,
				Project:      args[0],
				Email:        args[1],
			})
			if err != nil {
				return err
			}
			cmdutil.SuccessPrinter("Removed")
			return nil
		},
	}
	return removeCmd
}

func SetRoleCmd(cfg *config.Config) *cobra.Command {
	setRoleCmd := &cobra.Command{
		Use:   "set-role",
		Short: "Set role of Member",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			orgName := cfg.Org

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()
			_, err = client.SetProjectMemberRole(cmd.Context(), &adminv1.SetProjectMemberRoleRequest{
				Organization: orgName,
				Project:      args[0],
				Email:        args[1],
				Role:         args[2],
			})
			if err != nil {
				return err
			}
			cmdutil.SuccessPrinter("Updated")
			return nil
		},
	}
	return setRoleCmd
}
