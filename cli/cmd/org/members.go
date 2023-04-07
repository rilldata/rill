package org

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
	membersCmd.AddCommand(LeaveOrgCmd(cfg))

	membersCmd.PersistentFlags().StringVar(&cfg.Org, "org", cfg.Org, "Organization name")
	return membersCmd
}

func ListMembersCmd(cfg *config.Config) *cobra.Command {
	membersCmd := &cobra.Command{
		Use:   "list",
		Short: "List Members",
		RunE: func(cmd *cobra.Command, args []string) error {
			orgName := cfg.Org

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()
			resp, err := client.ListOrganizationMembers(cmd.Context(), &adminv1.ListOrganizationMembersRequest{
				Organization: orgName,
			})
			if err != nil {
				return err
			}

			cmdutil.PrintMembers(resp.Members)
			return nil
		},
	}

	return membersCmd
}

func AddCmd(cfg *config.Config) *cobra.Command {
	addCmd := &cobra.Command{
		Use:   "add <email> {admin|collaborator|reader}",
		Short: "Add Member",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			orgName := cfg.Org

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()
			_, err = client.AddOrganizationMember(cmd.Context(), &adminv1.AddOrganizationMemberRequest{
				Organization: orgName,
				Email:        args[0],
				Role:         args[1],
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
		Use:   "remove <email>",
		Short: "Remove Member",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			orgName := cfg.Org

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()
			_, err = client.RemoveOrganizationMember(cmd.Context(), &adminv1.RemoveOrganizationMemberRequest{
				Organization: orgName,
				Email:        args[0],
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
		Use:   "set-role <email> {admin|collaborator|reader}",
		Short: "Set role of Member",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			orgName := cfg.Org

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()
			_, err = client.SetOrganizationMemberRole(cmd.Context(), &adminv1.SetOrganizationMemberRoleRequest{
				Organization: orgName,
				Email:        args[0],
				Role:         args[1],
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

func LeaveOrgCmd(cfg *config.Config) *cobra.Command {
	leaveOrgCmd := &cobra.Command{
		Use:   "leave",
		Short: "Leave Organization",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			orgName := cfg.Org

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()
			_, err = client.LeaveOrganization(cmd.Context(), &adminv1.LeaveOrganizationRequest{
				Organization: orgName,
			})
			if err != nil {
				return err
			}
			cmdutil.SuccessPrinter("Removed")
			return nil
		},
	}
	return leaveOrgCmd
}
