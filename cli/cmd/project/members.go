package project

import (
	"fmt"

	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func MembersCmd(cfg *config.Config) *cobra.Command {
	var name, path string

	membersCmd := &cobra.Command{
		Use:   "members",
		Short: "Members",
		RunE: func(cmd *cobra.Command, args []string) error {
			sp := cmdutil.Spinner("Listing members...")
			sp.Start()

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if !cmd.Flags().Changed("project") {
				name, err = inferProjectName(cmd.Context(), client, cfg.Org, path)
				if err != nil {
					return err
				}
			}

			resp, err := client.ListProjectMembers(cmd.Context(), &adminv1.ListProjectMembersRequest{
				Organization: cfg.Org,
				Project:      name,
			})
			if err != nil {
				return err
			}

			cmdutil.PrintMembers(resp.Members)
			cmdutil.PrintInvites(resp.Invites)
			return nil
		},
	}
	membersCmd.AddCommand(ListMembersCmd(cfg))
	membersCmd.AddCommand(AddCmd(cfg))
	membersCmd.AddCommand(RemoveCmd(cfg))
	membersCmd.AddCommand(SetRoleCmd(cfg))

	membersCmd.Flags().SortFlags = false
	membersCmd.Flags().StringVar(&name, "project", "", "Name")
	membersCmd.Flags().StringVar(&path, "path", ".", "Project directory")

	return membersCmd
}

func ListMembersCmd(cfg *config.Config) *cobra.Command {
	var name, path string

	listMembersCmd := &cobra.Command{
		Use:   "list <project-name>",
		Short: "List Members",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			sp := cmdutil.Spinner("Listing members...")
			sp.Start()

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if !cmd.Flags().Changed("project") {
				name, err = inferProjectName(cmd.Context(), client, cfg.Org, path)
				if err != nil {
					return err
				}
			}

			resp, err := client.ListProjectMembers(cmd.Context(), &adminv1.ListProjectMembersRequest{
				Organization: cfg.Org,
				Project:      name,
			})
			if err != nil {
				return err
			}

			cmdutil.PrintMembers(resp.Members)
			cmdutil.PrintInvites(resp.Invites)

			return nil
		},
	}

	listMembersCmd.Flags().SortFlags = false
	listMembersCmd.Flags().StringVar(&name, "project", "", "Name")
	listMembersCmd.Flags().StringVar(&path, "path", ".", "Project directory")

	return listMembersCmd
}

func AddCmd(cfg *config.Config) *cobra.Command {
	addCmd := &cobra.Command{
		Use:   "add <project-name> <email> {admin|viewer}",
		Short: "Add Member",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			orgName := cfg.Org

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()
			res, err := client.AddProjectMember(cmd.Context(), &adminv1.AddProjectMemberRequest{
				Organization: orgName,
				Project:      args[0],
				Email:        args[1],
				Role:         args[2],
			})
			if err != nil {
				return err
			}

			if res.PendingSignup {
				cmdutil.SuccessPrinter(fmt.Sprintf("Invitation sent to %q to join project %q as %q", args[1], args[0], args[2]))
				return nil
			}
			cmdutil.SuccessPrinter(fmt.Sprintf("User %q added to the project %q as %q", args[1], args[0], args[2]))
			return nil
		},
	}
	return addCmd
}

func RemoveCmd(cfg *config.Config) *cobra.Command {
	removeCmd := &cobra.Command{
		Use:   "remove <project-name> <email>",
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
		Use:   "set-role <project-name> <email> {admin|viewer}",
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
