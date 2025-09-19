package service

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func DeleteCmd(ch *cmdutil.Helper) *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete <service-name>",
		Args:  cobra.ExactArgs(1),
		Short: "Delete service",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			_, err = client.DeleteService(cmd.Context(), &adminv1.DeleteServiceRequest{
				Name: args[0],
				Org:  ch.Org,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Deleted service: %q\n", args[0])

			return nil
		},
	}
	return deleteCmd
}
