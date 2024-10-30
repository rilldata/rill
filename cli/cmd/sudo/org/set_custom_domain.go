package org

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func SetCustomDomainCmd(ch *cmdutil.Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-custom-domain <org> <custom-domain>",
		Args:  cobra.ExactArgs(2),
		Short: "Set custom domain for an org (domain must not contain a scheme or path)",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			_, err = client.SudoUpdateOrganizationCustomDomain(cmd.Context(), &adminv1.SudoUpdateOrganizationCustomDomainRequest{
				Name:         args[0],
				CustomDomain: args[1],
			})
			if err != nil {
				return err
			}

			ch.Printf("Set custom domain %q for org %q. Remember to separately update DNS records and configure TLS on the load balancer (see runbook for details).\n", args[1], args[0])

			return nil
		},
	}

	return cmd
}
