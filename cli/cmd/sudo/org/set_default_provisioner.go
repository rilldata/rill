package org

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func SetDefaultProvisionerCmd(ch *cmdutil.Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-default-provisioner <org> [<provisioner>]",
		Args:  cobra.RangeArgs(1, 2),
		Short: "Set the default provisioner for an org's deployments (omit the provisioner to unset)",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			var provisioner string
			if len(args) > 1 {
				provisioner = args[1]
			}

			_, err = client.SudoUpdateOrganizationDefaultProvisioner(cmd.Context(), &adminv1.SudoUpdateOrganizationDefaultProvisionerRequest{
				Org:                args[0],
				DefaultProvisioner: provisioner,
			})
			if err != nil {
				return err
			}

			if provisioner == "" {
				ch.PrintfSuccess("Unset the default provisioner for org %q. It will now use the global default provisioner.\n", args[0])
			} else {
				ch.PrintfSuccess("Set default provisioner %q for org %q.\n", provisioner, args[0])
			}
			ch.PrintfWarn("Note: this only applies to deployments provisioned from now on. Existing deployments keep their current provisioner.\n")

			return nil
		},
	}

	return cmd
}
