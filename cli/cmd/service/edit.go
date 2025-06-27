package service

import (
	"encoding/json"
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func EditCmd(ch *cmdutil.Helper) *cobra.Command {
	var newName string
	var attributes string

	editCmd := &cobra.Command{
		Use:   "edit <service-name>",
		Args:  cobra.ExactArgs(1),
		Short: "edit service properties",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			// Parse attributes if provided
			var attrs map[string]string
			if attributes != "" {
				if err := json.Unmarshal([]byte(attributes), &attrs); err != nil {
					return fmt.Errorf("failed to parse --attributes as JSON: %w", err)
				}
			}

			req := &adminv1.UpdateServiceRequest{
				Name:             args[0],
				OrganizationName: ch.Org,
			}

			if newName != "" {
				req.NewName = &newName
			}

			if attrs != nil {
				req.Attributes = attrs
			}

			res, err := client.UpdateService(cmd.Context(), req)
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Updated service\n")
			ch.PrintServices([]*adminv1.Service{res.Service})

			return nil
		},
	}
	editCmd.Flags().SortFlags = false
	editCmd.Flags().StringVar(&newName, "new-name", "", "New service name")
	editCmd.Flags().StringVar(&attributes, "attributes", "", "JSON object of key-value pairs for service attributes")

	return editCmd
}
