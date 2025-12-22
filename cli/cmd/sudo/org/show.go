package org

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
)

func ShowCmd(ch *cmdutil.Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <org>",
		Args:  cobra.ExactArgs(1),
		Short: "Show all org details",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			res, err := client.GetOrganization(cmd.Context(), &adminv1.GetOrganizationRequest{
				Org:                  args[0],
				SuperuserForceAccess: true,
			})
			if err != nil {
				return err
			}

			enc := protojson.MarshalOptions{
				Multiline:       true,
				EmitUnpopulated: true,
			}

			data, err := enc.Marshal(res.Organization)
			if err != nil {
				return fmt.Errorf("failed to marshal org as JSON: %w", err)
			}

			fmt.Println(string(data))

			return nil
		},
	}

	return cmd
}
